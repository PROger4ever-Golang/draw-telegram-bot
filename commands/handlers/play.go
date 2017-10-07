package handlers

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/PROger4ever/telegramapi/mtproto"
	"github.com/go-telegram-bot-api/telegram-bot-api"

	"bitbucket.org/proger4ever/drawtelegrambot/commands/utils"
	"bitbucket.org/proger4ever/drawtelegrambot/config"
	"bitbucket.org/proger4ever/drawtelegrambot/userApi"
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
		return utils.SendError(c.Tool, c.Bot, int64(chat.ID), err)
	}
	channelInfo := r.Chats[0].(*mtproto.TLChannel)

	userTypesAdmins, err := c.Tool.ChannelsGetParticipants(channelInfo.ID, channelInfo.AccessHash, &mtproto.TLChannelParticipantsAdmins{},
		0, math.MaxInt32)
	if err != nil {
		return utils.SendError(c.Tool, c.Bot, int64(chat.ID), err)
	}
	admins := userapi.UserTypesToUsers(&userTypesAdmins.Users)
	adminsMap := map[int]*mtproto.TLUser{}
	for _, admin := range *admins {
		adminsMap[admin.ID] = admin
	}

	userTypesAll, err := c.Tool.ChannelsGetParticipants(channelInfo.ID, channelInfo.AccessHash, &mtproto.TLChannelParticipantsRecent{},
		0, math.MaxInt32)
	if err != nil {
		return utils.SendError(c.Tool, c.Bot, int64(chat.ID), err)
	}
	usersAll := userapi.UserTypesToUsers(&userTypesAll.Users)

	uAdmins := []*mtproto.TLUser{}
	uBots := []*mtproto.TLUser{}
	uRuleBreakers := []*mtproto.TLUser{}
	uParticipants := []*mtproto.TLUser{}
	for _, user := range *usersAll {
		_, isAdmin := adminsMap[user.ID]
		isBot := user.Bot()
		isRuleBreaker := (user.Username == "")
		if isAdmin {
			uAdmins = append(uAdmins, user)
		}
		if isBot {
			uBots = append(uBots, user)
		}
		if isRuleBreaker {
			uRuleBreakers = append(uRuleBreakers, user)
		}
		if !isAdmin && !isBot && !isRuleBreaker {
			uParticipants = append(uParticipants, user)
		}
	}

	bufferSelf := bytes.Buffer{}
	bufferSelf.WriteString(fmt.Sprintf("Перед розыгрышем.\n Пользователей на канале: __%d__\n **Админы:**\n", len(*usersAll)))
	utils.FormatUsers(&uAdmins, utils.FormatUserMarkdown, &bufferSelf)
	bufferSelf.WriteString("**Боты:**\n")
	utils.FormatUsers(&uBots, utils.FormatUserMarkdown, &bufferSelf)
	bufferSelf.WriteString("**Нарушители правил:**\n")
	utils.FormatUsers(&uRuleBreakers, utils.FormatUserMarkdown, &bufferSelf)
	bufferSelf.WriteString(fmt.Sprintf("\n\nВ розыгрыше учавствуют %v пользователей:\n", len(uParticipants)))
	utils.FormatUsers(&uParticipants, utils.FormatUserMarkdown, &bufferSelf)
	_, err = c.Tool.MessagesSendMessageSelf(bufferSelf.String())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error occured while c.Tool.MessagesSendMessageSelf()")
	}

	uParticipantsLen := len(uParticipants)
	if uParticipantsLen == 0 {
		resp := "```\nНет участников для розыгрыша.\nРозыгрыш только среди админов - не смешите мои байтики.\n```"
		err = utils.SendBotMessage(c.Bot, int64(chat.ID), resp)
		return err
	}

	err = utils.SendBotMessage(c.Bot, int64(chat.ID), "Начинаем розыгрыш.")
	if err != nil {
		return err
	}

	<-time.After(5 * time.Second)

	user := uParticipants[rand.Intn(uParticipantsLen)]
	bufferBot := bytes.Buffer{}
	bufferBot.WriteString("Итак, выигрывает...\n")
	bufferBot.WriteString(utils.FormatUserDog(user))
	bufferBot.WriteString("\nСпасибо всем за участие!")
	return utils.SendBotMessage(c.Bot, int64(chat.ID), bufferBot.String())
}
