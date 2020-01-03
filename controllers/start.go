package controllers

import (
	"github.com/amiraliio/tgbp/config"
	tb "gopkg.in/tucnak/telebot.v2"
	"log"
	"strings"
)

func (service *BotService) StartBot(app *config.App, bot *tb.Bot, message *tb.Message, request *Event) bool {
	if strings.TrimSpace(message.Text) == request.Command || strings.TrimSpace(message.Text) == request.Command1 {
		db := app.DB()
		defer db.Close()
		if message.Sender != nil {
			SaveUserLastState(db, app, bot, message.Text, message.Sender.ID, request.UserState)
		}
		newReplyModel := new(tb.ReplyMarkup)
		newReplyModel.ReplyKeyboard = StartBotKeys
		newSendOption := new(tb.SendOptions)
		newSendOption.ReplyMarkup = newReplyModel
		_ = bot.Delete(message)
		_, err := bot.Send(message.Sender, config.LangConfig.GetString("MESSAGES.WHAT_TO_YOU_WANT"), newSendOption)
		if err != nil {
			log.Println(err)
			return false
		}
		return true
	}
	return false
}

func (service *BotService) StartBotCallback(app *config.App, bot *tb.Bot, callback *tb.Callback, request *Event) bool {
	if strings.TrimSpace(callback.Data) == request.Command || strings.TrimSpace(callback.Data) == request.Command1 {
		db := app.DB()
		defer db.Close()
		if callback.Sender != nil {
			SaveUserLastState(db, app, bot, callback.Data, callback.Sender.ID, request.UserState)
		}
		newReplyModel := new(tb.ReplyMarkup)
		newReplyModel.ReplyKeyboard = StartBotKeys
		newSendOption := new(tb.SendOptions)
		newSendOption.ReplyMarkup = newReplyModel
		_ = bot.Delete(callback.Message)
		_, err := bot.Send(callback.Sender, config.LangConfig.GetString("MESSAGES.WHAT_TO_YOU_WANT"), newSendOption)
		if err != nil {
			log.Println(err)
			return false
		}
		return true
	}
	return false
}

func (service *BotService) AddAnonMessageToChannel(app *config.App, bot *tb.Bot, message *tb.Message, request *Event) bool {
	db := app.DB()
	defer db.Close()
	if message.Sender != nil {
		SaveUserLastState(db, app, bot, message.Text, message.Sender.ID, request.UserState)
	}
	bot.Send(message.Sender, config.LangConfig.GetString("MESSAGES.ALERT_FOR_ANON_MESSAGE"))
	return true
}
