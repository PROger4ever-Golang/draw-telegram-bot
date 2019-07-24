package botpkg

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/PROger4ever-Golang/draw-telegram-bot/commands/utils"
	"github.com/PROger4ever-Golang/draw-telegram-bot/config"
	"github.com/PROger4ever-Golang/draw-telegram-bot/error"
	"github.com/PROger4ever-Golang/draw-telegram-bot/userApi"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

const minRequestPeriod = 334 * time.Millisecond //should be changed after experiments later

const proxyUrlInvalid = "Proxy URL is invalid"
const cantConnectToBotApi = "Can't connect to Bot API"
const cantGetUpdatesChan = "Can't get channel to get updates from Bot API"

const cantSendBotMessage = "Ошибка при отправке сообщения от имени бота"
const cantQueryChatMember = "Ошибка при получении информации о участнике канала"
const cantQueryChatMembersCount = "Ошибка при получении информации о количестве участников канала"

type Bot struct {
	Conf *config.Config

	Tool    *userapi.Tool
	BotApi  *tgbotapi.BotAPI
	Updates tgbotapi.UpdatesChannel

	LastSendTime time.Time
}

func (b *Bot) Init(conf *config.Config) (err *eepkg.ExtendedError) {
	b.Conf = conf

	return b.initBotApi(&conf.BotApi)
}

func (b *Bot) initBotApi(bac *config.BotApiConfig) *eepkg.ExtendedError {
	var errStd error

	httpClient := &http.Client{
		Timeout: time.Duration(bac.ProxyTimeout) * time.Second,
	}

	//SOCKS5-proxy (alfa-stub)
	//if bac.ProxyUrl != "" {
	//	dialer, errStd := proxy.SOCKS5("tcp", bac.ProxyUrl, nil, proxy.Direct)
	//	if errStd != nil {
	//		return eepkg.Wrap(errStd, false, true, proxyUrlInvalid)
	//	}
	//	httpClient.Transport = &http.Transport{Dial:dialer.Dial}
	//}

	//http(s)-proxy
	if bac.ProxyUrl != "" {
		proxyUrl, errStd := url.Parse(bac.ProxyUrl)
		if errStd != nil {
			return eepkg.Wrap(errStd, false, true, proxyUrlInvalid)
		}
		httpClient.Transport = &http.Transport{Proxy: http.ProxyURL(proxyUrl)}
	}

	completeKey := fmt.Sprintf("%v:%v", bac.ID, bac.Key)
	b.BotApi, errStd = tgbotapi.NewBotAPIWithClient(completeKey, httpClient)
	if errStd != nil {
		return eepkg.Wrap(errStd, false, true, cantConnectToBotApi)
	}

	b.BotApi.Debug = bac.Debug
	u := tgbotapi.UpdateConfig{
		Timeout: 60,
	}
	b.Updates, errStd = b.BotApi.GetUpdatesChan(u)
	return eepkg.Wrap(errStd, false, true, cantGetUpdatesChan)
}

func (b *Bot) SendMessage(chatID int64, resp string) *eepkg.ExtendedError {
	msg := tgbotapi.NewMessage(chatID, resp)
	return b.send(&msg)
}
func (b *Bot) SendMessageMarkdown(chatID int64, resp string) *eepkg.ExtendedError {
	msg := tgbotapi.NewMessage(chatID, resp)
	msg.ParseMode = "Markdown"
	return b.send(&msg)
}
func (b *Bot) SendMessageUserKeyboard(chatID int64, resp string) *eepkg.ExtendedError {
	msg := tgbotapi.NewMessage(chatID, resp)
	msg.ReplyMarkup = utils.UserKeyboard
	return b.send(&msg)
}
func (b *Bot) SendError(chatID int64, err error) *eepkg.ExtendedError {
	resp := fmt.Sprintf("%s", err)
	msg := tgbotapi.NewMessage(chatID, resp)
	return b.send(&msg)
}
func (b *Bot) SendErrorUserKeyboard(chatID int64, err error) *eepkg.ExtendedError {
	resp := fmt.Sprintf("%s", err)
	msg := tgbotapi.NewMessage(chatID, resp)
	msg.ReplyMarkup = utils.UserKeyboard
	return b.send(&msg)
}

func (b *Bot) GetChatMember(telegramID int) (cm tgbotapi.ChatMember, err *eepkg.ExtendedError) {
	cm, errStd := b.BotApi.GetChatMember(tgbotapi.ChatConfigWithUser{
		SuperGroupUsername: "@" + b.Conf.Management.ChatUsername,
		UserID:             telegramID,
	})
	return cm, eepkg.Wrap(errStd, false, true, cantQueryChatMember)
}

func (b *Bot) GetChatMemberCount() (count int, err *eepkg.ExtendedError) {
	count, errStd := b.BotApi.GetChatMembersCount(tgbotapi.ChatConfig{
		SuperGroupUsername: "@" + b.Conf.Management.ChatUsername,
	})
	return count, eepkg.Wrap(errStd, false, true, cantQueryChatMembersCount)
}

// NOTE: User Api disabled
// func SendError(tool *userapi.Tool, bot *botpkg.Bot, chatID int64, err *eepkg.ExtendedError) *eepkg.ExtendedError {
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
func (b *Bot) send(msg *tgbotapi.MessageConfig) (err *eepkg.ExtendedError) {
	msg.DisableNotification = b.Conf.BotApi.DisableNotification

	b.waitRequestTime()
	_, errStd := b.BotApi.Send(msg)
	return eepkg.Wrap(errStd, false, true, cantSendBotMessage)
}
