package controllers

import (
	"database/sql"
	"github.com/amiraliio/tgbp/config"
	"github.com/amiraliio/tgbp/events"
	"github.com/amiraliio/tgbp/helpers"
	"github.com/amiraliio/tgbp/models"
	tb "gopkg.in/tucnak/telebot.v2"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

func (service *BotService) RegisterUserWithemail(db *sql.DB, app *config.App, bot *tb.Bot, m *tb.Message, request *events.Event, lastState *models.UserLastState, text string, userID int) bool {
	userModel := new(tb.User)
	userModel.ID = userID
	if lastState.State == "register_user_with_email" {
		if !strings.Contains(text, "@") {
			bot.Send(userModel, "Please Enter a Valid Email Address")
			return true
		}
		emails := []string{"gmail.com", "yahoo.com", "hotmail.com", "outlook.com", "zoho.com", "icloud.com", "mail.com", "aol.com", "yandex.com"}
		emailSuffix := strings.Split(text, "@")
		if helpers.SortAndSearchInStrings(emails, emailSuffix[1]) {
			bot.Send(userModel, "You Can't Enter a Free Email Service, Please Enter Your Email in Your Company:")
			return true
		}
		service.checkTheCompanyEmailSuffixExist(app, bot, text, "@"+emailSuffix[1], db, userModel)
		return true
	}
	SaveUserLastState(db, app, bot, text, userID, "register_user_with_email")
	bot.Send(userModel, "Please Enter Your Email in Your Company:", HomeKeyOption(db, app))
	return true
}

func (service *BotService) checkTheCompanyEmailSuffixExist(app *config.App, bot *tb.Bot, email, emailSuffix string, db *sql.DB, userModel *tb.User) {
	tempData, err := db.Prepare("SELECT co.companyName,ch.id,ch.channelName from `channels_email_suffixes` as cs inner join `channels` as ch on cs.channelID=ch.id inner join `companies_channels` as cc on ch.Id=cc.channelID inner join `companies` as co on cc.companyID=co.id where cs.suffix=? limit 1")
	if err != nil {
		log.Println(err)
		return
	}
	defer tempData.Close()
	companyModel := new(models.Company)
	channelModel := new(models.Channel)
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
		SaveUserLastState(db, app, bot, emailSuffix, userModel.ID, "confirm_register_company_email_address")
		bot.Send(userModel, "The company according to your email doesn't exist, do you confirm sending a registration request for your company?", options)
		return
	}
	SaveUserLastState(db, app, bot, strconv.FormatInt(channelModel.ID, 10)+"_"+email, userModel.ID, "register_user_for_the_company")
	bot.Send(userModel, "Do you confirm that you want to register to the channel/group "+channelModel.ChannelName+" blongs to the company "+companyModel.CompanyName+"?", options)
}

func (service *BotService) ConfirmRegisterCompanyRequest(db *sql.DB, app *config.App, bot *tb.Bot, m *tb.Message, request *events.Event, lastState *models.UserLastState) bool {
	userModel := new(tb.User)
	userModel.ID = m.Sender.ID
	switch m.Text {
	case "Yes":
		insertCompanyRequest, err := db.Query("INSERT INTO `companies_join_request` (`userID`,`emailSuffix`,`createdAt`) VALUES('" + strconv.FormatInt(lastState.UserID, 10) + "','" + lastState.Data + "','" + app.CurrentTime + "')")
		if err != nil {
			log.Println(err)
			return true
		}
		defer insertCompanyRequest.Close()
		SaveUserLastState(db, app, bot, "", m.Sender.ID, "join_request_added")
		bot.Send(userModel, "Your Request Sent To The Admin And Will be Active After Confirmation.", HomeKeyOption(db, app))
	case "No":
		SaveUserLastState(db, app, bot, "", m.Sender.ID, "join_request_dismissed")
		bot.Send(userModel, "You Can Continue in Bot by Pressing The Home Button", HomeKeyOption(db, app))
	}
	return true
}

