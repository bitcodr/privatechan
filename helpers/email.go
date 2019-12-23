package helpers

import (
	"crypto/tls"
	"github.com/spf13/viper"
	"gopkg.in/gomail.v2"
	"log"
)

func SendEmail(body, to string) {
	from := viper.GetString("EMAIL.FROM")
	pass := viper.GetString("EMAIL.PASSWORD")
	userName := viper.GetString("EMAIL.USERNAME")
	serverAddress := viper.GetString("EMAIL.PROVIDER")
	serverPort := viper.GetInt("EMAIL.PORT")
	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", "Candor Bot Activation Key")
	m.SetBody("text/html", "Your activation key is:   <b>"+body+"</b> this key will be expire about 10 minutes")
	d := gomail.NewDialer(serverAddress, serverPort, userName, pass)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	if err := d.DialAndSend(m); err != nil {
		log.Printf("smtp error: %s", err)
		return
	}
}
