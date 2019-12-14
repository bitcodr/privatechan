package controller

import (
	"github.com/amiraliio/tgbp/config"
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
				_, err = db.Query("INSERT INTO `temp_setup_flow` (`tableName`,`columnName`,`data`,`userID`,`relation`) VALUES ('" + tableName + "','" + columnName + "','" + strings.TrimSpace(text) + "','" + strconv.Itoa(userID) + "','setup_verified_company_account_" + strconv.Itoa(userID) + "_" + relationDate + "')")
				if err != nil {
					log.Println(err)
					return
				}
				if prevQuestionNo+1 > len(questions) {

					//TODO insert to final tables
					//TODO change status of temp_setup_flow

					sendMessageUserWithActionOnKeyboards(bot, userID, "The company channel created successfully, Please add this bot as admin in the created channel", false)
					SaveUserLastState(bot, text, userID, "done_setup_verified_company_account")
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
