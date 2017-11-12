package adminhelppkg

import (
	"bitbucket.org/proger4ever/draw-telegram-bot/bot"
	"bitbucket.org/proger4ever/draw-telegram-bot/config"
	"bitbucket.org/proger4ever/draw-telegram-bot/error"
	"bitbucket.org/proger4ever/draw-telegram-bot/userApi"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

const cantSendBotMessage = "Ошибка при отправке сообщения от имени бота"

const commandsList = `Список админских команд:
/adminHelp - список админских команд
стат - количество зарегистрированных

звук - состояние уведомлений
звук 0 - отключить уведомления
звук 1 - включить уведомления

розыграй - провести розыгрыш 1 приза
розыграй N - провести розыгрыш N призов`

type Handler struct {
	Bot  *botpkg.Bot
	Conf *config.Config
	Tool *userapi.Tool
}

func (h *Handler) GetAliases() []string {
	return []string{"adminHelp", "adminHelp"}
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

func (h *Handler) Execute(msg *tgbotapi.Message, params []string) (err error) {
	err = h.Bot.SendMessage(int64(msg.Chat.ID), commandsList, true)
	return eepkg.Wrap(err, false, true, cantSendBotMessage)
}
