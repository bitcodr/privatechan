package controllers

import (
	"database/sql"
	"github.com/amiraliio/tgbp/events"
	"github.com/amiraliio/tgbp/lang"
	"github.com/google/uuid"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/amiraliio/tgbp/config"
	"github.com/amiraliio/tgbp/models"
	"github.com/spf13/viper"
	tb "gopkg.in/tucnak/telebot.v2"
)

func (service event) RegisterGroup(app *config.App, bot *tb.Bot, m *tb.Message, request *events.Event) bool {
	db := app.DB()
	defer db.Close()
	if m.Sender != nil {
		SaveUserLastState(db, app, bot, m.Text, m.Sender.ID, request.UserState)
	}
	channelID := strconv.FormatInt(m.Chat.ID, 10)
	resultsStatement, err := db.Prepare("SELECT id FROM `channels` where channelID=?")
	if err != nil {
		log.Println(err)
		return true
	}
	defer resultsStatement.Close()
	results, err := resultsStatement.Query(channelID)
	if err != nil {
		log.Println(err)
		return true
	}
	if !results.Next() {
		//start transaction
		transaction, err := db.Begin()
		if err != nil {
			log.Println(err)
			return true
		}
		uniqueID := uuid.New().String()
		//insert channel
		channelInserted, err := transaction.Exec("INSERT INTO `channels` (`channelType`,`channelID`,`channelName`,`uniqueID`,`createdAt`,`updatedAt`) VALUES('group','" + channelID + "','" + m.Chat.Title + "','" + uniqueID + "','" + app.CurrentTime + "','" + app.CurrentTime + "')")
		if err != nil {
			transaction.Rollback()
			log.Println(err)
			return true
		}
		insertedChannelID, err := channelInserted.LastInsertId()
		if err == nil {
			//company name
			companyFlag := strconv.FormatInt(m.Chat.ID, 10)
			//check if company is not exist
			companyExistsStatement, err := db.Prepare("SELECT id FROM `companies` where `companyName`=?")
			if err != nil {
				transaction.Rollback()
				log.Println(err)
				return true
			}
			defer companyExistsStatement.Close()
			companyExists, err := companyExistsStatement.Query(companyFlag)
			if err != nil {
				transaction.Rollback()
				log.Println(err)
				return true
			}
			if !companyExists.Next() {
				//insert company
				companyInserted, err := transaction.Exec("INSERT INTO `companies` (`companyName`,`createdAt`,`updatedAt`) VALUES('" + companyFlag + "','" + app.CurrentTime + "','" + app.CurrentTime + "')")
				if err != nil {
					transaction.Rollback()
					log.Println(err)
					return true
				}
				insertedCompanyID, err := companyInserted.LastInsertId()
				if err == nil {
					companyModelID := strconv.FormatInt(insertedCompanyID, 10)
					channelModelID := strconv.FormatInt(insertedChannelID, 10)
					//insert company channel pivot
					_, err := transaction.Exec("INSERT INTO `companies_channels` (`companyID`,`channelID`,`createdAt`) VALUES('" + companyModelID + "','" + channelModelID + "','" + app.CurrentTime + "')")
					if err != nil {
						transaction.Rollback()
						log.Println(err)
						return true
					}
				}
			} else {
				companyModel := new(models.Company)
				if err := companyExists.Scan(&companyModel.ID); err != nil {
					transaction.Rollback()
					log.Println(err)
					return true
				}
				companyModelID := strconv.FormatInt(companyModel.ID, 10)
				channelModelID := strconv.FormatInt(insertedChannelID, 10)
				//insert company channel pivot
				_, err := transaction.Exec("INSERT INTO `companies_channels` (`companyID`,`channelID`,`createdAt`) VALUES('" + companyModelID + "','" + channelModelID + "','" + app.CurrentTime + "')")
				if err != nil {
					transaction.Rollback()
					log.Println(err)
					return true
				}
			}
			err = transaction.Commit()
			if err != nil {
				log.Println(err)
				return true
			}
			successMessage, _ := bot.Send(m.Chat, "You're group registered successfully")
			time.Sleep(2 * time.Second)
			if err := bot.Delete(successMessage); err != nil {
				log.Println(err)
				return true
			}
			sendOptionModel := new(tb.SendOptions)
			sendOptionModel.ParseMode = tb.ModeHTML
			_, err = bot.Send(m.Chat, "This is your group unique ID, you can save it and remove this message: <code> "+uniqueID+" </code>", sendOptionModel)
			if err != nil {
				log.Println(err)
				return true
			}
			time.Sleep(2 * time.Second)
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
			_, err = bot.Send(m.Chat, lang.StartGroup, newSendOption)
			if err != nil {
				log.Println(err)
				return true
			}
		}
	} else {
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
		_, err = bot.Send(m.Chat, lang.StartGroup, newSendOption)
		if err != nil {
			log.Println(err)
			return true
		}
		return true
	}
	return true
}

