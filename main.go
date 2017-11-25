package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"bitbucket.org/proger4ever/draw-telegram-bot/bot"
	"bitbucket.org/proger4ever/draw-telegram-bot/commands/handlers/add-me"
	"bitbucket.org/proger4ever/draw-telegram-bot/commands/handlers/adminhelp"
	"bitbucket.org/proger4ever/draw-telegram-bot/commands/handlers/help"
	"bitbucket.org/proger4ever/draw-telegram-bot/commands/handlers/notifications"
	"bitbucket.org/proger4ever/draw-telegram-bot/commands/handlers/play"
	"bitbucket.org/proger4ever/draw-telegram-bot/commands/handlers/stat"
	"bitbucket.org/proger4ever/draw-telegram-bot/commands/handlers/test"
	"bitbucket.org/proger4ever/draw-telegram-bot/commands/handlers/user"
	"bitbucket.org/proger4ever/draw-telegram-bot/commands/routing"
	"bitbucket.org/proger4ever/draw-telegram-bot/config"
	"bitbucket.org/proger4ever/draw-telegram-bot/error"
	"bitbucket.org/proger4ever/draw-telegram-bot/mongo"
	"bitbucket.org/proger4ever/draw-telegram-bot/mongo/models/user"
	appUtils "bitbucket.org/proger4ever/draw-telegram-bot/utils/app"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

const incomingMessage = `Got msg for me: %v
    at chat: id%d %s
    from:    id%d %s <%s %s>
`

const systemErrorText = `Извините, произошла системная ошибка бота.
Обратитесь к администраторам на канале @mazimota_chat`

var systemError = eepkg.New(true, false, systemErrorText)

var helpHandler *helppkg.Handler
var privateRouter *routing.BaseRouter
var publicRouter *routing.BaseRouter
var bot *botpkg.Bot

func main() {
	rand.NewSource(time.Now().UnixNano())

	conf, ee := config.LoadConfig("config.json")
	appUtils.PanicIfExtended(ee, "reading/decoding config file")

	//region mongo
	mongoConnection, ee := mongo.InitDefaultConnection(conf.Mongo.Host, conf.Mongo.Port)
	appUtils.PanicIfExtended(ee, "while connecting to mongo")
	defer mongoConnection.Close()
	fmt.Println("Mongo session open")

	// ee = settingState.NewCollectionDefault().EnsureIndexes()
	// common.PanicIfError(ee, "while ensuring setting-state indexes")
	ee = user.NewCollectionDefault().EnsureIndexes()
	appUtils.PanicIfExtended(ee, "while ensuring user indexes")
	fmt.Println("All indexes are ensured")

	// NOTE: User Api disabled
	//initUserApi(&conf.UserApi)

	bot = &botpkg.Bot{}
	ee = bot.Init(conf)
	appUtils.PanicIfExtended(ee, "initializing Bot API")
	fmt.Printf("Authorized on bot %s\n", bot.BotApi.Self.UserName)

	helpHandler = &helppkg.Handler{}
	privateHandlers := []routing.CommandHandler{
		&addmepkg.Handler{},
		helpHandler,
		&statpkg.Handler{},
		&notificationspkg.Handler{},
		&adminhelppkg.Handler{},
		&userpkg.Handler{},
		&testpkg.Handler{},
	}
	privateRouter = routing.New(bot.Conf, bot.Tool, bot, privateHandlers)

	publicHandlers := []routing.CommandHandler{
		&playpkg.Handler{},
	}
	publicRouter = routing.New(bot.Conf, bot.Tool, bot, publicHandlers)

	listen(bot)
}

func listen(bot *botpkg.Bot) {
	for {
		select {
		case update := <-bot.Updates:
			processUpdate(&update)
			break
			// NOTE: User Api disabled
			// case stateObj := <-tool.StateCh:
			// 	if uac.Debug > 0 {
			// 		fmt.Println("saving stateObj to mongo...")
			// 	}
			// 	state.Save(&stateObj)
		}
	}
}

