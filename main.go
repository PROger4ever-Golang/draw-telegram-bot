package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/PROger4ever/telegram-bot-api"

	pkgaddme "bitbucket.org/proger4ever/draw-telegram-bot/commands/handlers/add-me"
	pkghelp "bitbucket.org/proger4ever/draw-telegram-bot/commands/handlers/help"
	pkgplay "bitbucket.org/proger4ever/draw-telegram-bot/commands/handlers/play"
	"bitbucket.org/proger4ever/draw-telegram-bot/commands/routing"
	"bitbucket.org/proger4ever/draw-telegram-bot/common"
	"bitbucket.org/proger4ever/draw-telegram-bot/config"
	"bitbucket.org/proger4ever/draw-telegram-bot/mongo"
	"bitbucket.org/proger4ever/draw-telegram-bot/mongo/models/user"
	"bitbucket.org/proger4ever/draw-telegram-bot/userApi"
)

func Main() {
	main()
}

func main() {
	rand.NewSource(time.Now().UnixNano())

	conf, err := config.LoadConfig("config.json")
	common.PanicIfError(err, "reading/decoding config file")

	//region mongo
	mongoConnection, err := mongo.InitDefaultConnection(conf.Mongo.Host, conf.Mongo.Port)
	common.PanicIfError(err, "while connecting to mongo")
	defer mongoConnection.Close()
	fmt.Println("Mongo session open")

	// err = settingState.NewCollectionDefault().EnsureIndexes()
	// common.PanicIfError(err, "while ensuring setting-state indexes")
	err = user.NewCollectionDefault().EnsureIndexes()
	common.PanicIfError(err, "while ensuring user indexes")
	fmt.Println("All indexes are ensured")

	// NOTE: User Api disabled
	// stateObj, err := state.Load()
	// if err == nil {
	// 	// stateObj = settingState.DecodeValue()
	// 	//stateObj = stateSetting.Value
	// 	fmt.Println("Bot state loaded from mongo")
	// } else {
	// 	if err == mgo.ErrNotFound {
	// 		fmt.Println("Clear state created for bot")
	// 	} else {
	// 		common.PanicIfError(err, "loading saved bot state from mongo")
	// 	}
	// }
	// fmt.Printf("stateObj: %q", stateObj)
	//endregion

	//region user api
	// NOTE: User Api disabled
	var tool *userapi.Tool
	// uac := conf.UserApi
	// tool := &userapi.Tool{}
	// err = tool.Run(stateObj, uac.Host, uac.Port, uac.PublicKey, uac.ApiId, uac.ApiHash, uac.Debug)
	// common.PanicIfError(err, "connecting to Telegram User API")
	// //fmt.Println(tool)
	// fmt.Println("Connected to Telegram User API.")
	//endregion

	//region bot
	bac := conf.BotApi
	competeKey := fmt.Sprintf("%v:%v", bac.ID, bac.Key)
	bot, err := tgbotapi.NewBotAPI(competeKey)
	common.PanicIfError(err, "creating bot instance")
	bot.Debug = bac.Debug
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)
	common.PanicIfError(err, "getting updates chan")
	fmt.Printf("Authorized on bot %s\n", bot.Self.UserName)
	//endregion

	//region routing
	helpCommand := &pkghelp.Handler{}
	handlers := []routing.CommandHandler{
		&pkgaddme.Handler{},
		helpCommand,
		&pkgplay.Handler{},
		// &handlers.StartLoginHandler{},
		// &handlers.CompleteLoginWithCodeHandler{},
	}
	router := routing.Router{}
	router.Init(handlers, helpCommand, conf, tool, bot)
	router.InitCommands()
	//endregion

	for {
		select {
		case update := <-updates:
			router.ProcessUpdate(&update)
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
