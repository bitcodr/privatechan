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
