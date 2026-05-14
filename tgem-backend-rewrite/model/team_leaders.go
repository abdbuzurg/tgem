package model

type TeamLeaders struct {
	ID             uint `json:"id" gorm:"primaryKey"`
	TeamID         uint `json:"teamID"`
	LeaderWorkerID uint `json:"leaderWorkerID"`
}
