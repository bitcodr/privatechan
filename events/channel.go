//Package events ...
package events

import (
	"github.com/amiraliio/tgbp/config"
	tb "gopkg.in/tucnak/telebot.v2"
)

func channelEvents(app *config.App, bot *tb.Bot) {

	eventsHandler(app, bot, &Event{
		Event:      tb.OnChannelPost,
		UserState:  "register_channel",
		Command:    "/enable_anonymity_support",
		Controller: "RegisterChannel",
	})
}
