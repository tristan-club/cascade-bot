package service

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/redis/go-redis/v9"
	"github.com/tristan-club/cascade/model"
	"github.com/tristan-club/cascade/pconst"
	rdb "github.com/tristan-club/cascade/pkg/redisutil"
	"github.com/tristan-club/kit/config"
	"github.com/tristan-club/kit/customid"
	he "github.com/tristan-club/kit/error"
	tlog "github.com/tristan-club/kit/log"
	"github.com/tristan-club/kit/mdparse"
	"github.com/xssnick/tonutils-go/address"
	"time"
)

const (
	defaultSigninLayout = "20060102"
)

func HandleStart(ctx *Context) error {

	tlog.Info().Fields(map[string]interface{}{"action": "handle start", "openId": ctx.GetOpenId(), "ctx": ctx}).Send()

	if ctx.Update.Message.Text == "/start default" {
		tlog.Info().Fields(map[string]interface{}{"action": "add bot no permission"}).Send()
		return he.NewStandardBusinessError("Sorry, only admins can add this bot.")
	}

	ikm := tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonURL("Add me to your chat!", fmt.Sprintf("https://t.me/%s?startgroup=default", ctx.GetBotUsername()))})

	_, err := ctx.Client.SendMsg(ctx.Update.SentFrom().ID, mdparse.ParseV2(TextStart), ikm, true)
	if err != nil {
		tlog.Error().Fields(map[string]interface{}{"action": "send question error", "error": err.Error()}).Send()
		return he.NewStandardServerError(err)
	}

	return nil
}

func HandleHelp(ctx *Context) error {

	tlog.Info().Fields(map[string]interface{}{"action": "handle start", "openId": ctx.GetOpenId(), "ctx": ctx}).Send()

	ikm := tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("Group Message Statistics 📊", customid.NewCustomId(pconst.CustomAdminButtonGroupMsg, "", 0).String()),
		tgbotapi.NewInlineKeyboardButtonData("Configure Check-Ins 📝", customid.NewCustomId(pconst.CustomAdminButtonConfigSignIn, "", 0).String()),
	}, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("Send Points 💰", customid.NewCustomId(pconst.CustomAdminButtonDistributeToken, "", 0).String()),
		tgbotapi.NewInlineKeyboardButtonData("Member Info 📋", customid.NewCustomId(pconst.CustomAdminButtonExportMembers, "", 0).String()),
		tgbotapi.NewInlineKeyboardButtonData("Withdraw 💵", customid.NewCustomId(pconst.CustomAdminWithdraw, "", 0).String()),
	})

	_, err := ctx.Client.SendMsg(ctx.Update.SentFrom().ID, mdparse.ParseV2(TextHelp), ikm, true)
	if err != nil {
		tlog.Error().Fields(map[string]interface{}{"action": "send question error", "error": err.Error()}).Send()
		return he.NewStandardServerError(err)
	}

	return nil
}

