package config

import (
	"database/sql"
	tb "gopkg.in/tucnak/telebot.v2"
	"time"
)

type AppInterface interface {
	Environment()
	DB() (*sql.DB, error)
	Bot() *tb.Bot
	SetAppConfig() *App
}

type App struct {
	ProjectDir, BotToken, BotUsername, DBName, DBUserName, DBPass, CurrentTime, TgDomain, APIURL string
}

func (app *App) SetAppConfig() *App {
	app.BotToken = AppConfig.GetString("APP.TELEGRAM_API_TOKEN")
	app.DBName = AppConfig.GetString("DATABASES.MYSQL.DATABASE")
	app.DBUserName = AppConfig.GetString("DATABASES.MYSQL.USERNAME")
	app.DBPass = AppConfig.GetString("DATABASES.MYSQL.PASSWORD")
	app.CurrentTime = time.Now().UTC().Format("2006-01-02 03:04:05")
	app.BotUsername = AppConfig.GetString("APP.BOT_USERNAME")
	app.TgDomain = "https://t.me/"
	app.APIURL = AppConfig.GetString("APP.API_URL")
	return app
}
