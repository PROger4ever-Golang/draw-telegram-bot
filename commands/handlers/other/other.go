package other

// case "getDialogs":
// 	r, err := c.Tool.MessagesGetDialogs()
// 	if err == nil {
// 		bs, err := json.Marshal(r)
// 		resp := fmt.Sprintf("```\nMessagesGetDialogs result: %s, %q\n```", string(bs), err)
// 		err = SendBotMessage(c.Bot, int64(chat.ID), resp)
// 	} else {
// 		err = SendBotError(c.Bot, int64(chat.ID), err)
// 	}
// 	return err
// case "msgTest":
// 	resp := fmt.Sprintf("```\n/completeLoginWithCode@%v -\n```", bot.Self.UserName)
// 	return SendBotMessage(c.Bot, int64(chat.ID), resp)
// case "panic":
// 	panic("panic test")
