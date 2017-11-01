package routing

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/go-telegram-bot-api/telegram-bot-api"

	pkgaddme "bitbucket.org/proger4ever/draw-telegram-bot/commands/handlers/add-me"
	pkgplay "bitbucket.org/proger4ever/draw-telegram-bot/commands/handlers/play"
	"bitbucket.org/proger4ever/draw-telegram-bot/commands/utils"
	"bitbucket.org/proger4ever/draw-telegram-bot/common"
	"bitbucket.org/proger4ever/draw-telegram-bot/config"
	ee "bitbucket.org/proger4ever/draw-telegram-bot/errors"
	"bitbucket.org/proger4ever/draw-telegram-bot/userApi"
)

var cmdRegexp = regexp.MustCompile("^/([A-Za-z0-9_-]+)(@([A-Za-z0-9_-]+))? ?(.*)")

type CommandHandler interface {
	GetNames() []string
	IsForOwnersOnly() bool
	GetParamsCount() int

	Init(conf *config.Config, tool *userapi.Tool, bot *tgbotapi.BotAPI)
	Execute(msg *tgbotapi.Message, params []string) error
}

type Router struct {
	Bot  *tgbotapi.BotAPI
	Conf *config.Config
	Tool *userapi.Tool

	Handlers    []CommandHandler
	HandlersMap map[string]CommandHandler
}

func (r *Router) Init(conf *config.Config, tool *userapi.Tool, bot *tgbotapi.BotAPI) {
	r.Bot = bot
	r.Conf = conf
	r.Tool = tool
	r.initCommands()
}

func (r *Router) initCommands() {
	r.Handlers = []CommandHandler{
		&pkgaddme.Handler{},
		&pkgplay.Handler{},
		// &handlers.StartLoginHandler{},
		// &handlers.CompleteLoginWithCodeHandler{},
	}
	r.HandlersMap = make(map[string]CommandHandler)
	for _, handler := range r.Handlers {
		handler.Init(r.Conf, r.Tool, r.Bot)
		for _, name := range handler.GetNames() {
			name = strings.ToLower(name)
			r.HandlersMap[name] = handler
		}
	}
}

func (r *Router) ProcessUpdate(update *tgbotapi.Update) {
	defer common.TraceIfPanic("ProcessUpdate()", update)

	var msg *tgbotapi.Message
	if update.Message != nil {
		msg = update.Message
	} else if update.ChannelPost != nil {
		msg = update.ChannelPost
	}

	if msg != nil && len(msg.Text) > 0 {
		err := r.processMessage(msg)
		r.handleIfErrorMessage(msg, err)
	}
}

func (r *Router) processMessage(msg *tgbotapi.Message) error {
	defer func() {
		r.handleIfErrorMessage(msg, recover())
	}()

	cmdSubmatches := cmdRegexp.FindStringSubmatch(msg.Text)
	if len(cmdSubmatches) == 0 {
		return nil
	}

	cmdName := cmdSubmatches[1]
	cmdBot := cmdSubmatches[3]
	cmdParams := strings.Fields(cmdSubmatches[4])
	if cmdBot != "" && cmdBot != r.Bot.Self.UserName {
		return nil
	}
	return r.processCmd(msg, cmdName, cmdParams)
}

func (r *Router) processCmd(msg *tgbotapi.Message, name string, params []string) error {
	fmt.Printf("Got cmd for me: %v, params: %q\n  at chat @%s %d\n", name, params, msg.Chat.UserName, msg.Chat.ID)
	if msg.From != nil {
		fmt.Printf("  from: id%d %s <%s %s>\n", msg.From.ID, msg.From.UserName, msg.From.FirstName, msg.From.LastName)
	}

	h, hFound := r.HandlersMap[strings.ToLower(name)]
	if !hFound {
		return ee.Newf(true, false, "Неизвестная команда: %v", name)
	}

	if h.IsForOwnersOnly() && (msg.From == nil || msg.From.UserName != r.Conf.Management.OwnerUsername) {
		return ee.New(true, false, "Эта команда доступна только моему ПОВЕЛИТЕЛЮ! Я тебя не слушаюсь!")
	}

	if len(params) != h.GetParamsCount() {
		return ee.Newf(true, false, "Неверное количество параметров: %v. Ожидалось: %v", len(params), h.GetParamsCount())
	}

	return h.Execute(msg, params)
}

func (r *Router) handleIfErrorMessage(msg *tgbotapi.Message, errI interface{}) {
	if errI == nil {
		return
	}

	err := errI.(error)
	errActual := err

	var isUserCause bool
	ext, isEE := err.(*ee.ExtendedError)
	if isEE {
		isUserCause, _ = ext.Data().(bool)
	} else {
		errActual = ee.Wrap(err, false, true, "Unexpected error")
	}

	if isUserCause {
		utils.SendBotError(r.Bot, msg.Chat.ID, ext.GetRoot())
	} else {
		fmt.Fprintf(os.Stderr, "%+v\n", errActual)
		//TODO: send error to owner
		// utils.SendBotError(r.Bot, msg.Chat.ID, errActual)
	}
}
