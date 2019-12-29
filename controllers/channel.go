//Package controllers ...
package controllers

import (
	"database/sql"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/amiraliio/tgbp/lang"
	"github.com/google/uuid"

	"github.com/amiraliio/tgbp/config"
	"github.com/amiraliio/tgbp/models"
	tb "gopkg.in/tucnak/telebot.v2"
)

//TODO change query to queryRow
//TODO change value of queries to ?

//RegisterChannel
func (service *BotService) RegisterChannel(app *config.App, bot *tb.Bot, m *tb.Message, request *Event) bool {
	if strings.TrimSpace(m.Text) == request.Command {
		db := app.DB()
		defer db.Close()
		if m.Sender != nil {
			SaveUserLastState(db, app, bot, m.Text, m.Sender.ID, request.UserState)
		}
		//channel private url
		inviteLink, err := bot.GetInviteLink(m.Chat)
		if err != nil {
			log.Println(err)
			return true
		}
		channelURL := inviteLink
		channelID := strconv.FormatInt(m.Chat.ID, 10)
		statement, err := db.Prepare("SELECT id FROM `channels` where channelID=?")
		if err != nil {
			log.Println(err)
			return true
		}
		defer statement.Close()
		results, err := statement.Query(channelID)
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
			channelInserted, err := transaction.Exec("INSERT INTO `channels` (`channelType`,`channelURL`,`channelID`,`channelName`,`uniqueID`,`createdAt`,`updatedAt`) VALUES('channel','" + channelURL + "','" + channelID + "','" + m.Chat.Title + "','" + uniqueID + "','" + app.CurrentTime + "','" + app.CurrentTime + "')")
			if err != nil {
				transaction.Rollback()
				log.Println(err)
				return true
			}
			insertedChannelID, err := channelInserted.LastInsertId()
			if err == nil {
				//company name
				companyFlag := channelID
				//check if company is not exist
				companyStatement, err := db.Prepare("SELECT id FROM `companies` where `companyName`=?")
				if err != nil {
					transaction.Rollback()
					log.Println(err)
					return true
				}
				defer companyStatement.Close()
				companyExists, err := companyStatement.Query(companyFlag)
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
				transaction.Commit()
				successMessage, _ := bot.Send(m.Chat, "You're channel registered successfully")
				time.Sleep(2 * time.Second)
				if err := bot.Delete(successMessage); err != nil {
					log.Println(err)
					return true
				}
				sendOptionModel := new(tb.SendOptions)
				sendOptionModel.ParseMode = tb.ModeHTML
				_, err = bot.Send(m.Chat, "This is your channel unique ID, you can save it and remove this message: <code> "+uniqueID+" </code>", sendOptionModel)
				if err != nil {
					log.Println(err)
					return true
				}
				time.Sleep(2 * time.Second)
				compose := tb.InlineButton{
					Unique: "compose_message_in_group_" + channelID,
					Text:   "📝 New Anonymous Message 👻",
					URL:    "https://t.me/" + app.BotUsername + "?start=compose_message_in_group_" + channelID,
				}
				groupKeys := [][]tb.InlineButton{
					[]tb.InlineButton{compose},
				}
				newReplyModel := new(tb.ReplyMarkup)
				newReplyModel.InlineKeyboard = groupKeys
				newSendOption := new(tb.SendOptions)
				newSendOption.ReplyMarkup = newReplyModel
				newSendOption.ParseMode = tb.ModeMarkdown
				pinMessage, err := bot.Send(m.Chat, lang.StartGroup, newSendOption)
				if err != nil {
					log.Println(err)
					return true
				}
				if err := bot.Pin(pinMessage); err != nil {
					log.Println(err)
					return true
				}
				if err := bot.Delete(m); err != nil {
					log.Println(err)
					return true
				}
			}
		}
		return true
	}
	return false
}

