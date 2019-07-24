package routing

import (
	"fmt"
	"strings"

	"github.com/PROger4ever-Golang/draw-telegram-bot/bot"
	"github.com/PROger4ever-Golang/draw-telegram-bot/config"
	"github.com/PROger4ever-Golang/draw-telegram-bot/error"
	"github.com/PROger4ever-Golang/draw-telegram-bot/userApi"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

const adminsOnlyCommand = "Эта команда доступна только моему ПОВЕЛИТЕЛЮ! Я тебя не слушаюсь!"
const incorrectParamsLen = "Неверное количество параметров: %v. Ожидалось: %v"

const incomingCommand = `    cmd: %v, params: %q

`

const commandNotFoundText = "Команда не найдена"

var CommandNotFound = eepkg.New(true, false, commandNotFoundText)

type CommandHandler interface {
	GetAliases() []string
	IsForOwnersOnly() bool
	GetParamsMinCount() int

	Init(conf *config.Config, tool *userapi.Tool, bot *botpkg.Bot)
	Execute(msg *tgbotapi.Message, params []string) *eepkg.ExtendedError
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

func (r *BaseRouter) Execute(cmdName string, msg *tgbotapi.Message) (err *eepkg.ExtendedError) {
	h, found := r.GetHandler(cmdName)
	if !found {
		return CommandNotFound
	}
	params := GetParams(msg.Text, len(cmdName)+1)
	if err = CheckParams(h, params); err != nil {
		return err
	}

	if h.IsForOwnersOnly() {
		isOwner := msg.From != nil && msg.From.UserName == r.Conf.Management.OwnerUsername
		isOwnerChannel := msg.Chat.UserName == r.Conf.Management.ChannelUsername
		if !isOwner && !isOwnerChannel {
			return eepkg.New(true, false, adminsOnlyCommand)
		}
	}

	LogRequest(cmdName, params, msg)
	return h.Execute(msg, params)
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

func New(conf *config.Config, tool *userapi.Tool, bot *botpkg.Bot, handlers []CommandHandler) (r *BaseRouter) {
	r = &BaseRouter{}
	return r.Init(bot, conf, tool, handlers)
}

func LogRequest(cmdName string, params []string, msg *tgbotapi.Message) {
	fmt.Printf(incomingCommand, cmdName, params)
}
func GetFullCommand(text string) (cmd string) {
	cmdIndex := strings.Index(text, " ")
	if cmdIndex != -1 {
		cmd = text[:cmdIndex]
	} else {
		cmd = text
	}
	return
}
func ParseCommand(fullCmd string) (cmd string, bot string) {
	cmdLastIndex := strings.LastIndex(fullCmd, "@")
	if cmdLastIndex == -1 {
		cmdLastIndex = len(fullCmd)
	}
	cmd = fullCmd[0:cmdLastIndex]

	if cmdLastIndex+1 < len(fullCmd) {
		bot = fullCmd[cmdLastIndex+1:]
	}
	return
}
func GetParams(text string, start int) (params []string) {
	if start >= len(text) {
		return
	}
	paramsString := text[start:]
	return strings.Fields(paramsString)
}

func CheckParams(h CommandHandler, params []string) (err *eepkg.ExtendedError) {
	if len(params) < h.GetParamsMinCount() {
		return eepkg.Newf(true, false, incorrectParamsLen, len(params), h.GetParamsMinCount())
	}
	return
}
