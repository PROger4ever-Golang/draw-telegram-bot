package utils

import (
	"fmt"

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
