//Package repository ...
package repository

import (
	"database/sql"
	"fmt"
	"github.com/amiraliio/tgbp/config"
	"github.com/amiraliio/tgbp/model"
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
	tb "gopkg.in/tucnak/telebot.v2"
	"log"
	"strconv"
	"strings"
	"time"
)

//TODO some methods need transactions

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
			results, err := db.Query("SELECT id FROM `channels` where channelID=" + channelID)
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
				channelInserted, err := transaction.Exec("INSERT INTO `channels` (`channelURL`,`channelID`,`channelName`,`createdAt`,`updatedAt`) VALUES('" + channelURL + "','" + channelID + "','" + m.Chat.Title + "','" + time.Now().UTC().Format("2006-01-02 03:04:05") + "','" + time.Now().UTC().Format("2006-01-02 03:04:05") + "')")
				if err != nil {
					transaction.Rollback()
					log.Println(err)
				}
				insertedChannelID, err := channelInserted.LastInsertId()
				if err == nil {
					//company name
					companyFlag := channelDetails[1]
					//check if company is not exist
					companyExists, err := db.Query("SELECT id FROM `companies` where `companyName`='" + companyFlag + "'")
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

func JoinFromChannel(bot *tb.Bot, m *tb.Message) {
	data := strings.Split(m.Text, "join_group")
	if len(data) == 2 {
		channelID := data[1]
		db, err := config.DB()
		if err != nil {
			log.Println(err)
		}
		defer db.Close()
		userID := strconv.Itoa(m.Sender.ID)
		//check if user is not created
		results, err := db.Query("SELECT id FROM `users` where `status`= 'ACTIVE' and `userID`='" + userID + "'")
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
			checkAndInsertUserChannel(bot, m, insertedUserID, channelID, db, transaction)
			//TODO add channel and userId if channel for user is not exist
			//TODO show verification btn and send email
		} else {
			userModel := new(model.User)
			if err := results.Scan(&userModel.ID); err != nil {
				transaction.Rollback()
				log.Println(err)
			}
			checkAndInsertUserChannel(bot, m, userModel.ID, channelID, db, transaction)
		}
	}
}

func checkAndInsertUserChannel(bot *tb.Bot, m *tb.Message, queryUserID int64, channelID string, db *sql.DB, transaction *sql.Tx) {
	userModelID := strconv.FormatInt(queryUserID, 10)
	//check if channel for user is exists
	checkUserChannel, err := db.Query("SELECT ch.id as id FROM `channels` as ch inner join `users_channels` as uc on uc.channelID = ch.id and uc.userID='" + userModelID + "' and ch.channelID='" + channelID + "'")
	if err != nil {
		transaction.Rollback()
		log.Println(err)
	}
	if !checkUserChannel.Next() {
		getChannelID, err := db.Query("SELECT `id` FROM `channels` where `channelID`='" + channelID + "'")
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
	channelDetail, err := db.Query("SELECT ch.`id` as id, ch.channelName as channelName, co.companyName as companyName  FROM `channels` as ch inner join `companies_channels` as cc on ch.id = cc.channelID inner join `companies` as co on cc.companyID = co.id where ch.`channelID`='" + channelID + "'")
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
		checkUserChannelActivity, err := db.Query("SELECT `id`,`status` from `users_channels` where `userID`='" + userModelID + "' and `channelID`='" + channelModelID + "' and `status`='ACTIVE'")
		if err != nil {
			transaction.Rollback()
			log.Println(err)
		}
		transaction.Commit()
		if checkUserChannelActivity.Next() {
			//TODO show all keyboards
			fmt.Println("active")
		} else {
			_, err := bot.Send(m.Chat, "You trying to join to the channel "+channelModelData.ChannelName+" blongs to company "+companyModel.CompanyName+" If do want to send a message on channel you should send and verify your email address")
			if err != nil {
				log.Println(err)
			}
			//TODO show verification btn and get user email to send a code for activation
		}
	}
}
