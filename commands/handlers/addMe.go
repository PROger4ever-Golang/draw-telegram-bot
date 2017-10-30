package handlers

import (
	"errors"
	"time"

	"gopkg.in/mgo.v2"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"gopkg.in/mgo.v2/bson"

	"bitbucket.org/proger4ever/draw-telegram-bot/commands/utils"
	"bitbucket.org/proger4ever/draw-telegram-bot/config"
	"bitbucket.org/proger4ever/draw-telegram-bot/mongo/models/user"
	"bitbucket.org/proger4ever/draw-telegram-bot/userApi"
)

type AddMeHandler struct {
	Bot  *tgbotapi.BotAPI
	Conf *config.Config
	Tool *userapi.Tool
}

func (c *AddMeHandler) GetName() string {
	return "addMe"
}

func (c *AddMeHandler) IsForOwnersOnly() bool {
	return false
}

func (c *AddMeHandler) GetParamsCount() int {
	return 0
}

func (c *AddMeHandler) Init(conf *config.Config, tool *userapi.Tool, bot *tgbotapi.BotAPI) {
	c.Bot = bot
	c.Conf = conf
	c.Tool = tool
}

func (c *AddMeHandler) Execute(msg *tgbotapi.Message, params []string) error {
	// 1. От кого пришло?
	if msg.From == nil {
		err := errors.New("Команда недоступна в этом чате: невозможно определить отправителя сообщения")
		return utils.SendBotError(c.Bot, int64(msg.Chat.ID), err)
	}

	// 2. Имеется Username?
	if msg.From.UserName == "" {
		err := errors.New("Необходимо задать username/ник в настройках телеграма")
		return utils.SendBotError(c.Bot, int64(msg.Chat.ID), err)
	}

	// 2. Имеется Username?
	uc := user.NewCollectionDefault()
	u, err := uc.FindOne(bson.M{
		"telegram_id": msg.From.ID,
	})
	if err != nil && err != mgo.ErrNotFound {
		return err
	}

	// 3. Если недавно регали - то зачем так часто обновлять его данные в АПИ?
	fiveMinutesAgo := time.Now().Add(-5 * time.Minute)
	if err != mgo.ErrNotFound && u.UpdatedAt.Before(fiveMinutesAgo) {
		resp := "Вы уже зарегистрированы в розыгрышах @mazimota.\nВаши данные обновлялись недавно.\nПравила и подробности можно найти здесь: @mazimota_rules"
		return utils.SendBotMessage(c.Bot, int64(msg.Chat.ID), resp, false)
	}

	// 4. Подписан на канал?
	chatMember, err := c.Bot.GetChatMember(tgbotapi.ChatConfigWithUser{
		SuperGroupUsername: "@" + c.Conf.Management.ChannelUsername,
		UserID:             msg.From.ID,
	})
	if err != nil {
		return err
	}
	switch chatMember.Status {
	case "creator":
		fallthrough
	case "administrator":
		err := errors.New("Администраторы не могут учавствовать в розыгрыше")
		return utils.SendBotError(c.Bot, int64(msg.Chat.ID), err)
	case "kicked":
		err := errors.New("Вы были забанены на канале и не можете учавствовать в розыгрыше")
		return utils.SendBotError(c.Bot, int64(msg.Chat.ID), err)
	case "member":
	case "left":
		fallthrough
	default:
		err := errors.New("Вы должны быть подписаны на канал @mazimota")
		return utils.SendBotError(c.Bot, int64(msg.Chat.ID), err)
	}
	if err != nil {
		return err
	}

	// 5. Записать изменения в DB
	u.TelegramID = msg.From.ID
	u.Username = msg.From.UserName
	u.FirstName = msg.From.FirstName
	u.LastName = msg.From.LastName
	u.Status = chatMember.Status
	info, err := u.UpsertId()
	if err != nil {
		return err
	}

	// 5. Сообщить успех, правила
	var resp string
	if info.Updated > 0 {
		resp = "Вы уже зарегистрированы в розыгрышах @mazimota.\nВаши данные обновлены.\nПравила и подробности можно найти здесь: @mazimota_rules"
	} else {
		resp = "Вы зарегистрированы в розыгрышах @mazimota.\nПравила и подробности можно найти здесь: @mazimota_rules"
	}
	return utils.SendBotMessage(c.Bot, int64(msg.Chat.ID), resp, false)
}
