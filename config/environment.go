package config

import (
	"github.com/spf13/viper"
	"log"
)

func (app *App) Environment() {
	viper.SetConfigName("config")
	viper.AddConfigPath(app.ProjectDir)
	// viper.AddConfigPath("/var/www/privatechan")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalln(err)
	}
}


