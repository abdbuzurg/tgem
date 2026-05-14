package dto

import (
	"backend-v2/model"
	"time"
)

type InvoiceObjectPaginated struct {
	ID                  uint      `json:"id"`
	DeliveryCode        string    `json:"deliveryCode"`
	SupervisorName      string    `json:"supervisorName"`
	DistrictID          uint      `json:"districtID"`
	DistrictName        string    `json:"districtName"`
	ObjectName          string    `json:"objectName"`
	ObjectType          string    `json:"objectType"`
	TeamNumber          string    `json:"teamNumber"`
	DateOfInvoice       time.Time `json:"dateOfInvoice"`
	ConfirmedByOperator bool      `json:"confirmedByOperator"`
}

type InvoiceObjectItem struct {
	MaterialID    uint     `json:"materialID"`
	Amount        float64  `json:"amount"`
	SerialNumbers []string `json:"serialNumbers"`
	Notes         string   `json:"notes"`
}

type InvoiceObjectOperation struct {
	OperationID uint    `json:"operationID"`
	Amount      float64 `json:"amount"`
	Notes       string  `json:"notes"`
}

type InvoiceObjectCreate struct {
	Details    model.InvoiceObject `json:"details"`
	Items      []InvoiceObjectItem `json:"items"`
	Operations []InvoiceObjectOperation     `json:"operations"`
}

type InvoiceObjectCreateQueryData struct {
	Invoice               model.InvoiceObject
	InvoiceMaterials      []model.InvoiceMaterials
	InvoiceOperations      []model.InvoiceOperations
	SerialNumberMovements []model.SerialNumberMovement
}

type InvoiceObjectFullDataItem struct {
	ID           uint    `json:"id"`
	MaterialName string  `json:"materialName"`
	Amount       float64 `json:"amount"`
	Notes        string  `json:"notes"`
}

type InvoiceObjectFullData struct {
	Details InvoiceObjectPaginated      `json:"details"`
	Items   []InvoiceObjectFullDataItem `json:"items"`
}

type InvoiceObjectWithMaterialsDescriptive struct {
	InvoiceData                  InvoiceObjectPaginated                    `json:"invoiceData"`
	MaterialsWithSerialNumber    []InvoiceMaterialsWithSerialNumberView    `json:"materialsWithSN"`
	MaterialsWithoutSerialNumber []InvoiceMaterialsWithoutSerialNumberView `json:"materialsWithoutSN"`
}

type InvoiceObjectTeamMaterials struct {
	MaterialID      uint    `json:"materialID"`
	MaterialName    string  `json:"materialName"`
	MaterialUnit    string  `json:"materialUnit"`
	HasSerialNumber bool    `json:"hasSerialNumber"`
	Amount          float64 `json:"amount"`
}

type InvoiceObjectOperationsBasedOnTeam struct {
	OperationID   uint   `json:"operationID"`
	OperationName string `json:"operationName"`
	MaterialID    uint   `json:"materialID"`
  MaterialName string `json:"materialName"`
}
