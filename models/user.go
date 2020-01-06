package models

type User struct {
	ID        int64      `json:"id"`
	Status    string     `json:"status"`
	UserID    string     `json:"userId"`
	Username  string     `json:"username"`
	FirstName string     `json:"firstName"`
	LastName  string     `json:"lastName"`
	Lang      string     `json:"lang"`
	Email     string     `json:"email"`
	IsBot     string     `json:"isBot"`
	CustomID  string     `json:"customID"`
	CreatedAt string     `json:"createdAt"`
	UpdatedAt string     `json:"updatedAt"`
	Channels  []*Channel `json:"channels"`
}

type UserChannel struct {
	ID        int64    `json:"id"`
	Status    string   `json:"status"`
	UserID    int64    `json:"userID"`
	ChannelID int64    `json:"channelID"`
	CreatedAt string   `json:"createdAt"`
	UpdatedAt string   `json:"updatedAt"`
	User      *User    `json:"user"`
	Channel   *Channel `json:"channel"`
}

type UsersCurrentActiveChannel struct {
	ID        int64    `json:"id"`
	CreatedAt string   `json:"createdAt"`
	UpdatedAt string   `json:"updatedAt"`
	UserID    int64    `json:"userID"`
	ChannelID int64    `json:"channelID"`
	Status    string   `json:"status"`
	User      *User    `json:"user"`
	Channel   *Channel `json:"channel"`
}

type UserLastState struct {
	ID        int64  `json:"id"`
	Status    string `json:"status"`
	State     string `json:"state"`
	UserID    int64  `json:"userID"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	Data      string `json:"data"`
	User      *User  `json:"user"`
}

type UsersActivationKey struct {
	ID        int64  `json:"id"`
	UserID    int64  `json:"userID"`
	CreatedAt string `json:"createdAt"`
	ActiveKey string `json:"activeKey"`
	User      *User  `json:"user"`
}
