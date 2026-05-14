package dto

import (
	"backend-v2/model"
	"time"
)

type InvoiceReturnPaginated struct {
	ID            uint      `json:"id"`
	DeliveryCode  string    `json:"deliveryCode"`
	ProjectName   string    `json:"projectName"`
	ReturnerType  string    `json:"returnerType"`
	ReturnerName  string    `json:"returnerName"`
	DateOfInvoice time.Time `json:"dateOfInvoice"`
	Notes         string    `json:"notes"`
	Confirmation  bool      `json:"confirmation"`
}

type InvoiceReturnTeamPaginatedQueryData struct {
	ID             uint   `json:"id"`
	DeliveryCode   string `json:"deliveryCode"`
	DistrictName   string `json:"districtName"`
	AcceptorName   string `json:"acceptorName"`
	TeamNumber     string `json:"teamNumber"`
	TeamLeaderName string `json:"teamLeaderName"`
	DateOfInvoice  string `json:"dateOfInvoice"`
	Confirmation   bool   `json:"confirmation"`
}

type InvoiceReturnTeamPaginated struct {
	ID              uint     `json:"id"`
	DeliveryCode    string   `json:"deliveryCode"`
	DistrictName    string   `json:"districtName"`
	AcceptorName    string   `json:"acceptorName"`
	TeamNumber      string   `json:"teamNumber"`
	TeamLeaderNames []string `json:"teamLeaderNames"`
	DateOfInvoice   string   `json:"dateOfInvoice"`
	Confirmation    bool     `json:"confirmation"`
}

type InvoiceReturnObjectPaginatedQueryData struct {
	ID                   uint   `json:"id"`
	DeliveryCode         string `json:"deliveryCode"`
	AcceptorName         string `json:"acceptorName"`
	DistrictName         string `json:"districtName"`
	ObjectName           string `json:"objectName"`
	ObjectSupervisorName string `json:"objectSupervisorName"`
	ObjectType           string `json:"objectType"`
	TeamNumber           string `json:"teamNumber"`
	TeamLeaderName       string `json:"teamLeaderName"`
	DateOfInvoice        string `json:"dateOfInvoice"`
	Confirmation         bool
}

type InvoiceReturnObjectPaginated struct {
	ID                    uint     `json:"id"`
	DeliveryCode          string   `json:"deliveryCode"`
	AcceptorName          string   `json:"acceptorName"`
	DistrictName          string   `json:"districtName"`
	ObjectName            string   `json:"objectName"`
	ObjectSupervisorNames []string `json:"objectSupervisorNames"`
	ObjectType            string   `json:"objectType"`
	TeamNumber            string   `json:"teamNumber"`
	TeamLeaderName        string   `json:"teamLeaderName"`
	DateOfInvoice         string   `json:"dateOfInvoice"`
	Confirmation          bool     `json:"confirmation"`
}

type InvoiceReturnItem struct {
	MaterialID    uint     `json:"materialID"`
	Amount        float64  `json:"amount"`
	IsDefected    bool     `json:"isDefected"`
	SerialNumbers []string `json:"serialNumbers"`
	Notes         string   `json:"notes"`
}

type InvoiceReturn struct {
	Details model.InvoiceReturn `json:"details"`
	Items   []InvoiceReturnItem `json:"items"`
}

type InvoiceReturnReportFilterRequest struct {
	Code         string    `json:"code"`
	ReturnerType string    `json:"returnerType"`
	Returner     string    `json:"returner"`
	DateFrom     time.Time `json:"dateFrom"`
	DateTo       time.Time `json:"dateTo"`
}

type InvoiceReturnReportFilter struct {
	Code         string
	ReturnerType string
	ReturnerID   uint
	DateFrom     time.Time
	DateTo       time.Time
}

type InvoiceReturnCreateQueryData struct {
	Invoice               model.InvoiceReturn
	InvoiceMaterials      []model.InvoiceMaterials
	SerialNumberMovements []model.SerialNumberMovement
}

type InvoiceReturnMaterialsForExcel struct {
	MaterialCode     string
	MaterialName     string
	MaterialUnit     string
	MaterialDefected bool
	MaterialAmount   float64
	MaterialNotes    string
}

type InvoiceReturnTeamDataForExcel struct {
	ProjectName    string
	DeliveryCode   string
	DistrictName   string
	DateOfInvoice  time.Time
	TeamNumber     string
	TeamLeaderName string
	AcceptorName   string
}

type InvoiceReturnObjectDataForExcel struct {
	DeliveryCode   string
	DateOfInvoice  time.Time
	ProjectName    string
	DistrictName   string
	ObjectType     string
	ObjectName     string
	SupervisorName string
	TeamLeaderName string
}

type InvoiceReturnConfirmDataQuery struct {
	Invoice                                     model.InvoiceReturn
	MaterialsInReturnerLocation                 []model.MaterialLocation
	MaterialsInAcceptorLocation                 []model.MaterialLocation
	NewMaterialsInAcceptorLocationWithNewDefect []model.MaterialLocation
	MaterialsDefected                           []model.MaterialDefect
	NewMaterialsDefected                        []model.MaterialDefect
}

type InvoiceReturnMaterialForEdit struct {
	MaterialID      uint     `json:"materialID"`
	MaterialName    string   `json:"materialName"`
	Unit            string   `json:"uint"`
	HolderAmount    float64  `json:"holderAmount"`
	Amount          float64  `json:"amount"`
	HasSerialNumber bool     `json:"hasSerialNumber"`
	SerialNumbers   []string `json:"serialNumbers"`
	IsDefective     bool     `json:"isDefective"`
	Notes           string   `json:"notes"`
}

type InvoiceReturnMaterialForSelect struct {
	MaterialID      uint    `json:"materialID"`
	MaterialName    string  `json:"materialName"`
	MaterialUnit    string  `json:"materialUnit"`
	Amount          float64 `json:"amount"`
	HasSerialNumber bool    `json:"hasSerialNumber"`
}
