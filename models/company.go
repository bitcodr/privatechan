package models

type Company struct {
	ID          int64      `json:"id"`
	CreatedAt   string     `json:"createdAt"`
	UpdatedAt   string     `json:"updatedAt"`
	CompanyName string     `json:"companyName"`
	Channels    []*Channel `json:"channels"`
}


type CompanyEmailSuffixes struct {
	ID        int64    `json:"id"`
	CreatedAt string   `json:"createdAt"`
	UpdatedAt string   `json:"updatedAt"`
	CompanyID string   `json:"companyID"`
	Suffix    string   `json:"suffix"`
	Company   *Company `json:"company"`
}



type CompanyChannel struct {
	ID        int64    `json:"id"`
	CreatedAt string   `json:"createdAt"`
	ChannelID string   `json:"channelID"`
	CompanyID string   `json:"companyID"`
	Company   *Company `json:"company"`
	Channel   *Channel `json:"channel"`
}