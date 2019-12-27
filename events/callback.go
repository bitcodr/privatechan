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

	if eventsHandler(app, bot, &Event{
		Event:      &addAnonMessage,
		UserState:  "add_anon_message",
		Controller: "AddAnonMessageToChannel",
	}) {
		return
	}

	//TODO callback if necessary
	// if c.Data == "Home" || c.Data == "/start" {
	// 		if c.Sender != nil {
	// 			controllers.SaveUserLastState(bot, c.Data, c.Sender.ID, "home")
	// 		}
	// 		controllers.StartBot(bot, c.Message, startBotKeys)
	// 		return
	// 	}

}
