package model

import "time"

type InvoiceWriteOff struct {
	ID                 uint      `json:"id" gorm:"primaryKey"`
	ProjectID          uint      `json:"projectID"`
	ReleasedWorkerID   uint      `json:"releasedWorkerID"`
	WriteOffType       string    `json:"writeOffType"`
	WriteOffLocationID uint      `json:"writeOffLocationID"`
	DeliveryCode       string    `json:"deliveryCode"`
	DateOfInvoice      time.Time `json:"dateOfInvoice"`
	Confirmation       bool      `json:"confirmation"`
	DateOfConfirmation time.Time `json:"dateOfConfirmation"`
	Notes              string    `json:"notes"`
}
