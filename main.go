package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
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

func Abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	if x == 0 {
		return 0
	}
	return x
}

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

var cmdRegexp = regexp.MustCompile("^/([A-Za-z0-9_-]+)(@([A-Za-z0-9_-]+))? ?(.*)")

var conf config.Config
var bot *tgbotapi.BotAPI
var tool *userapi.Tool

func sendBotMessage(chatID int64, resp string) error {
	msg := tgbotapi.NewMessage(chatID, resp)
	msg.ParseMode = "Markdown"
	_, err := bot.Send(msg)
	return err
}
func sendBotError(chatID int64, err error) error {
	resp := fmt.Sprintf("```\nerror: %v\n```", err)
	return sendBotMessage(chatID, resp)
}

func processCmd(update tgbotapi.Update, chat *tgbotapi.Chat, name string, params []string) error {
	switch name {
	case "draw":
		r, err := tool.ContactsResolveUsername(chat.UserName)
		if err != nil {
			sendBotError(int64(chat.ID), err)
			return err
		}

		channelInfo := r.Chats[0].(*mtproto.TLChannel)

		userTypesAdmins, err := tool.ChannelsGetParticipants(channelInfo.ID, channelInfo.AccessHash, &mtproto.TLChannelParticipantsAdmins{},
			0, math.MaxInt32)
		if err != nil {
			sendBotError(int64(chat.ID), err)
			return err
		}
		admins := userapi.UserTypesToUsers(&userTypesAdmins.Users)
		adminsMap := map[int]*mtproto.TLUser{}
		for _, admin := range *admins {
			adminsMap[admin.ID] = admin
		}

		//fmt.Printf("admins: %q\n", *admins)

		userTypesAll, err := tool.ChannelsGetParticipants(channelInfo.ID, channelInfo.AccessHash, &mtproto.TLChannelParticipantsRecent{},
			0, math.MaxInt32)
		if err != nil {
			sendBotError(int64(chat.ID), err)
			return err
		}
		usersAll := userapi.UserTypesToUsers(&userTypesAll.Users)

		//fmt.Printf("usersAll: %q\n", *usersAll)

		usersOnly := []*mtproto.TLUser{}
		for _, user := range *usersAll {
			_, isAdmin := adminsMap[user.ID]
			if !isAdmin && !user.Bot() {
				usersOnly = append(usersOnly, user)
			}
		}

		//fmt.Printf("usersOnly: %q\n", usersOnly)

		usersOnlyLen := len(usersOnly)
		if usersOnlyLen == 0 {
			resp := "```\nНет участников для розыгрыша.\nРозыгрыш только среди админов - не смешите мои байтики.\n```"
			err = sendBotMessage(int64(chat.ID), resp)
			return err
		}

		var buffer bytes.Buffer
		buffer.WriteString(fmt.Sprintf("```\nВ розыгрыше учавствуют %v пользователей.", len(usersOnly)))
		// for _, user := range usersOnly {
		// 	userString := ""
		// 	if user.Username != "" {
		// 		userLink := fmt.Sprintf("https://t.me/%s", user.Username)
		// 		userString = fmt.Sprintf("\n[%v %v (%s, id%d)](%s)", user.FirstName, user.LastName, user.Username, user.ID, userLink)
		// 	} else {
		// 		userString = fmt.Sprintf("\n%v %v (id%d)", user.FirstName, user.LastName, user.ID)
		// 	}
		// 	buffer.WriteString(userString)
		// }
		buffer.WriteString("\nУдачи!\n```")
		err = sendBotMessage(int64(chat.ID), buffer.String())
		if err != nil {
			return err
		}

		<-time.After(5 * time.Second)

		user := usersOnly[rand.Intn(usersOnlyLen)]
		buffer = bytes.Buffer{}
		buffer.WriteString("Итак, выигрывает...\n")
		userLink := "tg://user?id=" + string(user.ID)
		userString := fmt.Sprintf("[%v %v, %s id%d](%s)\n", user.FirstName, user.LastName, user.Username, user.ID, userLink)
		buffer.WriteString(userString)
		buffer.WriteString("Спасибо всем за участие!\n")
		return sendBotMessage(int64(chat.ID), buffer.String())
	case "startLogin":
		// cmdLine := flag.NewFlagSet("", flag.PanicOnError)
		// phone := cmdLine.String("phone", "", "")
		// cmdLine.Parse(params)

		// fmt.Printf("phone: %v", *phone)

		if len(params) != 1 {
			return errors.New("Params for command startLogin are incorrect")
		}

		err := tool.StartLogin(params[0])

		if err == nil {
			resp := fmt.Sprintf("```\n/completeLoginWithCode@%v -\n```", bot.Self.UserName)
			err = sendBotMessage(int64(chat.ID), resp)
		} else {
			err = sendBotError(int64(chat.ID), err)
		}
		return err
	case "completeLoginWithCode":
		if len(params) != 1 {
			return errors.New("Params for command completeLoginWithCode are incorrect")
		}

		phoneCode := strings.Replace(params[0], "-", "", -1)
		user, err := tool.CompleteLoginWithCode(phoneCode)

		if err == nil {
			resp := fmt.Sprintf("```\nМы успешно авторизовались.\nUserID: %d\nUsername: %s\nName: %s %s\n```", user.ID, user.Username, user.FirstName, user.LastName)
			err = sendBotMessage(int64(chat.ID), resp)
		} else {
			err = sendBotError(int64(chat.ID), err)
		}
		return err
	case "getDialogs":
		r, err := tool.MessagesGetDialogs()
		if err == nil {
			bs, err := json.Marshal(r)
			resp := fmt.Sprintf("```\nMessagesGetDialogs result: %s, %q\n```", string(bs), err)
			err = sendBotMessage(int64(chat.ID), resp)
		} else {
			err = sendBotError(int64(chat.ID), err)
		}
		return err
	case "msgTest":
		resp := fmt.Sprintf("```\n/completeLoginWithCode@%v -\n```", bot.Self.UserName)
		return sendBotMessage(int64(chat.ID), resp)
	case "panic":
		panic("panic test")
	default:
		err := fmt.Errorf("Unknown cmd: %v", name)
		return sendBotError(int64(chat.ID), err)
	}
}

