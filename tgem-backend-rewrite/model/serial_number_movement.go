package model

type SerialNumberMovement struct {
	ID             uint   `json:"id" gorm:"primaryKey"`
	SerialNumberID uint   `json:"serialNumberID"`
	ProjectID      uint   `json:"projectID"`
	InvoiceID      uint   `json:"invoiceID"`
	InvoiceType    string `json:"invoiceType"`
	IsDefected     bool   `json:"isDefected"`
	Confirmation   bool   `json:"confirmation"`
}
