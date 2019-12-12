//Package controller ...
package controller

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/amiraliio/tgbp/config"
	"github.com/amiraliio/tgbp/model"
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
	tb "gopkg.in/tucnak/telebot.v2"
)

//TODO change query to queryRow

//RegisterChannel
func RegisterChannel(bot *tb.Bot, m *tb.Message) {
	if strings.Contains(m.Text, "https://t.me/joinchat/") && strings.Contains(m.Text, "@") {
		channelDetails := strings.SplitAfter(m.Text, "@")
		if len(channelDetails) == 2 && strings.Contains(channelDetails[0], "@") {
			//channel private url
			channelURL := strings.ReplaceAll(channelDetails[0], "@", "")
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
				//insert channel
				channelInserted, err := transaction.Exec("INSERT INTO `channels` (`channelType`,`channelURL`,`channelID`,`channelName`,`createdAt`,`updatedAt`) VALUES('channel','" + channelURL + "','" + channelID + "','" + m.Chat.Title + "','" + time.Now().UTC().Format("2006-01-02 03:04:05") + "','" + time.Now().UTC().Format("2006-01-02 03:04:05") + "')")
				if err != nil {
					transaction.Rollback()
					log.Println(err)
				}
				insertedChannelID, err := channelInserted.LastInsertId()
				if err == nil {
					//company name
					companyFlag := channelDetails[1]
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
					time.Sleep(3 * time.Second)
					if err := bot.Delete(successMessage); err != nil {
						log.Println(err)
					}
					pinMessage, err := bot.Send(m.Chat, "You can send a message in this channel, by https://t.me/"+viper.GetString("APP.BOTUSERNAME")+"?start=join_group"+channelID)
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
	}
}

func JoinFromChannel(bot *tb.Bot, m *tb.Message, inlineKeys [][]tb.InlineButton) {
	data := strings.Split(m.Text, "join_group")
	if len(data) == 2 {
		channelID := strings.TrimSpace(data[1])
		db, err := config.DB()
		if err != nil {
			log.Println(err)
		}
		defer db.Close()
		userID := strconv.Itoa(m.Sender.ID)
		//check if user is not created
		statement, err := db.Prepare("SELECT id FROM `users` where `status`= 'ACTIVE' and `userID`=?")
		if err != nil {
			log.Println(err)
		}
		defer statement.Close()
		results, err := statement.Query(userID)
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
			userInsert, err := transaction.Exec("INSERT INTO `users` (`userID`,`username`,`firstName`,`lastName`,`lang`,`isBot`,`createdAt`,`updatedAt`) VALUES('" + userID + "','" + m.Sender.Username + "','" + m.Sender.FirstName + "','" + m.Sender.LastName + "','" + m.Sender.LanguageCode + "','" + isBotValue + "','" + time.Now().UTC().Format("2006-01-02 03:04:05") + "','" + time.Now().UTC().Format("2006-01-02 03:04:05") + "')")
			if err != nil {
				transaction.Rollback()
				log.Println(err)
			}
			insertedUserID, err := userInsert.LastInsertId()
			if err != nil {
				transaction.Rollback()
				log.Println(err)
			}
			checkAndInsertUserChannel(bot, m, insertedUserID, channelID, db, transaction, inlineKeys)
		} else {
			userModel := new(model.User)
			if err := results.Scan(&userModel.ID); err != nil {
				transaction.Rollback()
				log.Println(err)
			}
			checkAndInsertUserChannel(bot, m, userModel.ID, channelID, db, transaction, inlineKeys)
		}
	}
}

func checkAndInsertUserChannel(bot *tb.Bot, m *tb.Message, queryUserID int64, channelID string, db *sql.DB, transaction *sql.Tx, inlineKeys [][]tb.InlineButton) {
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
			channelModel := new(model.Channel)
			if err := getChannelID.Scan(&channelModel.ID); err != nil {
				transaction.Rollback()
				log.Println(err)
			}
			channelModelID := strconv.FormatInt(channelModel.ID, 10)
			_, err := transaction.Exec("INSERT INTO `users_channels` (`userID`,`channelID`,`createdAt`,`updatedAt`) VALUES('" + userModelID + "','" + channelModelID + "','" + time.Now().UTC().Format("2006-01-02 03:04:05") + "','" + time.Now().UTC().Format("2006-01-02 03:04:05") + "')")
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
		channelModelData := new(model.Channel)
		companyModel := new(model.Company)
		if err := channelDetail.Scan(&channelModelData.ID, &channelModelData.ChannelName, &companyModel.CompanyName); err != nil {
			transaction.Rollback()
			log.Println(err)
		}
		channelModelID := strconv.FormatInt(channelModelData.ID, 10)
		//check the user is active or not
		checkUserChannelActivityStatement, err := db.Prepare("SELECT `id`,`status` from `users_channels` where `userID`=? and `channelID`=? and `status`='ACTIVE'")
		if err != nil {
			transaction.Rollback()
			log.Println(err)
		}
		defer checkUserChannelActivityStatement.Close()
		checkUserChannelActivity, err := checkUserChannelActivityStatement.Query(userModelID, channelModelID)
		if err != nil {
			transaction.Rollback()
			log.Println(err)
		}
		//inactive user last active channels
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
		transaction.Commit()
		if checkUserChannelActivity.Next() {
			//send welcome message and show inline keyboards
			_, err = bot.Send(m.Chat, "Welcome to the channel "+channelModelData.ChannelName+" the channel blongs to company "+companyModel.CompanyName+" You can use the inline keyboards to do an action on the channel or etc...", &tb.ReplyMarkup{
				InlineKeyboard: inlineKeys,
			})
			if err != nil {
				log.Println(err)
			}
		} else {
			//show a message for verification because the user isn't verify
			_, err := bot.Send(m.Chat, "You trying to join to the channel "+channelModelData.ChannelName+" blongs to company "+companyModel.CompanyName+" If do want to send a message on channel you should send and verify your email address")
			if err != nil {
				log.Println(err)
			}
			//TODO show verification btn and get user email to send a code for activation
			//TODO active user verification in the pivot table
			//TODO if user is active in one company channels the user should be active in other company channels
		}
	}
}

