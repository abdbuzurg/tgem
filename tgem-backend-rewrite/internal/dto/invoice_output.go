package dto

import (
	"backend-v2/model"
	"time"

	"github.com/shopspring/decimal"
)

type InvoiceOutputPaginated struct {
	ID                   uint      `json:"id"`
	WarehouseManagerID   uint      `json:"warehouseManagerID"`
	WarehouseManagerName string    `json:"warehouseManagerName"`
	ReleasedName         string    `json:"releasedName"`
	RecipientID          uint      `json:"recipientID"`
	RecipientName        string    `json:"recipientName"`
	TeamID               uint      `json:"teamID"`
	TeamName             string    `json:"teamName"`
	DistrictID           uint      `json:"districtID"`
	DistrictName         string    `json:"districtName"`
	DeliveryCode         string    `json:"deliveryCode"`
	Notes                string    `json:"notes"`
	DateOfInvoice        time.Time `json:"dateOfInvoice"`
	Confirmation         bool      `json:"confirmation"`
}

type InvoiceOutputItem struct {
	MaterialID    uint     `json:"materialID"`
	Amount        float64  `json:"amount"`
	SerialNumbers []string `json:"serialNumbers"`
	Notes         string   `json:"notes"`
}

type InvoiceOutput struct {
	Details model.InvoiceOutput `json:"details"`
	Items   []InvoiceOutputItem `json:"items"`
}

type InvoiceObject struct {
	ID             uint   `json:"id"`
	TeamLeaderName string `json:"teamLeaderName"`
	TeamNumber     string `json:"teamNumber"`
	ObjectName     string `json:"objectName"`
}

type InvoiceOutputCreateQueryData struct {
	Invoice               model.InvoiceOutput
	InvoiceMaterials      []model.InvoiceMaterials
	SerialNumberMovements []model.SerialNumberMovement
}

type InvoiceOutputConfirmationQueryData struct {
	InvoiceData        model.InvoiceOutput
	WarehouseMaterials []model.MaterialLocation
	TeamMaterials      []model.MaterialLocation
}

type InvoiceOutputReportFilterRequest struct {
	ProjectID          uint      `json:"projectID"`
	Code               string    `json:"code"`
	WarehouseManagerID uint      `json:"warehouseManagerID"`
	ReceivedID         uint      `json:"recievedID"`
	TeamID             uint      `json:"teamID"`
	DistrictID         uint      `json:"districtID"`
	DateFrom           time.Time `json:"dateFrom"`
	DateTo             time.Time `json:"dateTo"`
}

type InvoiceOutputReportFilter struct {
	Code               string
	WarehouseManagerID uint
	ReceivedID         uint
	TeamID             uint
	DistrictID         uint
	DateFrom           time.Time
	DateTo             time.Time
}

type AvailableMaterialsInWarehouse struct {
	ID              uint    `json:"id"`
	Name            string  `json:"name"`
	Unit            string  `json:"unit"`
	HasSerialNumber bool    `json:"hasSerialNumber"`
	Amount          float64 `json:"amount"`
}

type MaterialAmountSortedByCostM19QueryResult struct {
	MaterialID      uint
	MaterialCostID  uint
	MaterialCostM19 decimal.Decimal
	MaterialAmount  float64
}

type MaterialCostIDAndSNLocationIDQueryResult struct {
	MaterialCostID         uint
	SerialNumberID         uint
	SerialNumberLocationID uint
}

type InvoiceOutputDataForExcelQueryResult struct {
	ID                   uint
	ProjectName          string
	DeliveryCode         string
	DistrictName         string
	ObjectType           string
	ObjectName           string
	TeamLeaderName       string
	WarehouseManagerName string
	ReleasedName         string
	RecipientName        string
	DateOfInvoice        time.Time
}

type InvoiceOutputMaterialsForEdit struct {
	MaterialID      uint     `json:"materialID"`
	MaterialName    string   `json:"materialName"`
	Unit            string   `json:"unit"`
	WarehouseAmount float64  `json:"warehouseAmount"`
	Amount          float64  `json:"amount"`
	Notes           string   `json:"notes"`
	HasSerialNumber bool     `json:"hasSerialNumber"`
}

type InvoiceOutputDataForReport struct {
	ID                   uint      `json:"id"`
	DeliveryCode         string    `json:"deliveryCode"`
	WarehouseManagerName string    `json:"warehouseManagerName"`
	RecipientName        string    `json:"recipientName"`
	TeamNumber           string    `json:"teamNumber"`
	TeamLeaderName       string    `json:"teamLeaderName"`
	DateOfInvoice        time.Time `json:"dateOfInvoice"`
}

type InvoiceOutputMaterialDataForReport struct {
	MaterialName    string
	MaterialUnit    string
	MaterialCostM19 decimal.Decimal
	Notes           string
	Amount          float64
}

type InvoiceOutputImportData struct {
	Details model.InvoiceOutput
	Items   []model.InvoiceMaterials
}
