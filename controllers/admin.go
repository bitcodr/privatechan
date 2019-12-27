package controllers

import (
	"database/sql"
	"github.com/amiraliio/tgbp/config"
	"github.com/amiraliio/tgbp/events"
	"github.com/amiraliio/tgbp/models"
	"github.com/spf13/viper"
	tb "gopkg.in/tucnak/telebot.v2"
	"log"
	"strconv"
	"strings"
	"time"
)

func (service *events.BotService) SetUpCompanyByAdmin(db *sql.DB, app *config.App, bot *tb.Bot, m *tb.Message, request *events.Event, lastState *models.UserLastState, text string, userID int) bool {
	if lastState.Data != "" && lastState.State == "setup_verified_company_account" {
		questions := viper.GetStringMap("SUPERADMIN.COMPANY.SETUP.QUESTIONS")
		numberOfQuestion := strings.Split(lastState.Data, "_")
		if len(numberOfQuestion) == 2 {
			questioNumber := numberOfQuestion[0]
			relationDate := numberOfQuestion[1]
			prevQuestionNo, err := strconv.Atoi(questioNumber)
			if err == nil {
				tableName := viper.GetString("SUPERADMIN.COMPANY.SETUP.QUESTIONS.N" + questioNumber + ".TABLE_NAME")
				columnName := viper.GetString("SUPERADMIN.COMPANY.SETUP.QUESTIONS.N" + questioNumber + ".COLUMN_NAME")
				_, err = db.Query("INSERT INTO `temp_setup_flow` (`tableName`,`columnName`,`data`,`userID`,`relation`,`createdAt`) VALUES ('" + tableName + "','" + columnName + "','" + strings.TrimSpace(text) + "','" + strconv.Itoa(userID) + "','setup_verified_company_account_" + strconv.Itoa(userID) + "_" + relationDate + "','" + time.Now().UTC().Format("2006-01-02 03:04:05") + "')")
				if err != nil {
					log.Println(err)
					return true
				}
				if prevQuestionNo+1 > len(questions) {
					service.finalStage(app, bot, relationDate, db, text, userID)
					return true
				}
				service.nextQuestion(db, app, bot, m, lastState, relationDate, prevQuestionNo, text, userID)
			}
		}
		return true
	}
	initQuestion := viper.GetString("SUPERADMIN.COMPANY.SETUP.QUESTIONS.N1.QUESTION")
	service.sendMessageUserWithActionOnKeyboards(db, app, bot, userID, initQuestion, true)
	SaveUserLastState(db, app, bot, "1_"+strconv.FormatInt(time.Now().Unix(), 10), userID, "setup_verified_company_account")
	return true
}

//next question
func (service *events.BotService) nextQuestion(db *sql.DB, app *config.App, bot *tb.Bot, m *tb.Message, lastState *models.UserLastState, relationDate string, prevQuestionNo int, text string, userID int) {
	questionText := viper.GetString("SUPERADMIN.COMPANY.SETUP.QUESTIONS.N" + strconv.Itoa(prevQuestionNo+1) + ".QUESTION")
	answers := viper.GetString("SUPERADMIN.COMPANY.SETUP.QUESTIONS.N" + strconv.Itoa(prevQuestionNo+1) + ".ANSWERS")
	if answers != "" && strings.Contains(strings.TrimSpace(answers), ",") {
		splittedAnswers := strings.Split(answers, ",")
		replyKeysNested := []tb.ReplyButton{}
		for _, v := range splittedAnswers {
			replyBTN := tb.ReplyButton{
				Text: v,
			}
			replyKeysNested = append(replyKeysNested, replyBTN)
		}
		homeBTN := tb.ReplyButton{
			Text: "Home",
		}
		replyKeys := [][]tb.ReplyButton{
			replyKeysNested,
			[]tb.ReplyButton{homeBTN},
		}
		userModel := new(tb.User)
		userModel.ID = userID
		options := new(tb.SendOptions)
		replyMarkupModel := new(tb.ReplyMarkup)
		replyMarkupModel.ReplyKeyboard = replyKeys
		options.ReplyMarkup = replyMarkupModel
		_, _ = bot.Send(userModel, questionText, options)
	} else {
		userModel := new(tb.User)
		userModel.ID = userID
		homeBTN := tb.ReplyButton{
			Text: "Home",
		}
		replyKeys := [][]tb.ReplyButton{
			[]tb.ReplyButton{homeBTN},
		}
		options := new(tb.SendOptions)
		replyMarkupModel := new(tb.ReplyMarkup)
		replyMarkupModel.ReplyKeyboard = replyKeys
		options.ReplyMarkup = replyMarkupModel
		_, _ = bot.Send(userModel, questionText, options)
	}
	SaveUserLastState(db, app, bot, strconv.Itoa(prevQuestionNo+1)+"_"+relationDate, userID, "setup_verified_company_account")
}

