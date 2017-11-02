package routing

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/go-telegram-bot-api/telegram-bot-api"

	"bitbucket.org/proger4ever/draw-telegram-bot/commands/utils"
	"bitbucket.org/proger4ever/draw-telegram-bot/common"
	"bitbucket.org/proger4ever/draw-telegram-bot/config"
	ee "bitbucket.org/proger4ever/draw-telegram-bot/errors"
	"bitbucket.org/proger4ever/draw-telegram-bot/userApi"
)

const systemErrorText = `Извините, произошла системная ошибка бота.
Обратитесь к администраторам на канале @mazimota_chat`

var systemError = ee.New(true, false, systemErrorText)

var cmdRegexp = regexp.MustCompile("^/([A-Za-z0-9_-]+)(@([A-Za-z0-9_-]+))? ?(.*)")

type CommandHandler interface {
	GetAliases() []string
	IsForOwnersOnly() bool
	GetParamsMinCount() int

	Init(conf *config.Config, tool *userapi.Tool, bot *tgbotapi.BotAPI)
	Execute(msg *tgbotapi.Message, params []string) error
}

type Router struct {
	Bot         *tgbotapi.BotAPI
	Conf        *config.Config
	Tool        *userapi.Tool
	HelpCommand CommandHandler

	Handlers    []CommandHandler
	HandlersMap map[string]CommandHandler
}

func (r *Router) Init(handlers []CommandHandler, helpCommand CommandHandler, conf *config.Config, tool *userapi.Tool, bot *tgbotapi.BotAPI) {
	r.Handlers = handlers
	r.HelpCommand = helpCommand
	r.Bot = bot
	r.Conf = conf
	r.Tool = tool
}

func (r *Router) InitCommands() {
	r.HandlersMap = make(map[string]CommandHandler)
	for _, handler := range r.Handlers {
		handler.Init(r.Conf, r.Tool, r.Bot)
		for _, alias := range handler.GetAliases() {
			alias = strings.ToLower(alias)
			r.HandlersMap[alias] = handler
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
		return r.HelpCommand.Execute(msg, []string{name})
	}

	if h.IsForOwnersOnly() && (msg.From == nil || msg.From.UserName != r.Conf.Management.OwnerUsername) {
		return ee.New(true, false, "Эта команда доступна только моему ПОВЕЛИТЕЛЮ! Я тебя не слушаюсь!")
	}

	if len(params) < h.GetParamsMinCount() {
		return ee.Newf(true, false, "Неверное количество параметров: %v. Ожидалось: %v", len(params), h.GetParamsMinCount())
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
		_ = utils.SendBotError(r.Bot, msg.Chat.ID, ext.GetRoot()) //Игнорим error, как мы это любим
	} else {
		fmt.Fprintf(os.Stderr, "%+v\n", errActual)
		//TODO: send error to owner
		_ = utils.SendBotError(r.Bot, msg.Chat.ID, systemError) //Игнорим error, как мы это любим
	}
}
