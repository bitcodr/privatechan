package controllers

import (
	"database/sql"
	"github.com/amiraliio/tgbp/config"
	"github.com/amiraliio/tgbp/helpers"
	"github.com/amiraliio/tgbp/models"
	tb "gopkg.in/tucnak/telebot.v2"
	"strings"
)

func onTextEvents(app *config.App, bot *tb.Bot) {

	bot.Handle(tb.OnText, func(message *tb.Message) {
		if !message.Private() {
			return
		}

		db := app.DB()
		defer db.Close()
		lastState := GetUserLastState(db, app, bot, message, message.Sender.ID)

		//check incoming text
		incomingMessage := message.Text
		switch {
		case incomingMessage == config.LangConfig.GetString("GENERAL.HOME") || incomingMessage == "/start":
			goto StartBot
		case strings.Contains(incomingMessage, "reply_to_message_on_group_"):
			goto SendReply
		case strings.Contains(incomingMessage, config.LangConfig.GetString("STATE.REPLY_BY_DM")+"_"):
			goto SanedDM
		case strings.Contains(incomingMessage, config.LangConfig.GetString("STATE.COMPOSE_MESSAGE")+"_"):
			goto NewMessageGroupHandler
		default:
			goto CheckState
		}

	StartBot:
		if generalEventsHandler(app, bot, message, &Event{
			UserState:  "home",
			Command:    config.LangConfig.GetString("GENERAL.HOME"),
			Command1:   "/start",
			Controller: "StartBot",
		}) {
			Init(app, bot, true)
		}
		goto END

	SendReply:
		if generalEventsHandler(app, bot, message, &Event{
			UserState:  config.LangConfig.GetString("STATE.REPLY_TO_MESSAGE"),
			Command:    config.LangConfig.GetString("STATE.REPLY_TO_MESSAGE") + "_",
			Command1:   "/start reply_to_message_on_group_",
			Controller: "SendReply",
		}) {
			Init(app, bot, true)
		}
		goto END

	SanedDM:
		if generalEventsHandler(app, bot, message, &Event{
			UserState:  config.LangConfig.GetString("STATE.REPLY_BY_DM"),
			Command:    config.LangConfig.GetString("STATE.REPLY_BY_DM") + "_",
			Command1:   "/start reply_by_dm_to_user_on_group_",
			Controller: "SanedDM",
		}) {
			Init(app, bot, true)
		}
		goto END

	NewMessageGroupHandler:
		if generalEventsHandler(app, bot, message, &Event{
			UserState:  "new_message_to_group",
			Command:    config.LangConfig.GetString("STATE.COMPOSE_MESSAGE") + "_",
			Command1:   "/start compose_message_in_group_",
			Controller: "NewMessageGroupHandler",
		}) {
			Init(app, bot, true)
		}
		goto END

		/////////////////////////////////////////////
		////////check the user state////////////////
		///////////////////////////////////////////
	CheckState:
		switch {
		case lastState.State == config.LangConfig.GetString("STATE.SETUP_VERIFIED_COMPANY") || incomingMessage == setupVerifiedCompany.Text:
			goto SetUpCompanyByAdmin
		case lastState.State == "new_message_to_group" || strings.Contains(incomingMessage, config.LangConfig.GetString("STATE.COMPOSE_MESSAGE")+"_"):
			goto SaveAndSendMessage
		case lastState.State == config.LangConfig.GetString("STATE.REPLY_TO_MESSAGE") || strings.Contains(incomingMessage, "reply_to_message_on_group_"):
			goto SendAndSaveReplyMessage
		case lastState.State == config.LangConfig.GetString("STATE.REPLY_BY_DM") || strings.Contains(incomingMessage, config.LangConfig.GetString("STATE.REPLY_BY_DM")+"_"):
			goto SendAndSaveDirectMessage
		case lastState.State == config.LangConfig.GetString("STATE.ANSWER_TO_DM") || strings.Contains(incomingMessage, config.LangConfig.GetString("STATE.ANSWER_TO_DM")+"_"):
			goto SendAnswerAndSaveDirectMessage
		case lastState.State == "register_user_with_email" || incomingMessage == joinCompanyChannels.Text:
			goto RegisterUserWithemail
		case lastState.State == "confirm_register_company_email_address":
			goto ConfirmRegisterCompanyRequest
		case lastState.State == "register_user_for_the_company":
			goto ConfirmRegisterUserForTheCompany
		case lastState.State == "email_for_user_registration":
			goto RegisterUserWithEmailAndCode
		default:
			goto END
		}

	SetUpCompanyByAdmin:
		if inlineOnTextEventsHandler(app, bot, message, db, lastState, &Event{
			UserState:  config.LangConfig.GetString("STATE.SETUP_VERIFIED_COMPANY"),
			Command:    setupVerifiedCompany.Text,
			Controller: "SetUpCompanyByAdmin",
		}) {
			Init(app, bot, true)
		}
		goto END

	SaveAndSendMessage:
		if inlineOnTextEventsHandler(app, bot, message, db, lastState, &Event{
			UserState:  "new_message_to_group",
			Command:    config.LangConfig.GetString("STATE.COMPOSE_MESSAGE") + "_",
			Controller: "SaveAndSendMessage",
		}) {
			Init(app, bot, true)
		}
		goto END

	SendAndSaveReplyMessage:
		if inlineOnTextEventsHandler(app, bot, message, db, lastState, &Event{
			UserState:  config.LangConfig.GetString("STATE.REPLY_TO_MESSAGE"),
			Command:    config.LangConfig.GetString("STATE.REPLY_TO_MESSAGE") + "_",
			Command1: "/start reply_to_message_on_group_",
			Controller: "SendAndSaveReplyMessage",
		}) {
			Init(app, bot, true)
		}
		goto END

	SendAndSaveDirectMessage:
		if inlineOnTextEventsHandler(app, bot, message, db, lastState, &Event{
			UserState:  config.LangConfig.GetString("STATE.REPLY_BY_DM"),
			Command:    config.LangConfig.GetString("STATE.REPLY_BY_DM") + "_",
			Command1: "/start reply_by_dm_to_user_on_group_",
			Controller: "SendAndSaveDirectMessage",
		}) {
			Init(app, bot, true)
		}
		goto END

	SendAnswerAndSaveDirectMessage:
		if inlineOnTextEventsHandler(app, bot, message, db, lastState, &Event{
			UserState:  config.LangConfig.GetString("STATE.ANSWER_TO_DM"),
			Command:    config.LangConfig.GetString("STATE.ANSWER_TO_DM")+"_",
			Controller: "SendAnswerAndSaveDirectMessage",
		}) {
			Init(app, bot, true)
		}
		goto END

	RegisterUserWithemail:
		if inlineOnTextEventsHandler(app, bot, message, db, lastState, &Event{
			UserState:  "register_user_with_email",
			Command:    joinCompanyChannels.Text,
			Controller: "RegisterUserWithemail",
		}) {
			Init(app, bot, true)
		}
		goto END

	ConfirmRegisterCompanyRequest:
		if inlineOnTextEventsHandler(app, bot, message, db, lastState, &Event{
			UserState:  "confirm_register_company_email_address",
			Controller: "ConfirmRegisterCompanyRequest",
		}) {
			Init(app, bot, true)
		}
		goto END

	ConfirmRegisterUserForTheCompany:
		if inlineOnTextEventsHandler(app, bot, message, db, lastState, &Event{
			UserState:  "register_user_for_the_company",
			Controller: "ConfirmRegisterUserForTheCompany",
		}) {
			Init(app, bot, true)
		}
		goto END

	RegisterUserWithEmailAndCode:
		if inlineOnTextEventsHandler(app, bot, message, db, lastState, &Event{
			UserState:  "email_for_user_registration",
			Controller: "RegisterUserWithEmailAndCode",
		}) {
			Init(app, bot, true)
		}
		goto END

	END:
	})
}

func inlineOnTextEventsHandler(app *config.App, bot *tb.Bot, message *tb.Message, db *sql.DB, lastState *models.UserLastState, request *Event) bool {
	var result bool
	switch {
	case request.Controller == "RegisterUserWithemail" || request.Controller == "SetUpCompanyByAdmin":
		helpers.Invoke(new(BotService), &result, request.Controller, db, app, bot, message, request, lastState, strings.TrimSpace(message.Text), message.Sender.ID)
	default:
		helpers.Invoke(new(BotService), &result, request.Controller, db, app, bot, message, request, lastState)
	}
	return result
}