func processUpdate(update *tgbotapi.Update) {
	defer appUtils.TraceIfPanic("ProcessUpdate()", update)

	var msg *tgbotapi.Message
	if update.Message != nil {
		msg = update.Message
	} else if update.ChannelPost != nil {
		msg = update.ChannelPost
	}

	if msg != nil && len(msg.Text) > 0 {
		err := processMessage(msg)
		handleIfExtendedErrorMessage(msg, err)
	}
}
func processMessage(msg *tgbotapi.Message) (err *eepkg.ExtendedError) {
	defer func() {
		handleIfErrorMessage(msg, recover())
	}()

	logMessage(msg.Text, msg)
	cmd := routing.GetFullCommand(msg.Text)
	if len(cmd) == 0 {
		return
	}
	cmdName, cmdBot := routing.ParseCommand(cmd)
	if cmdBot != "" && cmdBot != bot.BotApi.Self.UserName {
		return
	}

	if msg.Chat.IsPrivate() {
		if cmdName[0] == '/' {
			cmdName = cmdName[1:]
		}
		err = privateRouter.Execute(cmdName, msg)
		if err == routing.CommandNotFound {
			err = helpHandler.Execute(msg, []string{cmdName})
		}
	} else if msg.Chat.IsChannel() {
		err = publicRouter.Execute(cmdName, msg)
		if err == routing.CommandNotFound {
			err = nil
		}
	}
	return
}

func logMessage(txt string, msg *tgbotapi.Message) {
	var (
		fromID                      int
		fromUserName                string
		fromFirstName, fromLastName string
	)
	if msg.From != nil {
		fromID = msg.From.ID
		fromUserName = msg.From.UserName
		fromFirstName = msg.From.FirstName
		fromLastName = msg.From.LastName
	}

	fmt.Printf(incomingMessage, txt, msg.Chat.ID, msg.Chat.UserName, fromID, fromUserName, fromFirstName, fromLastName)
}

func handleIfExtendedErrorMessage(msg *tgbotapi.Message, err *eepkg.ExtendedError) {
	if err == nil {
		return
	}

	isUserCause := err.Data().(bool)
	if isUserCause {
		if msg.Chat.IsPrivate() {
			_ = bot.SendErrorUserKeyboard(msg.Chat.ID, err.GetRoot()) //Игнорим error, как мы это любим
		} else {
			_ = bot.SendError(msg.Chat.ID, err.GetRoot()) //Игнорим error, как мы это любим
		}
	} else {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		//TODO: send error to owner
		_ = bot.SendError(msg.Chat.ID, systemError) //Игнорим error, как мы это любим
	}
}

func handleIfErrorMessage(msg *tgbotapi.Message, errI interface{}) {
	if errI == nil {
		return
	}
	err := eepkg.Wrap(errI.(error), false, true, "Unexpected error")
	handleIfExtendedErrorMessage(msg, err)
}

//func handleIfExtendedErrorMessage(msg *tgbotapi.Message, errI interface{}) {
//	if errI == nil {
//		return
//	}
//
//	err := errI.(error)
//	errActual := err
//
//	var isUserCause bool
//	ext, isEE := err.(*eepkg.ExtendedError)
//	if isEE {
//		isUserCause, _ = ext.Data().(bool)
//	} else {
//		errActual = eepkg.Wrap(err, false, true, "Unexpected error")
//	}
//
//	if isUserCause {
//		if msg.Chat.IsPrivate() {
//			_ = bot.SendErrorUserKeyboard(msg.Chat.ID, ext.GetRoot()) //Игнорим error, как мы это любим
//		} else {
//			_ = bot.SendError(msg.Chat.ID, ext.GetRoot()) //Игнорим error, как мы это любим
//		}
//	} else {
//		fmt.Fprintf(os.Stderr, "%+v\n", errActual)
//		//TODO: send error to owner
//		_ = bot.SendError(msg.Chat.ID, systemError) //Игнорим error, как мы это любим
//	}
//}

// NOTE: User Api disabled
//func initUserApiConf(bac *config.BotApiConfig) (err error) {
//	 stateObj, err := state.Load()
//	 if err == nil {
//	 	// stateObj = settingState.DecodeValue()
//	 	//stateObj = stateSetting.Value
//	 	fmt.Println("Bot state loaded from mongo")
//	 } else {
//	 	if err == mgo.ErrNotFound {
//	 		fmt.Println("Clear state created for bot")
//	 	} else {
//	 		common.PanicIfError(err, "loading saved bot state from mongo")
//	 	}
//	 }
//	 fmt.Printf("stateObj: %q", stateObj)
//}

// NOTE: User Api disabled
//func initUserApi(bac *config.BotApiConfig) (err error) {
//	 uac := conf.UserApi
//	 tool := &userapi.Tool{}
//	 err = tool.Run(stateObj, uac.Host, uac.Port, uac.PublicKey, uac.ApiId, uac.ApiHash, uac.Debug)
//	 common.PanicIfError(err, "connecting to Telegram User API")
//	 //fmt.Println(tool)
//	 fmt.Println("Connected to Telegram User API.")
//}
