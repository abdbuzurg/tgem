package model

import "time"

type InvoiceOutputOutOfProject struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	ProjectID        uint      `json:"ProjectID"`
	DeliveryCode     string    `json:"deliveryCode"`
	ReleasedWorkerID uint      `json:"releasedWorkerID"`
	NameOfProject    string    `json:"nameOfProject"`
	DateOfInvoice    time.Time `json:"dateOfInvoice"`
	Notes            string    `json:"notes"`
	Confirmation     bool      `json:"confirmation"`
}
