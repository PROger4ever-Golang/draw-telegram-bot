package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"gopkg.in/mgo.v2"

	"bitbucket.org/proger4ever/drawtelegrambot/commands/routing"
	"bitbucket.org/proger4ever/drawtelegrambot/common"
	"bitbucket.org/proger4ever/drawtelegrambot/config"
	"bitbucket.org/proger4ever/drawtelegrambot/stateSerializable"
	"bitbucket.org/proger4ever/drawtelegrambot/userApi"
)

func main() {
	rand.NewSource(time.Now().UnixNano())

	conf, err := config.LoadConfig("config.json")
	common.PanicIfError(err, "reading/decoding config file")

	//region mongo
	mongoSession, state, err := stateSerializable.Init(conf)
	defer mongoSession.Close()
	if err == nil {
		// state = settingState.DecodeValue()
		//state = stateSetting.Value
		fmt.Println("Bot state loaded from mongo")
	} else {
		if err == mgo.ErrNotFound {
			fmt.Println("Clear state created for bot")
		} else {
			common.PanicIfError(err, "loading saved bot state from mongo")
		}
	}
	fmt.Printf("state: %q", state)
	//endregion

	//region user api
	uac := conf.UserApi
	tool := &userapi.Tool{}
	err = tool.Run(state, uac.Host, uac.Port, uac.PublicKey, uac.ApiId, uac.ApiHash, uac.Debug)
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
		case state := <-tool.StateCh:
			if uac.Debug > 0 {
				fmt.Println("saving state to mongo...")
			}
			stateSerializable.Save(mongoSession, &state)
		}
	}
}
