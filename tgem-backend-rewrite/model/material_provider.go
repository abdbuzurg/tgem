package model

type MaterialProvider struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	ProjectID uint   `json:"projectID"`
	Name      string `json:"name"`
}