func (service *events.BotService) sendMessageUserWithActionOnKeyboards(db *sql.DB, app *config.App, bot *tb.Bot, userID int, message string, showKeyboard bool) {
	userModel := new(tb.User)
	userModel.ID = userID
	options := new(tb.SendOptions)
	homeBTN := tb.ReplyButton{
		Text: "Home",
	}
	replyKeys := [][]tb.ReplyButton{
		[]tb.ReplyButton{homeBTN},
	}
	replyModel := new(tb.ReplyMarkup)
	replyModel.ReplyKeyboardRemove = showKeyboard
	replyModel.ReplyKeyboard = replyKeys
	options.ReplyMarkup = replyModel
	_, _ = bot.Send(userModel, message, options)
}

func (service *events.BotService) finalStage(app *config.App, bot *tb.Bot, relationDate string, db *sql.DB, text string, userID int) {
	tempData, err := db.Prepare("SELECT id,tableName,columnName,data,relation,status,userID,createdAt from `temp_setup_flow` where status='ACTIVE' and relation=? and userID=?")
	if err != nil {
		log.Println(err)
		return
	}
	defer tempData.Close()
	results, err := tempData.Query("setup_verified_company_account_"+strconv.Itoa(userID)+"_"+relationDate, userID)
	if err == nil {
		var channelTableData []*models.TempSetupFlow
		var companyTableData []*models.TempSetupFlow
		var channelsEmailSuffixes []*models.TempSetupFlow
		var channelsSettings []*models.TempSetupFlow
		for results.Next() {
			tempSetupFlow := new(models.TempSetupFlow)
			err := results.Scan(&tempSetupFlow.ID, &tempSetupFlow.TableName, &tempSetupFlow.ColumnName, &tempSetupFlow.Data, &tempSetupFlow.Relation, &tempSetupFlow.Status, &tempSetupFlow.UserID, &tempSetupFlow.CreatedAt)
			if err != nil {
				log.Println(err)
				return
			}
			switch tempSetupFlow.TableName {
			case "companies":
				companyTableData = append(companyTableData, tempSetupFlow)
			case "channels":
				channelTableData = append(channelTableData, tempSetupFlow)
			case "channels_settings":
				channelsSettings = append(channelsSettings, tempSetupFlow)
			case "channels_email_suffixes":
				channelsEmailSuffixes = append(channelsEmailSuffixes, tempSetupFlow)
			}
		}
		transaction, err := db.Begin()
		if err != nil {
			log.Println(err)
			return
		}
		//insert company
		service.insertFinalStateData(app, bot, userID, transaction, channelTableData, companyTableData, channelsEmailSuffixes, channelsSettings, db)
		//update state of temp setup data
		_, err = transaction.Exec("update `temp_setup_flow` set `status`='INACTIVE' where status='ACTIVE' and relation=? and `userID`=?", "setup_verified_company_account_"+strconv.Itoa(userID)+"_"+relationDate, userID)
		if err != nil {
			_ = transaction.Rollback()
			log.Println(err)
			return
		}
		err = transaction.Commit()
		if err != nil {
			log.Println(err)
			return
		}
		service.sendMessageUserWithActionOnKeyboards(db, app, bot, userID, "The company registered successfully", false)
		SaveUserLastState(db, app, bot, text, userID, "done_setup_verified_company_account")
	}
}

