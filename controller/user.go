package controller

import (
	"database/sql"
	"github.com/amiraliio/tgbp/config"
	"github.com/amiraliio/tgbp/model"
	tb "gopkg.in/tucnak/telebot.v2"
	"log"
	"strings"
)

func RegisterUserWithemail(bot *tb.Bot, m *tb.Message, lastState *model.UserLastState, text string, userID int) {
	userModel := new(tb.User)
	userModel.ID = userID
	if strings.Contains(text, "@gmail.com") || strings.Contains(text, "@yahoo.com") || strings.Contains(text, "@hotmail.com") || strings.Contains(text, "@outlook.com") {
		bot.Send(userModel, "You Can't Enter Free Email services")
		return
	}
	db, err := config.DB()
	if err != nil {
		log.Println(err)
		return
	}
	defer db.Close()
	if lastState.State == "register_user_with_email" && strings.Contains(text, "@") {
		splitEmail := strings.Split(text, "@")
		emailSuffix := "@" + splitEmail[1]
		checkTheCompanyEmailSuffixExist(bot, emailSuffix, db, userModel)
	} else {
		bot.Send(userModel, "Please Enter Your Company Email:")
		SaveUserLastState(bot, text, userID, "register_user_with_email")
	}
}

func checkTheCompanyEmailSuffixExist(bot *tb.Bot, emailSuffix string, db *sql.DB, userModel *tb.User) {
	tempData, err := db.Prepare("SELECT co.companyName,ch.id,ch.channelName from `channels_email_suffixes` as cs inner join `channels` as ch on cs.channelID=ch.id inner join `companies_channels` as cc on ch.Id=cc.channelID inner join `companies` as co on cc.companyID=co.id where cs.suffix=?")
	if err != nil {
		log.Println(err)
		return
	}
	defer tempData.Close()
	companyModel := new(model.Company)
	channelModel := new(model.Channel)
	options := new(tb.SendOptions)
	yesBTN := tb.ReplyButton{
		Text: "Yes",
	}
	noBTN := tb.ReplyButton{
		Text: "No",
	}
	homeBTN := tb.ReplyButton{
		Text: "Home",
	}
	replyKeys := [][]tb.ReplyButton{
		[]tb.ReplyButton{yesBTN, noBTN},
		[]tb.ReplyButton{homeBTN},
	}
	replyModel := new(tb.ReplyMarkup)
	replyModel.ReplyKeyboard = replyKeys
	options.ReplyMarkup = replyModel
	if err = tempData.QueryRow(emailSuffix).Scan(&companyModel.CompanyName, &channelModel.ID, &channelModel.ChannelName); err != nil {
		bot.Send(userModel, "The company according to your email doesn't exist, do you confirm sending a registration request for your company?", options)
		return
	}
	bot.Send(userModel, "Do you confirm that you want to register to the channel/group "+channelModel.ChannelName+" blongs to the company "+companyModel.CompanyName+"?", options)
}
