package apiclient

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tristan-club/cascade/pkg/bot/ratelimiter"
	"github.com/tristan-club/kit/log"
	"golang.org/x/time/rate"
	"time"
)

type Client struct {
	botToken string
	*tgbotapi.BotAPI
	globalLimiter *rate.Limiter
	idLimiter     *ratelimiter.IdLimiter
}

func NewClientByToken(token string) (*Client, error) {
	return newClient(token, nil)
}

func NewClientByApi(api *tgbotapi.BotAPI) *Client {
	client, _ := newClient("", api)
	return client
}

func newClient(token string, api *tgbotapi.BotAPI) (client *Client, err error) {
	if api == nil {
		api, err = tgbotapi.NewBotAPI(token)
		if err != nil {
			log.Error().Msgf("Init tg bot error: %s", err.Error())
			return nil, err
		}
	}
	client = &Client{
		botToken:      api.Token,
		BotAPI:        api,
		globalLimiter: rate.NewLimiter(rate.Limit(30), 30),
		idLimiter:     ratelimiter.NewIdRateLimiter(rate.Limit(0.29), 19),
	}
	return client, nil
}

func (c *Client) GetUsername() string {
	return c.Self.UserName
}

func (c *Client) CheckLimit(chatId int64) {
	_ = c.globalLimiter.Wait(context.Background())
	_ = c.idLimiter.Wait(chatId)
}

func (c *Client) Start(h func(u *tgbotapi.Update)) error {

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	u.AllowedUpdates = []string{"update_id", "message", "edited_message", "channel_post", "edited_channel_post",
		"inline_query", "chosen_inline_result", "callback_query", "shipping_query", "pre_checkout_query", "poll", "poll_answer",
		"my_chat_member", "chat_member", "chat_join_request"}
	dw := tgbotapi.DeleteWebhookConfig{DropPendingUpdates: true}
	_, err := c.BotAPI.Request(dw)
	if err != nil {
		log.Error().Msgf("delete webhook error %s", err.Error())
		return err
	}

	//c.Debug = true
	updates := c.BotAPI.GetUpdatesChan(u)

	go func() {
		for update := range updates {
			update := update
			go h(&update)
		}
	}()
	return nil
}

func (c *Client) DeferDelMsg(chatId int64, msgId int, duration time.Duration) {
	go func() {
		if duration == 0 {
			duration = time.Second * 15
		}
		time.Sleep(duration)
		err := c.DelMsg(chatId, msgId)
		if err != nil {
			log.Error().Fields(map[string]interface{}{"action": "del msg error", "error": err.Error()}).Send()
		}
	}()
}

func (c *Client) SendMsg(chatId int64, content string, ikm interface{}, isMarkdown bool) (*tgbotapi.Message, error) {
	return c.SendMsgWithPhoto(chatId, content, ikm, "", "", isMarkdown, 0)
}

func (c *Client) ReplyMsg(chatId int64, replyMsgId int, content string, ikm interface{}, isMarkdown bool) (*tgbotapi.Message, error) {
	return c.SendMsgWithPhoto(chatId, content, ikm, "", "", isMarkdown, replyMsgId)
}

func (c *Client) SendMedia(chatId int64, path string) (*tgbotapi.Message, error) {

	_msg := tgbotapi.NewPhoto(chatId, tgbotapi.FileID(path))
	_msg.Caption = "ss"
	_, err := c.BotAPI.Send(_msg)
	return nil, err

	cfg := tgbotapi.NewMediaGroup(chatId, []interface{}{
		tgbotapi.NewInputMediaPhoto(tgbotapi.FilePath(path)),
	})

	messages, err := c.BotAPI.SendMediaGroup(cfg)
	if err != nil {
		log.Info().Fields(map[string]interface{}{"action": "s", "m": messages, "err": err}).Send()
		return nil, err
	}

	return &messages[0], nil

}

func (c *Client) SendDocument(chatId int64, path string) (*tgbotapi.Message, error) {
	c.CheckLimit(chatId)
	_msg := tgbotapi.NewDocument(chatId, tgbotapi.FilePath(path))
	msg, err := c.BotAPI.Send(_msg)
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "send document error", "error": err.Error()}).Send()
		return nil, err
	}
	return &msg, nil
}

func (c *Client) EditMsg(chatId int64, msgId int, content string, ikm *tgbotapi.InlineKeyboardMarkup, isMarkdown bool) error {
	c.CheckLimit(chatId)
	editMsg := tgbotapi.NewEditMessageText(chatId, msgId, content)
	if isNil(ikm) {
		editMsg.ReplyMarkup = ikm
	}
	if isMarkdown {
		editMsg.ParseMode = tgbotapi.ModeMarkdownV2
	}

	_, err := c.BotAPI.Send(editMsg)
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "edit msg error", "error": err.Error(), "editMsg": editMsg}).Send()
	}
	return err
}

func (c *Client) SendMsgWithPhoto(chatId int64, content string, ikm interface{}, photoUrl string, photoFileId string, isMarkdown bool, replyMsgId int) (*tgbotapi.Message, error) {

	c.CheckLimit(chatId)

	var msg tgbotapi.Chattable

	if photoUrl == "" && photoFileId == "" {
		_msg := tgbotapi.NewMessage(chatId, content)
		if ikm != nil {
			_msg.ReplyMarkup = ikm
		}
		if isMarkdown {
			_msg.ParseMode = tgbotapi.ModeMarkdownV2
		}
		_msg.DisableWebPagePreview = true
		_msg.ReplyToMessageID = replyMsgId
		msg = _msg
	} else {
		var file tgbotapi.RequestFileData
		if photoUrl != "" {
			file = tgbotapi.FileURL(photoUrl)
		} else {
			file = tgbotapi.FileID(photoFileId)
		}
		_msg := tgbotapi.NewPhoto(chatId, file)
		_msg.ReplyToMessageID = replyMsgId
		_msg.Caption = content
		if ikm != nil {
			_msg.ReplyMarkup = ikm
		}
		if isMarkdown {
			_msg.ParseMode = tgbotapi.ModeMarkdownV2
		}
		msg = _msg
	}

	resp, err := c.BotAPI.Send(msg)
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "tg bot api send msg error", "error": err.Error(), "msg": msg}).Send()
		return nil, err
	} else {
		return &resp, nil
	}
}

func (c *Client) DelMsg(chatId int64, msgId int) error {
	res, err := c.BotAPI.Request(tgbotapi.NewDeleteMessage(chatId, msgId))
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "delete msg error", "error": err.Error(), "chatId": chatId, "msgId": msgId}).Send()
		return err
	} else if !res.Ok {
		log.Error().Fields(map[string]interface{}{"action": "delete msg error", "error": res.Description, "chatId": chatId, "msgId": msgId}).Send()
		return fmt.Errorf(res.Description)
	}
	return nil
}
