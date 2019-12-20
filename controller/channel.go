//Package controller ...
package controller

import (
	"github.com/amiraliio/tgbp/lang"
	"github.com/google/uuid"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/amiraliio/tgbp/config"
	"github.com/amiraliio/tgbp/model"
	"github.com/spf13/viper"
	tb "gopkg.in/tucnak/telebot.v2"
)

//TODO change query to queryRow
//TODO change value of queries to ?

//RegisterChannel
func RegisterChannel(bot *tb.Bot, m *tb.Message) {
	//channel private url
	inviteLink, err := bot.GetInviteLink(m.Chat)
	if err != nil {
		log.Println(err)
		return
	}
	channelURL := inviteLink
	channelID := strconv.FormatInt(m.Chat.ID, 10)
	db, err := config.DB()
	if err != nil {
		log.Println(err)
	}
	defer db.Close()
	statement, err := db.Prepare("SELECT id FROM `channels` where channelID=?")
	if err != nil {
		log.Println(err)
	}
	defer statement.Close()
	results, err := statement.Query(channelID)
	if err != nil {
		log.Println(err)
	}
	if !results.Next() {
		//start transaction
		transaction, err := db.Begin()
		if err != nil {
			log.Println(err)
		}
		uniqueID := uuid.New().String()
		//insert channel
		channelInserted, err := transaction.Exec("INSERT INTO `channels` (`channelType`,`channelURL`,`channelID`,`channelName`,`uniqueID`,`createdAt`,`updatedAt`) VALUES('channel','" + channelURL + "','" + channelID + "','" + m.Chat.Title + "','" + uniqueID + "','" + time.Now().UTC().Format("2006-01-02 03:04:05") + "','" + time.Now().UTC().Format("2006-01-02 03:04:05") + "')")
		if err != nil {
			transaction.Rollback()
			log.Println(err)
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
			}
			defer companyStatement.Close()
			companyExists, err := companyStatement.Query(companyFlag)
			if err != nil {
				transaction.Rollback()
				log.Println(err)
			}
			if !companyExists.Next() {
				//insert company
				companyInserted, err := transaction.Exec("INSERT INTO `companies` (`companyName`,`createdAt`,`updatedAt`) VALUES('" + companyFlag + "','" + time.Now().UTC().Format("2006-01-02 03:04:05") + "','" + time.Now().UTC().Format("2006-01-02 03:04:05") + "')")
				if err != nil {
					transaction.Rollback()
					log.Println(err)
				}
				insertedCompanyID, err := companyInserted.LastInsertId()
				if err == nil {
					companyModelID := strconv.FormatInt(insertedCompanyID, 10)
					channelModelID := strconv.FormatInt(insertedChannelID, 10)
					//insert company channel pivot
					_, err := transaction.Exec("INSERT INTO `companies_channels` (`companyID`,`channelID`,`createdAt`) VALUES('" + companyModelID + "','" + channelModelID + "','" + time.Now().UTC().Format("2006-01-02 03:04:05") + "')")
					if err != nil {
						transaction.Rollback()
						log.Println(err)
					}
				}
			} else {
				companyModel := new(model.Company)
				if err := companyExists.Scan(&companyModel.ID); err != nil {
					transaction.Rollback()
					log.Println(err)
				}
				companyModelID := strconv.FormatInt(companyModel.ID, 10)
				channelModelID := strconv.FormatInt(insertedChannelID, 10)
				//insert company channel pivot
				_, err := transaction.Exec("INSERT INTO `companies_channels` (`companyID`,`channelID`,`createdAt`) VALUES('" + companyModelID + "','" + channelModelID + "','" + time.Now().UTC().Format("2006-01-02 03:04:05") + "')")
				if err != nil {
					transaction.Rollback()
					log.Println(err)
				}
			}
			transaction.Commit()
			successMessage, _ := bot.Send(m.Chat, "You're channel registered successfully")
			time.Sleep(2 * time.Second)
			if err := bot.Delete(successMessage); err != nil {
				log.Println(err)
			}
			sendOptionModel := new(tb.SendOptions)
			sendOptionModel.ParseMode = tb.ModeHTML
			_, err = bot.Send(m.Chat, "This is your channel unique ID, you can save it and remove this message: <code> "+uniqueID+" </code>", sendOptionModel)
			if err != nil {
				log.Println(err)
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
			pinMessage, err := bot.Send(m.Chat, lang.StartGroup, newSendOption)
			if err != nil {
				log.Println(err)
			}
			if err := bot.Pin(pinMessage); err != nil {
				log.Println(err)
			}
			if err := bot.Delete(m); err != nil {
				log.Println(err)
			}
		}
	}
}

