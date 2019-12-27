//Package events ...
package events

import (
	"github.com/amiraliio/tgbp/config"
	"github.com/amiraliio/tgbp/controllers"
	"github.com/amiraliio/tgbp/helpers"
	tb "gopkg.in/tucnak/telebot.v2"
	"log"
	"strconv"
	"strings"
)

func messageEvents(app *config.App, bot *tb.Bot) {

	messageEventsHandler(app, bot, &event{
		event:      tb.OnText,
		userState:  "reply_to_message_on_group",
		command:    "reply_to_message_on_group_",
		command1:   "/start reply_to_message_on_group_",
		controller: "SendReply",
	})

	directMessageEventsHandler(app, bot, &event{
		event:      tb.OnText,
		userState:  "reply_by_dm_to_user_on_group",
		command:    "reply_by_dm_to_user_on_group_",
		command1:   "/start reply_by_dm_to_user_on_group_",
		controller: "SanedDM",
	})

	
}

func messageEventsHandler(app *config.App, bot *tb.Bot, request *event) {
	bot.Handle(request.event, func(message *tb.Message) {
		db := app.DB()
		defer db.Close()
		if strings.Contains(message.Text, request.command) {
			if message.Sender != nil {
				controllers.SaveUserLastState(db, app, bot, message.Text, message.Sender.ID, request.userState)
			}
			ids := strings.TrimPrefix(message.Text, request.command1)
			data := strings.Split(ids, "_")
			channelID := strings.TrimSpace(data[0])
			messageID := strings.TrimSpace(data[2])
			helpers.Invoke(new(controllers.BotService), "JoinFromGroup", db, app, bot, message, channelID)
			helpers.Invoke(new(controllers.BotService), request.controller, db, app, bot, message.Sender, channelID, messageID)
		}
		return
	})
	return
}

func directMessageEventsHandler(app *config.App, bot *tb.Bot, request *event) {
	bot.Handle(request.event, func(message *tb.Message) {
		if strings.Contains(message.Text, request.command) {
			db := app.DB()
			defer db.Close()
			ids := strings.TrimPrefix(message.Text, request.command1)
			data := strings.Split(ids, "_")
			directSenderID, err := strconv.Atoi(data[1])
			if err != nil {
				log.Println(err)
				return
			}
			if message.Sender.ID == directSenderID {
				bot.Send(message.Sender, "You cannot send direct message to your self", controllers.HomeKeyOption(db, app))
				return
			}
			if message.Sender != nil {
				controllers.SaveUserLastState(db, app, bot, message.Text, message.Sender.ID, request.userState)
			}
			channelID := strings.TrimSpace(data[0])
			helpers.Invoke(new(controllers.BotService), "JoinFromGroup", db, app, bot, message, channelID)
			helpers.Invoke(new(controllers.BotService), request.controller, db, app, bot, message.Sender,  directSenderID, channelID)
		}
		return
	})
	return
}
