package main

import (
	"github.com/amiraliio/tgbp/controllers"
	"github.com/amiraliio/tgbp/events"
	"os"
	"strings"

	"github.com/amiraliio/tgbp/config"
	tb "gopkg.in/tucnak/telebot.v2"
	"log"
)

func main() {

	//get current directory
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}

	//initial app config
	app := new(config.App)
	app.ProjectDir = currentDir

	//initial environment variables
	app.Environment()

	//set other configs
	app = app.SetOtherConfigs()

	//init bot
	bot := app.Bot()

	//handle bot events
	events.Init(app, bot)

	botService := new(controllers.BotService)

	//on text handlers
	bot.Handle(tb.OnText, func(m *tb.Message) {
		lastState := botService.GetUserLastState(bot, m, m.Sender.ID)

		if strings.Contains(m.Text, "compose_message_in_group_") {
			botService.CheckUserRegisteredOrNot(bot, m, lastState, m.Text, m.Sender.ID)
			if m.Sender != nil {
				controllers.SaveUserLastState(bot, m.Text, m.Sender.ID, "new_message_to_group")
			}
			channelID := strings.ReplaceAll(m.Text, "/start compose_message_in_group_", "")
			botService.JoinFromGroup(bot, m, channelID)
			botService.NewMessageGroupHandler(bot, m.Sender, channelID)
			return
		}

		switch {
		case lastState.State == "new_message_to_group" && !strings.Contains(m.Text, "compose_message_in_group_"):
			botService.SaveAndSendMessage(bot, m)
			return
		case lastState.State == "reply_to_message_on_group" && !strings.Contains(m.Text, "reply_to_message_on_group_"):
			botService.SendAndSaveReplyMessage(bot, m, lastState)
			return
		case lastState.State == "reply_by_dm_to_user_on_group" && !strings.Contains(m.Text, "reply_by_dm_to_user_on_group_"):
			botService.SendAndSaveDirectMessage(bot, m, lastState)
			return
		case lastState.State == "answer_to_dm" && !strings.Contains(m.Text, "answer_to_dm_"):
			botService.SendAnswerAndSaveDirectMessage(bot, m, lastState)
			return
		case lastState.State == "setup_verified_company_account" || strings.Contains(m.Text, setupVerifiedCompany.Text):
			botService.SetUpCompanyByAdmin(bot, m, lastState, m.Text, m.Sender.ID)
			return
		case lastState.State == "register_user_with_email" || strings.Contains(m.Text, joinCompanyChannels.Text):
			botService.RegisterUserWithemail(bot, m, lastState, strings.TrimSpace(m.Text), m.Sender.ID)
			return
		case lastState.State == "confirm_register_company_email_address" && (strings.Contains(m.Text, "No") || strings.Contains(m.Text, "Yes")):
			botService.ConfirmRegisterCompanyRequest(bot, m, lastState, strings.TrimSpace(m.Text), m.Sender.ID)
			return
		case lastState.State == "register_user_for_the_company" && (strings.Contains(m.Text, "No") || strings.Contains(m.Text, "Yes")):
			botService.ConfirmRegisterUserForTheCompany(bot, m, lastState, strings.TrimSpace(m.Text), m.Sender.ID)
			return
		case lastState.State == "email_for_user_registration":
			botService.RegisterUserWithEmail(bot, m, lastState, strings.TrimSpace(m.Text), m.Sender.ID)
			return
		}
		return
	})

	//callback handlers
	bot.Handle(tb.OnCallback, func(c *tb.Callback) {
		if c.Data == "Home" || c.Data == "/start" {
			if c.Sender != nil {
				controllers.SaveUserLastState(bot, c.Data, c.Sender.ID, "home")
			}
			controllers.StartBot(bot, c.Message, startBotKeys)
			return
		}
		if strings.Contains(c.Data, "answer_to_dm_") {
			if c.Sender != nil {
				controllers.SaveUserLastState(bot, c.Data, c.Sender.ID, "answer_to_dm")
			}
			controllers.SanedAnswerDM(bot, c.Sender)
			return
		}
		lastState := controllers.GetUserLastState(bot, c.Message, c.Sender.ID)
		switch {
		case lastState.State == "setup_verified_company_account":
			controllers.SetUpCompanyByAdmin(bot, c.Message, lastState, c.Data, c.Sender.ID)
		case lastState.State == "register_user_with_email":
			controllers.RegisterUserWithemail(bot, c.Message, lastState, strings.TrimSpace(c.Data), c.Sender.ID)
		}
		return
	})

	bot.Start()
}
