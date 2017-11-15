package otherpkg

import (
	"fmt"

	"bitbucket.org/proger4ever/draw-telegram-bot/error"
	"github.com/go-telegram-bot-api/telegram-bot-api"

	"bitbucket.org/proger4ever/draw-telegram-bot/bot"
	"bitbucket.org/proger4ever/draw-telegram-bot/config"
	"bitbucket.org/proger4ever/draw-telegram-bot/userApi"
)

type StartLoginHandler struct {
	Bot  *botpkg.Bot
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

func (c *StartLoginHandler) Init(conf *config.Config, tool *userapi.Tool, bot *botpkg.Bot) {
	c.Bot = bot
	c.Conf = conf
	c.Tool = tool
}

func (c *StartLoginHandler) Execute(msg *tgbotapi.Message, params []string) *eepkg.ExtendedError {
	// defer common.WrapIfPanic("startLogin.execute()")
	err := c.Tool.StartLogin(params[0])
	if err == nil {
		resp := fmt.Sprintf("Отправь мне пришедший код, вставив в него минус:\n```\n/completeLoginWithCode -\n```")
		err = c.Bot.SendMessageMarkdown(int64(msg.Chat.ID), resp)
	} else {
		err = c.Bot.SendError(int64(msg.Chat.ID), err)
	}
	return err
}