func (service event) NewMessageGroupHandler(app *config.App, bot *tb.Bot, m *tb.Message, request *events.Event) bool {
	if strings.Contains(m.Text, request.Command) {
		db := app.DB()
		defer db.Close()
		lastState := events.GetUserLastState(db, app, bot, m, m.Sender.ID)
		service.CheckUserRegisteredOrNot(db, app, bot, m, lastState, m.Text, m.Sender.ID)
		if m.Sender != nil {
			SaveUserLastState(db, app, bot, m.Text, m.Sender.ID, request.UserState)
		}
		channelID := strings.ReplaceAll(m.Text, request.Command1, "")
		service.JoinFromGroup(db, app, bot, m, channelID)
		resultsStatement, err := db.Prepare("SELECT `channelName` FROM `channels` where `channelID`=?")
		if err != nil {
			log.Println(err)
			return true
		}
		defer resultsStatement.Close()
		channelModel := new(models.Channel)
		if err := resultsStatement.QueryRow(channelID).Scan(&channelModel.ChannelName); err != nil {
			log.Println(err)
			return true
		}
		options := new(tb.SendOptions)
		markup := new(tb.ReplyMarkup)
		homeBTN := tb.ReplyButton{
			Text: "Home",
		}
		replyKeys := [][]tb.ReplyButton{
			[]tb.ReplyButton{homeBTN},
		}
		markup.ReplyKeyboard = replyKeys
		options.ReplyMarkup = markup
		bot.Send(m.Sender, "Please draft your anonymous new message to the group / channel: "+channelModel.ChannelName, options)
		return true
	}
	return false
}

func (service event) JoinFromGroup(db *sql.DB, app *config.App, bot *tb.Bot, m *tb.Message, channelID string) {
	userID := strconv.Itoa(m.Sender.ID)
	//check if user is not created
	resultsStatement, err := db.Prepare("SELECT id FROM `users` where `status`= 'ACTIVE' and `userID`=?")
	if err != nil {
		log.Println(err)
	}
	defer resultsStatement.Close()
	results, err := resultsStatement.Query(userID)
	if err != nil {
		log.Println(err)
	}
	//start transaction
	transaction, err := db.Begin()
	if err != nil {
		log.Println(err)
	}
	if !results.Next() {
		//insert user
		var isBotValue string
		if m.Sender.IsBot {
			isBotValue = "1"
		} else {
			isBotValue = "0"
		}
		userInsert, err := transaction.Exec("INSERT INTO `users` (`userID`,`username`,`firstName`,`lastName`,`lang`,`isBot`,`createdAt`,`updatedAt`) VALUES('" + userID + "','" + m.Sender.Username + "','" + m.Sender.FirstName + "','" + m.Sender.LastName + "','" + m.Sender.LanguageCode + "','" + isBotValue + "','" + app.CurrentTime + "','" + app.CurrentTime + "')")
		if err != nil {
			transaction.Rollback()
			log.Println(err)
		}
		insertedUserID, err := userInsert.LastInsertId()
		if err != nil {
			transaction.Rollback()
			log.Println(err)
		}
		service.checkAndInsertUserGroup(app, bot, m, insertedUserID, channelID, db, transaction)
	} else {
		userModel := new(models.User)
		if err := results.Scan(&userModel.ID); err != nil {
			transaction.Rollback()
			log.Println(err)
		}
		service.checkAndInsertUserGroup(app, bot, m, userModel.ID, channelID, db, transaction)
	}
}

