package model

type InvoiceObjectOperator struct {
	ID               uint `json:"id"`
	OperatorWorkerID uint `json:"operatorWorkerID"`
	InvoiceObjectID  uint `json:"invoiceObjectID"`
}
