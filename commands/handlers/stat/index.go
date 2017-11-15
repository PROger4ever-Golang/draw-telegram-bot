package statpkg

import (
	"fmt"

	"bitbucket.org/proger4ever/draw-telegram-bot/bot"
	"bitbucket.org/proger4ever/draw-telegram-bot/config"
	"bitbucket.org/proger4ever/draw-telegram-bot/error"
	"bitbucket.org/proger4ever/draw-telegram-bot/mongo/models/user"
	"bitbucket.org/proger4ever/draw-telegram-bot/userApi"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

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
	return []string{"стат", "статистика", "stat", "stats"}
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

func (h *Handler) Execute(msg *tgbotapi.Message, params []string) (err *eepkg.ExtendedError) {
	channelCount, err := h.Bot.GetChatMemberCount()
	if err != nil {
		return
	}

	uc := user.NewCollectionDefault()
	dbCount, err := uc.CountInterface(nil)
	if err != nil {
		return err
	}

	percent := 100. * dbCount / channelCount

	resp := fmt.Sprintf(statFormat, channelCount, dbCount, percent)
	return h.Bot.SendMessageMarkdown(int64(msg.Chat.ID), resp)
}
