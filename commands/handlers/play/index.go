package playpkg

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"gopkg.in/mgo.v2/bson"

	"bitbucket.org/proger4ever/draw-telegram-bot/bot"
	"bitbucket.org/proger4ever/draw-telegram-bot/commands/utils"
	"bitbucket.org/proger4ever/draw-telegram-bot/config"
	"bitbucket.org/proger4ever/draw-telegram-bot/error"
	"bitbucket.org/proger4ever/draw-telegram-bot/mongo/models/user"
	snPkg "bitbucket.org/proger4ever/draw-telegram-bot/mongo/tools/SampleNavigator"
	"bitbucket.org/proger4ever/draw-telegram-bot/userApi"
)

//TODO: split common error messages into a package
const cantQueryDB = "Ошибка при операции с БД"
const cantQueryChatMember = "Ошибка при получении информации о участнике канала"
const cantSendBotMessage = "Ошибка при отправке сообщения от имени бота"

const commandUnavailable = "Команда недоступна в этом чате"
const notEnoughParticipants = `Недостаточно участников канала, подписавшихся у бота на розыгрыш.
Отправить приз в /dev/null - не смешите мои байтики! :)`

const contenderAnnouncement = `Итак, выигрывает...
%s
Поздравляем!`
const contendersAnnouncement = `Итак, выигрывают...
%s
Поздравляем!`

//var playPipeline = []bson.M{{
//	"$match": bson.M{
//		"username": bson.M{
//			// "$exists": true,
//			"$ne": "",
//		},
//		"status": "member",
//	},
//}, {
//	"$sample": bson.M{
//		"size": 1,
//	},
//}}

type Handler struct {
	Bot  *botpkg.Bot
	Conf *config.Config
	Tool *userapi.Tool
}

func (h *Handler) GetAliases() []string {
	return []string{"/play", "розыграй"}
}

func (h *Handler) IsForOwnersOnly() bool {
	return false
}

func (h *Handler) GetParamsMinCount() int {
	return 0
}

func (h *Handler) Init(conf *config.Config, tool *userapi.Tool, bot *botpkg.Bot) {
	h.Bot = bot
	h.Conf = conf
	h.Tool = tool
}

func (h *Handler) Execute(msg *tgbotapi.Message, params []string) (err error) {
	if msg.Chat.UserName != h.Conf.Management.ChannelUsername {
		return eepkg.New(true, false, commandUnavailable)
	}

	prizeCount := 1
	if len(params) > 0 {
		prizeCountTmp, err := strconv.Atoi(params[0])
		if err == nil && prizeCountTmp > 0 {
			prizeCount = prizeCountTmp
		}
	}

	<-time.After(5 * time.Second)

	contenders, err := h.getContenders(prizeCount)
	if err != nil {
		return err
	}
	if len(contenders) < prizeCount {
		return eepkg.New(true, false, notEnoughParticipants)
	}

	// Объявляем победителей
	resp := ""
	if len(contenders) == 1 {
		contenderString := utils.FormatUserDog(contenders[0])
		resp = fmt.Sprintf(contenderAnnouncement, contenderString)
	} else {
		contendersString := utils.FormatUsers(contenders, utils.FormatUserDog)
		resp = fmt.Sprintf(contendersAnnouncement, contendersString)
	}
	err = h.Bot.SendMessage(int64(msg.Chat.ID), resp, false)
	return eepkg.Wrap(err, false, true, cantSendBotMessage)
}

func (h *Handler) getContenders(count int) (contenders []*user.User, err error) {
	contenders = make([]*user.User, 0, count)

	uBufLen := int(math.Ceil(float64(count)*0.2)) + count
	uc := user.NewCollectionDefault()
	sn := snPkg.New(uc.BaseCollection, bson.M{}, uBufLen)

	for i := 0; i < count; i++ {
		u := user.New(uc)
		isVerified := false
		for !isVerified {
			err = sn.Next(u)
			if err == snPkg.ErrNotEnough {
				return contenders, nil
			}
			if err != nil {
				return contenders, eepkg.Wrap(err, false, true, cantQueryDB)
			}

			isVerified, err = h.verifyContender(u)
			if err != nil {
				return
			}
		}
		contenders = append(contenders, u)
	}
	return
}

func (h *Handler) verifyContender(c *user.User) (isVerified bool, err error) {
	// Обновляем данные пользователя
	chatMember, err := h.Bot.GetChatMember(tgbotapi.ChatConfigWithUser{
		SuperGroupUsername: "@" + h.Conf.Management.ChannelUsername,
		UserID:             c.TelegramID,
	})
	if err != nil {
		return false, eepkg.Wrap(err, false, true, cantQueryChatMember)
	}

	// Нет Username? Не подписан на канал? - удаляем в DB, continue
	if chatMember.User.UserName == "" || chatMember.Status != "member" {
		err = c.RemoveId()
		if err != nil {
			return false, eepkg.Wrap(err, false, true, cantQueryDB)
		}
		return false, nil
	}
	return true, nil
}
