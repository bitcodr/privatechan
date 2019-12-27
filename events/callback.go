//Package events ...
package events

import (
	"github.com/amiraliio/tgbp/config"
	"github.com/amiraliio/tgbp/helpers"
	tb "gopkg.in/tucnak/telebot.v2"
	"strings"
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

func onCallbackEventsHandler(app *config.App, bot *tb.Bot, request *Event) (result bool) {
	result = false
	bot.Handle(tb.OnCallback, func(c *tb.Callback) {
		helpers.Invoke(BotService{}, request.Controller, app, bot, c, request)
		result = true
	})
	return result
}

func inlineOnCallbackEventsHandler(app *config.App, bot *tb.Bot, request *Event) (result bool) {
	result = false
	bot.Handle(tb.OnCallback, func(c *tb.Callback) {
		db := app.DB()
		defer db.Close()
		lastState := GetUserLastState(db, app, bot, c.Message, c.Sender.ID)
		switch {
		case c.Data == request.Command || c.Data == request.Command1:
			helpers.Invoke(BotService{}, request.Controller, app, bot, c, request)
			result = true
		case lastState.State == request.UserState:
			helpers.Invoke(BotService{}, request.Controller, db, app, bot, c.Message, request, lastState, strings.TrimSpace(c.Data), c.Sender.ID)
			result = true
		default:
			result = false
		}
	})
	return result
}
