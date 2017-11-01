package utils

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/PROger4ever/telegramapi/mtproto"
	"github.com/go-telegram-bot-api/telegram-bot-api"

	"bitbucket.org/proger4ever/draw-telegram-bot/mongo/models/user"
)

func SendBotMessage(bot *tgbotapi.BotAPI, chatID int64, resp string, enableParsing bool) error {
	msg := tgbotapi.NewMessage(chatID, resp)
	if enableParsing {
		msg.ParseMode = "Markdown"
	}
	_, err := bot.Send(msg)
	return err
}
func SendBotError(bot *tgbotapi.BotAPI, chatID int64, err error) error {
	resp := fmt.Sprintf("%s", err)
	return SendBotMessage(bot, chatID, resp, false)
}

// NOTE: User Api disabled
// func SendError(tool *userapi.Tool, bot *tgbotapi.BotAPI, chatID int64, err error) error {
// 	resp := fmt.Sprintf("```\nОшибка: %v\n```", err)
// 	_, newErr := tool.MessagesSendMessageSelf(resp)
// 	if newErr != nil {
// 		newErr = SendBotError(bot, chatID, err)
// 	}
// 	return newErr
// }

func FormatUserMarkdown(user *user.User) string {
	userLink := "tg://user?id=" + strconv.Itoa(user.TelegramID)
	return fmt.Sprintf("[%v %v (id%d %s)](%s)", user.FirstName, user.LastName, user.ID, user.Username, userLink)
}
func FormatUserDog(user *user.User) string {
	return fmt.Sprintf("@%s", user.Username)
}

func FormatUsers(users *[]*mtproto.TLUser, formatter func(user *mtproto.TLUser) string, buffer *bytes.Buffer) {
	for i, user := range *users {
		buffer.WriteString(fmt.Sprintf("%d. ", i+1))
		buffer.WriteString(formatter(user))
		buffer.WriteString("\n")
	}
}
