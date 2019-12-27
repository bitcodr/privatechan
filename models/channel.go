package models


type Channel struct {
	ID                int64    `json:"id"`
	ChannelURL        string   `json:"channelURL"`
	ChannelID         string   `json:"channelID"`
	ChannelName       string   `json:"channelName"`
	UniqueID          string   `json:"uniqueID"`
	ManualChannelName string   `json:"manualChannelName"`
	CreatedAt         string   `json:"createdAt"`
	UpdatedAt         string   `json:"updatedAt"`
	Company           *Company `json:"company"`
	User              *User    `json:"user"`
}


type ChannelSetting struct {
	ID               int64    `json:"id"`
	CreatedAt        string   `json:"createdAt"`
	UpdatedAt        string   `json:"updatedAt"`
	JoinVerify       bool     `json:"joinVerify"`
	NewMessageVerify bool     `json:"newMessageVerify"`
	ReplyVerify      bool     `json:"replyVerify"`
	DirectVerify     bool     `json:"directVerify"`
	ChannelID        string   `json:"channelID"`
	Channel          *Channel `json:"channel"`
}