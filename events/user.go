//Package events ...
package events

import (
	"github.com/amiraliio/tgbp/config"
	tb "gopkg.in/tucnak/telebot.v2"
)

func userEvents(app *config.App, bot *tb.Bot) {

	if eventsHandler(app, bot, &Event{
		Event:      &addAnonMessage,
		UserState:  "add_anon_message",
		Controller: "AddAnonMessageToChannel",
	}) {
		return
	}
}
