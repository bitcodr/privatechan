//Package controllers ...
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

//TODO save user hashed id in db
//TODO in channel registration from super admin add channel type from company type in channelModel

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
		channelModel := new(models.Channel)
		err = db.QueryRow("SELECT id FROM `channels` where channelID=?", channelID).Scan(&channelModel.ID)
		if errors.Is(err, sql.ErrNoRows) {
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
				companyModel := new(models.Company)
				err := db.QueryRow("SELECT id FROM `companies` where `companyName`=?", companyFlag).Scan(&companyModel.ID)
				if errors.Is(err, sql.ErrNoRows) {
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
				successMessage, _ := bot.Send(m.Chat, config.LangConfig.GetString("MESSAGES.CHANNEL_REGISTERED_SUCCESSFULLY"))
				time.Sleep(2 * time.Second)
				if err := bot.Delete(successMessage); err != nil {
					log.Println(err)
					return true
				}
				sendOptionModel := new(tb.SendOptions)
				sendOptionModel.ParseMode = tb.ModeHTML
				_, err = bot.Send(m.Chat, config.LangConfig.GetString("MESSAGES.CHANNEL_UNIQUE_ID_MESSAGE")+" <code> "+uniqueID+" </code>", sendOptionModel)
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
		service.CheckIfBotIsAdmin(app, bot, m, db, request)
		if m.Sender != nil {
			SaveUserLastState(db, app, bot, m.Text, m.Sender.ID, request.UserState)
		}
		ids := strings.TrimPrefix(m.Text, request.Command1)
		data := strings.Split(ids, "_")
		channelID := strings.TrimSpace(data[0])
		messageID := strings.TrimSpace(data[2])
		service.JoinFromGroup(db, app, bot, m, channelID)
		channelModel := new(models.Channel)
		messageModel := new(models.Message)
		if err := db.QueryRow("SELECT ch.channelName,me.message FROM `channels` as ch inner join messages as me on ch.id=me.channelID and me.botMessageID=? where ch.channelID=?", messageID, channelID).Scan(&channelModel.ChannelName, &messageModel.Message); err != nil {
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
		_, err := bot.Send(m.Sender, config.LangConfig.GetString("MESSAGES.PLEASE_REPLY")+"'"+messageModel.Message+"...' on "+channelModel.ChannelName, options)
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
		service.CheckIfBotIsAdmin(app, bot, m, db, request)
		ids := strings.TrimPrefix(m.Text, request.Command1)
		data := strings.Split(ids, "_")
		directSenderID, err := strconv.Atoi(data[1])
		if err != nil {
			log.Println(err)
			return true
		}
		if m.Sender.ID == directSenderID {
			bot.Send(m.Sender, config.LangConfig.GetString("MESSAGES.BAN_DIRECT"), HomeKeyOption(db, app))
			if m.Sender != nil {
				SaveUserLastState(db, app, bot, config.LangConfig.GetString("STATE.NOT_DM_ACCESS"), m.Sender.ID, config.LangConfig.GetString("STATE.NOT_DM_ACCESS"))
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
			Text: config.LangConfig.GetString("GENERAL.HOME"),
		}
		replyKeys := [][]tb.ReplyButton{
			[]tb.ReplyButton{homeBTN},
		}
		markup.ReplyKeyboard = replyKeys
		options.ReplyMarkup = markup
		options.ParseMode = tb.ModeHTML
		user := service.GetUserByTelegramID(db, app, directSenderID)
		channel := service.GetChannelByTelegramID(db, app, channelID)
		_, err = bot.Send(m.Sender, config.LangConfig.GetString("MESSAGES.PLEASE_SEND_YOUR_DIRECT")+"<b>"+helpers.Hash(user.UserID+channelID)+"</b> "+config.LangConfig.GetString("GENERAL.FROM")+": <b>"+channel.ChannelName+"</b>", options)
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
			bot.Send(m.Sender, config.LangConfig.GetString("MESSAGES.BAN_DIRECT"), HomeKeyOption(db, app))
			if m.Sender != nil {
				SaveUserLastState(db, app, bot, config.LangConfig.GetString("STATE.NOT_DM_ACCESS"), m.Sender.ID, config.LangConfig.GetString("STATE.NOT_DM_ACCESS"))
			}
			return true
		}
		if m.Sender != nil {
			SaveUserLastState(db, app, bot, m.Data, m.Sender.ID, request.UserState)
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
		options.ParseMode = tb.ModeHTML
		channelID := strings.TrimSpace(data[0])
		user := service.GetUserByTelegramID(db, app, directSenderID)
		channel := service.GetChannelByTelegramID(db, app, channelID)
		_, err = bot.Send(m.Sender, config.LangConfig.GetString("MESSAGES.PLEASE_SEND_YOUR_DIRECT")+"<b>"+helpers.Hash(user.UserID+channelID)+"</b> "+config.LangConfig.GetString("GENERAL.FROM")+": <b>"+channel.ChannelName+"</b>", options)
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
			Unique: config.LangConfig.GetString("STATE.REPLY_TO_MESSAGE") + "_" + activeChannel.ChannelID + "_" + senderID + "_" + botMessageID,
			Text:   config.LangConfig.GetString("MESSAGES.REPLY"),
			URL:    app.TgDomain + app.BotUsername + "?start=" + config.LangConfig.GetString("STATE.REPLY_TO_MESSAGE") + "_" + activeChannel.ChannelID + "_" + senderID + "_" + botMessageID,
		}
		newM := tb.InlineButton{
			Unique: config.LangConfig.GetString("STATE.COMPOSE_MESSAGE") + "_" + activeChannel.ChannelID,
			Text:   config.LangConfig.GetString("MESSAGES.NEW"),
			URL:    app.TgDomain + app.BotUsername + "?start=" + config.LangConfig.GetString("STATE.COMPOSE_MESSAGE") + "_" + activeChannel.ChannelID,
		}
		newDM := tb.InlineButton{
			Unique: config.LangConfig.GetString("STATE.REPLY_BY_DM") + "_" + activeChannel.ChannelID + "_" + senderID + "_" + botMessageID,
			Text:   config.LangConfig.GetString("MESSAGES.DIRECT"),
			URL:    app.TgDomain + app.BotUsername + "?start=" + config.LangConfig.GetString("STATE.REPLY_BY_DM") + "_" + activeChannel.ChannelID + "_" + senderID + "_" + botMessageID,
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
				anotherMessage := tb.InlineButton{
					Unique: config.LangConfig.GetString("STATE.COMPOSE_MESSAGE") + "_" + activeChannel.ChannelID,
					Text:   config.LangConfig.GetString("MESSAGES.ANOTHER_NEW"),
				}
				var inlineKeys [][]tb.InlineButton
				if activeChannel.ChannelURL != "" {
					redirectBTN := tb.InlineButton{
						Text: config.LangConfig.GetString("MESSAGES.BACK_TO") + strings.Title(activeChannel.ChannelType),
						URL:  activeChannel.ChannelURL,
					}
					inlineKeys = [][]tb.InlineButton{
						[]tb.InlineButton{anotherMessage},
						[]tb.InlineButton{redirectBTN},
					}
				} else {
					inlineKeys = [][]tb.InlineButton{
						[]tb.InlineButton{anotherMessage},
					}
				}
				markup.InlineKeyboard = inlineKeys
				options.ReplyMarkup = markup
				bot.Send(m.Sender, config.LangConfig.GetString("MESSAGES.MESSAGE_HAS_BEEN_SENT")+activeChannel.ChannelName, options)
				SaveUserLastState(db, app, bot, "", m.Sender.ID, "message_sent")
			}
		}
	}
	return true
}

