package main

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"time"

	tuapi "github.com/PROger4ever/telegramapi"
	"github.com/PROger4ever/telegramapi/mtproto"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"bitbucket.org/proger4ever/drawtelegrambot/common"
	"bitbucket.org/proger4ever/drawtelegrambot/config"
	"bitbucket.org/proger4ever/drawtelegrambot/telegram/userapi"
)

type SettingState struct {
	ID    bson.ObjectId `bson:"_id,omitempty"`
	Name  string
	Value *StateSerializable
}

func (ss *SettingState) DecodeValue() *tuapi.State {
	ser := ss.Value
	st := &tuapi.State{
		PreferredDC: ser.PreferredDC,

		LoginState:    ser.LoginState,
		PhoneNumber:   ser.PhoneNumber,
		PhoneCodeHash: ser.PhoneCodeHash,

		UserID:    ser.UserID,
		FirstName: ser.FirstName,
		LastName:  ser.LastName,
		Username:  ser.Username,
	}

	st.DCs = make(map[int]*tuapi.DCState, len(st.DCs))
	for _, dc := range ser.DCs {
		st.DCs[dc.ID] = &tuapi.DCState{
			ID:          dc.ID,
			PrimaryAddr: dc.PrimaryAddr,
			FramerState: dc.FramerState,
			Auth: mtproto.AuthResult{
				Key:        dc.Auth.Key,
				KeyID:      (uint64(dc.Auth.KeyIDHigh) << 32) + uint64(dc.Auth.KeyIDLow),
				ServerSalt: dc.Auth.ServerSalt,
				TimeOffset: dc.Auth.TimeOffset,
				SessionID:  dc.Auth.SessionID,
			},
		}
	}
	return st
}

type StateSerializable struct {
	PreferredDC int

	DCs []*DCState

	LoginState    tuapi.LoginState
	PhoneNumber   string
	PhoneCodeHash string

	UserID    int
	FirstName string
	LastName  string
	Username  string
}

type DCState struct {
	ID int

	PrimaryAddr tuapi.Addr

	Auth        AuthResult
	FramerState mtproto.FramerState
}

type AuthResult struct {
	Key        []byte
	KeyIDHigh  uint32
	KeyIDLow   uint32
	ServerSalt [8]byte
	TimeOffset int
	SessionID  [8]byte
}

func NewStateSerializable(st *tuapi.State) *StateSerializable {
	ser := &StateSerializable{
		PreferredDC: st.PreferredDC,

		LoginState:    st.LoginState,
		PhoneNumber:   st.PhoneNumber,
		PhoneCodeHash: st.PhoneCodeHash,

		UserID:    st.UserID,
		FirstName: st.FirstName,
		LastName:  st.LastName,
		Username:  st.Username,
	}

	ser.DCs = make([]*DCState, len(st.DCs))
	i := 0
	for _, dc := range st.DCs {
		ser.DCs[i] = &DCState{
			ID:          dc.ID,
			PrimaryAddr: dc.PrimaryAddr,
			FramerState: dc.FramerState,
			Auth: AuthResult{
				Key:        dc.Auth.Key,
				KeyIDHigh:  uint32(dc.Auth.KeyID >> 32),
				KeyIDLow:   uint32(dc.Auth.KeyID),
				ServerSalt: dc.Auth.ServerSalt,
				TimeOffset: dc.Auth.TimeOffset,
				SessionID:  dc.Auth.SessionID,
			},
		}
		i++
	}
	return ser
}

var cmdRegexp = regexp.MustCompile("^/([A-Za-z0-9_-]+)@([A-Za-z0-9_-]+) ?(.*)")

var tool *userapi.Tool

