package routing

import (
	"strings"

	"bitbucket.org/proger4ever/draw-telegram-bot/bot"
	"bitbucket.org/proger4ever/draw-telegram-bot/config"
	"bitbucket.org/proger4ever/draw-telegram-bot/error"
	"bitbucket.org/proger4ever/draw-telegram-bot/userApi"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

const incorrectParamsLen = "Неверное количество параметров: %v. Ожидалось: %v"

type CommandHandler interface {
	GetAliases() []string
	IsForOwnersOnly() bool
	GetParamsMinCount() int

	Init(conf *config.Config, tool *userapi.Tool, bot *botpkg.Bot)
	Execute(msg *tgbotapi.Message, params []string) error
}

type BaseRouter struct {
	Bot  *botpkg.Bot
	Conf *config.Config
	Tool *userapi.Tool

	Handlers    []CommandHandler
	HandlersMap map[string]CommandHandler
}

func (r *BaseRouter) Init(bot *botpkg.Bot, conf *config.Config, tool *userapi.Tool, handlers []CommandHandler) *BaseRouter {
	r.Bot = bot
	r.Conf = conf
	r.Tool = tool
	r.Handlers = handlers
	r.initCommands()
	return r
}

func (r *BaseRouter) GetFullCommand(text string) (cmd string) {
	cmdIndex := strings.Index(text, " ")
	if cmdIndex != -1 {
		cmd = text[:cmdIndex]
	} else {
		cmd = text
	}
	return
}
func (r *BaseRouter) GetParams(text string, start int) (params []string) {
	if start >= len(text) {
		return
	}
	paramsString := text[start:]
	return strings.Fields(paramsString)
}

func (r *BaseRouter) CheckParams(h CommandHandler, params []string) (err error) {
	if len(params) < h.GetParamsMinCount() {
		return eepkg.Newf(true, false, incorrectParamsLen, len(params), h.GetParamsMinCount())
	}
	return
}

func (r *BaseRouter) GetHandler(cmd string) (handler CommandHandler, found bool) {
	cmd = strings.ToLower(cmd)
	handler, found = r.HandlersMap[cmd]
	return
}

func (r *BaseRouter) initCommands() {
	r.HandlersMap = make(map[string]CommandHandler)
	for _, handler := range r.Handlers {
		handler.Init(r.Conf, r.Tool, r.Bot)
		for _, alias := range handler.GetAliases() {
			alias = strings.ToLower(alias)
			r.HandlersMap[alias] = handler
		}
	}
}

func New(handlers []CommandHandler, conf *config.Config, tool *userapi.Tool, bot *botpkg.Bot) (r *BaseRouter) {
	r = &BaseRouter{}
	return r.Init(bot, conf, tool, handlers)
}
