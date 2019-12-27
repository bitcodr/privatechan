package events

import (
	"github.com/amiraliio/tgbp/config"
	"github.com/amiraliio/tgbp/controllers"
	"github.com/amiraliio/tgbp/helpers"
	tb "gopkg.in/tucnak/telebot.v2"
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

func eventsHandler(app *config.App, bot *tb.Bot, request *Event) {
	bot.Handle(request.Event, func(message *tb.Message) {
		helpers.Invoke(new(controllers.BotService), request.Controller, app, bot, message, request)
		return
	})
	return
}

//TODO need refactor
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
