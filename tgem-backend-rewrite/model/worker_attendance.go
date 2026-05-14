package model

import "time"

type WorkerAttendance struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	ProjectID uint      `json:"projectID"`
	WorkerID  uint      `json:"workerID"`
	Start     time.Time `json:"start"`
	End       time.Time `json:"end"`
}