func NewMessageHandler(bot *tb.Bot, c *tb.User) {
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
	_, err := bot.Send(c, "Please send your message:", options)
	if err != nil {
		log.Println(err)
		return
	}
}

func SendReply(bot *tb.Bot, m *tb.User) {
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
	_, err := bot.Send(m, "Please send your reply to the message:", options)
	if err != nil {
		log.Println(err)
		return
	}
}

func SanedDM(bot *tb.Bot, m *tb.User) {
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
	bot.Send(m, "Please send your direct message to the user:", options)
}

func SanedAnswerDM(bot *tb.Bot, m *tb.User) {
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
	bot.Send(m, "Please send your direct message to the user:", options)
}

func SaveAndSendMessage(bot *tb.Bot, m *tb.Message) {
	//TODO inactive user last state
	//TODO restart the bot and show keyboards again
	activeChannel := GetUserCurrentActiveChannel(bot, m)
	if activeChannel != nil {
		senderID := strconv.Itoa(m.Sender.ID)
		botMessageID := strconv.Itoa(m.ID)
		newReply := tb.InlineButton{
			Unique: "reply_to_message_on_group_" + activeChannel.ChannelID + "_" + senderID + "_" + botMessageID,
			Text:   "👻Reply",
			URL:    "https://t.me/" + viper.GetString("APP.BOTUSERNAME") + "?start=reply_to_message_on_group_" + activeChannel.ChannelID + "_" + senderID + "_" + botMessageID,
		}
		newM := tb.InlineButton{
			Unique: "compose_message_in_group_" + activeChannel.ChannelID,
			Text:   "📝New",
			URL:    "https://t.me/" + viper.GetString("APP.BOTUSERNAME") + "?start=compose_message_in_group_" + activeChannel.ChannelID,
		}
		newDM := tb.InlineButton{
			Unique: "reply_by_dm_to_user_on_group_" + activeChannel.ChannelID + "_" + senderID + "_" + botMessageID,
			Text:   "📲Direct",
			URL:    "https://t.me/" + viper.GetString("APP.BOTUSERNAME") + "?start=reply_by_dm_to_user_on_group_" + activeChannel.ChannelID + "_" + senderID + "_" + botMessageID,
		}
		inlineKeys := [][]tb.InlineButton{
			[]tb.InlineButton{newReply, newM, newDM},
		}
		activeChannelID, err := strconv.Atoi(activeChannel.ChannelID)
		if err == nil {
			user := new(tb.User)
			user.ID = activeChannelID
			message, err := bot.Send(user, m.Text, &tb.ReplyMarkup{
				InlineKeyboard: inlineKeys,
			})
			if err == nil {
				channelMessageID := strconv.Itoa(message.ID)
				channelID := strconv.FormatInt(activeChannel.ID, 10)
				db, err := config.DB()
				if err != nil {
					log.Println(err)
				}
				defer db.Close()
				insertedMessage, err := db.Query("INSERT INTO `messages` (`userID`,`channelID`,`channelMessageID`,`botMessageID`,`createdAt`) VALUES('" + senderID + "','" + channelID + "','" + channelMessageID + "','" + botMessageID + "','" + time.Now().UTC().Format("2006-01-02 03:04:05") + "')")
				if err != nil {
					log.Println(err)
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
				bot.Send(m.Sender, "Sent your message has been sent anonymously to the group / channel "+activeChannel.ChannelName, options)
				SaveUserLastState(bot, "", m.Sender.ID, "message_sent")
			}
		}
	}
}

func SendAndSaveReplyMessage(bot *tb.Bot, m *tb.Message, lastState *model.UserLastState) {
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
				db, err := config.DB()
				if err != nil {
					log.Println(err)
				}
				defer db.Close()
				messageStatement, err := db.Prepare("SELECT me.id,me.channelMessageID from `messages` as me inner join `channels` as ch on me.channelID=ch.id and ch.channelID=? where me.`botMessageID`=? and me.`userID`=?")
				if err != nil {
					log.Println(err)
				}
				defer messageStatement.Close()
				message := messageStatement.QueryRow(channelID, botMessageID, userID)
				messageModel := new(model.Message)
				if err := message.Scan(&messageModel.ID, &messageModel.ChannelMessageID); err == nil {
					channelIntValue, err := strconv.Atoi(channelID)
					if err == nil {
						newReply := tb.InlineButton{
							Unique: "reply_to_message_on_group_" + channelID + "_" + senderID + "_" + botMessageID,
							Text:   "👻Reply",
							URL:    "https://t.me/" + viper.GetString("APP.BOTUSERNAME") + "?start=reply_to_message_on_group_" + channelID + "_" + senderID + "_" + botMessageID,
						}
						newM := tb.InlineButton{
							Unique: "compose_message_in_group_" + channelID,
							Text:   "📝New",
							URL:    "https://t.me/" + viper.GetString("APP.BOTUSERNAME") + "?start=compose_message_in_group_" + channelID,
						}
						newDM := tb.InlineButton{
							Unique: "reply_by_dm_to_user_on_group_" + channelID + "_" + senderID + "_" + botMessageID,
							Text:   "📲Direct",
							URL:    "https://t.me/" + viper.GetString("APP.BOTUSERNAME") + "?start=reply_by_dm_to_user_on_group_" + channelID + "_" + senderID + "_" + botMessageID,
						}
						inlineKeys := [][]tb.InlineButton{
							[]tb.InlineButton{newReply, newM, newDM},
						}
						ChannelMessageDataID, err := strconv.Atoi(messageModel.ChannelMessageID)
						if err == nil {
							sendMessageModel := new(tb.Message)
							sendMessageModel.ID = ChannelMessageDataID
							newReplyModel := new(tb.ReplyMarkup)
							newReplyModel.InlineKeyboard = inlineKeys
							newSendOption := new(tb.SendOptions)
							newSendOption.ReplyTo = sendMessageModel
							newSendOption.ReplyMarkup = newReplyModel
							user := new(tb.User)
							user.ID = channelIntValue
							sendMessage, err := bot.Send(user, m.Text, newSendOption)
							if err == nil {
								newChannelMessageID := strconv.Itoa(sendMessage.ID)
								parentID := strconv.FormatInt(messageModel.ID, 10)
								currentChannelStatement, err := db.Prepare("SELECT id,channelName from `channels` where channelID=?")
								if err != nil {
									log.Println(err)
								}
								defer currentChannelStatement.Close()
								currentChannel := currentChannelStatement.QueryRow(channelID)
								newChannelModel := new(model.Channel)
								if err := currentChannel.Scan(&newChannelModel.ID, &newChannelModel.ChannelName); err == nil {
									newChannelModelID := strconv.FormatInt(newChannelModel.ID, 10)
									insertedMessage, err := db.Query("INSERT INTO `messages` (`userID`,`channelID`,`channelMessageID`,`botMessageID`,`parentID`,`createdAt`) VALUES('" + senderID + "','" + newChannelModelID + "','" + newChannelMessageID + "','" + newBotMessageID + "','" + parentID + "','" + time.Now().UTC().Format("2006-01-02 03:04:05") + "')")
									if err != nil {
										log.Println(err)
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
									SaveUserLastState(bot, "", m.Sender.ID, "reply_message_sent")
								}
							}
						}
					}
				}
			}
		}
	}
}

