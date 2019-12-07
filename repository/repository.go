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
				channelInserted, err := transaction.Exec("INSERT INTO `channels` (channelURL,channelID,channelName,createdAt,updatedAt) VALUES('" + channelURL + "','" + channelID + "','" + m.Chat.Title + "','" + time.Now().UTC().Format("2006-01-02 03:04:05") + "','" + time.Now().UTC().Format("2006-01-02 03:04:05") + "')")
				if err != nil {
					transaction.Rollback()
					log.Println(err)
				}
				insertedChannelID, err := channelInserted.LastInsertId()
				if err == nil {
					fmt.Println("ok1")
					//company name
					companyFlag := channelDetails[1]
					fmt.Println(companyFlag)
					//check if company is not exist
					companyExists, err := transaction.Query("SELECT id FROM `companies` where companyName=" + companyFlag)
					fmt.Println(companyExists)
					if err != nil {
						transaction.Rollback()
						log.Println(err)
					}
					fmt.Println("ok2")
					if !companyExists.Next() {
						fmt.Println("ok2")
						//insert company
						companyInserted, err := transaction.Exec("INSERT INTO `companies` (companyName,createdAt,updatedAt) VALUES('" + companyFlag + "','" + time.Now().UTC().Format("2006-01-02 03:04:05") + "','" + time.Now().UTC().Format("2006-01-02 03:04:05") + "')")
						if err != nil {
							transaction.Rollback()
							log.Println(err)
						}
						insertedCompanyID, err := companyInserted.LastInsertId()
						fmt.Println("ok3")
						if err == nil {
							companyModelID := strconv.FormatInt(insertedCompanyID, 10)
							channelModelID := strconv.FormatInt(insertedChannelID, 10)
							//insert company channel pivot
							_, err := transaction.Exec("INSERT INTO `companies_channels` (companyID,channelID,createdAt) VALUES('" + companyModelID + "','" + channelModelID + "','" + time.Now().UTC().Format("2006-01-02 03:04:05") + "')")
							if err != nil {
								transaction.Rollback()
								log.Println(err)
							}
						}
					}
					_ = transaction.Commit()
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
		isBot := strconv.FormatBool(m.Sender.IsBot)
		//check if user is not created
		results, err := db.Query("SELECT id FROM `users` where `status`= 'ACTIVE' and usersID=" + userID)
		if err != nil {
			log.Println(err)
		}
		if !results.Next() {
			//insert user
			userInsert, err := db.Query("INSERT INTO `users` (userID,username,firstName,lastName,lang,isBot,createdAt,updatedAt) VALUES('" + userID + "','" + m.Sender.Username + "','" + m.Sender.FirstName + "','" + m.Sender.LastName + "','" + m.Sender.LanguageCode + "','" + isBot + "','" + time.Now().UTC().Format("2006-01-02 03:04:05") + "','" + time.Now().UTC().Format("2006-01-02 03:04:05") + "')")
			if err != nil {
				log.Println(err)
			}
			defer userInsert.Close()
			checkAndInsertUserChannel(userInsert, channelID, db)
			//TODO add channel and userId if channel for user is not exist
			//TODO show verification btn and send email
		} else {
			checkAndInsertUserChannel(results, channelID, db)
		}
	}
}

func checkAndInsertUserChannel(results *sql.Rows, channelID string, db *sql.DB) {
	userModel := new(model.User)
	if err := results.Scan(userModel); err != nil {
		log.Println(err)
	}
	userModelID := strconv.FormatInt(userModel.ID, 10)
	//check if channel for user is exists
	checkUserChannel, err := db.Query("SELECT ch.id as id FROM `channels` as ch inner join `users_channels` as uc on uc.channelID = ch.id and uc.userID=" + userModelID + " and ch.channelID=" + channelID)
	if err != nil {
		log.Println(err)
	}
	if !checkUserChannel.Next() {
		getChannelID, err := db.Query("SELECT id FROM `channels` channelID=" + channelID)
		if err != nil {
			log.Println(err)
		}
		if getChannelID.Next() {
			channelModel := new(model.Channel)
			if err := getChannelID.Scan(channelModel); err != nil {
				log.Println(err)
			}
			channelModelID := strconv.FormatInt(channelModel.ID, 10)
			userChannelInserted, err := db.Query("INSERT INTO `users_channels` (userID,channelID,createdAt,updatedAt) VALUES('" + userModelID + "','" + channelModelID + "','" + time.Now().UTC().Format("2006-01-02 03:04:05") + "','" + time.Now().UTC().Format("2006-01-02 03:04:05") + "')")
			if err != nil {
				log.Println(err)
			}
			defer userChannelInserted.Close()
		}
	}
	channelModel := new(model.Channel)
	if err := checkUserChannel.Scan(channelModel); err != nil {
		log.Println(err)
	}
	channelModelID := strconv.FormatInt(channelModel.ID, 10)
	//check the user is active or not
	checkUserChannelActivity, err := db.Query("SELECT id,status from `users_channels` where userID=" + userModelID + " and channelID=" + channelModelID + " and status='ACTIVE'")
	if err != nil {
		log.Println(err)
	}
	if checkUserChannelActivity.Next() {
		//TODO show all keyboards
		fmt.Println("active")
	} else {
		fmt.Println("inactive")
		//TODO show verification btn and get user email to send a code for activation
	}
}
