package controller

import (
	"database/sql"
	"github.com/amiraliio/tgbp/config"
	"github.com/amiraliio/tgbp/lang"
	"github.com/amiraliio/tgbp/model"
	"github.com/spf13/viper"
	tb "gopkg.in/tucnak/telebot.v2"
	"log"
	"strconv"
	"strings"
	"time"
)

func SetUpCompanyByAdmin(bot *tb.Bot, m *tb.Message, lastState *model.UserLastState, text string, userID int) {
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
				db, err := config.DB()
				if err != nil {
					log.Println(err)
					return
				}
				defer db.Close()
				var textValue string
				if columnName != "channelID" {
					textValue = strings.TrimSpace(text)
				} else {
					textValue = strconv.FormatInt(m.OriginalChat.ID, 10)
					if !checkIsBotIsChannelAdminOrNot(bot, m, userID) {
						return
					}
					if !checkIsBotIsGroupAdminOrNot(bot, m, userID) {
						return
					}
				}
				_, err = db.Query("INSERT INTO `temp_setup_flow` (`tableName`,`columnName`,`data`,`userID`,`relation`,`createdAt`) VALUES ('" + tableName + "','" + columnName + "','" + textValue + "','" + strconv.Itoa(userID) + "','setup_verified_company_account_" + strconv.Itoa(userID) + "_" + relationDate + "','" + time.Now().UTC().Format("2006-01-02 03:04:05") + "')")
				if err != nil {
					log.Println(err)
					return
				}
				if prevQuestionNo+1 > len(questions) {
					finalStage(bot, relationDate, db, text, userID)
					return
				}
				nextQuestion(bot, m, lastState, relationDate, prevQuestionNo, text, userID)
			}
		}
		return
	}
	initQuestion := viper.GetString("SUPERADMIN.COMPANY.SETUP.QUESTIONS.N1.QUESTION")
	sendMessageUserWithActionOnKeyboards(bot, userID, initQuestion, true)
	SaveUserLastState(bot, "1_"+strconv.FormatInt(time.Now().Unix(), 10), userID, "setup_verified_company_account")
}

func checkIsBotIsChannelAdminOrNot(bot *tb.Bot, m *tb.Message, userID int) bool {
	if m.OriginalChat.Type == tb.ChatChannelPrivate || m.OriginalChat.Type == tb.ChatChannel {
		members, err := bot.AdminsOf(m.OriginalChat)
		if err != nil {
			log.Println(err)
			userModel := new(tb.User)
			userModel.ID = userID
			_, _ = bot.Send(userModel, "Bot Is Not Admin, Please add this bot as admin in the channel and then forward a message from the channel")
			return false
		}
		if members == nil {
			userModel := new(tb.User)
			userModel.ID = userID
			_, _ = bot.Send(userModel, "Bot Is Not Admin, Please add this bot as admin in the channel and then forward a message from the channel")
			return false
		}
		var isAdmin bool
		for _, v := range members {
			isAdmin = false
			if v.User.ID == viper.GetInt("APP.TELEGRAM_BOT_ID") {
				isAdmin = true
				break
			}
		}
		if !isAdmin {
			userModel := new(tb.User)
			userModel.ID = userID
			_, _ = bot.Send(userModel, "Bot Is Not Admin, Please add this bot as admin in the channel and then forward a message from the channel")
			return false
		}
		return true
	}
	return true
}

func checkIsBotIsGroupAdminOrNot(bot *tb.Bot, m *tb.Message, userID int) bool {
	if m.OriginalChat.Type == tb.ChatGroup || m.OriginalChat.Type == tb.ChatSuperGroup {
		admins, err := bot.AdminsOf(m.OriginalChat)
		if err != nil {
			log.Println(err)
			userModel := new(tb.User)
			userModel.ID = userID
			_, _ = bot.Send(userModel, "Bot Is Not Admin OR Member ,Please add this bot as admin in the group and then forward a message from the group")
			return false
		}
		userModel := new(tb.User)
		userModel.ID = viper.GetInt("APP.TELEGRAM_BOT_ID")
		members, err := bot.ChatMemberOf(m.OriginalChat, userModel)
		if err != nil {
			log.Println(err)
			userModel := new(tb.User)
			userModel.ID = userID
			_, _ = bot.Send(userModel, "Bot Is Not Admin OR Member ,Please add this bot as admin in the group and then forward a message from the group")
			return false
		}
		if admins == nil && members == nil {
			userModel := new(tb.User)
			userModel.ID = userID
			_, _ = bot.Send(userModel, "Bot Is Not Admin OR Member ,Please add this bot as admin in the group and then forward a message from the group")
			return false
		}
		var isAdmin bool
		for _, v := range admins {
			isAdmin = false
			if v.User.ID == viper.GetInt("APP.TELEGRAM_BOT_ID") {
				isAdmin = true
				break
			}
		}
		if !isAdmin && members == nil {
			userModel := new(tb.User)
			userModel.ID = userID
			_, _ = bot.Send(userModel, "Bot Is Not Admin OR Member ,Please add this bot as admin in the group and then forward a message from the group")
			return false
		}
		return true
	}
	return true
}