func NewMessageHandler(bot *tb.Bot, c *tb.Callback) {
	bot.Send(c.Sender, "Please send your message:")
}

func SendReply(bot *tb.Bot, m *tb.Message) {
	bot.Send(m.Sender, "Please send your reply to the message:")
}

func SanedDM(bot *tb.Bot, m *tb.Message) {
	bot.Send(m.Sender, "Please send your direct message to the user:")
}

func SanedAnswerDM(bot *tb.Bot, m *tb.Callback) {
	bot.Send(m.Sender, "Please send your direct message to the user:")
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
			Text:   "Reply",
			URL:    "https://t.me/" + viper.GetString("APP.BOTUSERNAME") + "?start=reply_to_message_on_group_" + activeChannel.ChannelID + "_" + senderID + "_" + botMessageID,
		}
		newDM := tb.InlineButton{
			Unique: "reply_by_dm_to_user_on_group_" + activeChannel.ChannelID + "_" + senderID + "_" + botMessageID,
			Text:   "Direct Message",
			URL:    "https://t.me/" + viper.GetString("APP.BOTUSERNAME") + "?start=reply_by_dm_to_user_on_group_" + activeChannel.ChannelID + "_" + senderID + "_" + botMessageID,
		}
		inlineKeys := [][]tb.InlineButton{
			[]tb.InlineButton{newReply, newDM},
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
				bot.Send(m.Sender, "Your Message Has Been Sent")
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
							Unique: "reply_to_message_on_group_" + channelID + "_" + senderID + "_" + newBotMessageID,
							Text:   "Reply",
							URL:    "https://t.me/" + viper.GetString("APP.BOTUSERNAME") + "?start=reply_to_message_on_group_" + channelID + "_" + senderID + "_" + newBotMessageID,
						}
						newDM := tb.InlineButton{
							Unique: "reply_by_dm_to_user_on_group_" + channelID + "_" + senderID + "_" + newBotMessageID,
							Text:   "Direct Message",
							URL:    "https://t.me/" + viper.GetString("APP.BOTUSERNAME") + "?start=reply_by_dm_to_user_on_group_" + channelID + "_" + senderID + "_" + newBotMessageID,
						}
						inlineKeys := [][]tb.InlineButton{
							[]tb.InlineButton{newReply, newDM},
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
									bot.Send(m.Sender, "Your Reply Message Has Been Sent")
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
							bot.Send(m.Sender, "Your Direct Message Has Been Sent To The User")
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
							fmt.Println(err)
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
				// botMessageID := data[2]
				senderID := strconv.Itoa(m.Sender.ID)
				newBotMessageID := strconv.Itoa(m.ID)
				userIDInInt, err := strconv.Atoi(userID)
				if err == nil {
					// message := db.QueryRow("SELECT me.id,me.channelMessageID from `messages` as me inner join `channels` as ch on me.channelID=ch.id and ch.channelID='" + channelID + "' where me.`botMessageID`='" + botMessageID + "' and me.`userID`='" + userID + "'")
					// messageModel := new(model.Message)
					// if err := message.Scan(&messageModel.ID, &messageModel.ChannelMessageID); err == nil {
					newReply := tb.InlineButton{
						Unique: "answer_to_dm_" + channelID + "_" + senderID + "_" + newBotMessageID,
						Text:   "Direct Reply",
					}
					inlineKeys := [][]tb.InlineButton{
						[]tb.InlineButton{newReply},
					}
					// ChannelMessageDataID, err := strconv.Atoi(messageModel.ChannelMessageID)
					// if err == nil {

					bot.Send(m.Sender, "Your Direct Message Has Been Sent To The User")
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
						db, err := config.DB()
						if err != nil {
							log.Println(err)
						}
						defer db.Close()
						newChannelMessageID := strconv.Itoa(sendMessage.ID)
						// parentID := strconv.FormatInt(messageModel.ID, 10)
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
					// }
					// }
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

func GetUserLastState(bot *tb.Bot, m *tb.Message) *model.UserLastState {
	db, err := config.DB()
	if err != nil {
		log.Println(err)
	}
	defer db.Close()
	userID := strconv.Itoa(m.Sender.ID)
	userLastStateQueryStatement, err := db.Prepare("SELECT ch.data,ch.state from `users_last_state` as ch inner join `users` as us on ch.userID=us.id and us.userID=? and us.`status`='ACTIVE' where ch.status='ACTIVE'")
	if err != nil {
		log.Println(err)
	}
	defer userLastStateQueryStatement.Close()
	userLastStateQuery, err := userLastStateQueryStatement.Query(userID)
	if err != nil {
		log.Println(err)
	}
	if userLastStateQuery.Next() {
		userLastState := new(model.UserLastState)
		if err := userLastStateQuery.Scan(&userLastState.Data, &userLastState.State); err != nil {
			log.Println(err)
		}
		return userLastState
	}
	return nil
}

func SaveUserLastState(bot *tb.Bot, data string, userDataID int, state string) {
	db, err := config.DB()
	if err != nil {
		log.Println(err)
	}
	defer db.Close()
	userID := strconv.Itoa(userDataID)
	resultsStatement, err := db.Prepare("SELECT id FROM `users` where `status`= 'ACTIVE' and `userID`=?")
	if err != nil {
		log.Println(err)
	}
	if err != nil {
		log.Println(err)
	}
	defer resultsStatement.Close()
	results, err := resultsStatement.Query(userID)
	if err != nil {
		log.Println(err)
	}
	if results.Next() {
		userModel := new(model.User)
		if err := results.Scan(&userModel.ID); err != nil {
			log.Println(err)
		}
		userModelID := strconv.FormatInt(userModel.ID, 10)
		updatedLastState, err := db.Query("update `users_last_state` set `status`='INACTIVE' where `userID`='" + userModelID + "'")
		if err != nil {
			log.Println(err)
		}
		defer updatedLastState.Close()
		insertedLastState, err := db.Query("INSERT INTO `users_last_state` (`userID`,`state`,`data`,`createdAt`,`updatedAt`) VALUES('" + userModelID + "','" + state + "','" + data + "','" + time.Now().UTC().Format("2006-01-02 03:04:05") + "','" + time.Now().UTC().Format("2006-01-02 03:04:05") + "')")
		if err != nil {
			log.Println(err)
		}
		defer insertedLastState.Close()
	}
}
