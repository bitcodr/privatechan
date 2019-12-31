//Package controllers ...
package controllers

import (
	tb "gopkg.in/tucnak/telebot.v2"
)

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
