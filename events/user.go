//Package events ...
package events

import (
	"github.com/amiraliio/tgbp/config"
	tb "gopkg.in/tucnak/telebot.v2"
)

func userEvents(app *config.App, bot *tb.Bot) {

	if inlineEventsHandler(app, bot, &Event{
		Event:      tb.OnText,
		UserState:  "register_user_with_email",
		Command:    joinCompanyChannels.Text,
		Controller: "RegisterUserWithemail",
	}) {
		return
	}

	if inlineEventsHandler(app, bot, &Event{
		Event:      tb.OnText,
		UserState:  "confirm_register_company_email_address",
		Controller: "ConfirmRegisterCompanyRequest",
	}) {
		return
	}

	if inlineEventsHandler(app, bot, &Event{
		Event:      tb.OnText,
		UserState:  "register_user_for_the_company",
		Controller: "ConfirmRegisterUserForTheCompany",
	}) {
		return
	}

	if inlineEventsHandler(app, bot, &Event{
		Event:      tb.OnText,
		UserState:  "email_for_user_registration",
		Controller: "RegisterUserWithEmailAndCode",
	}) {
		return
	}

	if inlineCallbackEventsHandler(app, bot, &Event{
		Event:      tb.OnCallback,
		UserState:  "register_user_with_email",
		Controller: "RegisterUserWithemail",
	}) {
		return
	}
}
