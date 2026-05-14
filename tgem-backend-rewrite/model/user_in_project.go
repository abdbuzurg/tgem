package model

type UserInProject struct {
	ID        uint `json:"id" gorm:"primaryKey"`
	ProjectID uint `json:"projectID"`
	UserID    uint `json:"userID"`
}
