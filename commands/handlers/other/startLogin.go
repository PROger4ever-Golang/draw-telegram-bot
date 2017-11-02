package other

import (
	"fmt"

	"github.com/PROger4ever/telegram-bot-api"

	"bitbucket.org/proger4ever/draw-telegram-bot/commands/utils"
	"bitbucket.org/proger4ever/draw-telegram-bot/config"
	"bitbucket.org/proger4ever/draw-telegram-bot/userApi"
)

type StartLoginHandler struct {
	Bot  *tgbotapi.BotAPI
	Conf *config.Config
	Tool *userapi.Tool
}

func (c *StartLoginHandler) GetName() string {
	return "startLogin"
}

func (c *StartLoginHandler) IsForOwnersOnly() bool {
	return true
}

func (c *StartLoginHandler) GetParamsMinCount() int {
	return 1
}

func (c *StartLoginHandler) Init(conf *config.Config, tool *userapi.Tool, bot *tgbotapi.BotAPI) {
	c.Bot = bot
	c.Conf = conf
	c.Tool = tool
}

func (c *StartLoginHandler) Execute(msg *tgbotapi.Message, params []string) error {
	// defer common.WrapIfPanic("startLogin.execute()")
	err := c.Tool.StartLogin(params[0])
	if err == nil {
		resp := fmt.Sprintf("Отправь мне пришедший код, вставив в него минус:\n```\n/completeLoginWithCode -\n```")
		err = utils.SendBotMessage(c.Bot, int64(msg.Chat.ID), resp, true)
	} else {
		err = utils.SendBotError(c.Bot, int64(msg.Chat.ID), err)
	}
	return err
}
