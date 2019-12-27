//Package events ...
package events

import (
	"github.com/amiraliio/tgbp/config"
	tb "gopkg.in/tucnak/telebot.v2"
)

func callbackEvents(app *config.App, bot *tb.Bot) {

	if eventsHandler(app, bot, &Event{
		Event:      tb.OnText,
		UserState:  "home",
		Command:    "Home",
		Command1:   "/start",
		Controller: "StartBot",
	}) {
		return
	}


	
}
