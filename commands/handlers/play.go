package handlers

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/PROger4ever/telegramapi/mtproto"
	"github.com/go-telegram-bot-api/telegram-bot-api"

	"bitbucket.org/proger4ever/drawtelegrambot/commands/utils"
	"bitbucket.org/proger4ever/drawtelegrambot/config"
	"bitbucket.org/proger4ever/drawtelegrambot/telegram/userapi"
)

type PlayHandler struct {
	Bot  *tgbotapi.BotAPI
	Conf *config.Config
	Tool *userapi.Tool
}

func (c *PlayHandler) GetName() string {
	return "play"
}

func (c *PlayHandler) GetParamsCount() int {
	return 0
}

func (c *PlayHandler) Init(conf *config.Config, tool *userapi.Tool, bot *tgbotapi.BotAPI) {
	c.Bot = bot
	c.Conf = conf
	c.Tool = tool
}

func (c *PlayHandler) Execute(chat *tgbotapi.Chat, params []string) error {
	if chat.UserName != c.Conf.Management.ChannelUsername {
		return utils.SendBotError(c.Bot, int64(chat.ID), errors.New("Розыгрыши недоступны в этом чате"))
	}

	r, err := c.Tool.ContactsResolveUsername(chat.UserName)
	if err != nil {
		utils.SendBotError(c.Bot, int64(chat.ID), err)
		return err
	}
	channelInfo := r.Chats[0].(*mtproto.TLChannel)

	userTypesAdmins, err := c.Tool.ChannelsGetParticipants(channelInfo.ID, channelInfo.AccessHash, &mtproto.TLChannelParticipantsAdmins{},
		0, math.MaxInt32)
	if err != nil {
		utils.SendBotError(c.Bot, int64(chat.ID), err)
		return err
	}
	admins := userapi.UserTypesToUsers(&userTypesAdmins.Users)
	adminsMap := map[int]*mtproto.TLUser{}
	for _, admin := range *admins {
		adminsMap[admin.ID] = admin
	}

	userTypesAll, err := c.Tool.ChannelsGetParticipants(channelInfo.ID, channelInfo.AccessHash, &mtproto.TLChannelParticipantsRecent{},
		0, math.MaxInt32)
	if err != nil {
		utils.SendBotError(c.Bot, int64(chat.ID), err)
		return err
	}
	usersAll := userapi.UserTypesToUsers(&userTypesAll.Users)

	usersOnly := []*mtproto.TLUser{}
	for _, user := range *usersAll {
		_, isAdmin := adminsMap[user.ID]
		if !isAdmin && !user.Bot() {
			usersOnly = append(usersOnly, user)
		}
	}

	usersOnlyLen := len(usersOnly)
	if usersOnlyLen == 0 {
		resp := "```\nНет участников для розыгрыша.\nРозыгрыш только среди админов - не смешите мои байтики.\n```"
		err = utils.SendBotMessage(c.Bot, int64(chat.ID), resp)
		return err
	}

	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("```\nВ розыгрыше учавствуют %v пользователей.", len(usersOnly)))
	// for _, user := range usersOnly {
	// 	userString := ""
	// 	if user.Username != "" {
	// 		userLink := fmt.Sprintf("https://t.me/%s", user.Username)
	// 		userString = fmt.Sprintf("\n[%v %v (%s, id%d)](%s)", user.FirstName, user.LastName, user.Username, user.ID, userLink)
	// 	} else {
	// 		userString = fmt.Sprintf("\n%v %v (id%d)", user.FirstName, user.LastName, user.ID)
	// 	}
	// 	buffer.WriteString(userString)
	// }
	buffer.WriteString("\nУдачи!\n```")
	err = utils.SendBotMessage(c.Bot, int64(chat.ID), buffer.String())
	if err != nil {
		return err
	}

	<-time.After(5 * time.Second)

	user := usersOnly[rand.Intn(usersOnlyLen)]
	buffer = bytes.Buffer{}
	buffer.WriteString("Итак, выигрывает...\n")
	userLink := "tg://user?id=" + string(user.ID)
	userString := fmt.Sprintf("[%v %v (id%d %s)](%s)\n", user.FirstName, user.LastName, user.ID, user.Username, userLink)
	buffer.WriteString(userString)
	buffer.WriteString("Спасибо всем за участие!\n")

	return utils.SendBotMessage(c.Bot, int64(chat.ID), buffer.String())
}
