//Package events ...
package events

import (
	"github.com/amiraliio/tgbp/config"
	tb "gopkg.in/tucnak/telebot.v2"
)

func adminEvents(app *config.App, bot *tb.Bot) {

	if inlineEventsHandler(app, bot, &Event{
		Event:      tb.OnText,
		UserState:  "setup_verified_company_account",
		Command:    setupVerifiedCompany.Text,
		Controller: "SetUpCompanyByAdmin",
	}) {
		return
	}

	if inlineCallbackEventsHandler(app, bot, &Event{
		Event:      tb.OnCallback,
		UserState:  "setup_verified_company_account",
		Controller: "SetUpCompanyByAdmin",
	}) {
		return
	}

}
