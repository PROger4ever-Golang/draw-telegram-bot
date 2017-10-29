package routing

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"gopkg.in/mgo.v2"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"gopkg.in/mgo.v2/bson"

	"bitbucket.org/proger4ever/draw-telegram-bot/commands/handlers"
	"bitbucket.org/proger4ever/draw-telegram-bot/commands/utils"
	"bitbucket.org/proger4ever/draw-telegram-bot/common"
	"bitbucket.org/proger4ever/draw-telegram-bot/config"
	"bitbucket.org/proger4ever/draw-telegram-bot/mongo/models/user"
	"bitbucket.org/proger4ever/draw-telegram-bot/userApi"
)

var cmdRegexp = regexp.MustCompile("^/([A-Za-z0-9_-]+)(@([A-Za-z0-9_-]+))? ?(.*)")

type CommandHandler interface {
	GetName() string
	GetParamsCount() int

	Init(conf *config.Config, tool *userapi.Tool, bot *tgbotapi.BotAPI)
	Execute(chat *tgbotapi.Chat, params []string) error
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
		&handlers.CompleteLoginWithCodeHandler{},
		&handlers.PlayHandler{},
		&handlers.StartLoginHandler{},
	}
	r.HandlersMap = make(map[string]CommandHandler)
	for _, h := range r.Handlers {
		h.Init(r.Conf, r.Tool, r.Bot)
		r.HandlersMap[h.GetName()] = h
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
		err := r.processMessage(update, msg.Chat, msg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Expected error while processing message: %v\n", err)
		}
	}

	if msg.NewChatMembers != nil {
		uc := user.NewCollectionDefault()
		for _, tuser := range *msg.NewChatMembers {
			chatMember, err := r.Bot.GetChatMember(tgbotapi.ChatConfigWithUser{
				ChatID: msg.Chat.ID,
				UserID: tuser.ID,
			})
			common.PanicIfError(err, "getting info about NewChatMembers")
			u := user.User{
				TelegramID: tuser.ID,
				Username:   tuser.UserName,
				LastName:   tuser.LastName,
				FirstName:  tuser.LastName,
				Status:     chatMember.Status,
			}
			u.Init(uc)
			err = uc.InsertOneOrUpdateModel(&u)
			common.PanicIfError(err, "saving NewChatMembers to db")
		}
	}

	if msg.LeftChatMember != nil {
		err := user.NewCollectionDefault().UpdateInterface(bson.M{
			"telegram_id": msg.LeftChatMember.ID,
		}, bson.M{
			"$set": bson.M{
				"status":     "left",
				"deleted_at": time.Now(),
			},
		})
		if err == mgo.ErrNotFound {
			err = nil
		}
		common.PanicIfError(err, "updating LeftChatMember in db")
	}
}

func (r *Router) processMessage(update *tgbotapi.Update, chat *tgbotapi.Chat, msg *tgbotapi.Message) error {
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

	fmt.Printf("Got cmd for me: %v, params: %q\n", cmdName, cmdParams)
	if msg.From != nil {
		fmt.Printf("From: id%d %s <%s %s>\n", msg.From.ID, msg.From.UserName, msg.From.FirstName, msg.From.LastName)
		if msg.From.UserName != r.Conf.Management.OwnerUsername {
			return utils.SendBotError(r.Bot, int64(msg.Chat.ID), errors.New("Ты не мой ПОВЕЛИТЕЛЬ! Я тебя не слушаюсь!"))
		}
	}
	return r.processCmd(update, chat, cmdName, cmdParams)
}

func (r *Router) processCmd(update *tgbotapi.Update, chat *tgbotapi.Chat, name string, params []string) error {
	h, hFound := r.HandlersMap[name]
	if !hFound {
		err := fmt.Errorf("Неизвестная команда: %v", name)
		return utils.SendBotError(r.Bot, int64(chat.ID), err)
	}
	if len(params) != h.GetParamsCount() {
		err := fmt.Errorf("Неверное количество параметров: %v. Ожидалось: %v", len(params), h.GetParamsCount())
		return utils.SendBotError(r.Bot, int64(chat.ID), err)
	}

	return h.Execute(chat, params)
}