func processMessage(update tgbotapi.Update, chat *tgbotapi.Chat, txt string) error {
	cmdSubmatches := cmdRegexp.FindStringSubmatch(txt)
	if len(cmdSubmatches) == 0 {
		return nil
	}

	cmdName := cmdSubmatches[1]
	cmdBot := cmdSubmatches[3]
	cmdParams := strings.Fields(cmdSubmatches[4])
	if cmdBot != "" && cmdBot != bot.Self.UserName {
		return nil
	}

	fmt.Printf("Got cmd for me: %v\nGot params: %q\n", cmdName, cmdParams)
	return processCmd(update, chat, cmdName, cmdParams)
}

func processUpdate(update tgbotapi.Update) {
	defer common.RepairIfError("processing update", update)

	var msg *tgbotapi.Message
	if update.Message != nil {
		msg = update.Message
	} else if update.ChannelPost != nil {
		msg = update.ChannelPost
	}

	if len(msg.Text) > 0 {
		err := processMessage(update, msg.Chat, msg.Text)
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

	var err error
	conf, err = config.LoadConfig("config.json")
	common.PanicIfError(err, "reading/decoding config file")

	//region mongo
	mongoSession, err := mgo.Dial(fmt.Sprintf("%v:%v", conf.Mongo.Host, conf.Mongo.Port))
	common.PanicIfError(err, "opening mongoSession")
	mongoSession.SetMode(mgo.Monotonic, true)
	defer mongoSession.Close()
	fmt.Println("MongoSession opened")

	state := &tuapi.State{}
	err = mgo.ErrNotFound
	// settingState := SettingState{}
	// err = mongoSession.DB("mazimotaBot").C("settings").Find(bson.M{
	// 	"name": "state",
	// }).One(&settingState)
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
	tool = &userapi.Tool{}
	err = tool.Run(state, uac.Host, uac.Port, uac.PublicKey, uac.ApiId, uac.ApiHash, 3)
	common.PanicIfError(err, "connecting to Telegram User API")
	//fmt.Println(tool)
	fmt.Println("Connected to Telegram User API.")
	//endregion

	//region bot
	bac := conf.BotApi
	competeKey := fmt.Sprintf("%v:%v", bac.ID, bac.Key)
	bot, err = tgbotapi.NewBotAPI(competeKey)
	common.PanicIfError(err, "creating bot instance")
	bot.Debug = true
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)
	common.PanicIfError(err, "getting updates chan")
	fmt.Printf("Authorized on bot %s\n", bot.Self.UserName)
	//endregion

	for {
		select {
		case update := <-updates:
			processUpdate(update)
			break
		// case _ = <-time.After(1 * time.Second):

		case state := <-tool.StateCh:
			_, err := mongoSession.DB("mazimotaBot").C("settings").Upsert(bson.M{
				"name": "state",
			}, SettingState{
				Name:  "state",
				Value: NewStateSerializable(&state),
			})
			common.PanicIfError(err, "saving bot state")
			fmt.Println("save state")
		}
	}
}
