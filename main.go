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

	//bot startup buttons
	addAnonMessage := tb.ReplyButton{
		Text: "Add Anonymous Message to a Channel/Group",
	}
	setupVerifiedCompany := tb.ReplyButton{
		Text: "Setup Verified Company Account",
	}
	joinCompanyChannels := tb.ReplyButton{
		Text: "Join To Company Anonymous Channel/Group",
	}
	startBotKeys := [][]tb.ReplyButton{
		[]tb.ReplyButton{addAnonMessage},
		[]tb.ReplyButton{setupVerifiedCompany},
		[]tb.ReplyButton{joinCompanyChannels},
	}

	//register a channel with the company name directly from channel
	bot.Handle(tb.OnChannelPost, func(m *tb.Message) {
		if m.Text == strings.TrimSpace("/enable_anonymity_support") {
			if m.Sender != nil {
				controller.SaveUserLastState(bot, m.Text, m.Sender.ID, "register_channel")
			}
			controller.RegisterChannel(bot, m)
		}
	})

	//register a channel with the company name directly from channel
	bot.Handle(tb.OnAddedToGroup, func(m *tb.Message) {
		if m.Sender != nil {
			controller.SaveUserLastState(bot, m.Text, m.Sender.ID, "register_group")
		}
		controller.RegisterGroup(bot, m)
	})

	//add anonymous message
	bot.Handle(&addAnonMessage, func(m *tb.Message) {
		if m.Sender != nil {
			controller.SaveUserLastState(bot, m.Text, m.Sender.ID, "add_anon_message")
		}
		controller.AddAnonMessageToChannel(bot, m.Sender)
	})

	//on text handlers
	bot.Handle(tb.OnText, func(m *tb.Message) {

		if m.Text == "Home" || m.Text == "/start" {
			if m.Sender != nil {
				controller.SaveUserLastState(bot, m.Text, m.Sender.ID, "home")
			}
			controller.StartBot(bot, m, startBotKeys)
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
		lastState := controller.GetUserLastState(bot, m, m.Sender.ID)
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
			case lastState.State == "setup_verified_company_account" || strings.Contains(m.Text, setupVerifiedCompany.Text):
				controller.SetUpCompanyByAdmin(bot, m, lastState, m.Text, m.Sender.ID)
			case lastState.State == "register_user_with_email" || strings.Contains(m.Text, joinCompanyChannels.Text):
				controller.RegisterUserWithemail(bot, m, lastState, strings.TrimSpace(m.Text), m.Sender.ID)
			case lastState.State == "confirm_register_company_email_address" && (strings.Contains(m.Text, "No") || strings.Contains(m.Text, "Yes")):
				controller.ConfirmRegisterCompanyRequest(bot, m, lastState, strings.TrimSpace(m.Text), m.Sender.ID)
			case lastState.State == "register_user_for_the_company" && (strings.Contains(m.Text, "No") || strings.Contains(m.Text, "Yes")):
				controller.ConfirmRegisterUserForTheCompany(bot, m, lastState, strings.TrimSpace(m.Text), m.Sender.ID)
			}
		}
	})

	//callback handlers
	bot.Handle(tb.OnCallback, func(c *tb.Callback) {
		if c.Data == "Home" || c.Data == "/start" {
			if c.Sender != nil {
				controller.SaveUserLastState(bot, c.Data, c.Sender.ID, "home")
			}
			controller.StartBot(bot, c.Message, startBotKeys)
		}
		if strings.Contains(c.Data, "answer_to_dm_") {
			if c.Sender != nil {
				controller.SaveUserLastState(bot, c.Data, c.Sender.ID, "answer_to_dm")
			}
			controller.SanedAnswerDM(bot, c.Sender)
		}
		lastState := controller.GetUserLastState(bot, c.Message, c.Sender.ID)
		if lastState != nil {
			switch {
			case lastState.State == "setup_verified_company_account":
				controller.SetUpCompanyByAdmin(bot, c.Message, lastState, c.Data, c.Sender.ID)
			case lastState.State == "register_user_with_email":
				controller.RegisterUserWithemail(bot, c.Message, lastState, strings.TrimSpace(c.Data), c.Sender.ID)
			}
		}
	})

	bot.Start()
}
