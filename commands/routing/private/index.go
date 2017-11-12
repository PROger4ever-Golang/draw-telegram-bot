package private

import (
	"fmt"

	"bitbucket.org/proger4ever/draw-telegram-bot/bot"
	"bitbucket.org/proger4ever/draw-telegram-bot/commands/handlers/add-me"
	"bitbucket.org/proger4ever/draw-telegram-bot/commands/handlers/adminhelp"
	"bitbucket.org/proger4ever/draw-telegram-bot/commands/handlers/help"
	"bitbucket.org/proger4ever/draw-telegram-bot/commands/handlers/notifications"
	"bitbucket.org/proger4ever/draw-telegram-bot/commands/handlers/stat"
	"bitbucket.org/proger4ever/draw-telegram-bot/commands/routing"
	"bitbucket.org/proger4ever/draw-telegram-bot/config"
	"bitbucket.org/proger4ever/draw-telegram-bot/error"
	"bitbucket.org/proger4ever/draw-telegram-bot/userApi"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

const incomingCommand = `Got private cmd for me: %v, params: %q
  from: id%d %s <%s %s>
`

type Router struct {
	*routing.BaseRouter

	HelpHandler *helppkg.Handler
}

func (r *Router) Execute(msg *tgbotapi.Message) (err error) {
	cmd := r.GetFullCommand(msg.Text)
	if len(cmd) == 0 {
		return
	}
	h, found := r.GetHandler(cmd)
	if !found {
		return r.HelpHandler.Execute(msg, []string{cmd})
	}
	params := r.GetParams(msg.Text, len(cmd)+1)
	if err = r.CheckParams(h, params); err != nil {
		return err
	}

	if h.IsForOwnersOnly() {
		isOwner := msg.From != nil && msg.From.UserName == r.Conf.Management.OwnerUsername
		isOwnerChannel := msg.Chat.UserName == r.Conf.Management.ChannelUsername
		if !isOwner && !isOwnerChannel {
			return eepkg.New(true, false, "Эта команда доступна только моему ПОВЕЛИТЕЛЮ! Я тебя не слушаюсь!")
		}
	}

	fmt.Printf(incomingCommand, cmd, params, msg.From.ID, msg.From.UserName, msg.From.FirstName, msg.From.LastName)

	return h.Execute(msg, params)
}

func New(conf *config.Config, tool *userapi.Tool, bot *botpkg.Bot) (r *Router) {
	r = &Router{
		BaseRouter:  &routing.BaseRouter{},
		HelpHandler: &helppkg.Handler{},
	}
	handlers := []routing.CommandHandler{
		&addmepkg.Handler{},
		r.HelpHandler,
		&statpkg.Handler{},
		&notificationspkg.Handler{},
		&adminhelppkg.Handler{},
		// &handlers.StartLoginHandler{},
		// &handlers.CompleteLoginWithCodeHandler{},
	}
	r.Init(bot, conf, tool, handlers)
	return r
}
