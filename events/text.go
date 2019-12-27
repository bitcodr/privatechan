//Package events ...
package events

import (
	"github.com/amiraliio/tgbp/config"
	"github.com/amiraliio/tgbp/helpers"
	tb "gopkg.in/tucnak/telebot.v2"
	"strings"
)

func onTextEvents(app *config.App, bot *tb.Bot) {

	if onTextEventsHandler(app, bot, &Event{
		UserState:  "home",
		Command:    "Home",
		Command1:   "/start",
		Controller: "StartBot",
	}) {
		return
	}

	if onTextEventsHandler(app, bot, &Event{
		UserState:  "reply_to_message_on_group",
		Command:    "reply_to_message_on_group_",
		Command1:   "/start reply_to_message_on_group_",
		Controller: "SendReply",
	}) {
		return
	}

	if onTextEventsHandler(app, bot, &Event{
		UserState:  "reply_by_dm_to_user_on_group",
		Command:    "reply_by_dm_to_user_on_group_",
		Command1:   "/start reply_by_dm_to_user_on_group_",
		Controller: "SanedDM",
	}) {
		return
	}

	if onTextEventsHandler(app, bot, &Event{
		UserState:  "new_message_to_group",
		Command:    "compose_message_in_group_",
		Command1:   "/start compose_message_in_group_",
		Controller: "NewMessageGroupHandler",
	}) {
		return
	}

	if inlineOnTextEventsHandler(app, bot, &Event{
		UserState:  "setup_verified_company_account",
		Command:    setupVerifiedCompany.Text,
		Controller: "SetUpCompanyByAdmin",
	}) {
		return
	}

	if inlineOnTextEventsHandler(app, bot, &Event{
		UserState:  "new_message_to_group",
		Command:    "compose_message_in_group_",
		Controller: "SaveAndSendMessage",
	}) {
		return
	}

	if inlineOnTextEventsHandler(app, bot, &Event{
		UserState:  "reply_to_message_on_group",
		Command:    "reply_to_message_on_group_",
		Controller: "SendAndSaveReplyMessage",
	}) {
		return
	}

	if inlineOnTextEventsHandler(app, bot, &Event{
		UserState:  "reply_by_dm_to_user_on_group",
		Command:    "reply_by_dm_to_user_on_group_",
		Controller: "SendAndSaveDirectMessage",
	}) {
		return
	}

	if inlineOnTextEventsHandler(app, bot, &Event{
		UserState:  "answer_to_dm",
		Command:    "answer_to_dm_",
		Controller: "SendAnswerAndSaveDirectMessage",
	}) {
		return
	}

	if inlineOnTextEventsHandler(app, bot, &Event{
		UserState:  "register_user_with_email",
		Command:    joinCompanyChannels.Text,
		Controller: "RegisterUserWithemail",
	}) {
		return
	}

	if inlineOnTextEventsHandler(app, bot, &Event{
		UserState:  "confirm_register_company_email_address",
		Controller: "ConfirmRegisterCompanyRequest",
	}) {
		return
	}

	if inlineOnTextEventsHandler(app, bot, &Event{
		UserState:  "register_user_for_the_company",
		Controller: "ConfirmRegisterUserForTheCompany",
	}) {
		return
	}

	if inlineOnTextEventsHandler(app, bot, &Event{
		UserState:  "email_for_user_registration",
		Controller: "RegisterUserWithEmailAndCode",
	}) {
		return
	}
}

func onTextEventsHandler(app *config.App, bot *tb.Bot, request *Event) (result bool) {
	result = false
	bot.Handle(tb.OnText, func(message *tb.Message) {
		helpers.Invoke(BotService{}, request.Controller, app, bot, message, request)
		result = true
	})
	return result
}

func inlineOnTextEventsHandler(app *config.App, bot *tb.Bot, request *Event) (result bool) {
	result = false
	bot.Handle(tb.OnText, func(message *tb.Message) {
		db := app.DB()
		defer db.Close()
		lastState := GetUserLastState(db, app, bot, message, message.Sender.ID)
		switch {
		case lastState.State == request.UserState && !strings.Contains(message.Text, request.Command):
			helpers.Invoke(BotService{}, request.Controller, db, app, bot, message, request, lastState)
			result = true
		case lastState.State == request.UserState || strings.Contains(message.Text, request.Command):
			helpers.Invoke(BotService{}, request.Controller, db, app, bot, message, request, lastState, strings.TrimSpace(message.Text), message.Sender.ID)
			result = true
		case lastState.State == request.UserState && (strings.Contains(message.Text, "No") || strings.Contains(message.Text, "Yes")):
			helpers.Invoke(BotService{}, request.Controller, db, app, bot, message, request, lastState)
			result = true
		case lastState.State == request.UserState:
			helpers.Invoke(BotService{}, request.Controller, db, app, bot, message, request, lastState)
			result = true
		default:
			result = false
		}
	})
	return result
}
