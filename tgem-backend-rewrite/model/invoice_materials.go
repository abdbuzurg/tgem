package model

type InvoiceMaterials struct {
	ID             uint    `json:"id" gorm:"primaryKey"`
	ProjectID      uint    `json:"projectID"`
	MaterialCostID uint    `json:"materialCostID"`
	InvoiceID      uint    `json:"invoiceID"`
	InvoiceType    string  `json:"invoiceType"`
	IsDefected     bool    `json:"isDefected"`
	Amount         float64 `json:"amount"`
	Notes          string  `json:"notes"`
}
