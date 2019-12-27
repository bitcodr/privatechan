package main

import (
	"log"
	"os"

	"github.com/amiraliio/tgbp/config"
	"github.com/amiraliio/tgbp/events"
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

	//start the bot
	bot.Start()
}
