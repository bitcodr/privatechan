package events

import (
	"database/sql"
	"github.com/amiraliio/tgbp/config"
	"github.com/amiraliio/tgbp/helpers"
	"github.com/amiraliio/tgbp/models"
	tb "gopkg.in/tucnak/telebot.v2"
	"log"
	"strings"
)

type BotService struct{}

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
	messageEvents(app, bot)
}

func eventsHandler(app *config.App, bot *tb.Bot, request *Event) bool {
	bot.Handle(request.Event, func(message *tb.Message) bool {
		return helpers.Invoke(new(BotService), request.Controller, app, bot, message, request)
	})
	return false
}

func callbackEventsHandler(app *config.App, bot *tb.Bot, request *Event) bool {
	bot.Handle(request.Event, func(c *tb.Callback) bool {
		return helpers.Invoke(new(BotService), request.Controller, app, bot, c, request)
	})
	return false
}

func inlineCallbackEventsHandler(app *config.App, bot *tb.Bot, request *Event) bool {
	bot.Handle(request.Event, func(c *tb.Callback) bool {
		db := app.DB()
		defer db.Close()
		lastState := GetUserLastState(db, app, bot, c.Message, c.Sender.ID)
		switch {
		case c.Data == request.Command || c.Data == request.Command1:
			return helpers.Invoke(new(BotService), request.Controller, app, bot, c, request)
		case lastState.State == request.UserState:
			return helpers.Invoke(new(BotService), request.Controller, db, app, bot, c.Message, request, lastState, strings.TrimSpace(c.Data), c.Sender.ID)
		default:
			return false
		}
	})
	return false
}

func inlineEventsHandler(app *config.App, bot *tb.Bot, request *Event) bool {
	bot.Handle(request.Event, func(message *tb.Message) bool {
		db := app.DB()
		defer db.Close()
		lastState := GetUserLastState(db, app, bot, message, message.Sender.ID)
		switch {
		case lastState.State == request.UserState && !strings.Contains(message.Text, request.Command):
			return helpers.Invoke(new(BotService), request.Controller, db, app, bot, message, request, lastState)
		case lastState.State == request.UserState || strings.Contains(message.Text, request.Command):
			return helpers.Invoke(new(BotService), request.Controller, db, app, bot, message, request, lastState, strings.TrimSpace(message.Text), message.Sender.ID)
		case lastState.State == request.UserState && (strings.Contains(message.Text, "No") || strings.Contains(message.Text, "Yes")):
			return helpers.Invoke(new(BotService), request.Controller, db, app, bot, message, request, lastState)
		case lastState.State == request.UserState:
			return helpers.Invoke(new(BotService), request.Controller, db, app, bot, message, request, lastState)
		default:
			return false
		}
	})
	return false
}

func GetUserLastState(db *sql.DB, app *config.App, bot *tb.Bot, m *tb.Message, user int) *models.UserLastState {
	userLastStateQueryStatement, err := db.Prepare("SELECT `data`,`state`,`userID` from `users_last_state` where `userId`=? order by `createdAt` DESC limit 1")
	if err != nil {
		log.Println(err)
	}
	defer userLastStateQueryStatement.Close()
	userLastStateQuery, err := userLastStateQueryStatement.Query(user)
	if err != nil {
		log.Println(err)
	}
	userLastState := new(models.UserLastState)
	if userLastStateQuery.Next() {
		if err := userLastStateQuery.Scan(&userLastState.Data, &userLastState.State, &userLastState.UserID); err != nil {
			log.Println(err)
		}
		return userLastState
	}
	return userLastState
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
