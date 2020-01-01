package config

import (
	tb "gopkg.in/tucnak/telebot.v2"
	"log"
	"strings"
	"time"
)

func (app *App) Bot() *tb.Bot {
	poller := &tb.LongPoller{Timeout: 15 * time.Second}
	spamProtected := tb.NewMiddlewarePoller(poller, func(upd *tb.Update) bool {
		if upd.Message == nil {
			return true
		}
		if strings.Contains(upd.Message.Text, "spam") {
			return false
		}
		return true
	})
	bot, err := tb.NewBot(tb.Settings{
		Token:  app.BotToken,
		Poller: spamProtected,
	})
	if err != nil {
		log.Fatalln(err)
	}
	return bot
}
