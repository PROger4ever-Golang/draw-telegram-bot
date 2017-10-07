package utils

import (
	"bytes"
	"fmt"
	"strconv"

	"bitbucket.org/proger4ever/drawtelegrambot/userApi"

	"github.com/PROger4ever/telegramapi/mtproto"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func SendBotMessage(bot *tgbotapi.BotAPI, chatID int64, resp string) error {
	msg := tgbotapi.NewMessage(chatID, resp)
	msg.ParseMode = "Markdown"
	_, err := bot.Send(msg)
	return err
}
func SendBotError(bot *tgbotapi.BotAPI, chatID int64, err error) error {
	resp := fmt.Sprintf("```\nОшибка: %v\n```", err)
	return SendBotMessage(bot, chatID, resp)
}

func SendError(tool *userapi.Tool, bot *tgbotapi.BotAPI, chatID int64, err error) error {
	resp := fmt.Sprintf("```\nОшибка: %v\n```", err)
	_, newErr := tool.MessagesSendMessageSelf(resp)
	if newErr != nil {
		newErr = SendBotError(bot, chatID, err)
	}
	return newErr
}

func FormatUserMarkdown(user *mtproto.TLUser) string {
	userLink := "tg://user?id=" + strconv.Itoa(user.ID)
	return fmt.Sprintf("[%v %v (id%d %s)](%s)", user.FirstName, user.LastName, user.ID, user.Username, userLink)
}
func FormatUserDog(user *mtproto.TLUser) string {
	return fmt.Sprintf("@%s", user.Username)
}

func FormatUsers(users *[]*mtproto.TLUser, formatter func(user *mtproto.TLUser) string, buffer *bytes.Buffer) {
	for i, user := range *users {
		buffer.WriteString(fmt.Sprintf("%d. ", i+1))
		buffer.WriteString(formatter(user))
		buffer.WriteString("\n")
	}
}
