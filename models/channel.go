package models

type Channel struct {
	ID                int64           `json:"id"`
	ChannelURL        string          `json:"channelURL"`
	ChannelID         string          `json:"channelID"`
	ChannelName       string          `json:"channelName"`
	UniqueID          string          `json:"uniqueID"`
	ChannelType       string          `json:"channelType"`
	ManualChannelName string          `json:"manualChannelName"`
	ChannelModel      string          `json:"channelModel"`
	CreatedAt         string          `json:"createdAt"`
	UpdatedAt         string          `json:"updatedAt"`
	Company           *Company        `json:"company"`
	User              *User           `json:"user"`
	Setting           *ChannelSetting `json:"setting"`
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
