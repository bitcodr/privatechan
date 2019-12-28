package controllers

import (
	"github.com/amiraliio/tgbp/config"
	"github.com/amiraliio/tgbp/helpers"
	tb "gopkg.in/tucnak/telebot.v2"
	"strings"
)

type BotService struct{}

type Event struct {
	UserState, Command, Command1, Controller string
	Event                                    interface{}
}

func Init(app *config.App, bot *tb.Bot) {
	triggersEvents(app, bot)
	keyboardsEvents(app, bot)
	onTextEvents(app, bot)
	onCallbackEvents(app, bot)
}

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
		helpers.Invoke(new(BotService), request.Controller, app, bot, c, request)
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
			helpers.Invoke(new(BotService), request.Controller, app, bot, c, request)
			result = true
		case lastState.State == request.UserState:
			helpers.Invoke(new(BotService), request.Controller, db, app, bot, c.Message, request, lastState, strings.TrimSpace(c.Data), c.Sender.ID)
			result = true
		default:
			result = false
		}
	})
	return result
}

func keyboardsEvents(app *config.App, bot *tb.Bot) {

	if keyboardsEventsHandler(app, bot, &Event{
		Event:      &addAnonMessage,
		UserState:  "add_anon_message",
		Controller: "AddAnonMessageToChannel",
	}) {
		return
	}

}

func keyboardsEventsHandler(app *config.App, bot *tb.Bot, request *Event) (result bool) {
	result = false
	bot.Handle(request.Event, func(message *tb.Message) {
		helpers.Invoke(new(BotService), request.Controller, app, bot, message, request)
		result = true
	})
	return result
}

func onTextEvents(app *config.App, bot *tb.Bot) {

	onTextEventsHandler(app, bot, &Event{
		UserState:  "home",
		Command:    "Home",
		Command1:   "/start",
		Controller: "StartBot",
	})

	return

	onTextEventsHandler(app, bot, &Event{
		UserState:  "reply_to_message_on_group",
		Command:    "reply_to_message_on_group_",
		Command1:   "/start reply_to_message_on_group_",
		Controller: "SendReply",
	})

	onTextEventsHandler(app, bot, &Event{
		UserState:  "reply_by_dm_to_user_on_group",
		Command:    "reply_by_dm_to_user_on_group_",
		Command1:   "/start reply_by_dm_to_user_on_group_",
		Controller: "SanedDM",
	})

	onTextEventsHandler(app, bot, &Event{
		UserState:  "new_message_to_group",
		Command:    "compose_message_in_group_",
		Command1:   "/start compose_message_in_group_",
		Controller: "NewMessageGroupHandler",
	})

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


func onTextEventsHandler(app *config.App, bot *tb.Bot, request *Event) {
	bot.Handle(tb.OnText, func(message *tb.Message) {
		helpers.Invoke(new(BotService), request.Controller, app, bot, message, request)
	})
}

func inlineOnTextEventsHandler(app *config.App, bot *tb.Bot, request *Event) (result bool) {
	result = false
	bot.Handle(tb.OnText, func(message *tb.Message) {
		db := app.DB()
		defer db.Close()
		lastState := GetUserLastState(db, app, bot, message, message.Sender.ID)
		switch {
		case lastState.State == request.UserState && !strings.Contains(message.Text, request.Command):
			helpers.Invoke(new(BotService), request.Controller, db, app, bot, message, request, lastState)
			result = true
		case lastState.State == request.UserState || (request.Command != "" && strings.Contains(message.Text, request.Command)):
			helpers.Invoke(new(BotService), request.Controller, db, app, bot, message, request, lastState, strings.TrimSpace(message.Text), message.Sender.ID)
			result = true
		case lastState.State == request.UserState && (strings.Contains(message.Text, "No") || strings.Contains(message.Text, "Yes")):
			helpers.Invoke(new(BotService), request.Controller, db, app, bot, message, request, lastState)
			result = true
		case lastState.State == request.UserState:
			helpers.Invoke(new(BotService), request.Controller, db, app, bot, message, request, lastState)
			result = true
		default:
			result = false
		}
	})
	return result
}

func triggersEvents(app *config.App, bot *tb.Bot) {

	if triggersEventsHandler(app, bot, &Event{
		Event:      tb.OnChannelPost,
		UserState:  "register_channel",
		Command:    "/enable_anonymity_support",
		Controller: "RegisterChannel",
	}) {
		return
	}

	if triggersEventsHandler(app, bot, &Event{
		Event:      tb.OnAddedToGroup,
		UserState:  "register_group",
		Controller: "RegisterGroup",
	}) {
		return
	}
}

func triggersEventsHandler(app *config.App, bot *tb.Bot, request *Event) (result bool) {
	result = false
	bot.Handle(request.Event, func(message *tb.Message) {
		helpers.Invoke(new(BotService), request.Controller, app, bot, message, request)
		result = true
	})
	return result
}

//TODO needs keyboard refactoring
//bot startup buttons
var addAnonMessage = tb.ReplyButton{
	Text: "Add Anonymous Message to a Channel/Group",
}
var setupVerifiedCompany = tb.ReplyButton{
	Text: "Setup Verified Company Account",
}
var joinCompanyChannels = tb.ReplyButton{
	Text: "Join To Company Anonymous Channel/Group",
}
var StartBotKeys = [][]tb.ReplyButton{
	[]tb.ReplyButton{addAnonMessage},
	[]tb.ReplyButton{setupVerifiedCompany},
	[]tb.ReplyButton{joinCompanyChannels},
}