func (service *BotService) SendAndSaveReplyMessage(db *sql.DB, app *config.App, bot *tb.Bot, m *tb.Message, request *Event, lastState *models.UserLastState) bool {
	if lastState.Data != "" {
		ids := strings.TrimPrefix(lastState.Data, request.Command1)
		if ids != "" {
			data := strings.Split(ids, "_")
			if len(data) == 3 {
				channelID := strings.TrimSpace(data[0])
				userID := strings.TrimSpace(data[1])
				botMessageID := strings.TrimSpace(data[2])
				senderID := strconv.Itoa(m.Sender.ID)
				newBotMessageID := strconv.Itoa(m.ID)
				messageModel := new(models.Message)
				if err := db.QueryRow("SELECT me.id,me.channelMessageID from `messages` as me inner join `channels` as ch on me.channelID=ch.id and ch.channelID=? where me.`botMessageID`=? and me.`userID`=?", channelID, botMessageID, userID).Scan(&messageModel.ID, &messageModel.ChannelMessageID); err == nil {
					channelIntValue, err := strconv.Atoi(channelID)
					if err == nil {
						newReply := tb.InlineButton{
							Unique: config.LangConfig.GetString("STATE.REPLY_TO_MESSAGE") + "_" + channelID + "_" + senderID + "_" + botMessageID,
							Text:   config.LangConfig.GetString("MESSAGES.REPLY"),
							URL:    app.TgDomain + app.BotUsername + "?start=" + config.LangConfig.GetString("STATE.REPLY_TO_MESSAGE") + "_" + channelID + "_" + senderID + "_" + botMessageID,
						}
						newM := tb.InlineButton{
							Unique: config.LangConfig.GetString("STATE.COMPOSE_MESSAGE") + "_" + channelID,
							Text:   config.LangConfig.GetString("MESSAGES.NEW"),
							URL:    app.TgDomain + app.BotUsername + "?start=" + config.LangConfig.GetString("STATE.COMPOSE_MESSAGE") + "_" + channelID,
						}
						newDM := tb.InlineButton{
							Unique: config.LangConfig.GetString("STATE.REPLY_BY_DM") + "_" + channelID + "_" + senderID + "_" + botMessageID,
							Text:   config.LangConfig.GetString("MESSAGES.DIRECT"),
							URL:    app.TgDomain + app.BotUsername + "?start=" + config.LangConfig.GetString("STATE.REPLY_BY_DM") + "_" + channelID + "_" + senderID + "_" + botMessageID,
						}
						inlineKeys := [][]tb.InlineButton{
							[]tb.InlineButton{newReply, newM, newDM},
						}
						ChannelMessageDataID, err := strconv.Atoi(messageModel.ChannelMessageID)
						if err == nil {
							activeChannel := service.GetUserCurrentActiveChannel(db, app, bot, m)
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
								newChannelModel := new(models.Channel)
								if err := db.QueryRow("SELECT id,channelName from `channels` where channelID=?", channelID).Scan(&newChannelModel.ID, &newChannelModel.ChannelName); err == nil {
									newChannelModelID := strconv.FormatInt(newChannelModel.ID, 10)
									insertedMessage, err := db.Query("INSERT INTO `messages` (`message`,`userID`,`channelID`,`channelMessageID`,`botMessageID`,`parentID`,`createdAt`) VALUES('" + helpers.ClearString(m.Text) + "','" + senderID + "','" + newChannelModelID + "','" + newChannelMessageID + "','" + newBotMessageID + "','" + parentID + "','" + app.CurrentTime + "')")
									if err != nil {
										log.Println(err)
										return true
									}
									defer insertedMessage.Close()
									options := new(tb.SendOptions)
									markup := new(tb.ReplyMarkup)
									if activeChannel.ChannelURL != "" {
										redirectBTN := tb.InlineButton{
											Text: config.LangConfig.GetString("MESSAGES.BACK_TO") + strings.Title(activeChannel.ChannelType),
											URL:  activeChannel.ChannelURL,
										}
										inlineKeys := [][]tb.InlineButton{
											[]tb.InlineButton{redirectBTN},
										}
										markup.InlineKeyboard = inlineKeys
									} else {
										homeBTN := tb.ReplyButton{
											Text: config.LangConfig.GetString("GENERAL.HOME"),
										}
										replyKeys := [][]tb.ReplyButton{
											[]tb.ReplyButton{homeBTN},
										}
										markup.ReplyKeyboard = replyKeys
									}
									options.ReplyMarkup = markup
									bot.Send(m.Sender, config.LangConfig.GetString("MESSAGES.REPLY_MESSAGE_HAS_BEEN_SENT")+newChannelModel.ChannelName, options)
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
		ids := strings.TrimPrefix(lastState.Data, request.Command1)
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
					messageModel := new(models.Message)
					channelModel := new(models.Channel)
					if err := db.QueryRow("SELECT me.id,me.channelMessageID,ch.channelName,ch.channelURL,ch.channelType from `messages` as me inner join `channels` as ch on me.channelID=ch.id and ch.channelID=? where me.`botMessageID`=? and me.`userID`=?", channelID, botMessageID, userID).Scan(&messageModel.ID, &messageModel.ChannelMessageID, &channelModel.ChannelName, &channelModel.ChannelURL, &channelModel.ChannelType); err == nil {
						_, err := strconv.Atoi(messageModel.ChannelMessageID)
						if err == nil {
							options := new(tb.SendOptions)
							markup := new(tb.ReplyMarkup)
							SendAnotherDM := tb.InlineButton{
								Unique: config.LangConfig.GetString("STATE.ANSWER_TO_DM") + "_" + channelID + "_" + userID + "_" + newBotMessageID,
								Text:   config.LangConfig.GetString("MESSAGES.ANOTHER_DIRECT_REPLY"),
							}
							var AnotherDMKeys [][]tb.InlineButton
							if channelModel.ChannelURL != "" {
								redirectBTN := tb.InlineButton{
									Text: config.LangConfig.GetString("MESSAGES.BACK_TO") + strings.Title(channelModel.ChannelType),
									URL:  channelModel.ChannelURL,
								}
								AnotherDMKeys = [][]tb.InlineButton{
									[]tb.InlineButton{SendAnotherDM},
									[]tb.InlineButton{redirectBTN},
								}

							} else {
								AnotherDMKeys = [][]tb.InlineButton{
									[]tb.InlineButton{SendAnotherDM},
								}
							}
							markup.InlineKeyboard = AnotherDMKeys
							options.ReplyMarkup = markup
							options.ParseMode = tb.ModeHTML
							bot.Send(m.Sender, config.LangConfig.GetString("MESSAGES.DIRECT_HAS_BEEN_SENT")+"<b>"+helpers.Hash(userID+channelID)+"</b>", options)
							newReplyModel := new(tb.ReplyMarkup)
							newReply := tb.InlineButton{
								Unique: config.LangConfig.GetString("STATE.ANSWER_TO_DM") + "_" + channelID + "_" + senderID + "_" + newBotMessageID,
								Text:   config.LangConfig.GetString("MESSAGES.DIRECT_REPLY"),
							}
							inlineKeys := [][]tb.InlineButton{
								[]tb.InlineButton{newReply},
							}
							newReplyModel.InlineKeyboard = inlineKeys
							newSendOption := new(tb.SendOptions)
							newSendOption.ReplyMarkup = newReplyModel
							newSendOption.ParseMode = tb.ModeHTML
							user := new(tb.User)
							user.ID = userIDInInt
							sendMessage, err := bot.Send(user, "<b>"+config.LangConfig.GetString("GENERAL.MESSAGE")+":</b> "+m.Text+" <b>"+config.LangConfig.GetString("GENERAL.FROM")+":</b> "+helpers.Hash(senderID+channelID)+"<b> "+config.LangConfig.GetString("MESSAGES.ON_CHANNEL_GROUP")+": </b> "+channelModel.ChannelName, newSendOption)
							if err == nil {
								newChannelMessageID := strconv.Itoa(sendMessage.ID)
								parentID := strconv.FormatInt(messageModel.ID, 10)
								newChannelModel := new(models.Channel)
								if err := db.QueryRow("SELECT id from `channels` where channelID=?", channelID).Scan(&newChannelModel.ID); err == nil {
									newChannelModelID := strconv.FormatInt(newChannelModel.ID, 10)
									insertedMessage, err := db.Query("INSERT INTO `messages` (`userID`,`channelID`,`channelMessageID`,`botMessageID`,`parentID`,`createdAt`) VALUES('" + senderID + "','" + newChannelModelID + "','" + newChannelMessageID + "','" + newBotMessageID + "','" + parentID + "','" + app.CurrentTime + "')")
									if err != nil {
										log.Println(err)
										return true
									}
									defer insertedMessage.Close()
									SaveUserLastState(db, app, bot, "", m.Sender.ID, config.LangConfig.GetString("STATE.DIRECT_MESSAGE_SENT"))
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
		ids := strings.ReplaceAll(lastState.Data, config.LangConfig.GetString("STATE.ANSWER_TO_DM")+"_", "")
		if ids != "" {
			data := strings.Split(ids, "_")
			if len(data) == 3 {
				channelID := strings.TrimSpace(data[0])
				userID := strings.TrimSpace(data[1])
				senderID := strconv.Itoa(m.Sender.ID)
				newBotMessageID := strconv.Itoa(m.ID)
				userIDInInt, err := strconv.Atoi(userID)
				if err == nil {
					channelModel := new(models.Channel)
					if err := db.QueryRow("SELECT channelURL,channelType from `channels` where channelID=?", channelID).Scan(&channelModel.ChannelURL, &channelModel.ChannelType); err == nil {
						options := new(tb.SendOptions)
						markup := new(tb.ReplyMarkup)
						SendAnotherDM := tb.InlineButton{
							Unique: config.LangConfig.GetString("STATE.ANSWER_TO_DM") + "_" + channelID + "_" + userID + "_" + newBotMessageID,
							Text:   config.LangConfig.GetString("MESSAGES.ANOTHER_DIRECT_REPLY"),
						}
						var AnotherDMKeys [][]tb.InlineButton
						if channelModel.ChannelURL != "" {
							redirectBTN := tb.InlineButton{
								Text: config.LangConfig.GetString("MESSAGES.BACK_TO") + strings.Title(channelModel.ChannelType),
								URL:  channelModel.ChannelURL,
							}
							AnotherDMKeys = [][]tb.InlineButton{
								[]tb.InlineButton{SendAnotherDM},
								[]tb.InlineButton{redirectBTN},
							}

						} else {
							AnotherDMKeys = [][]tb.InlineButton{
								[]tb.InlineButton{SendAnotherDM},
							}
						}
						markup.InlineKeyboard = AnotherDMKeys
						options.ReplyMarkup = markup
						options.ParseMode = tb.ModeHTML
						bot.Send(m.Sender, config.LangConfig.GetString("MESSAGES.DIRECT_HAS_BEEN_SENT")+" <b>"+helpers.Hash(userID+channelID)+"</b>", options)
						newChannelModel := new(models.Channel)
						if err := db.QueryRow("SELECT id,channelName from `channels` where `channelID`=?", channelID).Scan(&newChannelModel.ID, &newChannelModel.ChannelName); err == nil {
							newReply := tb.InlineButton{
								Unique: config.LangConfig.GetString("STATE.ANSWER_TO_DM") + "_" + channelID + "_" + senderID + "_" + newBotMessageID,
								Text:   config.LangConfig.GetString("MESSAGES.DIRECT_REPLY"),
							}
							inlineKeys := [][]tb.InlineButton{
								[]tb.InlineButton{newReply},
							}
							newReplyModel := new(tb.ReplyMarkup)
							newReplyModel.InlineKeyboard = inlineKeys
							newSendOption := new(tb.SendOptions)
							newSendOption.ReplyMarkup = newReplyModel
							newSendOption.ParseMode = tb.ModeHTML
							user := new(tb.User)
							user.ID = userIDInInt
							sendMessage, err := bot.Send(user, "<b>"+config.LangConfig.GetString("GENERAL.MESSAGE")+":</b> "+m.Text+" <b>"+config.LangConfig.GetString("GENERAL.FROM")+":</b> "+helpers.Hash(senderID+channelID)+"<b> "+config.LangConfig.GetString("MESSAGES.ON_CHANNEL_GROUP")+": </b> "+newChannelModel.ChannelName, newSendOption)
							if err == nil {
								newChannelMessageID := strconv.Itoa(sendMessage.ID)
								newChannelModelID := strconv.FormatInt(newChannelModel.ID, 10)
								insertedMessage, err := db.Query("INSERT INTO `messages` (`userID`,`channelID`,`channelMessageID`,`botMessageID`,`createdAt`) VALUES('" + senderID + "','" + newChannelModelID + "','" + newChannelMessageID + "','" + newBotMessageID + "','" + app.CurrentTime + "')")
								if err != nil {
									log.Println(err)
									return true
								}
								defer insertedMessage.Close()
								SaveUserLastState(db, app, bot, "", m.Sender.ID, config.LangConfig.GetString("STATE.DIRECT_MESSAGE_SENT"))
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
	userModel := new(models.User)
	channelModel := new(models.Channel)
	if err := db.QueryRow("SELECT ch.id,ch.channelID,ch.channelName,ch.channelURL,ch.channelType,us.id,us.userID from `channels` as ch inner join `users_current_active_channel` as uc on ch.id=uc.channelID and uc.status='ACTIVE' inner join `users` as us on uc.userID=us.id and us.userID=? and us.`status`='ACTIVE' ", userID).Scan(&channelModel.ID, &channelModel.ChannelID, &channelModel.ChannelName, &channelModel.ChannelURL, &channelModel.ChannelType, &userModel.ID, &userModel.UserID); err != nil {
		log.Println(err)
		return channelModel
	}
	channelModel.User = userModel
	return channelModel
}

func (service *BotService) GetChannelByTelegramID(db *sql.DB, app *config.App, channelID string) *models.Channel {
	channelModel := new(models.Channel)
	if err := db.QueryRow("SELECT `channelName` from `channels` where `channelID`=? ", channelID).Scan(&channelModel.ChannelName); err != nil {
		log.Println(err)
		return channelModel
	}
	return channelModel
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
