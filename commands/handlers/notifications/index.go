package notificationspkg

import (
	"fmt"

	"bitbucket.org/proger4ever/draw-telegram-bot/bot"
	"bitbucket.org/proger4ever/draw-telegram-bot/config"
	"bitbucket.org/proger4ever/draw-telegram-bot/error"
	"bitbucket.org/proger4ever/draw-telegram-bot/userApi"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

const cantSendBotMessage = "Ошибка при отправке сообщения от имени бота"

const soundStateFormat = "Состояние уведомлений: *%s*"
const soundNewStateFormat = "Новое состояние уведомлений: *%s*"
const soundOn = "включено"
const soundOff = "выключено"

const soundStateIncorrect = "Некорректное значение параметра"

type Handler struct {
	Bot  *botpkg.Bot
	Conf *config.Config
	Tool *userapi.Tool
}

func (h *Handler) GetAliases() []string {
	return []string{"звук", "/notifications", "/sound"}
}

func (h *Handler) IsForOwnersOnly() bool {
	return true
}

func (h *Handler) GetParamsMinCount() int {
	return 0
}

func (h *Handler) Init(conf *config.Config, tool *userapi.Tool, bot *botpkg.Bot) {
	h.Bot = bot
	h.Conf = conf
	h.Tool = tool
}

func (h *Handler) Execute(msg *tgbotapi.Message, params []string) error {
	if len(params) == 0 {
		return h.sendState(msg.Chat.ID, soundStateFormat)
	}

	newState := false
	switch params[0] {
	case "0":
		newState = false
	case "1":
		newState = true
	default:
		return eepkg.New(true, false, soundStateIncorrect)
	}
	h.Conf.BotApi.DisableNotification = !newState
	return h.sendState(msg.Chat.ID, soundNewStateFormat)
}

func (h *Handler) sendState(chatID int64, format string) error {
	soundState := soundOn
	if h.Conf.BotApi.DisableNotification {
		soundState = soundOff
	}
	resp := fmt.Sprintf(format, soundState)
	err := h.Bot.SendMessage(chatID, resp, true)
	return eepkg.Wrap(err, true, false, cantSendBotMessage)
}
