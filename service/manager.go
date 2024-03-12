package service

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tristan-club/cascade/model"
	"github.com/tristan-club/cascade/pconst"
	"github.com/tristan-club/cascade/pkg/bot/apiclient"
	"github.com/tristan-club/cascade/pkg/util"
	"github.com/tristan-club/cascade/pkg/xrpc"
	"github.com/tristan-club/cascade/pkg/xrpc/xproto/taskmonitor_pb"
	"github.com/tristan-club/kit/customid"
	"github.com/tristan-club/kit/dingding"
	he "github.com/tristan-club/kit/error"
	"github.com/tristan-club/kit/log"
	"golang.org/x/exp/slices"
	"runtime/debug"
	"strconv"
)

var mgr *Manager

func defaultManager() *Manager {
	return mgr
}

type Manager struct {
	Client *apiclient.Client
}

func InitBot(token string) error {

	cli, err := apiclient.NewClientByToken(token)
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "init client error", "error": err.Error()}).Send()
		return err
	}

	m := &Manager{Client: cli}

	groupCmds := make([]tgbotapi.BotCommand, 0)
	for _, cmd := range pconst.GroupCmdList {
		groupCmds = append(groupCmds, tgbotapi.BotCommand{
			Command:     cmd,
			Description: pconst.CmdDesc[cmd],
		})
	}

	scope := tgbotapi.NewBotCommandScopeAllGroupChats()
	_, err = m.Client.Request(tgbotapi.SetMyCommandsConfig{
		Commands:     groupCmds,
		Scope:        &scope,
		LanguageCode: "",
	})
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "register cmd error", "error": err.Error()}).Send()
		return err
	}

	privateCmds := make([]tgbotapi.BotCommand, 0)
	for _, cmd := range pconst.PrivateCmdList {
		privateCmds = append(privateCmds, tgbotapi.BotCommand{
			Command:     cmd,
			Description: pconst.CmdDesc[cmd],
		})
	}

	scope = tgbotapi.NewBotCommandScopeAllPrivateChats()
	_, err = m.Client.Request(tgbotapi.SetMyCommandsConfig{
		Commands:     privateCmds,
		Scope:        &scope,
		LanguageCode: "",
	})
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "register cmd error", "error": err.Error()}).Send()
		return err
	}

	err = m.Client.Start(m.Handle)
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "start client error", "error": err.Error()}).Send()
		return err
	}

	mgr = m

	log.Info().Msgf("fission bot %s manager init success", m.Client.BotAPI.Self.UserName)

	return nil

}

func (m *Manager) Handle(u *tgbotapi.Update) {
	defer func() {
		if err := recover(); err != nil {
			log.Error().Fields(map[string]interface{}{"action": "panic error", "error": err}).Send()
			dingding.Default().SendTextMessage(fmt.Sprintf("cascade panic error: %s", util.FastMarshal(err)), nil, false)
			debug.PrintStack()
		}
	}()
	_ = m.handle(u)
}

func syncUser(u *tgbotapi.Update, user *model.User) {
	tu := u.SentFrom()
	if user.Username != tu.UserName || user.FirstName != tu.FirstName || user.LastName != tu.LastName {
		user.Username = tu.UserName
		user.FirstName = tu.FirstName
		user.LastName = tu.LastName
		if err := model.SaveUser(user); err != nil {
			log.Error().Fields(map[string]interface{}{"action": "save user fission error", "error": err.Error(), "user": user, "tu": tu}).Send()
		}
	}
}

