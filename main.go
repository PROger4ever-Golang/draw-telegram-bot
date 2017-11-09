package main

import (
	"fmt"
	"math/rand"
	"time"

	"bitbucket.org/proger4ever/draw-telegram-bot/bot"
	"bitbucket.org/proger4ever/draw-telegram-bot/commands/handlers/add-me"
	"bitbucket.org/proger4ever/draw-telegram-bot/commands/handlers/help"
	"bitbucket.org/proger4ever/draw-telegram-bot/commands/handlers/notifications"
	"bitbucket.org/proger4ever/draw-telegram-bot/commands/handlers/play"
	"bitbucket.org/proger4ever/draw-telegram-bot/commands/routing"
	"bitbucket.org/proger4ever/draw-telegram-bot/common"
	"bitbucket.org/proger4ever/draw-telegram-bot/config"
	"bitbucket.org/proger4ever/draw-telegram-bot/mongo"
	"bitbucket.org/proger4ever/draw-telegram-bot/mongo/models/user"
	"bitbucket.org/proger4ever/draw-telegram-bot/userApi"
)

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

	bot := botpkg.Bot{}
	err = bot.Init(&conf.BotApi)
	common.PanicIfError(err, "initializing Bot API")
	fmt.Printf("Authorized on bot %s\n", bot.BotApi.Self.UserName)

	//region routing
	helpCommand := &helppkg.Handler{}
	handlers := []routing.CommandHandler{
		&addmepkg.Handler{},
		helpCommand,
		&playpkg.Handler{},
		&notificationspkg.Handler{},
		// &handlers.StartLoginHandler{},
		// &handlers.CompleteLoginWithCodeHandler{},
	}
	router := routing.Router{}
	router.Init(handlers, helpCommand, conf, tool, &bot)
	router.InitCommands()
	//endregion

	for {
		select {
		case update := <-bot.Updates:
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