func HandleSignIn(ctx *Context) error {

	tlog.Info().Fields(map[string]interface{}{"action": "handle signin", "ctx": ctx, "openId": ctx.GetOpenId()}).Send()

	sc, err := model.GetSignInConfig(ctx.Update.FromChat().ID)
	if err != nil {
		tlog.Error().Fields(map[string]interface{}{"action": "get signin config error", "error": err.Error()}).Send()
		return err
	} else if sc == nil || !sc.Active() {
		tlog.Info().Fields(map[string]interface{}{"action": "no signin config"}).Send()
		return nil
	}

	key := fmt.Sprintf("%s%d:%s:%d", pconst.RKSignInLimiter, ctx.Update.FromChat().ID, time.Now().Format(defaultSigninLayout), ctx.Update.SentFrom().ID)
	n, err := rdb.Default().Get(context.Background(), key).Int()
	if err != nil && err != redis.Nil {
		tlog.Error().Fields(map[string]interface{}{"action": "get event limit error", "error": err.Error()}).Send()
		return he.NewStandardServerError(fmt.Errorf("check event error: %s", err.Error()))
	}

	var content string

	limiter := 1
	if config.EnvIsDev() {
		limiter = 5
	}

	if n < limiter {

		err = rdb.Default().Incr(context.Background(), key).Err()
		if err != nil {
			tlog.Error().Fields(map[string]interface{}{"action": "set check in data error", "error": err.Error()}).Send()
			return err
		}

		if err = model.AddUserScore(ctx.Update.FromChat().ID, ctx.GetOpenId(), int64(sc.Reward)); err != nil {
			tlog.Error().Fields(map[string]interface{}{"action": "save user error", "error": err.Error()}).Send()
			return err
		}

		if sc.Text != "" {
			content = mdparse.ParseV2(sc.Text)
		} else {
			content = fmt.Sprintf(TextCheckInSuccess, fmt.Sprintf("[@%s](tg://user?id=%d)", mdparse.ParseV2(ctx.Update.SentFrom().FirstName), ctx.Update.SentFrom().ID), sc.Reward, sc.Symbol)
		}

	} else {
		content = TextAlreadyCheckIn
	}

	sr := model.NewSignInUser(ctx.Update.FromChat().ID, ctx.GetOpenId(), 0, uint64(sc.Reward))
	if err = model.SaveSignInRecord(sr); err != nil {
		tlog.Error().Fields(map[string]interface{}{"action": "save signin record error", "error": err.Error()}).Send()
	}

	tlog.Info().Msgf("user %d in group %d signin get reward %d", ctx.GetOpenId(), ctx.Update.FromChat().ID, sc.Reward)

	_, err = ctx.Client.ReplyMsg(ctx.Update.FromChat().ID, ctx.Update.Message.MessageID, content, nil, true)
	if err != nil {
		tlog.Error().Fields(map[string]interface{}{"action": "send msg error", "error": err.Error()}).Send()
		return err
	}

	return nil
}

func HandleScore(ctx *Context) error {

	tlog.Info().Fields(map[string]interface{}{"action": "handle score", "ctx": ctx, "openId": ctx.GetOpenId()}).Send()

	score, err := model.GetUserScore(ctx.Update.FromChat().ID, ctx.GetOpenId())
	if err != nil {
		tlog.Error().Fields(map[string]interface{}{"action": "get score error", "error": err.Error()}).Send()
		return err
	}

	_, err = ctx.Client.ReplyMsg(ctx.Update.FromChat().ID, ctx.Update.Message.MessageID, fmt.Sprintf(TextScore, score), nil, true)
	if err != nil {
		tlog.Error().Fields(map[string]interface{}{"action": "send msg error", "error": err.Error()}).Send()
		return err
	}

	return nil
}

func HandleRequestSubmitAddr(ctx *Context) error {

	tlog.Info().Fields(map[string]interface{}{"action": "handle submit addr", "ctx": ctx, "openId": ctx.GetOpenId()}).Send()

	UpdateUserStatus(ctx.GetOpenId(), pconst.CustomSubmitWallet)

	_, err := ctx.Client.SendMsg(ctx.Update.FromChat().ID, TextInputAddress, nil, false)
	if err != nil {
		tlog.Error().Fields(map[string]interface{}{"action": "send msg error", "error": err.Error()}).Send()
		return err
	}

	return nil
}
func HandleEnterSubmitAddr(ctx *Context) error {

	tlog.Info().Fields(map[string]interface{}{"action": "handle enter submit addr", "ctx": ctx, "openId": ctx.GetOpenId()}).Send()

	addr, err := address.ParseAddr(ctx.Update.Message.Text)
	if err != nil {
		tlog.Error().Fields(map[string]interface{}{"action": "parse addr error", "error": err.Error()}).Send()
		return fmt.Errorf("invalid address input")
	}

	user, err := model.GetUserByAddress(addr.String())
	if err != nil {
		tlog.Error().Fields(map[string]interface{}{"action": "get user error", "error": err.Error()}).Send()
		return err
	} else if user != nil {
		tlog.Info().Fields(map[string]interface{}{"action": "address bound", "addr": addr.String()}).Send()
		return he.NewStandardBusinessError("This wallet has already been bound.")
	}

	ctx.User.Addr = addr.String()
	if err = model.SaveUser(ctx.User); err != nil {
		tlog.Error().Fields(map[string]interface{}{"action": "save user error", "error": err.Error()}).Send()
		return err
	}

	_, err = ctx.Client.ReplyMsg(ctx.Update.FromChat().ID, ctx.Update.Message.MessageID, TextOpSuccess, nil, true)
	if err != nil {
		tlog.Error().Fields(map[string]interface{}{"action": "send msg error", "error": err.Error()}).Send()
		return err
	}

	ResetUserStatus(ctx.GetOpenId())

	return nil
}
