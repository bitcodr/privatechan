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

	//register a channel with the company name directly from channel
	bot.Handle(tb.OnChannelPost, func(m *tb.Message) {
		repository.RegisterChannel(bot, m)
	})

	bot.Handle(tb.OnText, func(m *tb.Message) {
		//TODO check user if registered then show keyboard
		//TODO check user is not registered show verification email
		if strings.Contains(m.Text, " join_group") {
			repository.JoinFromChannel(bot, m)
		}
	})

	bot.Handle("/join_group", func(m *tb.Message) {
		_, _ = bot.Send(m.Sender, "hello world")
	})

	bot.Handle("/new_group", func(m *tb.Message) {
		_, _ = bot.Send(m.Sender, "hello world")
	})

	bot.Handle("/new_message_to_group", func(m *tb.Message) {
		_, _ = bot.Send(m.Sender, "hello world")
	})

	bot.Handle("/new_survey_to_group", func(m *tb.Message) {
		_, _ = bot.Send(m.Sender, "hello world")
	})

	bot.Handle("/reply_to_message_on_group", func(m *tb.Message) {
		_, _ = bot.Send(m.Sender, "hello world")
	})

	bot.Handle("/reply_by_dm_to_user_on_group", func(m *tb.Message) {
		_, _ = bot.Send(m.Sender, "hello world")
	})

	bot.Start()
}
