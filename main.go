package main

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	mgo "gopkg.in/mgo.v2"

	"common"
	"config"
	"telegram/userapi"
)

var cmdRegexp = regexp.MustCompile("^/([A-Za-z0-9_-]+)@([A-Za-z0-9_-]+) ?(.*)")

var tool *userapi.Tool

func processCmd(bot *tgbotapi.BotAPI, update tgbotapi.Update, chatID int64, name string, params []string) error {
	switch name {
	case "draw":
		resp := "Starting a new drawing..."
		msg := tgbotapi.NewMessage(chatID, resp)
		bot.Send(msg)

		// membersCount, err := bot.GetChatMembersCount(tgbotapi.ChatConfig{
		// 	chatID,
		// 	"",
		// })
		// common.PanicIfError(err, "GetChatMembersCount")

		// memberID := rand.Intn(membersCount)

		// member, err := bot.GetChatMember(tgbotapi.ChatConfigWithUser{
		// 	chatID,
		// 	"",
		// 	memberID,
		// })
		// common.PanicIfError(err, "GetChatMember")
		// fmt.Printf("member: %v\n", member)
		break
	case "startLogin":
		// cmdLine := flag.NewFlagSet("", flag.PanicOnError)
		// phone := cmdLine.String("phone", "", "")
		// cmdLine.Parse(params)

		// fmt.Printf("phone: %v", *phone)

		if len(params) != 1 {
			return errors.New("Params for command startLogin are incorrect")
		}

		return tool.TG.StartLogin(params[0])
	case "completeLoginWithCode":
		if len(params) != 1 {
			return errors.New("Params for command completeLoginWithCode are incorrect")
		}

		return tool.TG.CompleteLoginWithCode(params[0])
	case "panic":
		panic("panic test")
	default:
		fmt.Fprint(os.Stderr, fmt.Errorf("Unknown cmd: %v", name))

		resp := "Unknown cmd"
		msg := tgbotapi.NewMessage(chatID, resp)
		bot.Send(msg)
	}

	return nil
}

func processMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update, chatID int64, txt string) error {
	cmdSubmatches := cmdRegexp.FindStringSubmatch(txt)

	if len(cmdSubmatches) == 0 {
		return nil
	}

	cmdName := cmdSubmatches[1]
	cmdBot := cmdSubmatches[2]
	cmdParams := strings.Fields(cmdSubmatches[3])
	if cmdBot != bot.Self.UserName {
		return nil
	}

	fmt.Printf("Got cmd for me: %v\nGot params: %q\n", cmdName, cmdParams)
	return processCmd(bot, update, chatID, cmdName, cmdParams)
}

func processUpdate(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	defer common.RepairIfError("processing update", update)

	var msg *tgbotapi.Message
	if update.Message != nil {
		msg = update.Message
	} else if update.ChannelPost != nil {
		msg = update.ChannelPost
	}

	if len(msg.Text) > 0 {
		err := processMessage(bot, update, msg.Chat.ID, msg.Text)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Expected error while processing message: %v\n", err)
		}
	}

	// if msg.NewChatMembers != nil {

	// }

	// if msg.LeftChatMember != nil {

	// }
}

func main() {
	rand.NewSource(time.Now().UnixNano())

	conf, err := config.LoadConfig("config.json")
	common.PanicIfError(err, "reading/decoding config file")

	//region mongo
	session, err := mgo.Dial(fmt.Sprintf("%v:%v", conf.Mongo.Host, conf.Mongo.Port))
	common.PanicIfError(err, "opening session for mongo")
	session.SetMode(mgo.Monotonic, true)
	defer session.Close()
	fmt.Println("Session for mongo opened")
	//endregion

	//region user api
	uac := conf.Telegram.UserApi
	tool = &userapi.Tool{}
	err = tool.Run(uac.Host, uac.Port, uac.PublicKey, uac.ApiId, uac.ApiHash, 500)
	common.PanicIfError(err, "connecting to Telegram User API")
	//fmt.Println(tool)
	fmt.Println("Connected to Telegram User API.")
	//endregion

	//region bot
	bac := conf.Telegram.BotApi
	competeKey := fmt.Sprintf("%v:%v", bac.ID, bac.Key)
	bot, err := tgbotapi.NewBotAPI(competeKey)
	common.PanicIfError(err, "creating bot instance")
	bot.Debug = true
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)
	fmt.Printf("Authorized on bot %s", bot.Self.UserName)
	//endregion

	for {
		select {
		case err = <-tool.ErrCh:
			common.PanicIfError(err, "working with Telegram User API")
		case update := <-updates:
			processUpdate(bot, update)
		}
	}
}
