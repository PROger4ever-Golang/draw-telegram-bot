package addmepkg

import (
	"time"

	"gopkg.in/mgo.v2"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"gopkg.in/mgo.v2/bson"

	"bitbucket.org/proger4ever/draw-telegram-bot/bot"
	"bitbucket.org/proger4ever/draw-telegram-bot/config"
	"bitbucket.org/proger4ever/draw-telegram-bot/error"
	"bitbucket.org/proger4ever/draw-telegram-bot/mongo/models/user"
	"bitbucket.org/proger4ever/draw-telegram-bot/userApi"
)

const cantQueryDB = "Ошибка при операции с БД"
const cantSendBotMessage = "Ошибка при отправке сообщения от имени бота"
const cantQueryChatMember = "Ошибка при получении информации о участнике канала"

const commandUnavailable = "Команда недоступна в этом чате: невозможно определить отправителя сообщения"
const noUsername = `Ошибка регистрации
У Вас нет имени пользователя. 
Пожалуйста, добавьте свое имя пользователя в настройках приложения`

const detailsInfo = `Правила @mazimota_rules
Чат @mazimota_chat`
const adminsMayntPlay = "Администраторы не могут учавствовать в розыгрыше"
const youBanned = "Вы были забанены на канале и не можете учавствовать в розыгрыше"
const noChannelSubscription = `Ошибка регистрации.
Вы не подписаны на канал @mazimota, ведь именно на нем проходят ежедневные розыгрыши призов! 
Пожалуйста, сначала подпишитесь на канал @mazimota, и обязательно возвращайтесь зарегистрироваться у бота)
` + detailsInfo
const registeredSuccessfully = `Вы зарегистрированы!
Желаем удачи!
` + detailsInfo
const alreadyRegisteredRecently = `Вы зарегистрированы!
Ваши данные обновлены недавно.
Желаем удачи!
` + detailsInfo
const alreadyRegistered = `Вы зарегистрированы!
Ваши данные обновлены.
Желаем удачи!
` + detailsInfo

type Handler struct {
	Bot  *botpkg.Bot
	Conf *config.Config
	Tool *userapi.Tool
}

func (h *Handler) GetAliases() []string {
	return []string{"addMe", "start"}
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

func (h *Handler) Execute(msg *tgbotapi.Message, params []string) error {
	// 1. От кого пришло?
	if msg.From == nil {
		return eepkg.New(true, false, commandUnavailable)
	}

	// 2. Имеется Username?
	if msg.From.UserName == "" {
		return eepkg.New(true, false, noUsername)
	}

	// 3. Ищем старую регистрацию в базе
	uc := user.NewCollectionDefault()
	u, err := uc.FindOne(bson.M{
		"telegram_id": msg.From.ID,
	})
	if err != nil && err != mgo.ErrNotFound {
		return eepkg.Wrap(err, false, true, cantQueryDB)
	}

	// 4. Если недавно регали - то зачем так часто обновлять его данные в АПИ?
	if err != mgo.ErrNotFound {
		err = nil
		minUpdateTime := time.Now().Add(-30 * time.Second)
		if minUpdateTime.Before(u.UpdatedAt) {
			err = h.Bot.SendMessage(int64(msg.Chat.ID), alreadyRegisteredRecently, false)
			return eepkg.Wrap(err, false, true, cantSendBotMessage)
		}
	}

	// 5. Подписан на канал?
	chatMember, err := h.Bot.GetChatMember(tgbotapi.ChatConfigWithUser{
		SuperGroupUsername: "@" + h.Conf.Management.ChannelUsername,
		UserID:             msg.From.ID,
	})
	if err != nil {
		return eepkg.Wrap(err, false, true, cantQueryChatMember)
	}
	switch chatMember.Status {
	case "creator":
		fallthrough
	case "administrator":
		return eepkg.New(true, false, adminsMayntPlay)
	case "kicked":
		return eepkg.New(true, false, youBanned)
	case "member":
	case "left":
		fallthrough
	default:
		return eepkg.New(true, false, noChannelSubscription)
	}

	// 6. Записать изменения в DB
	u.TelegramID = msg.From.ID
	u.Username = msg.From.UserName
	u.FirstName = msg.From.FirstName
	u.LastName = msg.From.LastName
	u.Status = chatMember.Status
	info, err := u.UpsertId()
	if err != nil {
		return eepkg.Wrap(err, false, true, cantQueryDB)
	}

	// 7. Сообщить успех, правила
	var resp string
	if info.Updated > 0 {
		resp = alreadyRegistered
	} else {
		resp = registeredSuccessfully
	}
	err = h.Bot.SendMessage(int64(msg.Chat.ID), resp, false)
	return eepkg.Wrap(err, false, true, cantSendBotMessage)
}
