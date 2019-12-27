//Package events ...
package events

import (
	"github.com/amiraliio/tgbp/config"
	"github.com/amiraliio/tgbp/helpers"
	tb "gopkg.in/tucnak/telebot.v2"
)

func keyboardsEvents(app *config.App, bot *tb.Bot) {

	if keyboardsEventsHandler(app, bot, &Event{
		Event:      &addAnonMessage,
		UserState:  "add_anon_message",
		Controller: "AddAnonMessageToChannel",
	}) {
		return
	}

}

func keyboardsEventsHandler(app *config.App, bot *tb.Bot, request *Event) (result bool) {
	result = false
	bot.Handle(request.Event, func(message *tb.Message) {
		helpers.Invoke(BotService{}, request.Controller, app, bot, message, request)
		result = true
	})
	return result
}
