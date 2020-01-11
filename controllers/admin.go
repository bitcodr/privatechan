package controllers

import (
	"database/sql"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/amiraliio/tgbp/config"
	"github.com/amiraliio/tgbp/models"
	tb "gopkg.in/tucnak/telebot.v2"
)

func (service *BotService) SetUpCompanyByAdmin(db *sql.DB, app *config.App, bot *tb.Bot, m *tb.Message, request *Event, lastState *models.UserLastState, text string, userID int) bool {
	if lastState.Data != "" && lastState.State == request.UserState {
		questions := config.QConfig.GetStringMap("SUPERADMIN.COMPANY.SETUP.QUESTIONS")
		numberOfQuestion := strings.Split(lastState.Data, "_")
		if len(numberOfQuestion) == 2 {
			questioNumber := numberOfQuestion[0]
			relationDate := numberOfQuestion[1]
			prevQuestionNo, err := strconv.Atoi(questioNumber)
			if err == nil {
				tableName := config.QConfig.GetString("SUPERADMIN.COMPANY.SETUP.QUESTIONS.N" + questioNumber + ".TABLE_NAME")
				columnName := config.QConfig.GetString("SUPERADMIN.COMPANY.SETUP.QUESTIONS.N" + questioNumber + ".COLUMN_NAME")
				_, err = db.Query("INSERT INTO `temp_setup_flow` (`tableName`,`columnName`,`data`,`userID`,`relation`,`createdAt`) VALUES ('" + tableName + "','" + columnName + "','" + strings.TrimSpace(text) + "','" + strconv.Itoa(userID) + "','" + config.LangConfig.GetString("STATE.SETUP_VERIFIED_COMPANY") + "_" + strconv.Itoa(userID) + "_" + relationDate + "','" + app.CurrentTime + "')")
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
	initQuestion := config.QConfig.GetString("SUPERADMIN.COMPANY.SETUP.QUESTIONS.N1.QUESTION")
	service.sendMessageUserWithActionOnKeyboards(db, app, bot, userID, initQuestion, true)
	SaveUserLastState(db, app, bot, "1_"+strconv.FormatInt(time.Now().Unix(), 10), userID, config.LangConfig.GetString("STATE.SETUP_VERIFIED_COMPANY"))
	return true
}

//next question
func (service *BotService) nextQuestion(db *sql.DB, app *config.App, bot *tb.Bot, m *tb.Message, lastState *models.UserLastState, relationDate string, prevQuestionNo int, text string, userID int) {
	questionText := config.QConfig.GetString("SUPERADMIN.COMPANY.SETUP.QUESTIONS.N" + strconv.Itoa(prevQuestionNo+1) + ".QUESTION")
	answers := config.QConfig.GetString("SUPERADMIN.COMPANY.SETUP.QUESTIONS.N" + strconv.Itoa(prevQuestionNo+1) + ".ANSWERS")
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
			Text: config.LangConfig.GetString("GENERAL.HOME"),
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
			Text: config.LangConfig.GetString("GENERAL.HOME"),
		}
		replyKeys := [][]tb.ReplyButton{
			[]tb.ReplyButton{homeBTN},
		}
		options := new(tb.SendOptions)
		replyMarkupModel := new(tb.ReplyMarkup)
		replyMarkupModel.ReplyKeyboard = replyKeys
		options.ReplyMarkup = replyMarkupModel
		bot.Send(userModel, questionText, options)
	}
	SaveUserLastState(db, app, bot, strconv.Itoa(prevQuestionNo+1)+"_"+relationDate, userID, config.LangConfig.GetString("STATE.SETUP_VERIFIED_COMPANY"))
}

func (service *BotService) sendMessageUserWithActionOnKeyboards(db *sql.DB, app *config.App, bot *tb.Bot, userID int, message string, showKeyboard bool) {
	userModel := new(tb.User)
	userModel.ID = userID
	homeBTN := tb.ReplyButton{
		Text: config.LangConfig.GetString("GENERAL.HOME"),
	}
	replyKeys := [][]tb.ReplyButton{
		[]tb.ReplyButton{homeBTN},
	}
	replyModel := new(tb.ReplyMarkup)
	replyModel.ReplyKeyboardRemove = showKeyboard
	replyModel.ReplyKeyboard = replyKeys
	options := new(tb.SendOptions)
	options.ReplyMarkup = replyModel
	bot.Send(userModel, message, options)
}

func (service *BotService) finalStage(app *config.App, bot *tb.Bot, relationDate string, db *sql.DB, text string, userID int) {
	results, err := db.Query("SELECT id,tableName,columnName,data,relation,status,userID,createdAt from `temp_setup_flow` where status='ACTIVE' and relation=? and userID=?", config.LangConfig.GetString("STATE.SETUP_VERIFIED_COMPANY")+"_"+strconv.Itoa(userID)+"_"+relationDate, userID)
	if err != nil {
		log.Println(err)
		return
	}
	defer results.Close()
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
			case config.LangConfig.GetString("GENERAL.COMPANIES"):
				companyTableData = append(companyTableData, tempSetupFlow)
			case config.LangConfig.GetString("GENERAL.CHANNELS"):
				channelTableData = append(channelTableData, tempSetupFlow)
			case config.LangConfig.GetString("GENERAL.CHANNELS_SETTINGS"):
				channelsSettings = append(channelsSettings, tempSetupFlow)
			case config.LangConfig.GetString("GENERAL.CHANNEL_EMAIL_SUFFIXES"):
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
		_, err = transaction.Exec("update `temp_setup_flow` set `status`='INACTIVE' where status='ACTIVE' and relation=? and `userID`=?", config.LangConfig.GetString("STATE.SETUP_VERIFIED_COMPANY")+"_"+strconv.Itoa(userID)+"_"+relationDate, userID)
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
		service.sendMessageUserWithActionOnKeyboards(db, app, bot, userID, config.LangConfig.GetString("MESSAGES.COMPANY_REGISTERED_SUCCESSFULLY"), false)
		SaveUserLastState(db, app, bot, text, userID, config.LangConfig.GetString("STATE.DONE_SETUP_VERIFIED_COMPANY"))
	}
}

