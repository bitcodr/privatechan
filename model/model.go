//Package model ...
package model

type Company struct {
	ID          int64  `json:"id"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
	CompanyName string `json:"companyName"`
}

type Channel struct {
	ID          int64  `json:"id"`
	ChannelURL  string `json:"channelURL"`
	ChannelID   string `json:"channelID"`
	ChannelName string `json:"channelName"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type CompanyChannel struct {
	ID        int64  `json:"id"`
	CreatedAt string `json:"createdAt"`
	ChannelID string `json:"channelID"`
	CompanyID string `json:"companyID"`
}

type User struct {
	ID        int64  `json:"id"`
	Status    string `json:"status"`
	UserID    string `json:"userId"`
	Username  string `json:"username"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Lang      string `json:"lang"`
	Email     string `json:"email"`
	IsBot     string `json:"isBot"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type UserChannel struct {
	ID        int64  `json:"id"`
	Status    string `json:"status"`
	UserID    int64  `json:"userID"`
	ChannelID int64  `json:"channelID"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}
