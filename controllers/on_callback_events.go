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

		db := app.DB()
		defer db.Close()
		lastState := GetUserLastState(db, app, bot, c.Message, c.Sender.ID)

		//check incoming text
		incomingMessage := c.Data
		switch {
		case incomingMessage == config.LangConfig.GetString("GENERAL.HOME") || incomingMessage == "/start":
			goto StartBotCallback
		case strings.Contains(incomingMessage, config.LangConfig.GetString("STATE.ANSWER_TO_DM")+"_"):
			goto SanedAnswerDM
		default:
			goto CheckState
		}

	SanedAnswerDM:
		if onCallbackEventsHandler(app, bot, c, &Event{
			UserState:  config.LangConfig.GetString("STATE.ANSWER_TO_DM"),
			Command:    config.LangConfig.GetString("STATE.ANSWER_TO_DM")+"_",
			Controller: "SanedAnswerDM",
		}) {
			Init(app, bot, true)
		}
		goto END

	StartBotCallback:
		if onCallbackEventsHandler(app, bot, c, &Event{
			UserState:  "home",
			Command:    config.LangConfig.GetString("GENERAL.HOME"),
			Command1:   "/start",
			Controller: "StartBotCallback",
		}) {
			Init(app, bot, true)
		}
		goto END

		/////////////////////////////////////////////
		////////check the user state////////////////
		///////////////////////////////////////////
	CheckState:
		switch lastState.State {
		case config.LangConfig.GetString("STATE.SETUP_VERIFIED_COMPANY"):
			goto SetUpCompanyByAdmin
		case "register_user_with_email":
			goto RegisterUserWithemail
		default:
			goto END
		}

	SetUpCompanyByAdmin:
		if inlineOnCallbackEventsHandler(app, bot, c, db, lastState, &Event{
			UserState:  config.LangConfig.GetString("STATE.SETUP_VERIFIED_COMPANY"),
			Controller: "SetUpCompanyByAdmin",
		}) {
			Init(app, bot, true)
		}
		goto END

	RegisterUserWithemail:
		if inlineOnCallbackEventsHandler(app, bot, c, db, lastState, &Event{
			UserState:  "register_user_with_email",
			Controller: "RegisterUserWithemail",
		}) {
			Init(app, bot, true)
		}
		goto END

	END:
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
	case request.Controller == "RegisterUserWithemail" || request.Controller == "SetUpCompanyByAdmin":
		helpers.Invoke(new(BotService), &result, request.Controller, db, app, bot, c.Message, request, lastState, strings.TrimSpace(c.Data), c.Sender.ID)
	default:
		helpers.Invoke(new(BotService), &result, request.Controller, app, bot, c, request)
	}
	return result
}