func (m *Manager) handle(u *tgbotapi.Update) error {

	if u.Message == nil && u.CallbackQuery == nil {
		return nil
	}

	if u.FromChat() != nil && !u.FromChat().IsPrivate() {

		if u.Message != nil && !u.Message.IsCommand() {
			if !u.SentFrom().IsBot && u.Message.Text != "" || len(u.Message.Photo) > 0 || u.Message.Sticker != nil {
				_, err := xrpc.GetTaskMonitorClient().AddUserMsgCount(context.Background(), &taskmonitor_pb.AddUserMsgCountReq{
					GroupId:   strconv.FormatInt(u.FromChat().ID, 10),
					OpenId:    strconv.FormatInt(u.SentFrom().ID, 10),
					MsgId:     strconv.Itoa(u.Message.MessageID),
					Username:  u.SentFrom().UserName,
					FirstName: u.SentFrom().FirstName,
					LastName:  u.SentFrom().LastName,
					//Profile : u.SentFrom().,
				})
				if err != nil {
					log.Error().Fields(map[string]interface{}{"action": "add user msg count error", "error": err.Error(), "ctx": m}).Send()
				} else {
					log.Debug().Msgf("bot record open_id %s group_id %s msg_id %s", strconv.FormatInt(u.SentFrom().ID, 10), strconv.FormatInt(u.FromChat().ID, 10), strconv.Itoa(u.Message.MessageID))
				}
			}
			return nil
		}

		if !u.Message.IsCommand() || !slices.Contains(pconst.GroupCmdList, u.Message.Command()) {
			return nil
		}

	}

	var replyMsgId int

	user, err := model.GetUser(u.SentFrom().ID)
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "db error", "error": err.Error()}).Send()
		return he.NewStandardServerError(err)
	} else if user == nil {
		user = model.NewUser(u.SentFrom().ID, u.SentFrom().UserName, u.SentFrom().FirstName, u.SentFrom().LastName)
		if err = model.AddUser(user); err != nil {
			log.Error().Fields(map[string]interface{}{"action": "save user error", "error": err.Error()}).Send()
			return err
		}
	} else {
		syncUser(u, user)
	}

	ctx := &Context{Update: u, User: user, Client: m.Client}

	if ctx.Update.FromChat().IsPrivate() {
		if u.Message != nil {
			if u.Message.IsCommand() {
				switch u.Message.Command() {
				case pconst.CmdStart:
					cid, ok := customid.ParseCustomId(u.Message.CommandArguments())
					if ok {
						switch cid.GetCustomType() {
						case pconst.CustomSubmitWallet:
							err = HandleRequestSubmitAddr(ctx)
						}
					} else {
						err = HandleStart(ctx)
					}

				case pconst.CmdHelp:
					err = HandleHelp(ctx)
				case pconst.CmdGroupMsgStat:
					err = HandleGroupMsgStat(ctx)
				case pconst.CmdUserData:
					err = HandleUserData(ctx)
				case pconst.CmdSetScore:
					err = HandleSetUserScore(ctx)
				case pconst.CmdConfigSignin:
					err = HandleConfigSignin(ctx)
				case pconst.CmdDisableSignin:
					err = HandleDisableSignin(ctx)
				case pconst.CmdWithdraw:
					err = HandleWithdraw(ctx)
				}
			} else {

				var notRkbMessage bool
				switch u.Message.Text {
				case RButtonMyAccount:
					err = HandleStart(ctx)

				default:
					notRkbMessage = true
				}

				if notRkbMessage {
					status := GetUserStatus(ctx.GetOpenId())
					switch status {
					case pconst.CustomSubmitWallet:
						err = HandleEnterSubmitAddr(ctx)
					}
				}
			}

		} else if u.CallbackQuery != nil {
			switch u.CallbackData() {
			case IButtonStart:
				err = HandleStart(ctx)
			default:
				if cid, ok := customid.ParseCustomId(u.CallbackData()); ok {
					switch cid.GetCustomType() {
					case pconst.CustomAdminButtonGroupMsg:
						err = HandleGroupMsgStat(ctx)
					case pconst.CustomAdminButtonConfigSignIn:
						err = HandleConfigSignin(ctx)
					case pconst.CustomAdminButtonDistributeToken:
						err = HandleSetUserScore(ctx)
					case pconst.CustomAdminButtonExportMembers:
						err = HandleUserData(ctx)
					case pconst.CustomAdminWithdraw:
						err = HandleWithdraw(ctx)
					}
				}
			}
		}
	} else {
		if ctx.Update.Message != nil {
			if ctx.Update.Message.IsCommand() && slices.Contains(pconst.GroupCmdList, ctx.Update.Message.Command()) {

				if ctx.User.HaveAddress() {
					replyMsgId = ctx.Update.Message.MessageID
					switch ctx.Update.Message.Command() {
					case pconst.CmdSignIn:
						err = HandleSignIn(ctx)
					case pconst.CmdScore:
						err = HandleScore(ctx)
					}
				} else {

					ikm := tgbotapi.NewInlineKeyboardMarkup(
						[]tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonURL("GO", fmt.Sprintf("https://t.me/%s?start=%s", ctx.Client.GetUsername(), customid.NewCustomId(pconst.CustomSubmitWallet, "", 0)))})
					mmsg, merr := ctx.Client.ReplyMsg(ctx.Update.FromChat().ID, replyMsgId, TextShouldSubmitWallet, ikm, false)
					if merr != nil {
						log.Error().Fields(map[string]interface{}{"action": "send msg error", "error": merr}).Send()
					} else {
						ctx.Client.DeferDelMsg(mmsg.Chat.ID, mmsg.MessageID, 0)
					}
				}

			}
		}
	}

	if err != nil {
		var content string
		if herr, ok := err.(he.Error); ok && herr.ErrorType() == he.BusinessError {
			content = herr.Msg()
			log.Info().Msgf("get business error, code:%d msg: %s, detail: %s", herr.Code(), herr.Msg(), herr.Error())
		} else {
			content = err.Error()
			log.Error().Fields(map[string]interface{}{"action": "get error", "error": err.Error()}).Send()
		}

		if ctx.Update.FromChat().IsPrivate() {
			_, err = m.Client.ReplyMsg(ctx.Update.FromChat().ID, replyMsgId, content, nil, false)
			if err != nil {
				log.Info().Fields(map[string]interface{}{"action": "send error msg error", "err": err, "text": content}).Send()
			}
		}

	}

	return nil
}
