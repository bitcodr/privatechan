package helpers

import (
	"github.com/spf13/viper"
	"log"
	"net/smtp"
)

func SendEmail(body, to string) {
	from := viper.GetString("EMAIL.FROM")
	pass := viper.GetString("EMAIL.PASSWORD")
	msg := "From: " + from + "\n" + "To: " + to + "\n" + "Subject: Joining to Channel/Group\n\n" +body
	err := smtp.SendMail(viper.GetString("EMAIL.PROVIDER")+":"+viper.GetString("EMAIL.PORT"),smtp.PlainAuth("", from, pass, viper.GetString("EMAIL.PROVIDER")),from, []string{to}, []byte(msg))
	if err != nil {
		log.Printf("smtp error: %s", err)
		return
	}
}
