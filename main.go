package main

import (
	"fmt"
	"strings"

	"github.com/amiraliio/tgbp/repository"
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
		repository.SaveUserLastState(bot, m.Sender.ID, "register_channel")
		repository.RegisterChannel(bot, m)
	})

	//redirect user from channel to bot for sending message or etc
	bot.Handle(tb.OnText, func(m *tb.Message) {
		if strings.Contains(m.Text, " join_group") {
			repository.SaveUserLastState(bot, m.Sender.ID, "join_group")
			repository.JoinFromChannel(bot, m, inlineKeys)
		}
		lastState := repository.GetUserLastState(bot, m)
		if lastState !=nil{
			switch lastState.State{
			case "new_message_to_group":
				repository.SaveAndSendMessage(bot, m)
			}
		}
	})

	//new message inline message handler
	bot.Handle(&newMessage, func(c *tb.Callback) {
		repository.SaveUserLastState(bot, c.Sender.ID, "new_message_to_group")
		repository.NewMessageHandler(bot, c)
	})

	bot.Start()
}
