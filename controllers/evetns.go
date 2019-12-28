//Package controllers ...
package controllers

import (
	"github.com/amiraliio/tgbp/config"
	"github.com/amiraliio/tgbp/helpers"
	tb "gopkg.in/tucnak/telebot.v2"
)

type BotService struct{}

type Event struct {
	UserState, Command, Command1, Controller string
	Event                                    interface{}
}

func Init(app *config.App, bot *tb.Bot, state interface{}) bool {
	if state != nil {
		return true
	}
	onActionEvents(app, bot)
	onTextEvents(app, bot)
	onCallbackEvents(app, bot)
	return false
}

func generalEventsHandler(app *config.App, bot *tb.Bot, message *tb.Message, request *Event) bool {
	var result bool
	helpers.Invoke(new(BotService), &result, request.Controller, app, bot, message, request)
	return result
}
