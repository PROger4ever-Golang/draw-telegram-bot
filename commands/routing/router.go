package routing

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/go-telegram-bot-api/telegram-bot-api"

	"bitbucket.org/proger4ever/draw-telegram-bot/commands/handlers"
	"bitbucket.org/proger4ever/draw-telegram-bot/commands/utils"
	"bitbucket.org/proger4ever/draw-telegram-bot/common"
	"bitbucket.org/proger4ever/draw-telegram-bot/config"
	"bitbucket.org/proger4ever/draw-telegram-bot/userApi"
)

var cmdRegexp = regexp.MustCompile("^/([A-Za-z0-9_-]+)(@([A-Za-z0-9_-]+))? ?(.*)")

type CommandHandler interface {
	GetName() string
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
		&handlers.AddMeHandler{},
		&handlers.PlayHandler{},
		// &handlers.StartLoginHandler{},
		// &handlers.CompleteLoginWithCodeHandler{},
	}
	r.HandlersMap = make(map[string]CommandHandler)
	for _, h := range r.Handlers {
		h.Init(r.Conf, r.Tool, r.Bot)
		r.HandlersMap[strings.ToLower(h.GetName())] = h
	}
}

func (r *Router) ProcessUpdate(update *tgbotapi.Update) {
	defer common.RepairIfError("processing update", update)

	var msg *tgbotapi.Message
	if update.Message != nil {
		msg = update.Message
	} else if update.ChannelPost != nil {
		msg = update.ChannelPost
	}

	if msg != nil && len(msg.Text) > 0 {
		err := r.processMessage(msg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Expected error while processing message: %v\n", err)
		}
	}
}

func (r *Router) processMessage(msg *tgbotapi.Message) error {
	txt := msg.Text
	cmdSubmatches := cmdRegexp.FindStringSubmatch(txt)
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
		err := fmt.Errorf("Неизвестная команда: %v", name)
		return utils.SendBotError(r.Bot, int64(msg.Chat.ID), err)
	}

	if h.IsForOwnersOnly() && (msg.From == nil || msg.From.UserName != r.Conf.Management.OwnerUsername) {
		err := errors.New("Эта команда доступна только моему ПОВЕЛИТЕЛЮ! Я тебя не слушаюсь!")
		return utils.SendBotError(r.Bot, int64(msg.Chat.ID), err)
	}

	if len(params) != h.GetParamsCount() {
		err := fmt.Errorf("Неверное количество параметров: %v. Ожидалось: %v", len(params), h.GetParamsCount())
		return utils.SendBotError(r.Bot, int64(msg.Chat.ID), err)
	}

	return h.Execute(msg, params)
}
