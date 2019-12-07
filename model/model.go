//Package model ...
package model

type Company struct {
	ID          int64     `json:"id"`
	CreatedAt   string    `json:"createdAt"`
	UpdatedAt   string    `json:"updatedAt"`
	CompanyName string    `json:"companyName"`
	Channels    []Channel `json:"channels"`
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
	ID        int64     `json:"id"`
	CreatedAt string    `json:"createdAt"`
	UpdatedAt string    `json:"updatedAt"`
	UserID    int64     `json:"userID"`
	ChannelID int64     `json:"channelID"`
	ParentID  int64     `json:"parentID"`
	Messages  []Message `json:"messages"`
	User      User      `json:"user"`
	Channel   Channel   `json:"channel"`
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
