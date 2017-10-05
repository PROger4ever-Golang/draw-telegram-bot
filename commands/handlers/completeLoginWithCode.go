package handlers

import (
	"fmt"
	"strings"

	"github.com/go-telegram-bot-api/telegram-bot-api"

	"bitbucket.org/proger4ever/drawtelegrambot/commands/utils"
	"bitbucket.org/proger4ever/drawtelegrambot/config"
	"bitbucket.org/proger4ever/drawtelegrambot/telegram/userapi"
)

type CompleteLoginWithCodeHandler struct {
	Bot  *tgbotapi.BotAPI
	Conf *config.Config
	Tool *userapi.Tool
}

func (c *CompleteLoginWithCodeHandler) GetName() string {
	return "completeLoginWithCode"
}

func (c *CompleteLoginWithCodeHandler) GetParamsCount() int {
	return 1
}

func (c *CompleteLoginWithCodeHandler) Init(conf *config.Config, tool *userapi.Tool, bot *tgbotapi.BotAPI) {
	c.Bot = bot
	c.Conf = conf
	c.Tool = tool
}

func (c *CompleteLoginWithCodeHandler) Execute(chat *tgbotapi.Chat, params []string) error {
	phoneCode := strings.Replace(params[0], "-", "", -1)
	user, err := c.Tool.CompleteLoginWithCode(phoneCode)

	if err == nil {
		resp := fmt.Sprintf("```\nМы успешно авторизовались.\nUserID: %d\nUsername: %s\nName: %s %s\n```", user.ID, user.Username, user.FirstName, user.LastName)
		err = utils.SendBotMessage(c.Bot, int64(chat.ID), resp)
	} else {
		err = utils.SendBotError(c.Bot, int64(chat.ID), err)
	}
	return err
}
