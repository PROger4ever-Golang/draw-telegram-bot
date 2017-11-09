package utils

import (
	"bytes"
	"fmt"
	"strconv"

	"bitbucket.org/proger4ever/draw-telegram-bot/mongo/models/user"
)

func FormatUserMarkdown(user *user.User) string {
	userLink := "tg://user?id=" + strconv.Itoa(user.TelegramID)
	return fmt.Sprintf("[%v %v (id%d %s)](%s)", user.FirstName, user.LastName, user.ID, user.Username, userLink)
}
func FormatUserDog(user *user.User) string {
	return fmt.Sprintf("@%s", user.Username)
}

func FormatUsers(users []*user.User, formatter func(user *user.User) string) string {
	buf := bytes.Buffer{}
	i := 0
	for ; i < len(users)-1; i++ {
		buf.WriteString(fmt.Sprintf("%d. ", i+1))
		buf.WriteString(formatter(users[i]))
		buf.WriteString("\n")
	}
	if len(users) > 0 {
		buf.WriteString(fmt.Sprintf("%d. ", i+1))
		buf.WriteString(formatter(users[i]))
	}
	return buf.String()
}
