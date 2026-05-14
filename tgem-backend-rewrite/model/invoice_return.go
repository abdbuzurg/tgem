package model

import "time"

type InvoiceReturn struct {
	ID                 uint      `json:"ID" gorm:"primaryKey"`
	ProjectID          uint      `json:"projectID"`
	DistrictID         uint      `json:"districtID"`
	ReturnerType       string    `json:"returnerType"`
	ReturnerID         uint      `json:"returnerID"`
	AcceptorType       string    `json:"acceptorType"`
	AcceptorID         uint      `json:"acceptorID"`
	AcceptedByWorkerID uint      `json:"acceptedByWorkerID"`
	DateOfInvoice      time.Time `json:"dateOfInvoice"`
	Notes              string    `json:"notes"`
	DeliveryCode       string    `json:"deliveryCode"`
	Confirmation       bool      `json:"confirmation"`
}
