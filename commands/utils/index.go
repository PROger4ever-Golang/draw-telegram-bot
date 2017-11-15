package utils

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	btn11        = tgbotapi.NewKeyboardButton("РЕГИСТРАЦИЯ")
	btn12        = tgbotapi.NewKeyboardButton("ПОМОЩЬ")
	row1         = tgbotapi.NewKeyboardButtonRow(btn11, btn12)
	UserKeyboard = tgbotapi.NewReplyKeyboard(row1)
)