func processCmd(bot *tgbotapi.BotAPI, update tgbotapi.Update, chatID int64, name string, params []string) error {
	switch name {
	case "draw":
		resp := "Starting a new drawing..."
		msg := tgbotapi.NewMessage(chatID, resp)
		bot.Send(msg)

		r, err := tool.MessagesGetFullChat(int(chatID))
		resp = ""
		if err == nil {
			resp = fmt.Sprintf("MessagesGetFullChat result: %q", r)
		} else {
			resp = fmt.Sprintf("error: %v", err)
		}
		msg = tgbotapi.NewMessage(chatID, resp)
		_, err = bot.Send(msg)

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
		return err
		break
	case "startLogin":
		// cmdLine := flag.NewFlagSet("", flag.PanicOnError)
		// phone := cmdLine.String("phone", "", "")
		// cmdLine.Parse(params)

		// fmt.Printf("phone: %v", *phone)

		if len(params) != 1 {
			return errors.New("Params for command startLogin are incorrect")
		}

		err := tool.StartLogin(params[0])

		resp := ""
		if err == nil {
			resp = fmt.Sprintf("/completeLoginWithCode@%v *", bot.Self.UserName)
		} else {
			resp = fmt.Sprintf("error: %v", err)
		}
		msg := tgbotapi.NewMessage(chatID, resp)
		_, err = bot.Send(msg)

		return err
	case "completeLoginWithCode":
		if len(params) != 1 {
			return errors.New("Params for command completeLoginWithCode are incorrect")
		}

		phoneCode := strings.Replace(params[0], "-", "", -1)
		user, err := tool.CompleteLoginWithCode(phoneCode)

		resp := ""
		if err == nil {
			resp = fmt.Sprintf("Мы успешно авторизовались.\nUserID: %d\nUsername: %s\nName: %s %s", user.ID, user.Username, user.FirstName, user.LastName)
		} else {
			resp = fmt.Sprintf("error: %v", err)
		}
		msg := tgbotapi.NewMessage(chatID, resp)
		_, err = bot.Send(msg)

		// tool.Conn.Shutdown()

		return err
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
	mongoSession, err := mgo.Dial(fmt.Sprintf("%v:%v", conf.Mongo.Host, conf.Mongo.Port))
	common.PanicIfError(err, "opening mongoSession")
	mongoSession.SetMode(mgo.Monotonic, true)
	defer mongoSession.Close()
	fmt.Println("MongoSession opened")

	state := &tuapi.State{}
	settingState := SettingState{}
	err = mongoSession.DB("mazimotaBot").C("settings").Find(bson.M{
		"name": "state",
	}).One(&settingState)
	if err == nil {
		state = settingState.DecodeValue()
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
	uac := conf.Telegram.UserApi
	tool = &userapi.Tool{}
	err = tool.Run(state, uac.Host, uac.Port, uac.PublicKey, uac.ApiId, uac.ApiHash, 3)
	common.PanicIfError(err, "connecting to Telegram User API")
	//fmt.Println(tool)
	fmt.Println("Connected to Telegram User API.")
	//endregion

	//region bot
	bac := conf.Telegram.BotApi
	competeKey := fmt.Sprintf("%v:%v", bac.ID, bac.Key)
	bot, err := tgbotapi.NewBotAPI(competeKey)
	common.PanicIfError(err, "creating bot instance")
	bot.Debug = false
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)
	common.PanicIfError(err, "getting updates chan")
	fmt.Printf("Authorized on bot %s\n", bot.Self.UserName)
	//endregion

	for {
		select {
		case err = <-tool.ErrCh:
			common.PanicIfError(err, "working with Telegram User API")
			break
		case update := <-updates:
			processUpdate(bot, update)
			break
		case state := <-tool.StateCh:
			_, err := mongoSession.DB("mazimotaBot").C("settings").Upsert(bson.M{
				"name": "state",
			}, SettingState{
				Name:  "state",
				Value: NewStateSerializable(&state),
			})
			common.PanicIfError(err, "saving bot state")
			// fmt.Printf("save state: %q\n", state)
		}
	}
}
