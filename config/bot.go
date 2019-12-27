package config

import (
	tb "gopkg.in/tucnak/telebot.v2"
	"log"
	"time"
)

func (app *App) Bot() *tb.Bot {
	bot, err := tb.NewBot(tb.Settings{
		Token:  app.BotToken,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		log.Fatalln(err)
	}
	return bot
}
