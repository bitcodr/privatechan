package helpers

import (
	"crypto/tls"
	"github.com/amiraliio/tgbp/config"
	"gopkg.in/gomail.v2"
	"log"
)

func SendEmail(body, to string) {
	from := config.AppConfig.GetString("EMAIL.FROM")
	pass := config.AppConfig.GetString("EMAIL.PASSWORD")
	userName := config.AppConfig.GetString("EMAIL.USERNAME")
	serverAddress := config.AppConfig.GetString("EMAIL.PROVIDER")
	serverPort := config.AppConfig.GetInt("EMAIL.PORT")
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
