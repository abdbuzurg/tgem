package model

import (
	"time"
)

type InvoiceOutput struct {
	ID                       uint      `json:"id" gorm:"primaryKey"`
	DistrictID               uint      `json:"districtID"`
	ProjectID                uint      `json:"projectID"`
	WarehouseManagerWorkerID uint      `json:"warehouseManagerWorkerID"`
	ReleasedWorkerID         uint      `json:"releasedWorkerID"`
	RecipientWorkerID        uint      `json:"recipientWorkerID"`
	TeamID                   uint      `json:"teamID"`
	DeliveryCode             string    `json:"deliveryCode"`
	DateOfInvoice            time.Time `json:"dateOfInvoice"`
	Notes                    string    `json:"notes"`
	Confirmation             bool      `json:"confirmation"`
}
