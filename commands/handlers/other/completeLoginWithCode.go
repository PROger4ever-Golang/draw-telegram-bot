package otherpkg

import (
	"fmt"
	"strings"

	"github.com/go-telegram-bot-api/telegram-bot-api"

	"bitbucket.org/proger4ever/draw-telegram-bot/bot"
	"bitbucket.org/proger4ever/draw-telegram-bot/config"
	"bitbucket.org/proger4ever/draw-telegram-bot/userApi"
)

type CompleteLoginWithCodeHandler struct {
	Bot  *botpkg.Bot
	Conf *config.Config
	Tool *userapi.Tool
}

func (c *CompleteLoginWithCodeHandler) GetName() string {
	return "completeLoginWithCode"
}

func (c *CompleteLoginWithCodeHandler) IsForOwnersOnly() bool {
	return true
}

func (c *CompleteLoginWithCodeHandler) GetParamsMinCount() int {
	return 1
}

func (c *CompleteLoginWithCodeHandler) Init(conf *config.Config, tool *userapi.Tool, bot *botpkg.Bot) {
	c.Bot = bot
	c.Conf = conf
	c.Tool = tool
}

func (c *CompleteLoginWithCodeHandler) Execute(msg *tgbotapi.Message, params []string) error {
	phoneCode := strings.Replace(params[0], "-", "", -1)
	user, err := c.Tool.CompleteLoginWithCode(phoneCode)
	if err == nil {
		resp := fmt.Sprintf("```\nМы успешно авторизовались.\nUserID: %d\nUsername: %s\nName: %s %s\n```", user.ID, user.Username, user.FirstName, user.LastName)
		err = c.Bot.SendMessage(int64(msg.Chat.ID), resp, true)
	} else {
		err = c.Bot.SendError(int64(msg.Chat.ID), err)
	}
	return err
}
