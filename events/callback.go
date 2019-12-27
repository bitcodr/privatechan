//Package events ...
package events

import (
	"strings"
	"github.com/amiraliio/tgbp/config"
	"github.com/amiraliio/tgbp/helpers"
	tb "gopkg.in/tucnak/telebot.v2"
)

func onCallbackEvents(app *config.App, bot *tb.Bot) {

	if inlineOnCallbackEventsHandler(app, bot, &Event{
		UserState:  "home",
		Command:    "Home",
		Command1:   "/start",
		Controller: "StartBotCallback",
	}) {
		return
	}

	if inlineOnCallbackEventsHandler(app, bot, &Event{
		UserState:  "setup_verified_company_account",
		Controller: "SetUpCompanyByAdmin",
	}) {
		return
	}

	if inlineOnCallbackEventsHandler(app, bot, &Event{
		UserState:  "register_user_with_email",
		Controller: "RegisterUserWithemail",
	}) {
		return
	}

	if onCallbackEventsHandler(app, bot, &Event{
		UserState:  "answer_to_dm",
		Command:    "answer_to_dm_",
		Controller: "SanedAnswerDM",
	}) {
		return
	}

}

func onCallbackEventsHandler(app *config.App, bot *tb.Bot, request *Event) bool {
	bot.Handle(tb.OnCallback, func(c *tb.Callback) bool {
		return helpers.Invoke(new(BotService), request.Controller, app, bot, c, request)
	})
	return false
}

func inlineOnCallbackEventsHandler(app *config.App, bot *tb.Bot, request *Event) bool {
	bot.Handle(tb.OnCallback, func(c *tb.Callback) bool {
		db := app.DB()
		defer db.Close()
		lastState := GetUserLastState(db, app, bot, c.Message, c.Sender.ID)
		switch {
		case c.Data == request.Command || c.Data == request.Command1:
			return helpers.Invoke(new(BotService), request.Controller, app, bot, c, request)
		case lastState.State == request.UserState:
			return helpers.Invoke(new(BotService), request.Controller, db, app, bot, c.Message, request, lastState, strings.TrimSpace(c.Data), c.Sender.ID)
		default:
			return false
		}
	})
	return false
}
