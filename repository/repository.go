//Package repository ...
package repository

import (
	"github.com/amiraliio/tgbp/config"
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
	tb "gopkg.in/tucnak/telebot.v2"
	"log"
	"strconv"
	"strings"
	"time"
)

//RegisterChannel
func RegisterChannel(bot *tb.Bot, m *tb.Message) {
	if strings.Contains(m.Text, "https://t.me/joinchat/") && strings.Contains(m.Text, "@") {
		channelDetails := strings.SplitAfter(m.Text, "@")
		if len(channelDetails) == 2 && strings.Contains(channelDetails[0], "@") {
			//channel private url
			channelURL := strings.ReplaceAll(channelDetails[0], "@", "")
			//company name
			companyName := channelDetails[1]
			db, err := config.DB()
			if err != nil {
				log.Println(err)
			}
			channelID := strconv.FormatInt(m.Chat.ID, 10)
			defer db.Close()
			results, err := db.Query("SELECT channelID name FROM channels where channelID=" + channelID)
			if err != nil {
					log.Println(err)
			}
			if !results.Next() {
				inserted, err := db.Query("INSERT INTO `channels` VALUES(1,'" + channelURL + "','" + companyName + "','" + channelID + "','" + m.Chat.Title + "','" + time.Now().UTC().Format("2006-01-02 03:04:05") + "','" + time.Now().UTC().Format("2006-01-02 03:04:05") + "')")
				if err != nil {
					log.Println(err)
				}
				defer inserted.Close()
				successMessage, _ := bot.Send(m.Chat, "You're channel registered successfully")
				time.Sleep(3 * time.Second)
				bot.Delete(successMessage)
				pinMessage, _ := bot.Send(m.Chat, "You can send a message in this channel, by https://t.me/"+viper.GetString("APP.BOTUSERNAME"))
				bot.Pin(pinMessage)
				bot.Delete(m)
			}
		}
	}
}
