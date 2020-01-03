package config

import (
	"github.com/spf13/viper"
	"log"
)

var (
	AppConfig  *viper.Viper
	LangConfig *viper.Viper
	QConfig    *viper.Viper
)

func (app *App) Environment() {
	app.appConfig()
	app.langConfig()
	app.questionConfig()
}

func (app *App) appConfig() {
	AppConfig = viper.New()
	AppConfig.SetConfigType("yaml")
	AppConfig.SetConfigName("config")
	AppConfig.AddConfigPath(app.ProjectDir)
	AppConfig.AddConfigPath("/var/www/privatechan")
	err := AppConfig.ReadInConfig()
	if err != nil {
		log.Fatalln(err)
	}
}

func (app *App) langConfig() {
	LangConfig = viper.New()
	LangConfig.SetConfigType("yaml")
	LangConfig.SetConfigName("lang")
	LangConfig.AddConfigPath(app.ProjectDir + "/lang")
	LangConfig.AddConfigPath("/var/www/privatechan/lang")
	err := LangConfig.ReadInConfig()
	if err != nil {
		log.Fatalln(err)
	}
}

func (app *App) questionConfig() {
	QConfig = viper.New()
	QConfig.SetConfigType("yaml")
	QConfig.SetConfigName("question")
	QConfig.AddConfigPath(app.ProjectDir + "/lang")
	QConfig.AddConfigPath("/var/www/privatechan/lang")
	err := QConfig.ReadInConfig()
	if err != nil {
		log.Fatalln(err)
	}
}
