package model

import "time"

type InvoiceInput struct {
	ID                       uint      `json:"id" gorm:"primaryKey"`
	ProjectID                uint      `json:"projectID"`
	WarehouseManagerWorkerID uint      `json:"warehouseManagerWorkerID"`
	ReleasedWorkerID         uint      `json:"releasedWorkerID"`
	DeliveryCode             string    `json:"deliveryCode" gorm:"tinyText"`
	Notes                    string    `json:"notes"`
	DateOfInvoice            time.Time `json:"dateOfInvoice"`
	Confirmed                bool      `json:"confirmation"`
}