func (service *events.BotService) insertFinalStateData(app *config.App, bot *tb.Bot, userID int, transaction *sql.Tx, channelTableData, companyTableData, channelsEmailSuffixes, channelsSettings []*models.TempSetupFlow, db *sql.DB) {
	if companyTableData == nil || channelsEmailSuffixes == nil || len(channelsEmailSuffixes) != 1 || channelTableData == nil || channelsSettings == nil {
		transaction.Rollback()
		log.Println("final data must not be null")
		return
	}

	//insert company
	var companyName, companyType string
	for _, v := range companyTableData {
		if v.ColumnName == "companyName" {
			companyName = v.Data
		}
		if v.ColumnName == "companyType" {
			companyType = v.Data
		}
	}
	companyResultsStatement, err := db.Prepare("SELECT id,companyName FROM `companies` where `companyName`=?")
	if err != nil {
		transaction.Rollback()
		log.Println(err)
		return
	}
	defer companyResultsStatement.Close()
	companyNewModel := new(models.Company)
	var companyID int64
	if err := companyResultsStatement.QueryRow(companyName).Scan(&companyNewModel.ID, &companyNewModel.CompanyName); err != nil {
		insertCompany, err := transaction.Exec("INSERT INTO `companies` (`companyName`,`companyType`,`createdAt`) VALUES('" + companyName + "','" + companyType + "','" + time.Now().UTC().Format("2006-01-02 03:04:05") + "')")
		if err != nil {
			_ = transaction.Rollback()
			log.Println(err)
			return
		}
		companyID, err = insertCompany.LastInsertId()
		if err != nil {
			_ = transaction.Rollback()
			log.Println(err)
			return
		}
	} else {
		companyID = companyNewModel.ID
	}

	//insert channel
	var channelType, manualChannelName, uniqueID, channelURL string
	for _, v := range channelTableData {
		if v.ColumnName == "channelType" {
			channelType = v.Data
		}
		if v.ColumnName == "manualChannelName" {
			manualChannelName = v.Data
		}
		if v.ColumnName == "uniqueID" {
			uniqueID = v.Data
		}
		if v.ColumnName == "channelURL" {
			channelURL = v.Data
		}
	}
	resultsStatement, err := db.Prepare("SELECT channelID,id FROM `channels` where `uniqueID`=?")
	if err != nil {
		transaction.Rollback()
		log.Println(err)
		return
	}
	defer resultsStatement.Close()
	channelModel := new(models.Channel)
	_ = resultsStatement.QueryRow(uniqueID).Scan(&channelModel.ChannelID, &channelModel.ID)
	_, err = transaction.Exec("update `channels` set `manualChannelName`='" + manualChannelName + "', `channelType`='" + channelType + "', `channelURL`='" + channelURL + "' where `uniqueID`='" + uniqueID + "'")
	if err != nil {
		transaction.Rollback()
		log.Println(err)
		return
	}

	//remove previous companies_channels, which create with channel id
	_, err = transaction.Exec("delete from `companies_channels` where `channelID`='" + strconv.FormatInt(channelModel.ID, 10) + "'")
	if err != nil {
		transaction.Rollback()
		log.Println(err)
		return
	}

	//insert company channel
	_, err = transaction.Exec("INSERT INTO `companies_channels` (`companyID`,`channelID`,`createdAt`) VALUES('" + strconv.FormatInt(companyID, 10) + "','" + strconv.FormatInt(channelModel.ID, 10) + "','" + time.Now().UTC().Format("2006-01-02 03:04:05") + "')")
	if err != nil {
		transaction.Rollback()
		log.Println(err)
		return
	}

	//remove previous company, which create with channel id
	_, err = transaction.Exec("delete from `companies` where `companyName`='" + channelModel.ChannelID + "'")
	if err != nil {
		transaction.Rollback()
		log.Println(err)
		return
	}

	//insert channelsEmailSuffixes
	emailSuffixed := channelsEmailSuffixes[0]
	if strings.Contains(emailSuffixed.Data, ",") {
		suffixes := strings.Split(emailSuffixed.Data, ",")
		for _, suffix := range suffixes {
			_, err := transaction.Exec("INSERT INTO `channels_email_suffixes` (`suffix`,`channelID`,`createdAt`) VALUES('" + suffix + "','" + strconv.FormatInt(channelModel.ID, 10) + "','" + time.Now().UTC().Format("2006-01-02 03:04:05") + "')")
			if err != nil {
				transaction.Rollback()
				log.Println(err)
				return
			}
		}
	} else {
		_, err := transaction.Exec("INSERT INTO `channels_email_suffixes` (`suffix`,`channelID`,`createdAt`) VALUES('" + emailSuffixed.Data + "','" + strconv.FormatInt(channelModel.ID, 10) + "','" + time.Now().UTC().Format("2006-01-02 03:04:05") + "')")
		if err != nil {
			_ = transaction.Rollback()
			log.Println(err)
			return
		}
	}

	//insert channel settings
	var joinVerify, newMessageVerify, replyVerify, directVerify string
	for _, v := range channelsSettings {
		if v.ColumnName == "joinVerify" {
			switch v.Data {
			case "Yes":
				joinVerify = "1"
			case "No":
				joinVerify = "0"
			}
		}
		if v.ColumnName == "newMessageVerify" {
			switch v.Data {
			case "Yes":
				newMessageVerify = "1"
			case "No":
				newMessageVerify = "0"
			}
		}
		if v.ColumnName == "replyVerify" {
			switch v.Data {
			case "Yes":
				replyVerify = "1"
			case "No":
				replyVerify = "0"
			}
		}
		if v.ColumnName == "directVerify" {
			switch v.Data {
			case "Yes":
				directVerify = "1"
			case "No":
				directVerify = "0"
			}
		}
	}
	_, err = transaction.Exec("INSERT INTO `channels_settings` (`joinVerify`,`newMessageVerify`,`replyVerify`,`directVerify`,`channelID`,`createdAt`) VALUES('" + joinVerify + "','" + newMessageVerify + "','" + replyVerify + "','" + directVerify + "','" + strconv.FormatInt(channelModel.ID, 10) + "','" + time.Now().UTC().Format("2006-01-02 03:04:05") + "')")
	if err != nil {
		transaction.Rollback()
		log.Println(err)
		return
	}
}