func (service *BotService) SendReply(app *config.App, bot *tb.Bot, m *tb.Message, request *Event) bool {
	if strings.Contains(m.Text, request.Command) {
		db := app.DB()
		defer db.Close()
		if m.Sender != nil {
			SaveUserLastState(db, app, bot, m.Text, m.Sender.ID, request.UserState)
		}
		ids := strings.TrimPrefix(m.Text, request.Command1)
		data := strings.Split(ids, "_")
		channelID := strings.TrimSpace(data[0])
		messageID := strings.TrimSpace(data[2])
		service.JoinFromGroup(db, app, bot, m, channelID)
		resultsStatement, err := db.Prepare("SELECT ch.channelName,me.message FROM `channels` as ch inner join messages as me on ch.id=me.channelID and me.botMessageID=? where ch.channelID=?")
		if err != nil {
			log.Println(err)
			return true
		}
		defer resultsStatement.Close()
		channelModel := new(models.Channel)
		messageModel := new(models.Message)
		if err := resultsStatement.QueryRow(messageID, channelID).Scan(&channelModel.ChannelName, &messageModel.Message); err != nil {
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
		_, err = bot.Send(m.Sender, "Please send your reply to the message: '"+messageModel.Message+"...' on "+channelModel.ChannelName, options)
		if err != nil {
			log.Println(err)
			return true
		}
		return true
	}
	return false
}

func (service *BotService) SanedDM(app *config.App, bot *tb.Bot, m *tb.Message, request *Event) bool {
	if strings.Contains(m.Text, request.Command) {
		db := app.DB()
		defer db.Close()
		ids := strings.TrimPrefix(m.Text, request.Command1)
		data := strings.Split(ids, "_")
		directSenderID, err := strconv.Atoi(data[1])
		if err != nil {
			log.Println(err)
			return true
		}
		if m.Sender.ID == directSenderID {
			bot.Send(m.Sender, "You cannot send direct message to your self", HomeKeyOption(db, app))
			if m.Sender != nil {
				SaveUserLastState(db, app, bot, "not_access_to_dm", m.Sender.ID, "not_access_to_dm")
			}
			return true
		}
		if m.Sender != nil {
			SaveUserLastState(db, app, bot, m.Text, m.Sender.ID, request.UserState)
		}
		channelID := strings.TrimSpace(data[0])
		service.JoinFromGroup(db, app, bot, m, channelID)
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
		options.ParseMode = tb.ModeHTML
		user := service.GetUserByTelegramID(db, app, directSenderID)
		channel := service.GetChannelByTelegramID(db, app, channelID)
		_, err = bot.Send(m.Sender, "<code>Please send your direct message to the user:</code><b>"+user.UserID+channelID+"</b> <code>From:</code> <b>"+channel.ChannelName+"</b>", options)
		if err != nil {
			log.Println(err)
			return true
		}
		return true
	}
	return false
}

func (service *BotService) SanedAnswerDM(app *config.App, bot *tb.Bot, m *tb.Callback, request *Event) bool {
	if strings.Contains(m.Data, request.Command) {
		db := app.DB()
		defer db.Close()
		text := strings.TrimPrefix(m.Data, request.Command1)
		ids := strings.ReplaceAll(text, request.Command, "")
		data := strings.Split(ids, "_")
		directSenderID, err := strconv.Atoi(data[1])
		if err != nil {
			log.Println(err)
			return true
		}
		if m.Sender.ID == directSenderID {
			bot.Send(m.Sender, "You cannot send direct message to your self", HomeKeyOption(db, app))
			if m.Sender != nil {
				SaveUserLastState(db, app, bot, "not_access_to_dm", m.Sender.ID, "not_access_to_dm")
			}
			return true
		}
		if m.Sender != nil {
			SaveUserLastState(db, app, bot, m.Data, m.Sender.ID, request.UserState)
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
		options.ParseMode = tb.ModeHTML
		channelID := strings.TrimSpace(data[0])
		user := service.GetUserByTelegramID(db, app, directSenderID)
		channel := service.GetChannelByTelegramID(db, app, channelID)
		_, err = bot.Send(m.Sender, "<code>Please send your direct message to the user:</code><b>"+user.UserID+channelID+"</b> <code>From:</code> <b>"+channel.ChannelName+"</b>", options)
		if err != nil {
			log.Println(err)
			return true
		}
		return true
	}
	return false
}

func (service *BotService) SaveAndSendMessage(db *sql.DB, app *config.App, bot *tb.Bot, m *tb.Message, request *Event, lastState *models.UserLastState) bool {
	activeChannel := service.GetUserCurrentActiveChannel(db, app, bot, m)
	if activeChannel != nil {
		senderID := strconv.Itoa(m.Sender.ID)
		botMessageID := strconv.Itoa(m.ID)
		newReply := tb.InlineButton{
			Unique: "reply_to_message_on_group_" + activeChannel.ChannelID + "_" + senderID + "_" + botMessageID,
			Text:   "👻Reply",
			URL:    "https://t.me/" + app.BotUsername + "?start=reply_to_message_on_group_" + activeChannel.ChannelID + "_" + senderID + "_" + botMessageID,
		}
		newM := tb.InlineButton{
			Unique: "compose_message_in_group_" + activeChannel.ChannelID,
			Text:   "📝New",
			URL:    "https://t.me/" + app.BotUsername + "?start=compose_message_in_group_" + activeChannel.ChannelID,
		}
		newDM := tb.InlineButton{
			Unique: "reply_by_dm_to_user_on_group_" + activeChannel.ChannelID + "_" + senderID + "_" + botMessageID,
			Text:   "📲Direct",
			URL:    "https://t.me/" + app.BotUsername + "?start=reply_by_dm_to_user_on_group_" + activeChannel.ChannelID + "_" + senderID + "_" + botMessageID,
		}
		inlineKeys := [][]tb.InlineButton{
			[]tb.InlineButton{newReply, newM, newDM},
		}
		activeChannelID, err := strconv.Atoi(activeChannel.ChannelID)
		if err == nil {
			user := new(tb.User)
			user.ID = activeChannelID
			options := new(tb.SendOptions)
			replyModel := new(tb.ReplyMarkup)
			replyModel.InlineKeyboard = inlineKeys
			options.ReplyMarkup = replyModel
			options.ParseMode = tb.ModeHTML
			message, err := bot.Send(user, m.Text, options)
			// message, err := bot.Send(user, "From: <b>"+strconv.FormatInt(activeChannel.User.ID, 10)+activeChannel.User.UserID+"</b> <pre>\n"+m.Text+"</pre>", options)
			if err == nil {
				channelMessageID := strconv.Itoa(message.ID)
				channelID := strconv.FormatInt(activeChannel.ID, 10)
				insertedMessage, err := db.Query("INSERT INTO `messages` (`message`,`userID`,`channelID`,`channelMessageID`,`botMessageID`,`createdAt`) VALUES('" + m.Text + "','" + senderID + "','" + channelID + "','" + channelMessageID + "','" + botMessageID + "','" + app.CurrentTime + "')")
				if err != nil {
					log.Println(err)
					return true
				}
				defer insertedMessage.Close()
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
				bot.Send(m.Sender, "Your message has been sent anonymously to the group / channel "+activeChannel.ChannelName, options)
				SaveUserLastState(db, app, bot, "", m.Sender.ID, "message_sent")
			}
		}
	}
	return true
}

func (service *BotService) SendAndSaveReplyMessage(db *sql.DB, app *config.App, bot *tb.Bot, m *tb.Message, request *Event, lastState *models.UserLastState) bool {
	if lastState.Data != "" {
		ids := strings.TrimPrefix(lastState.Data, "/start reply_to_message_on_group_")
		if ids != "" {
			data := strings.Split(ids, "_")
			if len(data) == 3 {
				channelID := strings.TrimSpace(data[0])
				userID := strings.TrimSpace(data[1])
				botMessageID := strings.TrimSpace(data[2])
				senderID := strconv.Itoa(m.Sender.ID)
				newBotMessageID := strconv.Itoa(m.ID)
				messageStatement, err := db.Prepare("SELECT me.id,me.channelMessageID from `messages` as me inner join `channels` as ch on me.channelID=ch.id and ch.channelID=? where me.`botMessageID`=? and me.`userID`=?")
				if err != nil {
					log.Println(err)
					return true
				}
				defer messageStatement.Close()
				message := messageStatement.QueryRow(channelID, botMessageID, userID)
				messageModel := new(models.Message)
				if err := message.Scan(&messageModel.ID, &messageModel.ChannelMessageID); err == nil {
					channelIntValue, err := strconv.Atoi(channelID)
					if err == nil {
						newReply := tb.InlineButton{
							Unique: "reply_to_message_on_group_" + channelID + "_" + senderID + "_" + botMessageID,
							Text:   "👻Reply",
							URL:    "https://t.me/" + app.BotUsername + "?start=reply_to_message_on_group_" + channelID + "_" + senderID + "_" + botMessageID,
						}
						newM := tb.InlineButton{
							Unique: "compose_message_in_group_" + channelID,
							Text:   "📝New",
							URL:    "https://t.me/" + app.BotUsername + "?start=compose_message_in_group_" + channelID,
						}
						newDM := tb.InlineButton{
							Unique: "reply_by_dm_to_user_on_group_" + channelID + "_" + senderID + "_" + botMessageID,
							Text:   "📲Direct",
							URL:    "https://t.me/" + app.BotUsername + "?start=reply_by_dm_to_user_on_group_" + channelID + "_" + senderID + "_" + botMessageID,
						}
						inlineKeys := [][]tb.InlineButton{
							[]tb.InlineButton{newReply, newM, newDM},
						}
						ChannelMessageDataID, err := strconv.Atoi(messageModel.ChannelMessageID)
						if err == nil {
							// activeChannel := service.GetUserCurrentActiveChannel(db, app, bot, m)
							sendMessageModel := new(tb.Message)
							sendMessageModel.ID = ChannelMessageDataID
							newReplyModel := new(tb.ReplyMarkup)
							newReplyModel.InlineKeyboard = inlineKeys
							newSendOption := new(tb.SendOptions)
							newSendOption.ReplyTo = sendMessageModel
							newSendOption.ReplyMarkup = newReplyModel
							newSendOption.ParseMode = tb.ModeHTML
							user := new(tb.User)
							user.ID = channelIntValue
							sendMessage, err := bot.Send(user, m.Text, newSendOption)
							// sendMessage, err := bot.Send(user, "From: <b>"+strconv.FormatInt(activeChannel.User.ID, 10)+activeChannel.User.UserID+"</b> <pre>\n"+m.Text+"</pre>", newSendOption)
							if err == nil {
								newChannelMessageID := strconv.Itoa(sendMessage.ID)
								parentID := strconv.FormatInt(messageModel.ID, 10)
								currentChannelStatement, err := db.Prepare("SELECT id,channelName from `channels` where channelID=?")
								if err != nil {
									log.Println(err)
									return true
								}
								defer currentChannelStatement.Close()
								currentChannel := currentChannelStatement.QueryRow(channelID)
								newChannelModel := new(models.Channel)
								if err := currentChannel.Scan(&newChannelModel.ID, &newChannelModel.ChannelName); err == nil {
									newChannelModelID := strconv.FormatInt(newChannelModel.ID, 10)
									insertedMessage, err := db.Query("INSERT INTO `messages` (`message`,`userID`,`channelID`,`channelMessageID`,`botMessageID`,`parentID`,`createdAt`) VALUES('" + m.Text + "','" + senderID + "','" + newChannelModelID + "','" + newChannelMessageID + "','" + newBotMessageID + "','" + parentID + "','" + app.CurrentTime + "')")
									if err != nil {
										log.Println(err)
										return true
									}
									defer insertedMessage.Close()
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
									bot.Send(m.Sender, "Your reply message has been sent anonymously to the group / channel "+newChannelModel.ChannelName, options)
									SaveUserLastState(db, app, bot, "", m.Sender.ID, "reply_message_sent")
								}
							}
						}
					}
				}
			}
		}
	}
	return true
}

func (service *BotService) SendAndSaveDirectMessage(db *sql.DB, app *config.App, bot *tb.Bot, m *tb.Message, request *Event, lastState *models.UserLastState) bool {
	if lastState.Data != "" {
		ids := strings.TrimPrefix(lastState.Data, "/start reply_by_dm_to_user_on_group_")
		if ids != "" {
			data := strings.Split(ids, "_")
			if len(data) == 3 {
				channelID := strings.TrimSpace(data[0])
				userID := strings.TrimSpace(data[1])
				botMessageID := strings.TrimSpace(data[2])
				senderID := strconv.Itoa(m.Sender.ID)
				newBotMessageID := strconv.Itoa(m.ID)
				userIDInInt, err := strconv.Atoi(userID)
				if err == nil {
					messageStatement, err := db.Prepare("SELECT me.id,me.channelMessageID from `messages` as me inner join `channels` as ch on me.channelID=ch.id and ch.channelID=? where me.`botMessageID`=? and me.`userID`=?")
					if err != nil {
						log.Println(err)
						return true
					}
					defer messageStatement.Close()
					message := messageStatement.QueryRow(channelID, botMessageID, userID)
					messageModel := new(models.Message)
					if err := message.Scan(&messageModel.ID, &messageModel.ChannelMessageID); err == nil {
						newReply := tb.InlineButton{
							Unique: "answer_to_dm_" + channelID + "_" + senderID + "_" + newBotMessageID,
							Text:   "Direct Reply",
						}
						inlineKeys := [][]tb.InlineButton{
							[]tb.InlineButton{newReply},
						}
						_, err := strconv.Atoi(messageModel.ChannelMessageID)
						if err == nil {
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
							bot.Send(m.Sender, "Your Direct Message Has Been Sent To The User: "+userID+channelID, options)
							// sendMessageModel := new(tb.Message)
							// sendMessageModel.ID = ChannelMessageDataID
							newReplyModel := new(tb.ReplyMarkup)
							newReplyModel.InlineKeyboard = inlineKeys
							newSendOption := new(tb.SendOptions)
							// newSendOption.ReplyTo = sendMessageModel
							newSendOption.ReplyMarkup = newReplyModel
							user := new(tb.User)
							user.ID = userIDInInt
							sendMessage, err := bot.Send(user, m.Text, newSendOption)
							if err == nil {
								newChannelMessageID := strconv.Itoa(sendMessage.ID)
								parentID := strconv.FormatInt(messageModel.ID, 10)
								currentChannelStatement, err := db.Prepare("SELECT id from `channels` where channelID=?")
								if err != nil {
									log.Println(err)
									return true
								}
								defer currentChannelStatement.Close()
								currentChannel := currentChannelStatement.QueryRow(channelID)
								newChannelModel := new(models.Channel)
								if err := currentChannel.Scan(&newChannelModel.ID); err == nil {
									newChannelModelID := strconv.FormatInt(newChannelModel.ID, 10)
									insertedMessage, err := db.Query("INSERT INTO `messages` (`userID`,`channelID`,`channelMessageID`,`botMessageID`,`parentID`,`createdAt`) VALUES('" + senderID + "','" + newChannelModelID + "','" + newChannelMessageID + "','" + newBotMessageID + "','" + parentID + "','" + app.CurrentTime + "')")
									if err != nil {
										log.Println(err)
										return true
									}
									defer insertedMessage.Close()
									SaveUserLastState(db, app, bot, "", m.Sender.ID, "direct_message_sent")
								}
							}
						}
					}
				}
			}
		}
	}
	return true
}

func (service *BotService) SendAnswerAndSaveDirectMessage(db *sql.DB, app *config.App, bot *tb.Bot, m *tb.Message, request *Event, lastState *models.UserLastState) bool {
	if lastState.Data != "" {
		ids := strings.ReplaceAll(lastState.Data, "answer_to_dm_", "")
		if ids != "" {
			data := strings.Split(ids, "_")
			if len(data) == 3 {
				channelID := strings.TrimSpace(data[0])
				userID := strings.TrimSpace(data[1])
				senderID := strconv.Itoa(m.Sender.ID)
				newBotMessageID := strconv.Itoa(m.ID)
				userIDInInt, err := strconv.Atoi(userID)
				if err == nil {
					newReply := tb.InlineButton{
						Unique: "answer_to_dm_" + channelID + "_" + senderID + "_" + newBotMessageID,
						Text:   "Direct Reply",
					}
					inlineKeys := [][]tb.InlineButton{
						[]tb.InlineButton{newReply},
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
					bot.Send(m.Sender, "Your Direct Message Has Been Sent To The User: "+userID+channelID, options)
					newReplyModel := new(tb.ReplyMarkup)
					newReplyModel.InlineKeyboard = inlineKeys
					newSendOption := new(tb.SendOptions)
					newSendOption.ReplyMarkup = newReplyModel
					user := new(tb.User)
					user.ID = userIDInInt
					sendMessage, err := bot.Send(user, m.Text, newSendOption)
					if err == nil {
						newChannelMessageID := strconv.Itoa(sendMessage.ID)
						currentChannelStatement, err := db.Prepare("SELECT id from `channels` where `channelID`=?")
						if err != nil {
							log.Println(err)
							return true
						}
						defer currentChannelStatement.Close()
						currentChannel, err := currentChannelStatement.Query(channelID)
						if err != nil {
							log.Println(err)
							return true
						}
						if currentChannel.Next() {
							newChannelModel := new(models.Channel)
							if err := currentChannel.Scan(&newChannelModel.ID); err == nil {
								newChannelModelID := strconv.FormatInt(newChannelModel.ID, 10)
								insertedMessage, err := db.Query("INSERT INTO `messages` (`userID`,`channelID`,`channelMessageID`,`botMessageID`,`createdAt`) VALUES('" + senderID + "','" + newChannelModelID + "','" + newChannelMessageID + "','" + newBotMessageID + "','" + app.CurrentTime + "')")
								if err != nil {
									log.Println(err)
									return true
								}
								defer insertedMessage.Close()
								SaveUserLastState(db, app, bot, "", m.Sender.ID, "direct_message_sent")
							}
						}
					}
				}
			}
		}
	}
	return true
}

func (service *BotService) GetUserCurrentActiveChannel(db *sql.DB, app *config.App, bot *tb.Bot, m *tb.Message) *models.Channel {
	userID := strconv.Itoa(m.Sender.ID)
	userActiveStatement, err := db.Prepare("SELECT ch.id,ch.channelID,ch.channelName,us.id,us.userID from `channels` as ch inner join `users_current_active_channel` as uc on ch.id=uc.channelID and uc.status='ACTIVE' inner join `users` as us on uc.userID=us.id and us.userID=? and us.`status`='ACTIVE'")
	if err != nil {
		log.Println(err)
	}
	defer userActiveStatement.Close()
	userActiveChannel, err := userActiveStatement.Query(userID)
	if err != nil {
		log.Println(err)
	}
	if userActiveChannel.Next() {
		channelModel := new(models.Channel)
		userModel := new(models.User)
		if err := userActiveChannel.Scan(&channelModel.ID, &channelModel.ChannelID, &channelModel.ChannelName, &userModel.ID, &userModel.UserID); err != nil {
			log.Println(err)
		}
		channelModel.User = userModel
		return channelModel
	}
	return nil
}

func SaveUserLastState(db *sql.DB, app *config.App, bot *tb.Bot, data string, userDataID int, state string) {
	userID := strconv.Itoa(userDataID)
	insertedState, err := db.Query("INSERT INTO `users_last_state` (`userID`,`state`,`data`,`createdAt`) VALUES('" + userID + "','" + state + "','" + strings.TrimSpace(data) + "','" + app.CurrentTime + "')")
	if err != nil {
		log.Println(err)
		return
	}
	defer insertedState.Close()
}

func (service *BotService) GetChannelByTelegramID(db *sql.DB, app *config.App, channelID string) *models.Channel {
	userLastStateQueryStatement, err := db.Prepare("SELECT `channelName` from `channels` where `channelID`=? ")
	if err != nil {
		log.Println(err)
	}
	defer userLastStateQueryStatement.Close()
	userLastStateQuery, err := userLastStateQueryStatement.Query(channelID)
	if err != nil {
		log.Println(err)
	}
	channelModel := new(models.Channel)
	if userLastStateQuery.Next() {
		if err := userLastStateQuery.Scan(&channelModel.ChannelName); err != nil {
			log.Println(err)
		}
		return channelModel
	}
	return channelModel
}