func HomeKeyOption(db *sql.DB, app *config.App) *tb.SendOptions {
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

func (service *BotService) ConfirmRegisterUserForTheCompany(db *sql.DB, app *config.App, bot *tb.Bot, m *tb.Message, request *events.Event, lastState *models.UserLastState) bool {
	userModel := new(tb.User)
	userModel.ID = m.Sender.ID
	switch m.Text {
	case "Yes":
		if !strings.Contains(lastState.Data, "_") {
			log.Println("string must be two part, channelID and userEmail")
			return true
		}
		channelData := strings.Split(lastState.Data, "_")
		if len(channelData) != 2 {
			log.Println("length of channel data must be 2")
			return true
		}
		userExistOrNot, err := db.Prepare("SELECT us.id FROM `users` as us inner join `users_channels` as uc on us.id=uc.userID and uc.channelID=? where us.userID=?")
		if err != nil {
			log.Println(err)
			return true
		}
		defer userExistOrNot.Close()
		userDBModel := new(models.User)
		if err := userExistOrNot.QueryRow(channelData[0], m.Sender.ID).Scan(&userDBModel.ID); err == nil {
			bot.Send(userModel, "You've been registered in the channel/group")
			return true
		}
		rand.Seed(time.Now().UnixNano())
		randomeNumber := rand.Intn(100000)
		hashedRandomNumber, err := helpers.HashPassword(strconv.Itoa(randomeNumber))
		if err != nil {
			log.Println(err)
			return true
		}
		insertCompanyRequest, err := db.Query("INSERT INTO `users_activation_key` (`userID`,`activeKey`,`createdAt`) VALUES('" + strconv.FormatInt(lastState.UserID, 10) + "','" + hashedRandomNumber + "','" + app.CurrentTime + "')")
		if err != nil {
			log.Println(err)
			return true
		}
		defer insertCompanyRequest.Close()
		go helpers.SendEmail(strconv.Itoa(randomeNumber), channelData[1])
		SaveUserLastState(db, app, bot, lastState.Data, m.Sender.ID, "email_for_user_registration")
		bot.Send(userModel, "Please Enter The Code That Sent To Your Email Address", HomeKeyOption(db, app))
	case "No":
		SaveUserLastState(db, app, bot, "", m.Sender.ID, "cancel_user_registration_for_the_company")
		bot.Send(userModel, "You Can Continue in Bot by Pressing The Home Button", HomeKeyOption(db, app))
	}
	return true
}

func (service *BotService) RegisterUserWithEmailAndCode(db *sql.DB, app *config.App, bot *tb.Bot, m *tb.Message, request *events.Event, lastState *models.UserLastState) bool {
	userModel := new(tb.User)
	userModel.ID = m.Sender.ID
	userActiveKey, err := db.Prepare("SELECT `activeKey`,`createdAt` FROM `users_activation_key` where userID=? order by `createdAt` DESC limit 1")
	if err != nil {
		log.Println(err)
		return true
	}
	defer userActiveKey.Close()
	userActiveKeyModel := new(models.UsersActivationKey)
	if err := userActiveKey.QueryRow(m.Sender.ID).Scan(&userActiveKeyModel.ActiveKey, &userActiveKeyModel.CreatedAt); err != nil {
		log.Println(err)
		return true
	}
	//TODO check token expire time
	if !helpers.CheckPasswordHash(m.Text, userActiveKeyModel.ActiveKey) {
		bot.Send(userModel, "The key Is Invalid", HomeKeyOption(db, app))
		return true
	}
	if !strings.Contains(lastState.Data, "_") {
		log.Println("string must be two part, channelID and userEmail")
		return true
	}
	channelData := strings.Split(lastState.Data, "_")
	if len(channelData) != 2 {
		log.Println("length of channel data must be 2")
		return true
	}
	resultsStatement, err := db.Prepare("SELECT channelID,channelURL,manualChannelName FROM `channels` where id=?")
	if err != nil {
		log.Println(err)
		return true
	}
	defer resultsStatement.Close()
	channelModel := new(models.Channel)
	channelID, err := strconv.ParseInt(channelData[0], 10, 0)
	if err != nil {
		log.Println(err)
		return true
	}
	if err := resultsStatement.QueryRow(channelID).Scan(&channelModel.ChannelID, &channelModel.ChannelURL, &channelModel.ManualChannelName); err != nil {
		log.Println(err)
		return true
	}
	service.JoinFromGroup(db, app, bot, m, channelModel.ChannelID)
	_, err = db.Query("update `users` set `email`=? where `userID`=?", channelData[1], m.Sender.ID)
	if err != nil {
		log.Println(err)
		return true
	}
	SaveUserLastState(db, app, bot, "", m.Sender.ID, "join_request_added")
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
	return true
}

func (service *BotService) GetUserByTelegramID(db *sql.DB, app *config.App, userID int) *models.User {
	userLastStateQueryStatement, err := db.Prepare("SELECT `id`,`userID` from `users` where `userID`=? ")
	if err != nil {
		log.Println(err)
	}
	defer userLastStateQueryStatement.Close()
	userLastStateQuery, err := userLastStateQueryStatement.Query(userID)
	if err != nil {
		log.Println(err)
	}
	userModel := new(models.User)
	if userLastStateQuery.Next() {
		if err := userLastStateQuery.Scan(&userModel.ID, &userModel.UserID); err != nil {
			log.Println(err)
		}
		return userModel
	}
	return userModel
}

func (service *BotService) CheckUserRegisteredOrNot(db *sql.DB, app *config.App, bot *tb.Bot, m *tb.Message, lastState *models.UserLastState, text string, userID int) {
	//TODO check the channel is registered or not
	//TODO if the channel is one of the company that user is registered verification is not necessary
	//TODO also check it according to event channel is required a action for instance reply is mandatory or not
	//TODO check if user is registered to company or not

}
