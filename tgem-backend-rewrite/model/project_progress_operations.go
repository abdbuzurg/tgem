package model

import "time"

type ProjectProgressOperations struct {
	ID          uint `gorm:"primaryKey"`
	ProjectID   uint
	OperationID uint
	Installed   float64
	Date        time.Time
}
