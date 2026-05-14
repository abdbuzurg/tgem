package model

type InvoiceCount struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	ProjectID   uint   `json:"projectID"`
	InvoiceType string `json:"invoiceType"`
	Count       uint   `json:"count"`
}
