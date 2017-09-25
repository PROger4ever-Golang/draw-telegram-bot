package main

import (
    "errors"
    "fmt"
    "math/rand"
    "os"
    "regexp"
    "strings"
    "time"

    "gopkg.in/mgo.v2"
    "github.com/go-telegram-bot-api/telegram-bot-api"
    tuapi "github.com/andreyvit/telegramapi"
    "github.com/andreyvit/telegramapi/mtproto"

    "./common"
    "./config"
    "./telegram/userapi"
    "gopkg.in/mgo.v2/bson"
    _"strconv"
    "strconv"
)

type StateSetting struct {
    ID    bson.ObjectId `bson:"_id,omitempty"`
    Name  string
    Value *tuapi.State
}

//type State struct {
//    PreferredDC int
//
//    DCs map[int]*DCState
//
//    LoginState    LoginState
//    PhoneNumber   string
//    PhoneCodeHash string
//
//    UserID    int
//    FirstName string
//    LastName  string
//    Username  string
//}

//func (set *StateSetting) Decode(m *bson.M) {
//    x := *m
//
//    set.ID = x["_id"].(bson.ObjectId)
//    set.Name = x["name"].(string)
//
//    val := x["value"].(bson.M)
//    set.Value = &tuapi.State{
//        PreferredDC: val["prefereddc"].(int),
//
//        LoginState:    val["loginstate"].(tuapi.LoginState),
//        PhoneNumber:   val["phonenumber"].(string),
//        PhoneCodeHash: val["phonecodehash"].(string),
//
//        UserID:    val["userid"].(int),
//        FirstName: val["firstname"].(string),
//        LastName:  val["lastname"].(string),
//        Username:  val["username"].(string),
//    }
//
//    dcs := val["dcs"].(bson.M)
//    for km, vm := range dcs {
//        kc, err := strconv.Atoi(km)
//        common.PanicIfError(err, "convert km to kc")
//
//        dcm := vm.(bson.M)
//        dcc := tuapi.DCState{}
//
//        dcc.ID = dcm["id"].(int)
//        primaryAddrM := dcm["primaryaddr"].(bson.M)
//        dcc.PrimaryAddr = tuapi.Addr{
//            IP: primaryAddrM["ip"].(string),
//            Port: primaryAddrM["port"].(int),
//        }
//        authM := dcm["auth"].(bson.M)
//
//        keyB := authM["key"].(bson.Binary)
//        serverSaltB := authM["serversalt"].(bson.Binary)
//        serverSaltS := serverSaltB.Data
//        serverSalt := [8]byte{}
//        copy(serverSalt[:], serverSaltS)
//        dcc.Auth = mtproto.AuthResult {
//            Key: keyB.Data,
//            KeyID: authM["keyid"].(uint64),
//            ServerSalt: serverSalt,
//            KeyID: authM["timeoffset"].(uint64),
//        }
//
//        set.Value.DCs[kc] =
//    }
//}

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
            resp = fmt.Sprintf("Кажется, мы успешно авторизовались: id %d name <%s %s>\n", user.ID, user.FirstName, user.LastName)
        } else {
            resp = fmt.Sprintf("error: %v", err)
        }
        msg := tgbotapi.NewMessage(chatID, resp)
        _, err = bot.Send(msg)

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

    state := tuapi.State{}
    stateSettingM := bson.M{}
    stateSetting := StateSetting{}
    err = mongoSession.DB("mazimotaBot").C("settings").Find(bson.M{
        "name": "state",
    }).One(&stateSettingM)
    if err == nil {
        fmt.Printf("stateSetting: %q", stateSettingM)
        stateSetting.Decode(&stateSettingM)
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
    err = tool.Run(&state, uac.Host, uac.Port, uac.PublicKey, uac.ApiId, uac.ApiHash, 0)
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
        case _ = <-tool.StateCh:
            //_, err := mongoSession.DB("mazimotaBot").C("settings").Upsert(bson.M{
            //    "name": "state",
            //}, StateSetting{
            //    Name:  "state",
            //    Value: state,
            //})
            //common.PanicIfError(err, "saving bot state")
            //fmt.Printf("save state: %q\n", state)
        }
    }
}
