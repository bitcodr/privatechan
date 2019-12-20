package controller

import (
	"database/sql"
	"github.com/amiraliio/tgbp/config"
	"github.com/amiraliio/tgbp/helpers"
	"github.com/amiraliio/tgbp/model"
	tb "gopkg.in/tucnak/telebot.v2"
	"log"
	"strconv"
	"strings"
	"time"
)

func RegisterUserWithemail(bot *tb.Bot, m *tb.Message, lastState *model.UserLastState, text string, userID int) {
	userModel := new(tb.User)
	userModel.ID = userID
	if lastState.State == "register_user_with_email" {
		if !strings.Contains(text, "@") {
			bot.Send(userModel, "Please Enter a Valid Email Address")
			return
		}
		emails := []string{"gmail.com", "yahoo.com", "hotmail.com", "outlook.com", "zoho.com", "icloud.com", "mail.com", "aol.com", "yandex.com"}
		emailSuffix := strings.Split(text, "@")
		if helpers.SortAndSearchInStrings(emails, emailSuffix[1]) {
			bot.Send(userModel, "You Can't Enter an Free Email Services")
			return
		}
		db, err := config.DB()
		if err != nil {
			log.Println(err)
			return
		}
		defer db.Close()
		checkTheCompanyEmailSuffixExist(bot, "@"+emailSuffix[1], db, userModel)
		return
	}
	SaveUserLastState(bot, text, userID, "register_user_with_email")
	bot.Send(userModel, "Please Enter Your Company Email:", homeKeyOption())
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
		SaveUserLastState(bot, emailSuffix, userModel.ID, "confirm_register_company_email_address")
		bot.Send(userModel, "The company according to your email doesn't exist, do you confirm sending a registration request for your company?", options)
		return
	}
	SaveUserLastState(bot, strconv.FormatInt(channelModel.ID, 10), userModel.ID, "register_user_for_the_company")
	bot.Send(userModel, "Do you confirm that you want to register to the channel/group "+channelModel.ChannelName+" blongs to the company "+companyModel.CompanyName+"?", options)
}

func ConfirmRegisterCompanyRequest(bot *tb.Bot, m *tb.Message, lastState *model.UserLastState, text string, userID int) {
	userModel := new(tb.User)
	userModel.ID = userID
	switch text {
	case "Yes":
		db, err := config.DB()
		if err != nil {
			log.Println(err)
			return
		}
		defer db.Close()
		insertCompanyRequest, err := db.Query("INSERT INTO `companies_join_request` (`userID`,`emailSuffix`,`createdAt`) VALUES('" + strconv.FormatInt(lastState.UserID, 10) + "','" + lastState.Data + "','" + time.Now().UTC().Format("2006-01-02 03:04:05") + "')")
		if err != nil {
			log.Println(err)
			return
		}
		defer insertCompanyRequest.Close()
		SaveUserLastState(bot, "", userID, "join_request_added")
		bot.Send(userModel, "Your Request Sent To The Admin And Will be Active After Confirmation.", homeKeyOption())
	case "No":
		SaveUserLastState(bot, "", userID, "join_request_dismissed")
		bot.Send(userModel, "You Can Continue in Bot by Pressing The Home Button", homeKeyOption())
	}
}

func homeKeyOption() *tb.SendOptions {
	options := new(tb.SendOptions)
	homeBTN := tb.ReplyButton{
		Text: "Home",
	}
	replyKeys := [][]tb.ReplyButton{
		[]tb.ReplyButton{homeBTN},
	}
	replyModel := new(tb.ReplyMarkup)
	replyModel.ReplyKeyboard = replyKeys
	options.ReplyMarkup = replyModel
	return options
}

func ConfirmRegisterUserForTheCompany(bot *tb.Bot, m *tb.Message, lastState *model.UserLastState, text string, userID int) {
	userModel := new(tb.User)
	userModel.ID = userID
	switch text {
	case "Yes":
		JoinFromGroup(bot, m, lastState.Data)
		db, err := config.DB()
		if err != nil {
			log.Println(err)
			return
		}
		defer db.Close()
		resultsStatement, err := db.Prepare("SELECT channelURL,manualChannelName FROM `channels` where channelID=?")
		if err != nil {
			log.Println(err)
			return
		}
		defer resultsStatement.Close()
		channelModel := new(model.Channel)
		if err := resultsStatement.QueryRow(lastState.Data).Scan(&channelModel.ChannelURL, &channelModel.ManualChannelName); err != nil {
			log.Println(err)
			return
		}
		SaveUserLastState(bot, "", userID, "join_request_added")
		options := new(tb.SendOptions)
		startBTN := tb.InlineButton{
			Text: "Click Here To Start Communication",
		}
		replyKeys := [][]tb.InlineButton{
			[]tb.InlineButton{startBTN},
		}
		replyModel := new(tb.ReplyMarkup)
		replyModel.InlineKeyboard = replyKeys
		options.ReplyMarkup = replyModel
		bot.Send(userModel, "You are now member of channel/group "+channelModel.ManualChannelName, options)
	case "No":
		SaveUserLastState(bot, "", userID, "cancel_user_registration_for_the_company")
		bot.Send(userModel, "You Can Continue in Bot by Pressing The Home Button", homeKeyOption())
	}
}
