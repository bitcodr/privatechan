//Package events ...
package events

import (
	"github.com/amiraliio/tgbp/config"
	tb "gopkg.in/tucnak/telebot.v2"
)

func messageEvents(app *config.App, bot *tb.Bot) {

	if eventsHandler(app, bot, &Event{
		Event:      tb.OnText,
		UserState:  "reply_to_message_on_group",
		Command:    "reply_to_message_on_group_",
		Command1:   "/start reply_to_message_on_group_",
		Controller: "SendReply",
	}) {
		return
	}

	if eventsHandler(app, bot, &Event{
		Event:      tb.OnText,
		UserState:  "reply_by_dm_to_user_on_group",
		Command:    "reply_by_dm_to_user_on_group_",
		Command1:   "/start reply_by_dm_to_user_on_group_",
		Controller: "SanedDM",
	}) {
		return
	}

	if eventsHandler(app, bot, &Event{
		Event:      tb.OnText,
		UserState:  "new_message_to_group",
		Command:    "compose_message_in_group_",
		Command1:   "/start compose_message_in_group_",
		Controller: "NewMessageGroupHandler",
	}) {
		return
	}

	if eventsHandler(app, bot, &Event{
		Event:      tb.OnText,
		UserState:  "new_message_to_group",
		Command:    "compose_message_in_group_",
		Controller: "SaveAndSendMessage",
	}) {
		return
	}

}
