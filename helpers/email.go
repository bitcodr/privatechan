package helpers

import (
	"github.com/spf13/viper"
	"log"
	"net/smtp"
)

func SendEmail(body, to string) {
	from := viper.GetString("EMAIL.FROM")
	pass := viper.GetString("EMAIL.PASSWORD")
	msg := []byte("From: " + from + "\n" + "To: " + to + "\n" + "Subject: Joining to Channel/Group\n\n" + body)
	auth := smtp.PlainAuth("", from, pass, viper.GetString("EMAIL.PROVIDER"))
	serverAddress := viper.GetString("EMAIL.PROVIDER") + ":" + viper.GetString("EMAIL.PORT")
	receiver := []string{to}
	err := smtp.SendMail(serverAddress, auth, from, receiver, msg)
	if err != nil {
		log.Printf("smtp error: %s", err)
		return
	}
}
