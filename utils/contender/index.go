package contender

import (
	"bytes"
	"fmt"

	"github.com/PROger4ever/draw-telegram-bot/bot"
	"github.com/PROger4ever/draw-telegram-bot/error"
	"github.com/PROger4ever/draw-telegram-bot/mongo/models/user"
)

const userCompleteFormat = `UserID: %d
Username: %s
Имя: %s %s

Статус:     %s
Создание:   %v
Обновление: %v
Удаление:   %v (если было)`

func FormatUserComplete(user *user.User) string {
	//userLink := "tg://user?id=" + strconv.Itoa(user.TelegramID)
	//return fmt.Sprintf("[%v %v (id%d %s)](%s)", user.FirstName, user.LastName, user.ID, user.Username, userLink)
	return fmt.Sprintf(userCompleteFormat, user.TelegramID,
		user.Username, user.FirstName, user.LastName,
		user.Status, user.CreatedAt, user.UpdatedAt, user.DeletedAt)
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

func RefreshUser(bot *botpkg.Bot, u *user.User) (isVerified bool, err *eepkg.ExtendedError) {
	// Обновляем данные пользователя
	chatMember, err := bot.GetChatMember(u.TelegramID)
	if err != nil {
		return
	}

	// Нет Username? Не подписан на канал? - удаляем в DB, continue
	if chatMember.User.UserName == "" || chatMember.Status != "member" {
		err = u.RemoveId()
		return
	}
	isVerified = true

	//Записать изменения в DB
	u.Username = chatMember.User.UserName
	u.FirstName = chatMember.User.FirstName
	u.LastName = chatMember.User.LastName
	u.Status = chatMember.Status
	_, err = u.UpsertId()
	return
}
