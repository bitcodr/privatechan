//Package controllers ...
package controllers

import (
	"github.com/amiraliio/tgbp/config"
	tb "gopkg.in/tucnak/telebot.v2"
)

//TODO needs keyboard refactoring
//bot startup buttons
var addAnonMessage = tb.ReplyButton{
	Text: config.LangConfig.GetString("MESSAGES.ADD_MESSAGE_TO_CHANNEL"),
}
var setupVerifiedCompany = tb.ReplyButton{
	Text: config.LangConfig.GetString("MESSAGES.SETUP_VERIFIED_COMPANY"),
}
var joinCompanyChannels = tb.ReplyButton{
	Text: config.LangConfig.GetString("MESSAGES.JOIN_TO_COMPANY_CHANNEL"),
}
var StartBotKeys = [][]tb.ReplyButton{
	[]tb.ReplyButton{addAnonMessage},
	[]tb.ReplyButton{setupVerifiedCompany},
	[]tb.ReplyButton{joinCompanyChannels},
}
