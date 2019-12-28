package controllers

import (
	"database/sql"
	"github.com/amiraliio/tgbp/config"
	"github.com/amiraliio/tgbp/models"
	tb "gopkg.in/tucnak/telebot.v2"
	"log"
)

type BotService struct{}

type Event struct {
	UserState, Command, Command1, Controller string
	Event                                    interface{}
}

func Init(app *config.App, bot *tb.Bot) {
	triggersEvents(app, bot)
	keyboardsEvents(app, bot)
	onTextEvents(app, bot)
	onCallbackEvents(app, bot)
}

func GetUserLastState(db *sql.DB, app *config.App, bot *tb.Bot, m *tb.Message, user int) *models.UserLastState {
	userLastStateQueryStatement, err := db.Prepare("SELECT `data`,`state`,`userID` from `users_last_state` where `userId`=? order by `createdAt` DESC limit 1")
	if err != nil {
		log.Println(err)
	}
	defer userLastStateQueryStatement.Close()
	userLastStateQuery, err := userLastStateQueryStatement.Query(user)
	if err != nil {
		log.Println(err)
	}
	userLastState := new(models.UserLastState)
	if userLastStateQuery.Next() {
		if err := userLastStateQuery.Scan(&userLastState.Data, &userLastState.State, &userLastState.UserID); err != nil {
			log.Println(err)
		}
		return userLastState
	}
	return userLastState
}

//TODO needs keyboard refactoring
//bot startup buttons
var addAnonMessage = tb.ReplyButton{
	Text: "Add Anonymous Message to a Channel/Group",
}
var setupVerifiedCompany = tb.ReplyButton{
	Text: "Setup Verified Company Account",
}
var joinCompanyChannels = tb.ReplyButton{
	Text: "Join To Company Anonymous Channel/Group",
}
var StartBotKeys = [][]tb.ReplyButton{
	[]tb.ReplyButton{addAnonMessage},
	[]tb.ReplyButton{setupVerifiedCompany},
	[]tb.ReplyButton{joinCompanyChannels},
}
