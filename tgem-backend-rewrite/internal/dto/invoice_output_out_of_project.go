package dto

import (
	"backend-v2/model"
	"time"
)

type InvoiceOutputOutOfProjectPaginated struct {
	ID                 uint      `json:"id" gorm:"primaryKey"`
	NameOfProject      string    `json:"nameOfProject"`
	DeliveryCode       string    `json:"deliveryCode"`
	ReleasedWorkerName string    `json:"releasedWorkerName"`
	DateOfInvoice      time.Time `json:"dateOfInvoice"`
	Confirmation       bool      `json:"confirmation"`
}

type InvoiceOutputOutOfProjectCreateQueryData struct {
	Invoice          model.InvoiceOutputOutOfProject
	InvoiceMaterials []model.InvoiceMaterials
}

type InvoiceOutputOutOfProjectConfirmationQueryData struct {
	InvoiceData           model.InvoiceOutputOutOfProject
	WarehouseMaterials    []model.MaterialLocation
	OutOfProjectMaterials []model.MaterialLocation
}

type InvoiceOutputOutOfProject struct {
	Details model.InvoiceOutputOutOfProject `json:"details"`
	Items   []InvoiceOutputItem             `json:"items"`
}

type InvoiceOutputOutOfProjectSearchParameters struct {
	ProjectID        uint
	NameOfProject    string
	ReleasedWorkerID uint
}

type InvoiceOutputOutOfProjectReportFilter struct {
	DateFrom  time.Time `json:"dateFrom"`
	DateTo    time.Time `json:"dateTo"`
	ProjectID uint
}

type InvoiceOutputOutOfProjectReportData struct {
	ID                 uint
	NameOfProject      string
	DeliveryCode       string
	ReleasedWorkerName string
	DateOfInvoice      time.Time
}