func SendAndSaveDirectMessage(bot *tb.Bot, m *tb.Message, lastState *model.UserLastState) {
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
					db, err := config.DB()
					if err != nil {
						log.Println(err)
					}
					defer db.Close()
					messageStatement, err := db.Prepare("SELECT me.id,me.channelMessageID from `messages` as me inner join `channels` as ch on me.channelID=ch.id and ch.channelID=? where me.`botMessageID`=? and me.`userID`=?")
					if err != nil {
						log.Println(err)
					}
					defer messageStatement.Close()
					message := messageStatement.QueryRow(channelID, botMessageID, userID)
					messageModel := new(model.Message)
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
							bot.Send(m.Sender, "Your Direct Message Has Been Sent To The User", options)
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
								}
								defer currentChannelStatement.Close()
								currentChannel := currentChannelStatement.QueryRow(channelID)
								newChannelModel := new(model.Channel)
								if err := currentChannel.Scan(&newChannelModel.ID); err == nil {
									newChannelModelID := strconv.FormatInt(newChannelModel.ID, 10)
									insertedMessage, err := db.Query("INSERT INTO `messages` (`userID`,`channelID`,`channelMessageID`,`botMessageID`,`parentID`,`createdAt`) VALUES('" + senderID + "','" + newChannelModelID + "','" + newChannelMessageID + "','" + newBotMessageID + "','" + parentID + "','" + time.Now().UTC().Format("2006-01-02 03:04:05") + "')")
									if err != nil {
										log.Println(err)
									}
									defer insertedMessage.Close()
									SaveUserLastState(bot, "", m.Sender.ID, "direct_message_sent")
								}
							}
						}
					}
				}
			}
		}
	}
}

