package model

type MaterialDefect struct {
	ID                 uint    `json:"id" gorm:"primaryKey"`
	Amount             float64 `json:"amount"`
	MaterialLocationID uint    `json:"materialLocationID"`
}
