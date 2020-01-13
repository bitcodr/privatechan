package controllers

import (
	"database/sql"
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/amiraliio/tgbp/config"
	"github.com/amiraliio/tgbp/helpers"
	"github.com/amiraliio/tgbp/lang"
	"github.com/amiraliio/tgbp/models"
	tb "gopkg.in/tucnak/telebot.v2"
)

func (service *BotService) RegisterGroup(app *config.App, bot *tb.Bot, m *tb.Message, request *Event) bool {
	db := app.DB()
	defer db.Close()
	if m.Sender != nil {
		SaveUserLastState(db, app, bot, m.Text, m.Sender.ID, request.UserState)
	}
	channelID := strconv.FormatInt(m.Chat.ID, 10)
	channelModel := new(models.Channel)
	err := db.QueryRow("SELECT id FROM `channels` where channelID=?", channelID).Scan(&channelModel.ID)
	if errors.Is(err, sql.ErrNoRows) {
		//start transaction
		transaction, err := db.Begin()
		if err != nil {
			log.Println(err)
			return true
		}
		uniqueID := uuid.New().String()
		//insert channel
		channelInserted, err := transaction.Exec("INSERT INTO `channels` (`channelType`,`channelID`,`channelName`,`uniqueID`,`createdAt`,`updatedAt`) VALUES(?,?,?,?,?,?)", "group", channelID, m.Chat.Title, uniqueID, app.CurrentTime, app.CurrentTime)
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
			companyModel := new(models.Company)
			err := db.QueryRow("SELECT id FROM `companies` where `companyName`=?", companyFlag).Scan(&companyModel.ID)
			if errors.Is(err, sql.ErrNoRows) {
				//insert company
				companyInserted, err := transaction.Exec("INSERT INTO `companies` (`companyName`,`createdAt`,`updatedAt`) VALUES(?,?,?)", companyFlag, app.CurrentTime, app.CurrentTime)
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
					_, err := transaction.Exec("INSERT INTO `companies_channels` (`companyID`,`channelID`,`createdAt`) VALUES(?,?,?)", companyModelID, channelModelID, app.CurrentTime)
					if err != nil {
						transaction.Rollback()
						log.Println(err)
						return true
					}
				}
			} else {
				companyModelID := strconv.FormatInt(companyModel.ID, 10)
				channelModelID := strconv.FormatInt(insertedChannelID, 10)
				//insert company channel pivot
				_, err := transaction.Exec("INSERT INTO `companies_channels` (`companyID`,`channelID`,`createdAt`) VALUES(?,?,?)", companyModelID, channelModelID, app.CurrentTime)
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
			successMessage, _ := bot.Send(m.Chat, config.LangConfig.GetString("MESSAGES.GROUP_REGISTERED_SUCCESSFULLY"))
			time.Sleep(2 * time.Second)
			if err := bot.Delete(successMessage); err != nil {
				log.Println(err)
				return true
			}
			sendOptionModel := new(tb.SendOptions)
			sendOptionModel.ParseMode = tb.ModeHTML
			_, err = bot.Send(m.Chat, config.LangConfig.GetString("MESSAGES.GROUP_UNIQUE_ID_MESSAGE")+"<code> "+uniqueID+" </code>", sendOptionModel)
			if err != nil {
				log.Println(err)
				return true
			}
			time.Sleep(2 * time.Second)
			compose := tb.InlineButton{
				Unique: config.LangConfig.GetString("STATE.COMPOSE_MESSAGE") + "_" + channelID,
				Text:   config.LangConfig.GetString("MESSAGES.COMPOSE_MESSAGE"),
				URL:    app.TgDomain + app.BotUsername + "?start=" + config.LangConfig.GetString("STATE.COMPOSE_MESSAGE") + "_" + channelID,
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
			Unique: config.LangConfig.GetString("STATE.COMPOSE_MESSAGE") + "_" + channelID,
			Text:   config.LangConfig.GetString("MESSAGES.COMPOSE_MESSAGE"),
			URL:    app.TgDomain + app.BotUsername + "?start=" + config.LangConfig.GetString("STATE.COMPOSE_MESSAGE") + "_" + channelID,
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

func (service *BotService) NewMessageGroupHandler(app *config.App, bot *tb.Bot, m *tb.Message, request *Event) bool {
	if strings.Contains(m.Text, request.Command) {
		db := app.DB()
		defer db.Close()
		service.CheckIfBotIsAdmin(app, bot, m, db, request)
		lastState := GetUserLastState(db, app, bot, m, m.Sender.ID)
		if service.CheckUserRegisteredOrNot(db, app, bot, m, request, lastState, m.Text, m.Sender.ID) {
			return true
		}
		if m.Sender != nil {
			SaveUserLastState(db, app, bot, m.Text, m.Sender.ID, request.UserState)
		}
		channelID := strings.ReplaceAll(m.Text, request.Command1, "")
		service.JoinFromGroup(db, app, bot, m, channelID)
		channelModel := new(models.Channel)
		if err := db.QueryRow("SELECT `channelName`,`channelType` FROM `channels` where `channelID`=?", channelID).Scan(&channelModel.ChannelName, &channelModel.ChannelType); err != nil {
			log.Println(err)
			return true
		}
		options := new(tb.SendOptions)
		markup := new(tb.ReplyMarkup)
		homeBTN := tb.ReplyButton{
			Text: config.LangConfig.GetString("GENERAL.HOME"),
		}
		replyKeys := [][]tb.ReplyButton{
			[]tb.ReplyButton{homeBTN},
		}
		markup.ReplyKeyboard = replyKeys
		options.ReplyMarkup = markup
		bot.Send(m.Sender, config.LangConfig.GetString("MESSAGES.PLEASE_DRAFT_YOUR_MESSAGE")+channelModel.ChannelType+" "+channelModel.ChannelName, options)
		return true
	}
	return false
}

func (service *BotService) NewMessageGroupHandlerCallback(app *config.App, bot *tb.Bot, c *tb.Callback, request *Event) bool {
	if strings.Contains(c.Data, request.Command) {
		db := app.DB()
		defer db.Close()
		lastState := GetUserLastState(db, app, bot, c.Message, c.Sender.ID)
		if service.CheckUserRegisteredOrNot(db, app, bot, c.Message, request, lastState, c.Data, c.Sender.ID) {
			return true
		}
		if c.Sender != nil {
			SaveUserLastState(db, app, bot, c.Data, c.Sender.ID, request.UserState)
		}
		channelID := strings.ReplaceAll(c.Data, request.Command, "")
		channelModel := new(models.Channel)
		if err := db.QueryRow("SELECT `channelName`,`channelType` FROM `channels` where `channelID`=?", strings.TrimLeft(channelID, "\f")).Scan(&channelModel.ChannelName, &channelModel.ChannelType); err != nil {
			log.Println(err)
			return true
		}
		options := new(tb.SendOptions)
		markup := new(tb.ReplyMarkup)
		homeBTN := tb.ReplyButton{
			Text: config.LangConfig.GetString("GENERAL.HOME"),
		}
		replyKeys := [][]tb.ReplyButton{
			[]tb.ReplyButton{homeBTN},
		}
		markup.ReplyKeyboard = replyKeys
		options.ReplyMarkup = markup
		bot.Send(c.Sender, config.LangConfig.GetString("MESSAGES.PLEASE_DRAFT_YOUR_MESSAGE")+channelModel.ChannelType+" "+channelModel.ChannelName, options)
		return true
	}
	return false
}

func (service *BotService) JoinFromGroup(db *sql.DB, app *config.App, bot *tb.Bot, m *tb.Message, channelID string) {
	userID := strconv.Itoa(m.Sender.ID)
	//check if user is not created
	userModel := new(models.User)
	err := db.QueryRow("SELECT id FROM `users` where `status`= 'ACTIVE' and `userID`=?", userID).Scan(&userModel.ID)
	if errors.Is(err, sql.ErrNoRows) {
		//insert user
		var isBotValue string
		if m.Sender.IsBot {
			isBotValue = "1"
		} else {
			isBotValue = "0"
		}
		transaction, err := db.Begin()
		if err != nil {
			log.Println(err)
			return
		}
		userInsert, err := transaction.Exec("INSERT INTO `users` (`userID`,`username`,`firstName`,`lastName`,`lang`,`isBot`,`customID`,`createdAt`,`updatedAt`) VALUES(?,?,?,?,?,?,?,?,?)", userID, m.Sender.Username, m.Sender.FirstName, m.Sender.LastName, m.Sender.LanguageCode, isBotValue, helpers.Hash(userID+channelID), app.CurrentTime, app.CurrentTime)
		if err != nil {
			transaction.Rollback()
			log.Println(err)
			return
		}
		insertedUserID, err := userInsert.LastInsertId()
		if err != nil {
			transaction.Rollback()
			log.Println(err)
			return
		}
		service.checkAndInsertUserGroup(app, bot, m, insertedUserID, channelID, db, transaction)
	} else {
		transaction, err := db.Begin()
		if err != nil {
			log.Println(err)
			return
		}
		service.checkAndInsertUserGroup(app, bot, m, userModel.ID, channelID, db, transaction)
	}
}

func (service *BotService) checkAndInsertUserGroup(app *config.App, bot *tb.Bot, m *tb.Message, queryUserID int64, channelID string, db *sql.DB, transaction *sql.Tx) {
	userModelID := strconv.FormatInt(queryUserID, 10)
	//check if channel for user is exists
	channelModel := new(models.Channel)
	err := db.QueryRow("SELECT ch.id as id FROM `channels` as ch inner join `users_channels` as uc on uc.channelID = ch.id and uc.userID=? and ch.channelID=?", userModelID, channelID).Scan(&channelModel.ID)
	if errors.Is(err, sql.ErrNoRows) {
		channelModel := new(models.Channel)
		if err := db.QueryRow("SELECT `id` FROM `channels` where `channelID`=?", channelID).Scan(&channelModel.ID); err != nil {
			transaction.Rollback()
			log.Println(err)
			return
		}
		channelModelID := strconv.FormatInt(channelModel.ID, 10)
		_, err := transaction.Exec("INSERT INTO `users_channels` (`userID`,`channelID`,`createdAt`,`updatedAt`) VALUES(?,?,?,?)", userModelID, channelModelID, app.CurrentTime, app.CurrentTime)
		if err != nil {
			transaction.Rollback()
			log.Println(err)
			return
		}
	}
	channelModelData := new(models.Channel)
	companyModel := new(models.Company)
	if err := db.QueryRow("SELECT ch.`id` as id, ch.channelName as channelName, co.companyName as companyName  FROM `channels` as ch inner join `companies_channels` as cc on ch.id = cc.channelID inner join `companies` as co on cc.companyID = co.id where ch.`channelID`=?", channelID).Scan(&channelModelData.ID, &channelModelData.ChannelName, &companyModel.CompanyName); err != nil {
		transaction.Rollback()
		log.Println(err)
		return
	}
	channelModelID := strconv.FormatInt(channelModelData.ID, 10)
	_, err = transaction.Exec("update `users_current_active_channel` set `status`='INACTIVE' where `userID`=?", userModelID)
	if err != nil {
		transaction.Rollback()
		log.Println(err)
		return
	}
	//set user active channel
	_, err = transaction.Exec("INSERT INTO `users_current_active_channel` (`userID`,`channelID`,`createdAt`,`updatedAt`) VALUES(?,?,?,?)", userModelID, channelModelID, app.CurrentTime, app.CurrentTime)
	if err != nil {
		transaction.Rollback()
		log.Println(err)
		return
	}
	err = transaction.Commit()
	if err != nil {
		log.Println(err)
		return
	}
}

func (service *BotService) UpdateGroupTitle(app *config.App, bot *tb.Bot, m *tb.Message, request *Event) bool {
	db := app.DB()
	defer db.Close()
	if m.Sender != nil {
		SaveUserLastState(db, app, bot, m.Text, m.Sender.ID, request.UserState)
	}
	_, err := db.Query("update channels set channelName=? where channelID=?", m.NewGroupTitle, m.Chat.ID)
	if err != nil {
		log.Println(err)
		return true
	}
	return true
}

func (service *BotService) CheckIfBotIsAdmin(app *config.App, bot *tb.Bot, m *tb.Message, db *sql.DB, request *Event) {
	ids := strings.TrimPrefix(m.Text, request.Command1)
	var channelID string
	if strings.Contains(ids, "_") {
		data := strings.Split(ids, "_")
		if len(data) > 0 {
			channelID = strings.TrimSpace(data[0])
		}
	} else {
		channelID = ids
	}
	chatModel := new(tb.Chat)
	id, err := strconv.ParseInt(channelID, 10, 0)
	if err != nil {
		log.Println(err)
		return
	}
	chatModel.ID = id
	admins, err := bot.AdminsOf(chatModel)
	if err != nil {
		log.Println(err)
		return
	}
	if len(admins) > 0 {
		for _, admin := range admins {
			if admin.User.ID == bot.Me.ID {
				channelModel := new(models.Channel)
				row := db.QueryRow("select id from `channels` where `channelID`=? and (`channelURL` is NULL OR `channelURL` = '')", id)
				if err := row.Scan(channelModel.ID); err == nil {
					inviteLink, err := bot.GetInviteLink(chatModel)
					if err != nil {
						log.Println(err)
						return
					}
					_, err = db.Query("update channels set channelURL=? where channelID=?", inviteLink, id)
					if err != nil {
						log.Println(err)
						return
					}
				}
			}
		}
	}
}