func SendAnswerAndSaveDirectMessage(bot *tb.Bot, m *tb.Message, lastState *model.UserLastState) {
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
					bot.Send(m.Sender, "Your Direct Message Has Been Sent To The User", options)
					newReplyModel := new(tb.ReplyMarkup)
					newReplyModel.InlineKeyboard = inlineKeys
					newSendOption := new(tb.SendOptions)
					newSendOption.ReplyMarkup = newReplyModel
					user := new(tb.User)
					user.ID = userIDInInt
					sendMessage, err := bot.Send(user, m.Text, newSendOption)
					if err == nil {
						db, err := config.DB()
						if err != nil {
							log.Println(err)
						}
						defer db.Close()
						newChannelMessageID := strconv.Itoa(sendMessage.ID)
						currentChannelStatement, err := db.Prepare("SELECT id from `channels` where `channelID`=?")
						if err != nil {
							log.Println(err)
						}
						defer currentChannelStatement.Close()
						currentChannel, err := currentChannelStatement.Query(channelID)
						if err != nil {
							log.Println(err)
						}
						if currentChannel.Next() {
							newChannelModel := new(model.Channel)
							if err := currentChannel.Scan(&newChannelModel.ID); err == nil {
								newChannelModelID := strconv.FormatInt(newChannelModel.ID, 10)
								insertedMessage, err := db.Query("INSERT INTO `messages` (`userID`,`channelID`,`channelMessageID`,`botMessageID`,`createdAt`) VALUES('" + senderID + "','" + newChannelModelID + "','" + newChannelMessageID + "','" + newBotMessageID + "','" + time.Now().UTC().Format("2006-01-02 03:04:05") + "')")
								if err != nil {
									log.Println(err)
								}
								defer insertedMessage.Close()
								SaveUserLastState(bot, "", m.Sender.ID, "direct_message_sent")
							}
						}
					}
				}
			}
		}
	}
}

func GetUserCurrentActiveChannel(bot *tb.Bot, m *tb.Message) *model.Channel {
	db, err := config.DB()
	if err != nil {
		log.Println(err)
	}
	defer db.Close()
	userID := strconv.Itoa(m.Sender.ID)
	userActiveStatement, err := db.Prepare("SELECT ch.id,ch.channelID,ch.channelName from `channels` as ch inner join `users_current_active_channel` as uc on ch.id=uc.channelID and uc.status='ACTIVE' inner join `users` as us on uc.userID=us.id and us.userID=? and us.`status`='ACTIVE'")
	if err != nil {
		log.Println(err)
	}
	defer userActiveStatement.Close()
	userActiveChannel, err := userActiveStatement.Query(userID)
	if err != nil {
		log.Println(err)
	}
	if userActiveChannel.Next() {
		channelModel := new(model.Channel)
		if err := userActiveChannel.Scan(&channelModel.ID, &channelModel.ChannelID, &channelModel.ChannelName); err != nil {
			log.Println(err)
		}
		return channelModel
	}
	return nil
}

func GetUserLastState(bot *tb.Bot, m *tb.Message, user int) *model.UserLastState {
	db, err := config.DB()
	if err != nil {
		log.Println(err)
	}
	defer db.Close()
	userLastStateQueryStatement, err := db.Prepare("SELECT `data`,`state` from `users_last_state` where `userId`=? order by `createdAt` DESC limit 1")
	if err != nil {
		log.Println(err)
	}
	defer userLastStateQueryStatement.Close()
	userLastStateQuery, err := userLastStateQueryStatement.Query(user)
	if err != nil {
		log.Println(err)
	}
	userLastState := new(model.UserLastState)
	if userLastStateQuery.Next() {
		if err := userLastStateQuery.Scan(&userLastState.Data, &userLastState.State); err != nil {
			log.Println(err)
		}
		return userLastState
	}
	return userLastState
}

func SaveUserLastState(bot *tb.Bot, data string, userDataID int, state string) {
	db, err := config.DB()
	if err != nil {
		log.Println(err)
		return
	}
	defer db.Close()
	userID := strconv.Itoa(userDataID)
	insertedState, err := db.Query("INSERT INTO `users_last_state` (`userID`,`state`,`data`,`createdAt`) VALUES('" + userID + "','" + state + "','" + strings.TrimSpace(data) + "','" + time.Now().UTC().Format("2006-01-02 03:04:05") + "')")
	if err != nil {
		log.Println(err)
		return
	}
	defer insertedState.Close()
}