//next question
func nextQuestion(bot *tb.Bot, m *tb.Message, lastState *model.UserLastState, relationDate string, prevQuestionNo int, text string, userID int) {
	questionText := viper.GetString("SUPERADMIN.COMPANY.SETUP.QUESTIONS.N" + strconv.Itoa(prevQuestionNo+1) + ".QUESTION")
	answers := viper.GetString("SUPERADMIN.COMPANY.SETUP.QUESTIONS.N" + strconv.Itoa(prevQuestionNo+1) + ".ANSWERS")
	if answers != "" && strings.Contains(strings.TrimSpace(answers), ",") {
		splittedAnswers := strings.Split(answers, ",")
		inlineKeysNested := []tb.InlineButton{}
		for _, v := range splittedAnswers {
			inlineBTN := tb.InlineButton{
				Text:   v,
				Unique: v,
			}
			inlineKeysNested = append(inlineKeysNested, inlineBTN)
		}
		inlineKeys := [][]tb.InlineButton{
			inlineKeysNested,
		}
		userModel := new(tb.User)
		userModel.ID = userID
		options := new(tb.SendOptions)
		replyMarkupModel := new(tb.ReplyMarkup)
		replyMarkupModel.InlineKeyboard = inlineKeys
		options.ReplyMarkup = replyMarkupModel
		_, _ = bot.Send(userModel, questionText, options)
	} else {
		userModel := new(tb.User)
		userModel.ID = userID
		_, _ = bot.Send(userModel, questionText)
	}
	SaveUserLastState(bot, strconv.Itoa(prevQuestionNo+1)+"_"+relationDate, userID, "setup_verified_company_account")
}

func sendMessageUserWithActionOnKeyboards(bot *tb.Bot, userID int, message string, showKeyboard bool) {
	userModel := new(tb.User)
	userModel.ID = userID
	options := new(tb.SendOptions)
	replyModel := new(tb.ReplyMarkup)
	replyModel.ReplyKeyboardRemove = showKeyboard
	options.ReplyMarkup = replyModel
	_, _ = bot.Send(userModel, message, options)
}

func finalStage(bot *tb.Bot, relationDate string, db *sql.DB, text string, userID int) {
	tempData, err := db.Prepare("SELECT id,tableName,columnName,data,relation,status,userID,createdAt from `temp_setup_flow` where status='ACTIVE' and relation=? and userID=?")
	if err != nil {
		log.Println(err)
		return
	}
	defer tempData.Close()
	results, err := tempData.Query("setup_verified_company_account_"+strconv.Itoa(userID)+"_"+relationDate, userID)
	if err == nil {
		var channelTableData []*model.TempSetupFlow
		var companyTableData []*model.TempSetupFlow
		var companies_email_suffixes []*model.TempSetupFlow
		var channels_settings []*model.TempSetupFlow
		for results.Next() {
			tempSetupFlow := new(model.TempSetupFlow)
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
				channels_settings = append(channels_settings, tempSetupFlow)
			case "companies_email_suffixes":
				companies_email_suffixes = append(companies_email_suffixes, tempSetupFlow)
			}
		}
		transaction, err := db.Begin()
		if err == nil {
			//insert company
			insertFinalStateData(bot, userID, transaction, channelTableData, companyTableData, companies_email_suffixes, channels_settings)
			//update state of temp setup data
			_, err = transaction.Exec("update `temp_setup_flow` set `status`='INACTIVE' where status='ACTIVE' and relation=? and `userID`=?", "setup_verified_company_account_"+strconv.Itoa(userID)+"_"+relationDate, userID)
			if err != nil {
				transaction.Rollback()
				log.Println(err)
				return
			}
			transaction.Commit()
			sendMessageUserWithActionOnKeyboards(bot, userID, "The company registered successfully", false)
			SaveUserLastState(bot, text, userID, "done_setup_verified_company_account")
		}
	}
}

