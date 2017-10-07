package handlers

import (
	"fmt"

	"github.com/go-telegram-bot-api/telegram-bot-api"

	"bitbucket.org/proger4ever/drawtelegrambot/commands/utils"
	"bitbucket.org/proger4ever/drawtelegrambot/config"
	"bitbucket.org/proger4ever/drawtelegrambot/userApi"
)

type StartLoginHandler struct {
	Bot  *tgbotapi.BotAPI
	Conf *config.Config
	Tool *userapi.Tool
}

func (c *StartLoginHandler) GetName() string {
	return "startLogin"
}

func (c *StartLoginHandler) GetParamsCount() int {
	return 1
}

func (c *StartLoginHandler) Init(conf *config.Config, tool *userapi.Tool, bot *tgbotapi.BotAPI) {
	c.Bot = bot
	c.Conf = conf
	c.Tool = tool
}

func (c *StartLoginHandler) Execute(chat *tgbotapi.Chat, params []string) error {
	err := c.Tool.StartLogin(params[0])
	if err == nil {
		resp := fmt.Sprintf("Отправь мне пришедший код, вставив в него минус:\n```\n/completeLoginWithCode -\n```")
		err = utils.SendBotMessage(c.Bot, int64(chat.ID), resp)
	} else {
		err = utils.SendBotError(c.Bot, int64(chat.ID), err)
	}
	return err
}
