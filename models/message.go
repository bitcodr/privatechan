package models

type Message struct {
	ID               int64    `json:"id"`
	CreatedAt        string   `json:"createdAt"`
	UpdatedAt        string   `json:"updatedAt"`
	UserID           int64    `json:"userID"`
	ChannelID        int64    `json:"channelID"`
	ParentID         int64    `json:"parentID"`
	ChannelMessageID string   `json:"channelMessageID"`
	BotMessageID     string   `json:"botMessageID"`
	Message          string   `json:"message"`
	Receiver         int64    `json:"receiver"`
	Type             string   `json:"type"`
	User             *User    `json:"user"`
	Channel          *Channel `json:"channel"`
}
