package botpkg

import (
	"fmt"
	"time"

	"bitbucket.org/proger4ever/draw-telegram-bot/config"
	"bitbucket.org/proger4ever/draw-telegram-bot/error"
	"bitbucket.org/proger4ever/draw-telegram-bot/userApi"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

const minRequestPeriod = 334 * time.Millisecond //should be changed after experiments later

const cantConnectToBotApi = "Can't connect to Bot API"
const cantGetUpdatesChan = "Can't get channel to get updates from Bot API"

type Bot struct {
	Conf *config.Config

	Tool    *userapi.Tool
	BotApi  *tgbotapi.BotAPI
	Updates tgbotapi.UpdatesChannel

	LastSendTime time.Time
}

func (b *Bot) Init(conf *config.Config) (err error) {
	b.Conf = conf

	return b.initBotApi(&conf.BotApi)
}

func (b *Bot) initBotApi(bac *config.BotApiConfig) (err error) {
	competeKey := fmt.Sprintf("%v:%v", bac.ID, bac.Key)
	b.BotApi, err = tgbotapi.NewBotAPI(competeKey)
	if err != nil {
		return eepkg.Wrap(err, false, true, cantConnectToBotApi)
	}

	b.BotApi.Debug = bac.Debug
	u := tgbotapi.UpdateConfig{
		Timeout: 60,
	}
	b.Updates, err = b.BotApi.GetUpdatesChan(u)
	if err != nil {
		return eepkg.Wrap(err, false, true, cantGetUpdatesChan)
	}
	return
}

func (b *Bot) SendMessage(chatID int64, resp string, enableParsing bool) error {
	msg := tgbotapi.NewMessage(chatID, resp)
	msg.DisableNotification = b.Conf.BotApi.DisableNotification

	//btn11 := tgbotapi.NewKeyboardButton("Регистрация")
	//btn12 := tgbotapi.NewKeyboardButton("Помощь")
	//row1 := tgbotapi.NewKeyboardButtonRow(btn11, btn12)
	//msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(row1)

	if enableParsing {
		msg.ParseMode = "Markdown"
	}
	_, err := b.send(msg)
	return err
}
func (b *Bot) SendError(chatID int64, err error) error {
	resp := fmt.Sprintf("%s", err)
	return b.SendMessage(chatID, resp, false)
}
func (b *Bot) GetChatMember(config tgbotapi.ChatConfigWithUser) (tgbotapi.ChatMember, error) {
	return b.BotApi.GetChatMember(config)
}

// NOTE: User Api disabled
// func SendError(tool *userapi.Tool, bot *botpkg.Bot, chatID int64, err error) error {
// 	resp := fmt.Sprintf("```\nОшибка: %v\n```", err)
// 	_, newErr := tool.MessagesSendMessageSelf(resp)
// 	if newErr != nil {
// 		newErr = SendError(bot, chatID, err)
// 	}
// 	return newErr
// }

func (b *Bot) waitRequestTime() {
	now := time.Now()
	minBound := b.LastSendTime.Add(minRequestPeriod)
	if now.Before(minBound) {
		sleepTime := minBound.Sub(now)
		<-time.After(sleepTime)
	}
	b.LastSendTime = time.Now() // it could take longer due to high load
}
func (b *Bot) send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	b.waitRequestTime()
	return b.BotApi.Send(c)
}
