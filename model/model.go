//Package model ...
package model

type Company struct {
	ID          int64     `json:"id"`
	CreatedAt   string    `json:"createdAt"`
	UpdatedAt   string    `json:"updatedAt"`
	CompanyName string    `json:"companyName"`
	Channels    []Channel `json:"channels"`
}

type CompanyEmailSuffixes struct {
	ID        int64   `json:"id"`
	CreatedAt string  `json:"createdAt"`
	UpdatedAt string  `json:"updatedAt"`
	CompanyID string  `json:"companyID"`
	Suffix    string  `json:"suffix"`
	Company   Company `json:"company"`
}

type Channel struct {
	ID          int64   `json:"id"`
	ChannelURL  string  `json:"channelURL"`
	ChannelID   string  `json:"channelID"`
	ChannelName string  `json:"channelName"`
	CreatedAt   string  `json:"createdAt"`
	UpdatedAt   string  `json:"updatedAt"`
	Company     Company `json:"company"`
}

type CompanyChannel struct {
	ID        int64   `json:"id"`
	CreatedAt string  `json:"createdAt"`
	ChannelID string  `json:"channelID"`
	CompanyID string  `json:"companyID"`
	Company   Company `json:"company"`
	Channel   Channel `json:"channel"`
}

type ChannelSetting struct {
	ID               int64   `json:"id"`
	CreatedAt        string  `json:"createdAt"`
	UpdatedAt        string  `json:"updatedAt"`
	JoinVerify       bool    `json:"joinVerify"`
	NewMessageVerify bool    `json:"newMessageVerify"`
	ReplyVerify      bool    `json:"replyVerify"`
	DirectVerify     bool    `json:"directVerify"`
	ChannelID        string  `json:"channelID"`
	Channel          Channel `json:"channel"`
}

type User struct {
	ID        int64     `json:"id"`
	Status    string    `json:"status"`
	UserID    string    `json:"userId"`
	Username  string    `json:"username"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Lang      string    `json:"lang"`
	Email     string    `json:"email"`
	IsBot     string    `json:"isBot"`
	CreatedAt string    `json:"createdAt"`
	UpdatedAt string    `json:"updatedAt"`
	Channels  []Channel `json:"channels"`
}

type UserChannel struct {
	ID        int64   `json:"id"`
	Status    string  `json:"status"`
	UserID    int64   `json:"userID"`
	ChannelID int64   `json:"channelID"`
	CreatedAt string  `json:"createdAt"`
	UpdatedAt string  `json:"updatedAt"`
	User      User    `json:"user"`
	Channel   Channel `json:"channel"`
}

type Message struct {
	ID               int64   `json:"id"`
	CreatedAt        string  `json:"createdAt"`
	UpdatedAt        string  `json:"updatedAt"`
	UserID           int64   `json:"userID"`
	ChannelID        int64   `json:"channelID"`
	ParentID         int64   `json:"parentID"`
	ChannelMessageID string  `json:"channelMessageID"`
	BotMessageID     string  `json:"botMessageID"`
	User             User    `json:"user"`
	Channel          Channel `json:"channel"`
}

type UsersCurrentActiveChannel struct {
	ID        int64   `json:"id"`
	CreatedAt string  `json:"createdAt"`
	UpdatedAt string  `json:"updatedAt"`
	UserID    int64   `json:"userID"`
	ChannelID int64   `json:"channelID"`
	Status    string  `json:"status"`
	User      User    `json:"user"`
	Channel   Channel `json:"channel"`
}

type UserLastState struct {
	ID        int64  `json:"id"`
	Status    string `json:"status"`
	State     string `json:"state"`
	UserID    int64  `json:"userID"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	Data      string `json:"data"`
	User      User   `json:"user"`
}
