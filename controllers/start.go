package controllers

import (
	"github.com/amiraliio/tgbp/config"
	"github.com/amiraliio/tgbp/events"
	tb "gopkg.in/tucnak/telebot.v2"
	"log"
	"strings"
)

func (service BotService) StartBot(app *config.App, bot *tb.Bot, message *tb.Message, request *events.Event) bool {
	if strings.TrimSpace(message.Text) == request.Command || strings.TrimSpace(message.Text) == request.Command1 {
		db := app.DB()
		defer db.Close()
		if message.Sender != nil {
			SaveUserLastState(db, app, bot, message.Text, message.Sender.ID, request.UserState)
		}
		newReplyModel := new(tb.ReplyMarkup)
		newReplyModel.ReplyKeyboard = events.StartBotKeys
		newSendOption := new(tb.SendOptions)
		newSendOption.ReplyMarkup = newReplyModel
		_ = bot.Delete(message)
		_, err := bot.Send(message.Sender, "What Do You Want To Do?", newSendOption)
		if err != nil {
			log.Println(err)
			return false
		}
		return true
	}
	return false
}

func (service BotService) StartBotCallback(app *config.App, bot *tb.Bot, callback *tb.Callback, request *events.Event) bool {
	db := app.DB()
	defer db.Close()
	if callback.Sender != nil {
		SaveUserLastState(db, app, bot, callback.Data, callback.Sender.ID, request.UserState)
	}
	newReplyModel := new(tb.ReplyMarkup)
	newReplyModel.ReplyKeyboard = events.StartBotKeys
	newSendOption := new(tb.SendOptions)
	newSendOption.ReplyMarkup = newReplyModel
	_ = bot.Delete(callback.Message)
	_, err := bot.Send(callback.Sender, "What Do You Want To Do?", newSendOption)
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

func (service BotService) AddAnonMessageToChannel(app *config.App, bot *tb.Bot, message *tb.Message, request *events.Event) bool {
	db := app.DB()
	defer db.Close()
	if message.Sender != nil {
		SaveUserLastState(db, app, bot, message.Text, message.Sender.ID, request.UserState)
	}
	bot.Send(message.Sender, "For anonymous message add the bot to your group and start messaging")
	return true
}
