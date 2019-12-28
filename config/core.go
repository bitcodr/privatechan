package config

import (
	"database/sql"
	"github.com/spf13/viper"
	tb "gopkg.in/tucnak/telebot.v2"
	"time"
)

type AppInterface interface {
	Environment()
	DB() (*sql.DB, error)
	Bot() *tb.Bot
	SetOtherConfigs() *App
}

type App struct {
	ProjectDir, BotToken, BotUsername, DBName, DBUserName, DBPass, CurrentTime string
}

func (app *App) SetAppConfig() *App {
	app.BotToken = viper.GetString("APP.TELEGRAM_API_TOKEN")
	app.DBName = viper.GetString("DATABASES.MYSQL.DATABASE")
	app.DBUserName = viper.GetString("DATABASES.MYSQL.USERNAME")
	app.DBPass = viper.GetString("DATABASES.MYSQL.PASSWORD")
	app.CurrentTime = time.Now().UTC().Format("2006-01-02 03:04:05")
	app.BotUsername = viper.GetString("APP.APP.BOT_USERNAME")
	return app
}
