//Package events ...
package events

import (
	"github.com/amiraliio/tgbp/config"
	tb "gopkg.in/tucnak/telebot.v2"
)

func groupEvents(app *config.App, bot *tb.Bot) {

	eventsHandler(app, bot, &Event{
		Event:      tb.OnAddedToGroup,
		UserState:  "register_group",
		Controller: "RegisterGroup",
	})

}
