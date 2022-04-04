package addmepkg

import (
	"time"

	"github.com/PROger4ever-Golang/draw-telegram-bot/bot"
	"github.com/PROger4ever-Golang/draw-telegram-bot/config"
	"github.com/PROger4ever-Golang/draw-telegram-bot/error"
	"github.com/PROger4ever-Golang/draw-telegram-bot/mongo/models/user"
	"github.com/PROger4ever-Golang/draw-telegram-bot/userApi"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const commandUnavailable = "Команда недоступна в этом чате: невозможно определить отправителя сообщения"
const noUsername = `Ошибка регистрации
У Вас нет имени пользователя. 
Пожалуйста, добавьте свое имя пользователя в настройках приложения`

const adminsMayntPlay = "Администраторы не могут учавствовать в розыгрыше"
const youBanned = "Вы были забанены на канале и не можете учавствовать в розыгрыше"
const noChannelSubscription = `Ошибка регистрации.
Вы не подписаны на канал @StartupMarket_rus, ведь именно на нем проходят ежедневные розыгрыши призов! 
Пожалуйста, сначала подпишитесь на канал @StartupMarket_rus, и обязательно возвращайтесь зарегистрироваться у бота)
`
const registeredSuccessfully = `Вы зарегистрированы!
Информация о Вас и время вашего последнего "онлайна у бота" сохранены.
Желаем удачи!
`
const alreadyRegisteredRecently = `Вы уже зарегистрированы!
Информация о Вас и время вашего последнего "онлайна у бота" обновлены совсем недавно.
Желаем удачи!
`
const alreadyRegistered = `Вы уже зарегистрированы!
Информация о Вас и время вашего последнего "онлайна у бота" обновлены.
Желаем удачи!
`

type Handler struct {
	Bot  *botpkg.Bot
	Conf *config.Config
	Tool *userapi.Tool
}

func (h *Handler) GetAliases() []string {
	return []string{"регистрация", "addMe"}
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

func (h *Handler) Execute(msg *tgbotapi.Message, params []string) *eepkg.ExtendedError {
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
	if err != nil && err.GetRoot() != mgo.ErrNotFound {
		return err
	}

	// 4. Если недавно регали - то зачем так часто обновлять его данные в АПИ?
	if err != mgo.ErrNotFound {
		err = nil
		minUpdateTime := time.Now().Add(-30 * time.Second)
		if minUpdateTime.Before(u.LastAdditionAt) {
			return h.Bot.SendMessageUserKeyboard(int64(msg.Chat.ID), alreadyRegisteredRecently)
		}
	}

	// 5. Подписан на канал?
	chatMember, err := h.Bot.GetChatMember(msg.From.ID)
	if err != nil {
		return err
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
	u.LastAdditionAt = time.Now()
	info, err := u.UpsertId()
	if err != nil {
		return err
	}

	// 7. Сообщить успех, правила
	var resp string
	if info.Updated > 0 {
		resp = alreadyRegistered
	} else {
		resp = registeredSuccessfully
	}
	return h.Bot.SendMessageUserKeyboard(int64(msg.Chat.ID), resp)
}
