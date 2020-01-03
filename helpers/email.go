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
	m.SetHeader("Subject", config.AppConfig.GetString("APP.BOT_USERNAME")+config.LangConfig.GetString("MESSAGES.ACTIVE_KEY"))
	m.SetBody("text/html", config.LangConfig.GetString("MESSAGES.YOU_ACTIVE_KEY")+"<b>"+body+"</b>"+config.LangConfig.GetString("MESSAGES.ACTIVE_EXPIRE_PEROID"))
	d := gomail.NewDialer(serverAddress, serverPort, userName, pass)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	if err := d.DialAndSend(m); err != nil {
		log.Println(err.Error())
		return
	}
}
