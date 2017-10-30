package handlers

import (
	"bytes"
	"errors"

	"gopkg.in/mgo.v2"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"gopkg.in/mgo.v2/bson"

	"bitbucket.org/proger4ever/draw-telegram-bot/commands/utils"
	"bitbucket.org/proger4ever/draw-telegram-bot/config"
	"bitbucket.org/proger4ever/draw-telegram-bot/mongo/models/user"
	"bitbucket.org/proger4ever/draw-telegram-bot/userApi"
)

var playPipeline = []bson.M{{
	"$match": bson.M{
		"username": bson.M{
			// "$exists": true,
			"$ne": "",
		},
		"status": "member",
	},
}, {
	"$sample": bson.M{
		"size": 1,
	},
}}

type PlayHandler struct {
	Bot  *tgbotapi.BotAPI
	Conf *config.Config
	Tool *userapi.Tool
}

func (c *PlayHandler) GetName() string {
	return "play"
}

func (c *PlayHandler) IsForOwnersOnly() bool {
	return false
}

func (c *PlayHandler) GetParamsCount() int {
	return 0
}

func (c *PlayHandler) Init(conf *config.Config, tool *userapi.Tool, bot *tgbotapi.BotAPI) {
	c.Bot = bot
	c.Conf = conf
	c.Tool = tool
}

func (c *PlayHandler) Execute(msg *tgbotapi.Message, params []string) error {
	if msg.Chat.UserName != c.Conf.Management.ChannelUsername {
		err := errors.New("Розыгрыши недоступны в этом чате")
		return utils.SendBotError(c.Bot, int64(msg.Chat.ID), err)
	}

	uc := user.NewCollectionDefault()
	for {
		u, err := uc.PipeOne(playPipeline)
		if err == mgo.ErrNotFound {
			err := errors.New("Нет участников канала, подписавшихся у бота на розыгрыш.\nРозыгрыш среди НИКОГО - не смешите мои байтики")
			return utils.SendBotError(c.Bot, int64(msg.Chat.ID), err)
		}
		if err != nil {
			return err
		}

		// 2. Обновляем данные пользователя
		chatMember, err := c.Bot.GetChatMember(tgbotapi.ChatConfigWithUser{
			SuperGroupUsername: "@" + c.Conf.Management.ChannelUsername,
			UserID:             u.TelegramID,
		})
		if err != nil {
			return err
		}

		// 3. Нет Username? Не подписан на канал? - удаляем в DB, continue
		if chatMember.User.UserName == "" || chatMember.Status != "member" {
			err = u.RemoveId()
			if err != nil {
				return err
			}
			continue
		}

		// 4. Объявляем победителя, break for
		bufferBot := bytes.Buffer{}
		bufferBot.WriteString("Итак, выигрывает...\n")
		bufferBot.WriteString(utils.FormatUserDog(u))
		bufferBot.WriteString("\nСпасибо всем за участие!")
		return utils.SendBotMessage(c.Bot, int64(msg.Chat.ID), bufferBot.String(), false)
	}
}
