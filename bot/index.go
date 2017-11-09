package botpkg

import (
	"bitbucket.org/proger4ever/draw-telegram-bot/config"
	"bitbucket.org/proger4ever/draw-telegram-bot/error"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"time"
)

const minRequestPeriod = 334 * time.Millisecond //should be changed after experiments later

const cantConnectToBotApi = "Can't connect to Bot API"
const cantGetUpdatesChan = "Can't get channel to get updates from Bot API"

type Bot struct {
	Conf *config.BotApiConfig

	BotApi  *tgbotapi.BotAPI
	Updates tgbotapi.UpdatesChannel

	LastSendTime time.Time
}

func (b *Bot) Init(bac *config.BotApiConfig) (err error) {
	b.Conf = bac
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
	msg.DisableNotification = b.Conf.DisableNotification
	if enableParsing {
		msg.ParseMode = "Markdown"
	}
	_, err := b.Send(msg)
	return err
}
func (b *Bot) SendError(chatID int64, err error) error {
	resp := fmt.Sprintf("%s", err)
	return b.SendMessage(chatID, resp, false)
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

func (b *Bot) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	b.waitRequestTime()
	return b.BotApi.Send(c)
}

func (b *Bot) GetChatMember(config tgbotapi.ChatConfigWithUser) (tgbotapi.ChatMember, error) {
	return b.BotApi.GetChatMember(config)
}

func (b *Bot) waitRequestTime() {
	now := time.Now()
	minBound := b.LastSendTime.Add(minRequestPeriod)
	if now.Before(minBound) {
		sleepTime := minBound.Sub(now)
		<-time.After(sleepTime)
	}
	b.LastSendTime = time.Now() // it could take longer due to high load
}
