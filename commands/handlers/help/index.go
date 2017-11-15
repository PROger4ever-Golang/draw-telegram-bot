package helppkg

import (
	"fmt"

	"bitbucket.org/proger4ever/draw-telegram-bot/error"
	"github.com/go-telegram-bot-api/telegram-bot-api"

	"bitbucket.org/proger4ever/draw-telegram-bot/bot"
	"bitbucket.org/proger4ever/draw-telegram-bot/config"
	"bitbucket.org/proger4ever/draw-telegram-bot/userApi"
)

const helpText = `Чтобы участвовать в розыгрышах Вам нужно:
1. Зарегистрироваться у mazimota бот, нажмите РЕГИСТРАЦИЯ
2. Подписаться на канал @mazimota
3. Прочитать правила @mazimota_rules
По желанию вступить в наш чат @mazimota_chat`

const unknownCommand = `Вы указали несуществующую команду: %s

` + helpText

type Handler struct {
	Bot  *botpkg.Bot
	Conf *config.Config
	Tool *userapi.Tool
}

func (h *Handler) GetAliases() []string {
	return []string{"помощь", "старт", "help", "start"}
}

func (h *Handler) IsForOwnersOnly() bool {
	return false
}

func (h *Handler) GetParamsMinCount() int {
	return 0
}

func (h *Handler) Init(conf *config.Config, tool *userapi.Tool, bot *botpkg.Bot) {
	h.Bot = bot
	h.Conf = conf
	h.Tool = tool
}

func (h *Handler) Execute(msg *tgbotapi.Message, params []string) *eepkg.ExtendedError {
	text := ""
	if len(params) >= 1 {
		text = fmt.Sprintf(unknownCommand, params[0])
	} else {
		text = helpText
	}
	return h.Bot.SendMessageUserKeyboard(int64(msg.Chat.ID), text)
}
