package playonlinepkg

import (
	"fmt"
	"strconv"
	"time"

	"github.com/PROger4ever-Golang/draw-telegram-bot/bot"
	"github.com/PROger4ever-Golang/draw-telegram-bot/config"
	"github.com/PROger4ever-Golang/draw-telegram-bot/error"
	"github.com/PROger4ever-Golang/draw-telegram-bot/mongo/models/user"
	snPkg "github.com/PROger4ever-Golang/draw-telegram-bot/mongo/tools/SampleNavigator"
	"github.com/PROger4ever-Golang/draw-telegram-bot/userApi"
	contenderUtils "github.com/PROger4ever-Golang/draw-telegram-bot/utils/contender"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"gopkg.in/mgo.v2/bson"
)

const cantNavigateSample = "Can't navigate mongodb sample"

const commandUnavailable = "Команда недоступна в этом чате"
const invalidDuration = `Максимальная длительность оффлайна указана неверно.
Примеры корректного формата:
1) 15m
2) 1h15m30s
`
const notEnoughParticipants = `Недостаточно участников канала, подписавшихся у бота на онлайн-розыгрыш.
Отправить приз в /dev/null - не смешите мои байтики! :)`

const contenderAnnouncement = `Итак, выигрывает...
%s
Поздравляем!`
const contendersAnnouncement = `Итак, выигрывают...
%s
Поздравляем!`

type Handler struct {
	Bot  *botpkg.Bot
	Conf *config.Config
	Tool *userapi.Tool
}

func (h *Handler) GetAliases() []string {
	return []string{"/playOnline", "розыграйОнлайн"}
}

func (h *Handler) IsForOwnersOnly() bool {
	return false
}

func (h *Handler) GetParamsMinCount() int {
	return 1
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

	maxOfflineDuration, errStd := time.ParseDuration(params[0])
	if errStd != nil {
		return eepkg.New(true, false, invalidDuration)
	}

	prizeCount := 1
	if len(params) >= 2 {
		prizeCountTmp, err := strconv.Atoi(params[1])
		if err == nil && prizeCountTmp > 0 {
			prizeCount = prizeCountTmp
		}
	}

	<-time.After(5 * time.Second)

	contenders, err := h.getContenders(maxOfflineDuration, prizeCount)
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

func (h *Handler) getContenders(maxOfflineDuration time.Duration, count int) (contenders []*user.User, err *eepkg.ExtendedError) {
	contenders = make([]*user.User, 0, count)
	uc := user.NewCollectionDefault()

	maxOfflineTime := time.Now().Add(-maxOfflineDuration)
	var match = bson.M{
		"last_addition_at": bson.M{
			"$gte": maxOfflineTime,
		},
	}
	fmt.Printf("maxOfflineTime: %v", maxOfflineTime)
	sn := snPkg.New(uc.BaseCollection, match, count)

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

			isVerified, err = contenderUtils.RefreshUser(h.Bot, u)
			if err != nil {
				return
			}
		}
		contenders = append(contenders, u)
	}
	return
}
