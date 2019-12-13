package main

import (
	"fmt"
	"strings"

	"github.com/amiraliio/tgbp/controller"
	"github.com/spf13/viper"
	tb "gopkg.in/tucnak/telebot.v2"
	"log"
	"time"
)

func main() {
	//config
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %s", err))
	}
	//init bot
	bot, err := tb.NewBot(tb.Settings{
		Token:  viper.GetString("APP.TELEGRAM_APITOKEN"),
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		log.Fatal(err)
		return
	}

	joinGroup := tb.InlineButton{
		Unique: "join_group",
		Text:   "Join To The Other Company Channels",
	}
	newSurvey := tb.InlineButton{
		Unique: "new_survey_to_group",
		Text:   "New Survey To The Channel",
	}
	newMessage := tb.InlineButton{
		Unique: "new_message_to_group",
		Text:   "New Message To The Channel",
	}

	inlineKeys := [][]tb.InlineButton{
		[]tb.InlineButton{joinGroup},
		[]tb.InlineButton{newMessage},
		[]tb.InlineButton{newSurvey},
	}


	//register a channel with the company name directly from channel
	bot.Handle(tb.OnChannelPost, func(m *tb.Message) {
		if m.Sender != nil {
			controller.SaveUserLastState(bot, m.Text, m.Sender.ID, "register_channel")
		}
		controller.RegisterChannel(bot, m)
	})

	//register a channel with the company name directly from channel
	bot.Handle(tb.OnAddedToGroup, func(m *tb.Message) {
		if m.Sender != nil {
			controller.SaveUserLastState(bot, m.Text, m.Sender.ID, "register_group")
		}
		controller.RegisterGroup(bot, m)
	})

	//redirect user from channel to bot for sending message or etc
	bot.Handle(tb.OnText, func(m *tb.Message) {
		// fmt.Println(m.Text)
		if strings.Contains(m.Text, " join_group") {
			if m.Sender != nil {
				controller.SaveUserLastState(bot, m.Text, m.Sender.ID, "join_group")
			}
			controller.JoinFromChannel(bot, m, inlineKeys)
		}

		if strings.Contains(m.Text, "reply_to_message_on_group_") {
			if m.Sender != nil {
				controller.SaveUserLastState(bot, m.Text, m.Sender.ID, "reply_to_message_on_group")
			}
			ids := strings.TrimPrefix(m.Text, "/start reply_to_message_on_group_")
			data := strings.Split(ids, "_")
			channelID := strings.TrimSpace(data[0])
			controller.JoinFromGroup(bot, m, channelID)
			controller.SendReply(bot, m.Sender)
		}

		if strings.Contains(m.Text, "reply_by_dm_to_user_on_group_") {
			if m.Sender != nil {
				controller.SaveUserLastState(bot, m.Text, m.Sender.ID, "reply_by_dm_to_user_on_group")
			}
			ids := strings.TrimPrefix(m.Text, "/start reply_by_dm_to_user_on_group_")
			data := strings.Split(ids, "_")
			channelID := strings.TrimSpace(data[0])
			controller.JoinFromGroup(bot, m, channelID)
			controller.SanedDM(bot, m.Sender)
		}

		if strings.Contains(m.Text, "compose_message_in_group_") {
			if m.Sender != nil {
				controller.SaveUserLastState(bot, m.Text, m.Sender.ID, "new_message_to_group")
			}
			channelID := strings.ReplaceAll(m.Text, "/start compose_message_in_group_", "")
			controller.JoinFromGroup(bot, m, channelID)
			controller.NewMessageGroupHandler(bot, m.Sender)
		}

		if strings.Contains(m.Text, "more_from_group_") {
			if m.Sender != nil {
				controller.SaveUserLastState(bot, m.Text, m.Sender.ID, "more_from_group")
			}
			ids := strings.TrimPrefix(m.Text, "/start more_from_group_")
			if ids != "" {
				data := strings.Split(ids, "_")
				if len(data) == 3 {
					channelID := strings.TrimSpace(data[0])
					controller.JoinFromGroup(bot, m, channelID)
					controller.ShowGroupAnotherbuttons(bot, m, data)
				}
			}
		}

		lastState := controller.GetUserLastState(bot, m)
		if lastState != nil {
			switch {
			case lastState.State == "new_message_to_group" && !strings.Contains(m.Text, "compose_message_in_group_"):
				controller.SaveAndSendMessage(bot, m)
			case lastState.State == "reply_to_message_on_group" && !strings.Contains(m.Text, "reply_to_message_on_group_"):
				controller.SendAndSaveReplyMessage(bot, m, lastState)
			case lastState.State == "reply_by_dm_to_user_on_group" && !strings.Contains(m.Text, "reply_by_dm_to_user_on_group_"):
				controller.SendAndSaveDirectMessage(bot, m, lastState)
			case lastState.State == "answer_to_dm" && !strings.Contains(m.Text, "answer_to_dm_"):
				controller.SendAnswerAndSaveDirectMessage(bot, m, lastState)
			}
		}
	})

	//new message inline message handler
	bot.Handle(&newMessage, func(c *tb.Callback) {
		if c.Sender != nil {
			controller.SaveUserLastState(bot, c.Message.Text, c.Sender.ID, "new_message_to_group")
		}
		controller.NewMessageHandler(bot, c.Sender)
	})

	bot.Handle(tb.OnCallback, func(c *tb.Callback) {
		if strings.Contains(c.Data, "answer_to_dm_") {
			if c.Sender != nil {
				controller.SaveUserLastState(bot, c.Data, c.Sender.ID, "answer_to_dm")
			}
			controller.SanedAnswerDM(bot, c.Sender)
		}
		if strings.Contains(c.Data, "reply_to_message_on_group_") {
			if c.Sender != nil {

				controller.SaveUserLastState(bot, c.Data, c.Sender.ID, "reply_to_message_on_group")
			}
			ids := strings.TrimPrefix(strings.TrimSpace(c.Data), "reply_to_message_on_group_")
			data := strings.Split(ids, "_")
			channelID := strings.TrimSpace(data[0])
			controller.JoinFromGroup(bot, c.Message, channelID)
			controller.SendReply(bot, c.Sender)
		}

		if strings.Contains(c.Data, "reply_by_dm_to_user_on_group_") {
			if c.Sender != nil {
				controller.SaveUserLastState(bot, c.Data, c.Sender.ID, "reply_by_dm_to_user_on_group")
			}
			ids := strings.TrimPrefix(strings.TrimSpace(c.Data), "reply_by_dm_to_user_on_group_")
			data := strings.Split(ids, "_")
			channelID := strings.TrimSpace(data[0])
			controller.JoinFromGroup(bot, c.Message, channelID)
			controller.SanedDM(bot, c.Sender)
		}

		if strings.Contains(c.Data, "compose_message_in_group_") {
			if c.Sender != nil {
				controller.SaveUserLastState(bot, c.Data, c.Sender.ID, "new_message_to_group")
			}
			channelID := strings.ReplaceAll(strings.TrimSpace(c.Data), "compose_message_in_group_", "")
			controller.JoinFromGroup(bot, c.Message, channelID)
			controller.NewMessageGroupHandler(bot, c.Sender)
		}
	})

	bot.Start()
}
