package userpkg

import (
	"fmt"
	"strconv"

	"github.com/PROger4ever/draw-telegram-bot/bot"
	"github.com/PROger4ever/draw-telegram-bot/config"
	"github.com/PROger4ever/draw-telegram-bot/error"
	"github.com/PROger4ever/draw-telegram-bot/mongo/models/user"
	"github.com/PROger4ever/draw-telegram-bot/userApi"
	contenderUtils "github.com/PROger4ever/draw-telegram-bot/utils/contender"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"gopkg.in/mgo.v2"
)

const cantParseTelegramID = "Не могу разобрать UserID"
const incorrectIdentifier = "Не могу разобрать идентификатор пользователя"

var incorrectIdentifierError = eepkg.New(true, false, incorrectIdentifier)

const userNotRegistered = "Такой пользователь не зарегистрирован у бота"

var userNotRegisteredError = eepkg.New(true, false, userNotRegistered)

const userNotRegisteredAnymore = `Только что такой пользователь был зарегистрирован у бота.
НО! В ходе дополнительной проверки выяснилось, что на текущий момент с его стороны выполнены не все условия участия в розыгрыша.
Теперь этот пользователь удалён из БД бота.`

var userNotRegisteredAnymoreError = eepkg.New(true, false, userNotRegisteredAnymore)

const userRegistered = `Пользователь зарегистрирован у бота.
На текущий момент выполняет все необходимые условия для участия в розыгрышах.
Свежие данные пользователя обновлены с серверов телеграма.
` + "```" + `
%s
` + "```"

type Handler struct {
	Bot  *botpkg.Bot
	Conf *config.Config
	Tool *userapi.Tool
}

func (h *Handler) GetAliases() []string {
	return []string{"пользователь", "user"}
}

func (h *Handler) IsForOwnersOnly() bool {
	return true
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
	identifier := params[0]

	if len(identifier) < 2 {
		return incorrectIdentifierError
	}

	identifierType := identifier[0]
	identifierValue := identifier[1:]
	uc := user.NewCollectionDefault()
	var u *user.User
	switch identifierType {
	case '@':
		u, err = uc.FindOneByUsername(identifierValue)
	case '#':
		telegramID, errStd := strconv.Atoi(identifierValue)
		if errStd != nil {
			return eepkg.Wrap(errStd, false, true, cantParseTelegramID)
		}
		u, err = uc.FindOneByTelegramID(telegramID)
	default:
		return incorrectIdentifierError
	}
	if err != nil && err.GetRoot() == mgo.ErrNotFound {
		return userNotRegisteredError
	}
	if err != nil {
		return err
	}

	isVerified, err := contenderUtils.RefreshUser(h.Bot, u)
	if !isVerified {
		return userNotRegisteredAnymoreError
	}

	resp := fmt.Sprintf(userRegistered, contenderUtils.FormatUserComplete(u))
	return h.Bot.SendMessageMarkdown(int64(msg.Chat.ID), resp)
}
