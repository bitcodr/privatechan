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

func Init(app *config.App, bot *tb.Bot, state interface{}) {
	if state != nil {
		return
	}

	bot.Handle(tb.OnChannelPost, func(message *tb.Message) {
		generalEventsHandler(app, bot, message, &Event{
			Event:      tb.OnChannelPost,
			UserState:  "register_channel",
			Command:    "/enable_anonymity_support",
			Controller: "RegisterChannel",
		})

	})

	bot.Handle(tb.OnAddedToGroup, func(message *tb.Message) {
		generalEventsHandler(app, bot, message, &Event{
			Event:      tb.OnAddedToGroup,
			UserState:  "register_group",
			Controller: "RegisterGroup",
		})
	})

	bot.Handle(&addAnonMessage, func(message *tb.Message) {
		generalEventsHandler(app, bot, message, &Event{
			Event:      &addAnonMessage,
			UserState:  "add_anon_message",
			Controller: "AddAnonMessageToChannel",
		})
	})

	//////////////////////////////////////////////////
	//on Text handler
	//////////////////////////////////////////////////
	bot.Handle(tb.OnText, func(message *tb.Message) {
		generalEventsHandler(app, bot, message, &Event{
			UserState:  "home",
			Command:    "Home",
			Command1:   "/start",
			Controller: "StartBot",
		})

		generalEventsHandler(app, bot, message, &Event{
			UserState:  "reply_to_message_on_group",
			Command:    "reply_to_message_on_group_",
			Command1:   "/start reply_to_message_on_group_",
			Controller: "SendReply",
		})

		generalEventsHandler(app, bot, message, &Event{
			UserState:  "reply_by_dm_to_user_on_group",
			Command:    "reply_by_dm_to_user_on_group_",
			Command1:   "/start reply_by_dm_to_user_on_group_",
			Controller: "SanedDM",
		})

		generalEventsHandler(app, bot, message, &Event{
			UserState:  "new_message_to_group",
			Command:    "compose_message_in_group_",
			Command1:   "/start compose_message_in_group_",
			Controller: "NewMessageGroupHandler",
		})

		inlineOnTextEventsHandler(app, bot, message, &Event{
			UserState:  "setup_verified_company_account",
			Command:    setupVerifiedCompany.Text,
			Controller: "SetUpCompanyByAdmin",
		})

		inlineOnTextEventsHandler(app, bot, message, &Event{
			UserState:  "new_message_to_group",
			Command:    "compose_message_in_group_",
			Controller: "SaveAndSendMessage",
		})

		inlineOnTextEventsHandler(app, bot, message, &Event{
			UserState:  "reply_to_message_on_group",
			Command:    "reply_to_message_on_group_",
			Controller: "SendAndSaveReplyMessage",
		})

		inlineOnTextEventsHandler(app, bot, message, &Event{
			UserState:  "reply_by_dm_to_user_on_group",
			Command:    "reply_by_dm_to_user_on_group_",
			Controller: "SendAndSaveDirectMessage",
		})

		inlineOnTextEventsHandler(app, bot, message, &Event{
			UserState:  "answer_to_dm",
			Command:    "answer_to_dm_",
			Controller: "SendAnswerAndSaveDirectMessage",
		})

		inlineOnTextEventsHandler(app, bot, message, &Event{
			UserState:  "register_user_with_email",
			Command:    joinCompanyChannels.Text,
			Controller: "RegisterUserWithemail",
		})

		inlineOnTextEventsHandler(app, bot, message, &Event{
			UserState:  "confirm_register_company_email_address",
			Controller: "ConfirmRegisterCompanyRequest",
		})

		inlineOnTextEventsHandler(app, bot, message, &Event{
			UserState:  "register_user_for_the_company",
			Controller: "ConfirmRegisterUserForTheCompany",
		})

		inlineOnTextEventsHandler(app, bot, message, &Event{
			UserState:  "email_for_user_registration",
			Controller: "RegisterUserWithEmailAndCode",
		})
	})

	//////////////////////////////////////////////////
	//CAllback handler
	//////////////////////////////////////////////////
	bot.Handle(tb.OnCallback, func(c *tb.Callback) {
		onCallbackEventsHandler(app, bot, c, &Event{
			UserState:  "answer_to_dm",
			Command:    "answer_to_dm_",
			Controller: "SanedAnswerDM",
		})
		inlineOnCallbackEventsHandler(app, bot, c, &Event{
			UserState:  "home",
			Command:    "Home",
			Command1:   "/start",
			Controller: "StartBotCallback",
		})

		inlineOnCallbackEventsHandler(app, bot, c, &Event{
			UserState:  "setup_verified_company_account",
			Controller: "SetUpCompanyByAdmin",
		})

		inlineOnCallbackEventsHandler(app, bot, c, &Event{
			UserState:  "register_user_with_email",
			Controller: "RegisterUserWithemail",
		})
	})

}

func generalEventsHandler(app *config.App, bot *tb.Bot, message *tb.Message, request *Event) {
	var result bool
	helpers.Invoke(new(BotService), &result, request.Controller, app, bot, message, request)
	if result {
		Init(app, bot, true)
	}
}

func onCallbackEventsHandler(app *config.App, bot *tb.Bot, c *tb.Callback, request *Event) {
	var result bool
	helpers.Invoke(new(BotService), &result, request.Controller, app, bot, c, request)
	if result {
		Init(app, bot, true)
	}
}

func inlineOnCallbackEventsHandler(app *config.App, bot *tb.Bot, c *tb.Callback, request *Event) {
	var result bool
	db := app.DB()
	defer db.Close()
	lastState := GetUserLastState(db, app, bot, c.Message, c.Sender.ID)
	switch {
	case c.Data == request.Command || c.Data == request.Command1:
		helpers.Invoke(new(BotService), &result, request.Controller, app, bot, c, request)
	case lastState.State == request.UserState:
		helpers.Invoke(new(BotService), &result, request.Controller, db, app, bot, c.Message, request, lastState, strings.TrimSpace(c.Data), c.Sender.ID)
	}
	if result {
		Init(app, bot, true)
	}
}

func inlineOnTextEventsHandler(app *config.App, bot *tb.Bot, message *tb.Message, request *Event) {
	var result bool
	db := app.DB()
	defer db.Close()
	lastState := GetUserLastState(db, app, bot, message, message.Sender.ID)
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
	if result {
		Init(app, bot, true)
	}
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
