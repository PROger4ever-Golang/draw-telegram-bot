package statpkg

import (
	"bitbucket.org/proger4ever/draw-telegram-bot/bot"
	"bitbucket.org/proger4ever/draw-telegram-bot/config"
	"bitbucket.org/proger4ever/draw-telegram-bot/error"
	"bitbucket.org/proger4ever/draw-telegram-bot/mongo/models/user"
	"bitbucket.org/proger4ever/draw-telegram-bot/userApi"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

const cantQueryDB = "Ошибка при операции с БД"
const cantSendBotMessage = "Ошибка при отправке сообщения от имени бота"
const cantQueryChatMembersCount = "Ошибка при получении информации о количестве участников канала"

const statFormat = `*Напоминаю, что в БД могут быть отписавшиеся ранее пользователи.*
Для обновления статусов всех пользователей в будущем будет команда /verifyUsers.
` + "```" + `
На канале пользователей: %d
В БД бота пользователей: %d
Зарегистрированных:      %v%%
` + "```"

type Handler struct {
	Bot  *botpkg.Bot
	Conf *config.Config
	Tool *userapi.Tool
}

func (h *Handler) GetAliases() []string {
	return []string{"stat", "stats"}
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
	channelCount, err := h.Bot.BotApi.GetChatMembersCount(tgbotapi.ChatConfig{
		SuperGroupUsername: "@" + h.Conf.Management.ChannelUsername,
	})
	if err != nil {
		return eepkg.Wrap(err, false, true, cantQueryChatMembersCount)
	}

	uc := user.NewCollectionDefault()
	dbCount, err := uc.CountInterface(nil)
	if err != nil {
		return eepkg.Wrap(err, false, true, cantQueryDB)
	}

	percent := 100. * dbCount / channelCount

	resp := fmt.Sprintf(statFormat, channelCount, dbCount, percent)
	err = h.Bot.SendMessage(int64(msg.Chat.ID), resp, true)
	return eepkg.Wrap(err, false, true, cantSendBotMessage)
}
