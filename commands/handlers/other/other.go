package otherpkg

// case "getDialogs":
// 	r, err := c.Tool.MessagesGetDialogs()
// 	if err == nil {
// 		bs, err := json.Marshal(r)
// 		resp := fmt.Sprintf("```\nMessagesGetDialogs result: %s, %q\n```", string(bs), err)
// 		err = SendMessage(c.Bot, int64(chat.ID), resp)
// 	} else {
// 		err = SendError(c.Bot, int64(chat.ID), err)
// 	}
// 	return err
// case "msgTest":
// 	resp := fmt.Sprintf("```\n/completeLoginWithCode@%v -\n```", bot.Self.UserName)
// 	return SendMessage(c.Bot, int64(chat.ID), resp)
// case "panic":
// 	panic("panic test")
