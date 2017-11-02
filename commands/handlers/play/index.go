package play

import (
	"fmt"
	"time"

	"gopkg.in/mgo.v2"

	"github.com/PROger4ever/telegram-bot-api"
	"gopkg.in/mgo.v2/bson"

	"bitbucket.org/proger4ever/draw-telegram-bot/commands/utils"
	"bitbucket.org/proger4ever/draw-telegram-bot/config"
	ee "bitbucket.org/proger4ever/draw-telegram-bot/errors"
	"bitbucket.org/proger4ever/draw-telegram-bot/mongo/models/user"
	"bitbucket.org/proger4ever/draw-telegram-bot/userApi"
)

//TODO: split common error messages into a package
const cantQueryDB = "Ошибка при операции с БД"
const cantQueryChatMember = "Ошибка при получении информации о участнике канала"
const cantSendBotMessage = "Ошибка при отправке сообщения от имени бота"

const commandUnavailable = "Команда недоступна в этом чате"
const noParticipants = `Нет участников канала, подписавшихся у бота на розыгрыш.
Розыгрыш среди НИКОГО - не смешите мои байтики`

const startingDraw = "Начинаем розыгрыш"
const contenderAnnouncement = `Итак, выигрывает...
%s
Поздравляем!`

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

type Handler struct {
	Bot  *tgbotapi.BotAPI
	Conf *config.Config
	Tool *userapi.Tool
}

func (h *Handler) GetAliases() []string {
	return []string{"play"}
}

func (h *Handler) IsForOwnersOnly() bool {
	return false
}

func (h *Handler) GetParamsMinCount() int {
	return 0
}

func (h *Handler) Init(conf *config.Config, tool *userapi.Tool, bot *tgbotapi.BotAPI) {
	h.Bot = bot
	h.Conf = conf
	h.Tool = tool
}

func (h *Handler) Execute(msg *tgbotapi.Message, params []string) error {
	if msg.Chat.UserName != h.Conf.Management.ChannelUsername {
		return ee.New(true, false, commandUnavailable)
	}

	err := utils.SendBotMessage(h.Bot, int64(msg.Chat.ID), startingDraw, false)
	if err != nil {
		return ee.Wrap(err, false, true, cantSendBotMessage)
	}
	<-time.After(5 * time.Second)

	uc := user.NewCollectionDefault()
	for {
		u, err := uc.PipeOne(playPipeline)
		if err == mgo.ErrNotFound {
			return ee.New(true, false, noParticipants)
		}
		if err != nil {
			return ee.Wrap(err, false, true, cantQueryDB)
		}

		// Обновляем данные пользователя
		chatMember, err := h.Bot.GetChatMember(tgbotapi.ChatConfigWithUser{
			SuperGroupUsername: "@" + h.Conf.Management.ChannelUsername,
			UserID:             u.TelegramID,
		})
		if err != nil {
			return ee.Wrap(err, false, true, cantQueryChatMember)
		}

		// Нет Username? Не подписан на канал? - удаляем в DB, continue
		if chatMember.User.UserName == "" || chatMember.Status != "member" {
			err = u.RemoveId()
			if err != nil {
				return ee.Wrap(err, false, true, cantQueryDB)
			}
			continue
		}

		// Объявляем победителя, break for
		resp := fmt.Sprintf(contenderAnnouncement, utils.FormatUserDog(u))
		err = utils.SendBotMessage(h.Bot, int64(msg.Chat.ID), resp, false)
		return ee.Wrap(err, false, true, cantSendBotMessage)
	}
}
