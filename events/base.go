package events

import (
	"github.com/amiraliio/tgbp/config"
	"github.com/amiraliio/tgbp/controllers"
	"github.com/amiraliio/tgbp/helpers"
	tb "gopkg.in/tucnak/telebot.v2"
	"strings"
)

type Event struct {
	UserState, Command, Command1, Controller string
	Event                                    interface{}
	Args                                     []interface{}
}

func Init(app *config.App, bot *tb.Bot) {
	channelEvents(app, bot)
	callbackEvents(app, bot)
	userEvents(app, bot)
	groupEvents(app, bot)
	adminEvents(app, bot)
}

func eventsHandler(app *config.App, bot *tb.Bot, request *Event) bool {
	bot.Handle(request.Event, func(message *tb.Message) bool {
		return helpers.Invoke(new(controllers.BotService), request.Controller, app, bot, message, request)
	})
	return false
}

func inlineEventsHandler(app *config.App, bot *tb.Bot, request *Event) bool {
	bot.Handle(request.Event, func(message *tb.Message) bool {
		db := app.DB()
		defer db.Close()
		botService := new(controllers.BotService)
		lastState := botService.GetUserLastState(db, app, bot, message, message.Sender.ID)
		switch {
		case lastState.State == request.UserState && !strings.Contains(message.Text, request.Command):
			return helpers.Invoke(new(controllers.BotService), request.Controller, db, app, bot, message, request, lastState)
		case lastState.State == request.UserState || strings.Contains(message.Text, request.Command):
			return helpers.Invoke(new(controllers.BotService), request.Controller, db, app, bot, message, request, lastState, strings.TrimSpace(message.Text), message.Sender.ID)
		case lastState.State == request.UserState && (strings.Contains(message.Text, "No") || strings.Contains(message.Text, "Yes")):
			return helpers.Invoke(new(controllers.BotService), request.Controller, db, app, bot, message, request, lastState)
		case lastState.State == request.UserState:
			return helpers.Invoke(new(controllers.BotService), request.Controller, db, app, bot, message, request, lastState)
		default:
			return false
		}
	})
	return false
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
