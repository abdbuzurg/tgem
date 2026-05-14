package model

type InvoiceOperations struct {
	ID          uint    `json:"id" gorm:"primaryKey"`
	ProjectID   uint    `json:"projectID"`
	OperationID uint    `json:"operationID"`
	InvoiceID   uint    `json:"invoiceID"`
	InvoiceType string  `json:"invoiceType"`
	Amount      float64 `json:"amount"`
	Notes       string  `json:"notes"`
}
