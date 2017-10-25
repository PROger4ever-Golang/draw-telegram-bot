package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"gopkg.in/mgo.v2"

	"bitbucket.org/proger4ever/draw-telegram-bot/commands/routing"
	"bitbucket.org/proger4ever/draw-telegram-bot/common"
	"bitbucket.org/proger4ever/draw-telegram-bot/config"
	"bitbucket.org/proger4ever/draw-telegram-bot/mongo"
	"bitbucket.org/proger4ever/draw-telegram-bot/state"
	"bitbucket.org/proger4ever/draw-telegram-bot/userApi"
)

func main() {
	rand.NewSource(time.Now().UnixNano())

	conf, err := config.LoadConfig("config.json")
	common.PanicIfError(err, "reading/decoding config file")

	//region mongo
	mongoConnection, err := mongo.NewConnection(conf.Mongo.Host, conf.Mongo.Port)
	common.PanicIfError(err, "while connecting to mongo")
	defer mongoConnection.Close()
	stateObj, err := state.Load(mongoConnection)
	if err == nil {
		// stateObj = settingState.DecodeValue()
		//stateObj = stateSetting.Value
		fmt.Println("Bot state loaded from mongo")
	} else {
		if err == mgo.ErrNotFound {
			fmt.Println("Clear state created for bot")
		} else {
			common.PanicIfError(err, "loading saved bot state from mongo")
		}
	}
	fmt.Printf("stateObj: %q", stateObj)
	//endregion

	// A case of using models
	// uc := user.NewCollection(mongoConnection)
	// us := user.New(uc)
	// us.TelegramID = 10555555
	// us.LastName = "LastName"
	// us.FirstName = "FirstName"
	// us.Username = "Username"
	// _, err = us.UpsertId()
	// common.PanicIfError(err, "user saving 1")

	// us.LastName = "LastName Changed"
	// _, err = us.UpsertId()
	// common.PanicIfError(err, "user saving 2")
	// os.Exit(0)

	//region user api
	uac := conf.UserApi
	tool := &userapi.Tool{}
	err = tool.Run(stateObj, uac.Host, uac.Port, uac.PublicKey, uac.ApiId, uac.ApiHash, uac.Debug)
	common.PanicIfError(err, "connecting to Telegram User API")
	//fmt.Println(tool)
	fmt.Println("Connected to Telegram User API.")
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

	router := routing.Router{}
	router.Init(conf, tool, bot)

	for {
		select {
		case update := <-updates:
			router.ProcessUpdate(&update)
			break
		case stateObj := <-tool.StateCh:
			if uac.Debug > 0 {
				fmt.Println("saving stateObj to mongo...")
			}
			state.Save(mongoConnection, &stateObj)
		}
	}
}
