package service

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tristan-club/cascade/model"
	"github.com/tristan-club/cascade/pkg/bot/apiclient"
)

type Context struct {
	botUsername string
	Update      *tgbotapi.Update `json:"-"`
	User        *model.User
	Client      *apiclient.Client
}

func (c *Context) GetOpenId() int64 {
	return c.Update.SentFrom().ID
}

func (c *Context) GetChatId() int64 {
	return c.Update.FromChat().ID
}
func (c *Context) GetBotUsername() string {
	return c.Client.GetUsername()
}
