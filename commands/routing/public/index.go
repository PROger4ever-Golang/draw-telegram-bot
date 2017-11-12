package public

import (
	"fmt"
	"regexp"

	"bitbucket.org/proger4ever/draw-telegram-bot/bot"
	"bitbucket.org/proger4ever/draw-telegram-bot/commands/handlers/play"
	"bitbucket.org/proger4ever/draw-telegram-bot/commands/routing"
	"bitbucket.org/proger4ever/draw-telegram-bot/config"
	"bitbucket.org/proger4ever/draw-telegram-bot/userApi"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

const incomingCommand = "Got public cmd for me: %v, params: %q\n"

var cmdPattern = regexp.MustCompile("^(/?[^@ ]+)(@([A-Za-z0-9_-]+))?")

type Router struct {
	*routing.BaseRouter
}

func (r *Router) Execute(msg *tgbotapi.Message) (err error) {
	cmd := r.GetFullCommand(msg.Text)
	if len(cmd) == 0 {
		return
	}

	cmdSubmatches := cmdPattern.FindStringSubmatch(cmd)
	if len(cmdSubmatches) == 0 {
		return
	}
	cmdName := cmdSubmatches[1]
	cmdBot := cmdSubmatches[3]
	if cmdBot != "" && cmdBot != r.Bot.BotApi.Self.UserName {
		return
	}
	h, found := r.GetHandler(cmdName)
	if !found {
		return
	}
	params := r.GetParams(msg.Text, len(cmd)+1)
	if err = r.CheckParams(h, params); err != nil {
		return err
	}

	fmt.Printf(incomingCommand, cmdName, params)
	return h.Execute(msg, params)
}

func New(conf *config.Config, tool *userapi.Tool, bot *botpkg.Bot) (r *Router) {
	r = &Router{
		BaseRouter: &routing.BaseRouter{},
	}
	handlers := []routing.CommandHandler{
		&playpkg.Handler{},
	}
	r.Init(bot, conf, tool, handlers)
	return r
}
