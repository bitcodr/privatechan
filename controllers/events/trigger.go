//Package events ...
package controllers

import (
	"github.com/amiraliio/tgbp/config"
	"github.com/amiraliio/tgbp/helpers"
	tb "gopkg.in/tucnak/telebot.v2"
)

func triggersEvents(app *config.App, bot *tb.Bot) {

	if triggersEventsHandler(app, bot, &Event{
		Event:      tb.OnChannelPost,
		UserState:  "register_channel",
		Command:    "/enable_anonymity_support",
		Controller: "RegisterChannel",
	}) {
		return
	}

	if triggersEventsHandler(app, bot, &Event{
		Event:      tb.OnAddedToGroup,
		UserState:  "register_group",
		Controller: "RegisterGroup",
	}) {
		return
	}
}

func triggersEventsHandler(app *config.App, bot *tb.Bot, request *Event) (result bool) {
	result = false
	bot.Handle(request.Event, func(message *tb.Message) {
		helpers.Invoke(BotService{}, request.Controller, app, bot, message, request)
		result = true
	})
	return result
}
