package contender

import (
	"bytes"
	"fmt"

	"bitbucket.org/proger4ever/draw-telegram-bot/bot"
	"bitbucket.org/proger4ever/draw-telegram-bot/error"
	"bitbucket.org/proger4ever/draw-telegram-bot/mongo/models/user"
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

func RefreshUser(bot *botpkg.Bot, telegramID int) (c *user.User, isVerified bool, err *eepkg.ExtendedError) {
	// Обновляем данные пользователя
	chatMember, err := bot.GetChatMember(telegramID)
	if err != nil {
		return
	}

	uc := user.NewCollectionDefault()
	c = &user.User{
		TelegramID: chatMember.User.ID,
		Username:   chatMember.User.UserName,
		FirstName:  chatMember.User.FirstName,
		LastName:   chatMember.User.LastName,
		Status:     chatMember.Status,
	}
	c.Init(uc)
	// Нет Username? Не подписан на канал? - удаляем в DB, continue
	if chatMember.User.UserName == "" || chatMember.Status != "member" {
		err = uc.RemoveByTelegramID(telegramID)
		return
	}
	isVerified = true

	//Записать изменения в DB
	_, err = c.UpdateOneOrInsertTelegramId()
	return
}
