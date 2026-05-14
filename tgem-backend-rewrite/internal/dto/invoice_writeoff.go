package dto

import (
	"backend-v2/model"
	"time"
)

type InvoiceWriteOffPaginated struct {
	ID                   uint      `json:"id" gorm:"primaryKey"`
	WriteOffType         string    `json:"writeOffType"`
	WriteOffLocationID   uint      `json:"writeOffLocationID"`
	WriteOffLocationName string    `json:"writeOffLocationName"`
	ReleasedWorkerID     uint      `json:"releasedWorkerID"`
	ReleasedWorkerName   string    `json:"releasedWorkerName"`
	DeliveryCode         string    `json:"deliveryCode"`
	DateOfInvoice        time.Time `json:"dateOfInvoice"`
	Confirmation         bool      `json:"confirmation"`
	DateOfConfirmation   time.Time `json:"dateOfConfirmation"`
}

type InvoiceWriteOffItem struct {
	MaterialID    uint     `json:"materialID"`
	Amount        float64  `json:"amount"`
	Notes         string   `json:"notes"`
	SerialNumbers []string `json:"serialNumbers"`
}

type InvoiceWriteOff struct {
	Details model.InvoiceWriteOff `json:"details"`
	Items   []InvoiceWriteOffItem `json:"items"`
}

type InvoiceWriteOffSearchParameters struct {
	ProjectID    uint
	WriteOffType string
}

type InvoiceWriteOffMutationData struct {
	InvoiceWriteOff  model.InvoiceWriteOff
	InvoiceMaterials []model.InvoiceMaterials
}

type InvoiceWriteOffMaterialsForEdit struct {
	MaterialID      uint     `json:"materialID"`
	MaterialName    string   `json:"materialName"`
	Unit            string   `json:"unit"`
	Amount          float64  `json:"amount"`
	MaterialCostID  uint     `json:"materialCostID"`
	MaterialCost    float64  `json:"materialCost"`
	Notes           string   `json:"notes"`
	HasSerialNumber bool     `json:"hasSerialNumber"`
	SerialNumbers   []string `json:"serialNumbers"`
	LocationAmount  float64  `json:"locationAmount"`
}

type InvoiceWriteOffConfirmationData struct {
	InvoiceWriteOff     model.InvoiceWriteOff
	MaterialsInLocation []model.MaterialLocation
	MaterialsInWriteOff []model.MaterialLocation
}

type InvoiceWriteOffReportParameters struct {
	ProjectID          uint
	WriteOffType       string    `json:"writeOffType"`
	WriteOffLocationID uint      `json:"writeOffLocationID"`
	DateFrom           time.Time `json:"dateFrom"`
	DateTo             time.Time `json:"dateTo"`
}

type InvoiceWriteOffReportData struct {
	ID                 uint
	DeliveryCode       string
	ReleasedWorkerName string
	DateOfInvoice      time.Time
}
