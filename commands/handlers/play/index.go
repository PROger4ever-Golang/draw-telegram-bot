package playpkg

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"bitbucket.org/proger4ever/draw-telegram-bot/bot"
	"bitbucket.org/proger4ever/draw-telegram-bot/config"
	"bitbucket.org/proger4ever/draw-telegram-bot/error"
	"bitbucket.org/proger4ever/draw-telegram-bot/mongo/models/user"
	snPkg "bitbucket.org/proger4ever/draw-telegram-bot/mongo/tools/SampleNavigator"
	"bitbucket.org/proger4ever/draw-telegram-bot/userApi"
	contenderUtils "bitbucket.org/proger4ever/draw-telegram-bot/utils/contender"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"gopkg.in/mgo.v2/bson"
)

const cantNavigateSample = "Can't navigate mongodb sample"

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

func (h *Handler) Execute(msg *tgbotapi.Message, params []string) (err *eepkg.ExtendedError) {
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
		contenderString := contenderUtils.FormatUserDog(contenders[0])
		resp = fmt.Sprintf(contenderAnnouncement, contenderString)
	} else {
		contendersString := contenderUtils.FormatUsers(contenders, contenderUtils.FormatUserDog)
		resp = fmt.Sprintf(contendersAnnouncement, contendersString)
	}
	return h.Bot.SendMessage(int64(msg.Chat.ID), resp)
}

func (h *Handler) getContenders(count int) (contenders []*user.User, err *eepkg.ExtendedError) {
	contenders = make([]*user.User, 0, count)

	uBufLen := int(math.Ceil(float64(count)*0.2)) + count
	uc := user.NewCollectionDefault()
	sn := snPkg.New(uc.BaseCollection, bson.M{}, uBufLen)

	for i := 0; i < count; i++ {
		u := user.New(uc)
		isVerified := false
		for !isVerified {
			errStd := sn.Next(u)
			if errStd == snPkg.ErrNotEnough {
				return contenders, nil
			}
			if errStd != nil {
				return contenders, eepkg.Wrap(errStd, false, true, cantNavigateSample)
			}

			u, isVerified, err = contenderUtils.RefreshUser(h.Bot, u.TelegramID)
			if err != nil {
				return
			}
		}
		contenders = append(contenders, u)
	}
	return
}
