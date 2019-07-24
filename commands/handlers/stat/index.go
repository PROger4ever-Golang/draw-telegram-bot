package statpkg

import (
	"fmt"

	"github.com/PROger4ever-Golang/draw-telegram-bot/bot"
	"github.com/PROger4ever-Golang/draw-telegram-bot/config"
	"github.com/PROger4ever-Golang/draw-telegram-bot/error"
	"github.com/PROger4ever-Golang/draw-telegram-bot/mongo/models/user"
	"github.com/PROger4ever-Golang/draw-telegram-bot/userApi"
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
