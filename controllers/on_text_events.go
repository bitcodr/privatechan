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

		//check incoming text
		incomingMessage := message.Text
		switch {
		case incomingMessage == "Home" || incomingMessage == "/start":
			goto StartBot
		case strings.Contains(incomingMessage, "reply_to_message_on_group_"):
			goto SendReply
		case strings.Contains(incomingMessage, "reply_by_dm_to_user_on_group_"):
			goto SanedDM
		case strings.Contains(incomingMessage, "compose_message_in_group_"):
			goto NewMessageGroupHandler
		default:
			goto CheckState
		}

	StartBot:
		if generalEventsHandler(app, bot, message, &Event{
			UserState:  "home",
			Command:    "Home",
			Command1:   "/start",
			Controller: "StartBot",
		}) {
			Init(app, bot, true)
		}

	SendReply:
		if generalEventsHandler(app, bot, message, &Event{
			UserState:  "reply_to_message_on_group",
			Command:    "reply_to_message_on_group_",
			Command1:   "/start reply_to_message_on_group_",
			Controller: "SendReply",
		}) {
			Init(app, bot, true)
		}

	SanedDM:
		if generalEventsHandler(app, bot, message, &Event{
			UserState:  "reply_by_dm_to_user_on_group",
			Command:    "reply_by_dm_to_user_on_group_",
			Command1:   "/start reply_by_dm_to_user_on_group_",
			Controller: "SanedDM",
		}) {
			Init(app, bot, true)
		}

	NewMessageGroupHandler:
		if generalEventsHandler(app, bot, message, &Event{
			UserState:  "new_message_to_group",
			Command:    "compose_message_in_group_",
			Command1:   "/start compose_message_in_group_",
			Controller: "NewMessageGroupHandler",
		}) {
			Init(app, bot, true)
		}

		/////////////////////////////////////////////
		////////check the user state////////////////
		///////////////////////////////////////////
	CheckState:
		db := app.DB()
		defer db.Close()
		lastState := GetUserLastState(db, app, bot, message, message.Sender.ID)
		switch {
		case incomingMessage == setupVerifiedCompany.Text:
			goto SetUpCompanyByAdmin
		case strings.Contains(incomingMessage, "compose_message_in_group_"):
			goto SaveAndSendMessage
		case strings.Contains(incomingMessage, "reply_to_message_on_group_"):
			goto SendAndSaveReplyMessage
		case strings.Contains(incomingMessage, "reply_by_dm_to_user_on_group_"):
			goto SendAndSaveDirectMessage
		case strings.Contains(incomingMessage, "answer_to_dm_"):
			goto SendAnswerAndSaveDirectMessage
		case incomingMessage == joinCompanyChannels.Text:
			goto RegisterUserWithemail
		case lastState.State == "confirm_register_company_email_address":
			goto ConfirmRegisterCompanyRequest
		case lastState.State == "register_user_for_the_company":
			goto ConfirmRegisterUserForTheCompany
		case lastState.State == "email_for_user_registration":
			goto RegisterUserWithEmailAndCode
		}

	SetUpCompanyByAdmin:
		if inlineOnTextEventsHandler(app, bot, message, db, lastState, &Event{
			UserState:  "setup_verified_company_account",
			Command:    setupVerifiedCompany.Text,
			Controller: "SetUpCompanyByAdmin",
		}) {
			Init(app, bot, true)
		}

	SaveAndSendMessage:
		if inlineOnTextEventsHandler(app, bot, message, db, lastState, &Event{
			UserState:  "new_message_to_group",
			Command:    "compose_message_in_group_",
			Controller: "SaveAndSendMessage",
		}) {
			Init(app, bot, true)
		}

	SendAndSaveReplyMessage:
		if inlineOnTextEventsHandler(app, bot, message, db, lastState, &Event{
			UserState:  "reply_to_message_on_group",
			Command:    "reply_to_message_on_group_",
			Controller: "SendAndSaveReplyMessage",
		}) {
			Init(app, bot, true)
		}

	SendAndSaveDirectMessage:
		if inlineOnTextEventsHandler(app, bot, message, db, lastState, &Event{
			UserState:  "reply_by_dm_to_user_on_group",
			Command:    "reply_by_dm_to_user_on_group_",
			Controller: "SendAndSaveDirectMessage",
		}) {
			Init(app, bot, true)
		}

	SendAnswerAndSaveDirectMessage:
		if inlineOnTextEventsHandler(app, bot, message, db, lastState, &Event{
			UserState:  "answer_to_dm",
			Command:    "answer_to_dm_",
			Controller: "SendAnswerAndSaveDirectMessage",
		}) {
			Init(app, bot, true)
		}

	RegisterUserWithemail:
		if inlineOnTextEventsHandler(app, bot, message, db, lastState, &Event{
			UserState:  "register_user_with_email",
			Command:    joinCompanyChannels.Text,
			Controller: "RegisterUserWithemail",
		}) {
			Init(app, bot, true)
		}

	ConfirmRegisterCompanyRequest:
		if inlineOnTextEventsHandler(app, bot, message, db, lastState, &Event{
			UserState:  "confirm_register_company_email_address",
			Controller: "ConfirmRegisterCompanyRequest",
		}) {
			Init(app, bot, true)
		}

	ConfirmRegisterUserForTheCompany:
		if inlineOnTextEventsHandler(app, bot, message, db, lastState, &Event{
			UserState:  "register_user_for_the_company",
			Controller: "ConfirmRegisterUserForTheCompany",
		}) {
			Init(app, bot, true)
		}

	RegisterUserWithEmailAndCode:
		if inlineOnTextEventsHandler(app, bot, message, db, lastState, &Event{
			UserState:  "email_for_user_registration",
			Controller: "RegisterUserWithEmailAndCode",
		}) {
			Init(app, bot, true)
		}

	})
}

func inlineOnTextEventsHandler(app *config.App, bot *tb.Bot, message *tb.Message, db *sql.DB, lastState *models.UserLastState, request *Event) bool {
	var result bool
	switch {
	case lastState.State == request.UserState && !strings.Contains(message.Text, request.Command):
		helpers.Invoke(new(BotService), &result, request.Controller, db, app, bot, message, request, lastState)
	case lastState.State == request.UserState || (request.Command != "" && strings.Contains(message.Text, request.Command)):
		helpers.Invoke(new(BotService), &result, request.Controller, db, app, bot, message, request, lastState, strings.TrimSpace(message.Text), message.Sender.ID)
	case lastState.State == request.UserState && (strings.Contains(message.Text, "No") || strings.Contains(message.Text, "Yes")):
		helpers.Invoke(new(BotService), &result, request.Controller, db, app, bot, message, request, lastState)
	case lastState.State == request.UserState:
		helpers.Invoke(new(BotService), &result, request.Controller, db, app, bot, message, request, lastState)
	}
	return result
}
