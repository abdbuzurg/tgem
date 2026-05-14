package model

type OperatorErrorFound struct {
	ID                 uint    `gorm:"primaryKey" json:"id"`
	InvoiceMaterialsID uint    `json:"invoiceMaterialID"`
	MaterialCostID     uint    `json:"materialCostID"`
	Amount             float64 `json:"amount"`
	Notes              string  `json:"notes"`
}