func insertFinalStateData(bot *tb.Bot, userID int, transaction *sql.Tx, channelTableData, companyTableData, companies_email_suffixes, channels_settings []*model.TempSetupFlow) {
	if companyTableData == nil || companies_email_suffixes == nil || len(companies_email_suffixes) != 1 || channelTableData == nil || channels_settings == nil {
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
	insertCompany, err := transaction.Exec("INSERT INTO `companies` (`companyName`,`companyType`,`createdAt`) VALUES('" + companyName + "','" + companyType + "','" + time.Now().UTC().Format("2006-01-02 03:04:05") + "')")
	if err != nil {
		transaction.Rollback()
		log.Println(err)
		return
	}
	companyID, err := insertCompany.LastInsertId()
	if err != nil {
		transaction.Rollback()
		log.Println(err)
		return
	}

	//insert company_email_suffixes
	emailSuffixed := companies_email_suffixes[0]
	if strings.Contains(emailSuffixed.Data, ",") {
		suffixes := strings.Split(emailSuffixed.Data, ",")
		for _, suffix := range suffixes {
			_, err := transaction.Exec("INSERT INTO `companies_email_suffixes` (`suffix`,`companyID`,`createdAt`) VALUES('" + suffix + "','" + strconv.FormatInt(companyID, 10) + "','" + time.Now().UTC().Format("2006-01-02 03:04:05") + "')")
			if err != nil {
				transaction.Rollback()
				log.Println(err)
				return
			}
		}
	}

	//insert channel
	var channelType, channelName, channelURL, channelID string
	for _, v := range channelTableData {
		if v.ColumnName == "channelType" {
			channelType = v.Data
		}
		if v.ColumnName == "channelName" {
			channelName = v.Data
		}
		if v.ColumnName == "channelURL" {
			channelURL = v.Data
		}
		if v.ColumnName == "channelID" {
			channelID = v.Data
		}
	}
	insertChannel, err := transaction.Exec("INSERT INTO `channels` (`channelURL`,`channelName`,`channelType`,`channelID`,`createdAt`) VALUES('" + channelURL + "','" + channelName + "','" + channelType + "','" + channelID + "','" + time.Now().UTC().Format("2006-01-02 03:04:05") + "')")
	if err != nil {
		transaction.Rollback()
		log.Println(err)
		return
	}
	channelInsertedID, err := insertChannel.LastInsertId()
	if err != nil {
		transaction.Rollback()
		log.Println(err)
		return
	}

	//insert channel settings
	var joinVerify, newMessageVerify, replyVerify, directVerify string
	for _, v := range channels_settings {
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
	_, err = transaction.Exec("INSERT INTO `channels_settings` (`joinVerify`,`newMessageVerify`,`replyVerify`,`directVerify`,`channelID`,`createdAt`) VALUES('" + joinVerify + "','" + newMessageVerify + "','" + replyVerify + "','" + directVerify + "','" + strconv.FormatInt(channelInsertedID, 10) + "','" + time.Now().UTC().Format("2006-01-02 03:04:05") + "')")
	if err != nil {
		transaction.Rollback()
		log.Println(err)
		return
	}

	//send  message to channel to start conversation
	compose := tb.InlineButton{
		Unique: "compose_message_in_group_" + channelID,
		Text:   "📝 New Anonymous Message 👻",
		URL:    "https://t.me/" + viper.GetString("APP.BOTUSERNAME") + "?start=compose_message_in_group_" + channelID,
	}
	groupKeys := [][]tb.InlineButton{
		[]tb.InlineButton{compose},
	}
	newReplyModel := new(tb.ReplyMarkup)
	newReplyModel.InlineKeyboard = groupKeys
	newSendOption := new(tb.SendOptions)
	newSendOption.ReplyMarkup = newReplyModel
	newSendOption.ParseMode = tb.ModeMarkdown
	idOfChannel, err := strconv.Atoi(channelID)
	if err != nil {
		transaction.Rollback()
		log.Println(err)
	}
	user := new(tb.User)
	user.ID = idOfChannel
	_, err = bot.Send(user, lang.StartGroup, newSendOption)
	if err != nil {
		transaction.Rollback()
		log.Println(err)
	}
}
