package controllers

import (
	"database/sql"
	"github.com/amiraliio/tgbp/config"
	"github.com/amiraliio/tgbp/helpers"
	"github.com/amiraliio/tgbp/models"
	tb "gopkg.in/tucnak/telebot.v2"
	"strings"
)

func onCallbackEvents(app *config.App, bot *tb.Bot) {
	bot.Handle(tb.OnCallback, func(c *tb.Callback) {

		//check incoming text
		incomingMessage := c.Data
		switch {
		case incomingMessage == "Home" || incomingMessage == "/start":
			goto StartBotCallback
		case strings.Contains(incomingMessage, "answer_to_dm_"):
			goto SanedAnswerDM
		default:
			goto CheckState
		}

	SanedAnswerDM:
		if onCallbackEventsHandler(app, bot, c, &Event{
			UserState:  "answer_to_dm",
			Command:    "answer_to_dm_",
			Controller: "SanedAnswerDM",
		}) {
			Init(app, bot, true)
		}

	StartBotCallback:
		if onCallbackEventsHandler(app, bot, c, &Event{
			UserState:  "home",
			Command:    "Home",
			Command1:   "/start",
			Controller: "StartBotCallback",
		}) {
			Init(app, bot, true)
		}

		/////////////////////////////////////////////
		////////check the user state////////////////
		///////////////////////////////////////////
	CheckState:
		db := app.DB()
		defer db.Close()
		lastState := GetUserLastState(db, app, bot, c.Message, c.Sender.ID)
		switch lastState.State {
		case "setup_verified_company_account":
			goto SetUpCompanyByAdmin
		case "register_user_with_email":
			goto RegisterUserWithemail
		}

	SetUpCompanyByAdmin:
		if inlineOnCallbackEventsHandler(app, bot, c, db, lastState, &Event{
			UserState:  "setup_verified_company_account",
			Controller: "SetUpCompanyByAdmin",
		}) {
			Init(app, bot, true)
		}

	RegisterUserWithemail:
		if inlineOnCallbackEventsHandler(app, bot, c, db, lastState, &Event{
			UserState:  "register_user_with_email",
			Controller: "RegisterUserWithemail",
		}) {
			Init(app, bot, true)
		}
	})
}

func onCallbackEventsHandler(app *config.App, bot *tb.Bot, c *tb.Callback, request *Event) bool {
	var result bool
	helpers.Invoke(new(BotService), &result, request.Controller, app, bot, c, request)
	return result
}

func inlineOnCallbackEventsHandler(app *config.App, bot *tb.Bot, c *tb.Callback, db *sql.DB, lastState *models.UserLastState, request *Event) bool {
	var result bool
	switch {
	case c.Data == request.Command || c.Data == request.Command1:
		helpers.Invoke(new(BotService), &result, request.Controller, app, bot, c, request)
	case lastState.State == request.UserState:
		helpers.Invoke(new(BotService), &result, request.Controller, db, app, bot, c.Message, request, lastState, strings.TrimSpace(c.Data), c.Sender.ID)
	}
	return result
}
