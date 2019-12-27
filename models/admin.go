//Package model ...
package models

type TempSetupFlow struct {
	ID         int64  `json:"id"`
	TableName  string `json:"tableName"`
	ColumnName string `json:"columnName"`
	Data       string `json:"data"`
	Relation   string `json:"relation"`
	Status     string `json:"status"`
	UserID     int64  `json:"userID"`
	CreatedAt  string `json:"createdAt"`
	UpdatedAt  string `json:"updatedAt"`
	User       *User  `json:"user"`
}
