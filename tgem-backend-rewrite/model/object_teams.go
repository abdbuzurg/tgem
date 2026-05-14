package model

type ObjectTeams struct {
	ID       uint `json:"id" gorm:"primaryKey"`
	TeamID   uint `json:"teamID"`
	ObjectID uint `json:"objectID"`
}
