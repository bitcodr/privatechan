package controller

import (
	tb "gopkg.in/tucnak/telebot.v2"
	"log"
)

func StartBot(bot *tb.Bot, m *tb.Message, replyKeys [][]tb.ReplyButton) {
	newReplyModel := new(tb.ReplyMarkup)
	newReplyModel.ReplyKeyboard = replyKeys
	newSendOption := new(tb.SendOptions)
	newSendOption.ReplyMarkup = newReplyModel
	_ = bot.Delete(m)
	_, err := bot.Send(m.Sender, "What Do You Want To Do?", newSendOption)
	if err != nil {
		log.Println(err)
	}
}

func AddAnonMessageToChannel(bot *tb.Bot, m *tb.User) {
	bot.Send(m, "For anonymous message add the bot to your group and start messaging")
}
