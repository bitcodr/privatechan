package controller

import (
	"database/sql"
	"github.com/amiraliio/tgbp/config"
	"github.com/amiraliio/tgbp/helpers"
	"github.com/amiraliio/tgbp/model"
	tb "gopkg.in/tucnak/telebot.v2"
	"log"
	"math/rand"
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
			bot.Send(userModel, "You Can't Enter a Free Email Service, Please Enter Your Email in Your Company:")
			return
		}
		db, err := config.DB()
		if err != nil {
			log.Println(err)
			return
		}
		defer db.Close()
		checkTheCompanyEmailSuffixExist(bot, text, "@"+emailSuffix[1], db, userModel)
		return
	}
	SaveUserLastState(bot, text, userID, "register_user_with_email")
	bot.Send(userModel, "Please Enter Your Email in Your Company:", homeKeyOption())
}

func checkTheCompanyEmailSuffixExist(bot *tb.Bot, email, emailSuffix string, db *sql.DB, userModel *tb.User) {
	tempData, err := db.Prepare("SELECT co.companyName,ch.id,ch.channelName from `channels_email_suffixes` as cs inner join `channels` as ch on cs.channelID=ch.id inner join `companies_channels` as cc on ch.Id=cc.channelID inner join `companies` as co on cc.companyID=co.id where cs.suffix=? limit 1")
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
	SaveUserLastState(bot, strconv.FormatInt(channelModel.ID, 10)+"_"+email, userModel.ID, "register_user_for_the_company")
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
		if !strings.Contains(lastState.Data, "_") {
			log.Println("string must be two part, channelID and userEmail")
			return
		}
		channelData := strings.Split(lastState.Data, "_")
		if len(channelData) != 2 {
			log.Println("length of channel data must be 2")
			return
		}
		db, err := config.DB()
		if err != nil {
			log.Println(err)
			return
		}
		defer db.Close()
		userExistOrNot, err := db.Prepare("SELECT us.id FROM `users` as us inner join `users_channels` as uc on us.id=uc.userID and uc.channelID=? where us.userID=?")
		if err != nil {
			log.Println(err)
			return
		}
		defer userExistOrNot.Close()
		userDBModel := new(model.User)
		if err := userExistOrNot.QueryRow(channelData[0], userID).Scan(&userDBModel.ID); err == nil {
			bot.Send(userModel, "You've been registered in the channel/group")
			return
		}
		rand.Seed(time.Now().UnixNano())
		randomeNumber := rand.Intn(100000)
		hashedRandomNumber, err := helpers.HashPassword(strconv.Itoa(randomeNumber))
		if err != nil {
			log.Println(err)
			return
		}
		insertCompanyRequest, err := db.Query("INSERT INTO `users_activation_key` (`userID`,`activeKey`,`createdAt`) VALUES('" + strconv.FormatInt(lastState.UserID, 10) + "','" + hashedRandomNumber + "','" + time.Now().UTC().Format("2006-01-02 03:04:05") + "')")
		if err != nil {
			log.Println(err)
			return
		}
		defer insertCompanyRequest.Close()
		go helpers.SendEmail(strconv.Itoa(randomeNumber), channelData[1])
		SaveUserLastState(bot, lastState.Data, userID, "email_for_user_registration")
		bot.Send(userModel, "Please Enter The Code That Sent To Your Email Address", homeKeyOption())
	case "No":
		SaveUserLastState(bot, "", userID, "cancel_user_registration_for_the_company")
		bot.Send(userModel, "You Can Continue in Bot by Pressing The Home Button", homeKeyOption())
	}
}

func RegisterUserWithEmail(bot *tb.Bot, m *tb.Message, lastState *model.UserLastState, text string, userID int) {
	db, err := config.DB()
	if err != nil {
		log.Println(err)
		return
	}
	defer db.Close()
	userModel := new(tb.User)
	userModel.ID = userID
	userActiveKey, err := db.Prepare("SELECT `activeKey`,`createdAt` FROM `users_activation_key` where userID=? order by `createdAt` DESC limit 1")
	if err != nil {
		log.Println(err)
		return
	}
	defer userActiveKey.Close()
	userActiveKeyModel := new(model.UsersActivationKey)
	if err := userActiveKey.QueryRow(userID).Scan(&userActiveKeyModel.ActiveKey, &userActiveKeyModel.CreatedAt); err != nil {
		log.Println(err)
		return
	}
	//TODO check token expire time
	if !helpers.CheckPasswordHash(text, userActiveKeyModel.ActiveKey) {
		bot.Send(userModel, "The key Is Invalid", homeKeyOption())
		return
	}
	if !strings.Contains(lastState.Data, "_") {
		log.Println("string must be two part, channelID and userEmail")
		return
	}
	channelData := strings.Split(lastState.Data, "_")
	if len(channelData) != 2 {
		log.Println("length of channel data must be 2")
		return
	}
	resultsStatement, err := db.Prepare("SELECT channelID,channelURL,manualChannelName FROM `channels` where id=?")
	if err != nil {
		log.Println(err)
		return
	}
	defer resultsStatement.Close()
	channelModel := new(model.Channel)
	channelID, err := strconv.ParseInt(channelData[0], 10, 0)
	if err != nil {
		log.Println(err)
		return
	}
	if err := resultsStatement.QueryRow(channelID).Scan(&channelModel.ChannelID, &channelModel.ChannelURL, &channelModel.ManualChannelName); err != nil {
		log.Println(err)
		return
	}
	JoinFromGroup(bot, m, channelModel.ChannelID)
	_, err = db.Query("update `users` set `email`=? where `userID`=?", channelData[1], userID)
	if err != nil {
		log.Println(err)
		return
	}
	SaveUserLastState(bot, "", userID, "join_request_added")
	options := new(tb.SendOptions)
	homeBTN := tb.ReplyButton{
		Text: "Home",
	}
	replyBTN := [][]tb.ReplyButton{
		[]tb.ReplyButton{homeBTN},
	}
	startBTN := tb.InlineButton{
		Text: "Click Here To Start Communication",
		URL:  channelModel.ChannelURL,
	}
	replyKeys := [][]tb.InlineButton{
		[]tb.InlineButton{startBTN},
	}
	replyModel := new(tb.ReplyMarkup)
	replyModel.ReplyKeyboard = replyBTN
	replyModel.InlineKeyboard = replyKeys
	options.ReplyMarkup = replyModel
	bot.Send(userModel, "You are now member of channel/group "+channelModel.ManualChannelName, options)
}



func CheckUserRegisteredOrNot(bot *tb.Bot, m *tb.Message, lastState *model.UserLastState, text string, userID int){
	//TODO check the channel is registered or not
	//TODO if the channel is one of the company that user is registered verification is not necessary
	//TODO also check it according to event channel is required a action for instance reply is mandatory or not
	//TODO check if user is registered to company or not

}