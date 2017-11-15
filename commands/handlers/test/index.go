package testpkg

import (
	"fmt"
	"strconv"

	"bitbucket.org/proger4ever/draw-telegram-bot/bot"
	"bitbucket.org/proger4ever/draw-telegram-bot/config"
	"bitbucket.org/proger4ever/draw-telegram-bot/error"
	"bitbucket.org/proger4ever/draw-telegram-bot/userApi"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

type Handler struct {
	Bot  *botpkg.Bot
	Conf *config.Config
	Tool *userapi.Tool
}

func (h *Handler) GetAliases() []string {
	return []string{"test"}
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
	userLink := "tg://user?id=" + strconv.Itoa(msg.From.ID)
	resp := fmt.Sprintf("[test](%s)", userLink)
	return h.Bot.SendMessageMarkdown(int64(msg.Chat.ID), resp)
}