func (service *BotService) insertFinalStateData(app *config.App, bot *tb.Bot, userID int, transaction *sql.Tx, channelTableData, companyTableData, channelsEmailSuffixes, channelsSettings []*models.TempSetupFlow, db *sql.DB) {
	if companyTableData == nil || channelsEmailSuffixes == nil || len(channelsEmailSuffixes) != 1 || channelTableData == nil || channelsSettings == nil {
		transaction.Rollback()
		log.Println(config.LangConfig.GetString("MESSAGES.DATA_MUST_NOT_BE_NULL"))
		return
	}

	//insert company
	var companyName, companyType string
	for _, v := range companyTableData {
		if v.ColumnName == config.LangConfig.GetString("GENERAL.COMPANY_NAME") {
			companyName = v.Data
		}
		if v.ColumnName == config.LangConfig.GetString("GENERAL.COMPANY_TYPE") {
			companyType = v.Data
		}
	}
	companyNewModel := new(models.Company)
	var companyID int64
	if err := db.QueryRow("SELECT id,companyName FROM `companies` where `companyName`=?", companyName).Scan(&companyNewModel.ID, &companyNewModel.CompanyName); err != nil {
		insertCompany, err := transaction.Exec("INSERT INTO `companies` (`companyName`,`companyType`,`createdAt`) VALUES(?,?,?)", companyName, companyType, app.CurrentTime)
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
	var channelModelField, manualChannelName, uniqueID, channelURL string
	for _, v := range channelTableData {
		if v.ColumnName == config.LangConfig.GetString("GENERAL.CHANNEL_MODEL") {
			channelModelField = v.Data
		}
		if v.ColumnName == config.LangConfig.GetString("GENERAL.MANUAL_CHANNEL_NAME") {
			manualChannelName = v.Data
		}
		if v.ColumnName == config.LangConfig.GetString("GENERAL.UNIQUE_ID") {
			uniqueID = v.Data
		}
		if v.ColumnName == config.LangConfig.GetString("GENERAL.CHANNEL_URL") {
			channelURL = v.Data
		}
	}
	channelModel := new(models.Channel)
	if err := db.QueryRow("SELECT channelID,id FROM `channels` where `uniqueID`=?", uniqueID).Scan(&channelModel.ChannelID, &channelModel.ID); err != nil {
		transaction.Rollback()
		log.Println(err)
		return
	}
	_, err := transaction.Exec("update `channels` set `manualChannelName`=? `channelModel`=? `channelURL`=? where `uniqueID`=?", manualChannelName, channelModelField, channelURL, uniqueID)
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
	_, err = transaction.Exec("INSERT INTO `companies_channels` (`companyID`,`channelID`,`createdAt`) VALUES('" + strconv.FormatInt(companyID, 10) + "','" + strconv.FormatInt(channelModel.ID, 10) + "','" + app.CurrentTime + "')")
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
			_, err := transaction.Exec("INSERT INTO `channels_email_suffixes` (`suffix`,`channelID`,`createdAt`) VALUES('" + suffix + "','" + strconv.FormatInt(channelModel.ID, 10) + "','" + app.CurrentTime + "')")
			if err != nil {
				transaction.Rollback()
				log.Println(err)
				return
			}
		}
	} else {
		_, err := transaction.Exec("INSERT INTO `channels_email_suffixes` (`suffix`,`channelID`,`createdAt`) VALUES('" + emailSuffixed.Data + "','" + strconv.FormatInt(channelModel.ID, 10) + "','" + app.CurrentTime + "')")
		if err != nil {
			_ = transaction.Rollback()
			log.Println(err)
			return
		}
	}

	//insert channel settings
	var joinVerify, newMessageVerify, replyVerify, directVerify string
	for _, v := range channelsSettings {
		if v.ColumnName == config.LangConfig.GetString("GENERAL.JOIN_VERIFY") {
			switch v.Data {
			case config.LangConfig.GetString("GENERAL.YES_TEXT"):
				joinVerify = "1"
			case config.LangConfig.GetString("GENERAL.NO_TEXT"):
				joinVerify = "0"
			}
		}
		if v.ColumnName == config.LangConfig.GetString("GENERAL.NEW_MESSAGE_VERIFY") {
			switch v.Data {
			case config.LangConfig.GetString("GENERAL.YES_TEXT"):
				newMessageVerify = "1"
			case config.LangConfig.GetString("GENERAL.NO_TEXT"):
				newMessageVerify = "0"
			}
		}
		if v.ColumnName == config.LangConfig.GetString("GENERAL.REPLY_VERIFY") {
			switch v.Data {
			case config.LangConfig.GetString("GENERAL.YES_TEXT"):
				replyVerify = "1"
			case config.LangConfig.GetString("GENERAL.NO_TEXT"):
				replyVerify = "0"
			}
		}
		if v.ColumnName == "directVerify" {
			switch v.Data {
			case config.LangConfig.GetString("GENERAL.YES_TEXT"):
				directVerify = "1"
			case config.LangConfig.GetString("GENERAL.NO_TEXT"):
				directVerify = "0"
			}
		}
	}
	_, err = transaction.Exec("INSERT INTO `channels_settings` (`joinVerify`,`newMessageVerify`,`replyVerify`,`directVerify`,`channelID`,`createdAt`) VALUES('" + joinVerify + "','" + newMessageVerify + "','" + replyVerify + "','" + directVerify + "','" + strconv.FormatInt(channelModel.ID, 10) + "','" + app.CurrentTime + "')")
	if err != nil {
		transaction.Rollback()
		log.Println(err)
		return
	}
}
