package controllers

import (
	"github.com/amiraliio/tgbp/config"
	tb "gopkg.in/tucnak/telebot.v2"
)

func onActionEvents(app *config.App, bot *tb.Bot) {

	bot.Handle(tb.OnChannelPost, func(message *tb.Message) {
		if generalEventsHandler(app, bot, message, &Event{
			Event:      tb.OnChannelPost,
			UserState:  "register_channel",
			Command:    "/enable_anonymity_support",
			Controller: "RegisterChannel",
		}) {
			Init(app, bot, true)
		}
	})

	bot.Handle(tb.OnAddedToGroup, func(message *tb.Message) {
		if generalEventsHandler(app, bot, message, &Event{
			Event:      tb.OnAddedToGroup,
			UserState:  "register_group",
			Controller: "RegisterGroup",
		}) {
			Init(app, bot, true)
		}
	})

	bot.Handle(tb.OnNewGroupTitle, func(message *tb.Message) {
		if generalEventsHandler(app, bot, message, &Event{
			Event:      tb.OnNewGroupTitle,
			UserState:  "updateGrouptitle",
			Controller: "UpdateGroupTitle",
		}) {
			Init(app, bot, true)
		}
	})

	bot.Handle(&addAnonMessage, func(message *tb.Message) {
		if generalEventsHandler(app, bot, message, &Event{
			Event:      &addAnonMessage,
			UserState:  "add_anon_message",
			Controller: "AddAnonMessageToChannel",
		}) {
			Init(app, bot, true)
		}
	})

}