func (service event) checkAndInsertUserGroup(app *config.App, bot *tb.Bot, m *tb.Message, queryUserID int64, channelID string, db *sql.DB, transaction *sql.Tx) {
	userModelID := strconv.FormatInt(queryUserID, 10)
	//check if channel for user is exists
	checkUserChannelStatement, err := db.Prepare("SELECT ch.id as id FROM `channels` as ch inner join `users_channels` as uc on uc.channelID = ch.id and uc.userID=? and ch.channelID=?")
	if err != nil {
		transaction.Rollback()
		log.Println(err)
	}
	defer checkUserChannelStatement.Close()
	checkUserChannel, err := checkUserChannelStatement.Query(userModelID, channelID)
	if err != nil {
		transaction.Rollback()
		log.Println(err)
	}
	if !checkUserChannel.Next() {
		getChannelIDStatement, err := db.Prepare("SELECT `id` FROM `channels` where `channelID`=?")
		if err != nil {
			transaction.Rollback()
			log.Println(err)
		}
		defer getChannelIDStatement.Close()
		getChannelID, err := getChannelIDStatement.Query(channelID)
		if err != nil {
			transaction.Rollback()
			log.Println(err)
		}
		if getChannelID.Next() {
			channelModel := new(models.Channel)
			if err := getChannelID.Scan(&channelModel.ID); err != nil {
				transaction.Rollback()
				log.Println(err)
			}
			channelModelID := strconv.FormatInt(channelModel.ID, 10)
			_, err := transaction.Exec("INSERT INTO `users_channels` (`status`,`userID`,`channelID`,`createdAt`,`updatedAt`) VALUES('ACTIVE','" + userModelID + "','" + channelModelID + "','" + time.Now().UTC().Format("2006-01-02 03:04:05") + "','" + time.Now().UTC().Format("2006-01-02 03:04:05") + "')")
			if err != nil {
				transaction.Rollback()
				log.Println(err)
			}
		}
	}
	channelDetailStatement, err := db.Prepare("SELECT ch.`id` as id, ch.channelName as channelName, co.companyName as companyName  FROM `channels` as ch inner join `companies_channels` as cc on ch.id = cc.channelID inner join `companies` as co on cc.companyID = co.id where ch.`channelID`=?")
	if err != nil {
		transaction.Rollback()
		log.Println(err)
	}
	defer channelDetailStatement.Close()
	channelDetail, err := channelDetailStatement.Query(channelID)
	if err != nil {
		transaction.Rollback()
		log.Println(err)
	}
	if channelDetail.Next() {
		channelModelData := new(models.Channel)
		companyModel := new(models.Company)
		if err := channelDetail.Scan(&channelModelData.ID, &channelModelData.ChannelName, &companyModel.CompanyName); err != nil {
			transaction.Rollback()
			log.Println(err)
		}
		channelModelID := strconv.FormatInt(channelModelData.ID, 10)
		_, err = transaction.Exec("update `users_current_active_channel` set `status`='INACTIVE' where `userID`='" + userModelID + "'")
		if err != nil {
			transaction.Rollback()
			log.Println(err)
		}
		//set user active channel
		_, err = transaction.Exec("INSERT INTO `users_current_active_channel` (`userID`,`channelID`,`createdAt`,`updatedAt`) VALUES('" + userModelID + "','" + channelModelID + "','" + time.Now().UTC().Format("2006-01-02 03:04:05") + "','" + time.Now().UTC().Format("2006-01-02 03:04:05") + "')")
		if err != nil {
			transaction.Rollback()
			log.Println(err)
		}
		err := transaction.Commit()
		if err != nil {
			log.Println(err)
			return
		}

	}
}
