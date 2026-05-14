package model

import "time"

type ProjectProgressMaterials struct {
	ID                uint `gorm:"primaryKey"`
	ProjectID         uint
	MaterialCostID    uint
	Received          float64
	Installed         float64
	AmountInWarehouse float64
	AmountInTeams     float64
	AmountInObjects   float64
	AmountWriteOff    float64
	Date              time.Time
}
