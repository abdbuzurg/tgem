package dto

import "github.com/shopspring/decimal"

type ReportBalanceFilterRequest struct {
	Type     string `json:"type"`
	TeamID   uint   `json:"teamID"`
	ObjectID uint   `json:"objectID"`
}

type ReportBalanceFilter struct {
	LocationType string
	LocationID   uint
}

type BalanceReportQueryResult struct {
	LocationID      uint
	MaterialCode    string
	MaterialName    string
	MaterialUnit    string
	TotalAmount     float64
	DefectAmount    float64
	MaterialCostM19 decimal.Decimal
	TotalCost       decimal.Decimal
	TotalDefectCost decimal.Decimal
}

type MaterialLocationLiveSearchParameters struct {
	ProjectID    uint
	LocationType string
	LocationID   uint
	MaterialID   uint
}

type MaterialLocationLiveView struct {
	MaterialID      uint    `json:"materialID"`
	MaterialName    string  `json:"materialName"`
	MaterialUnit    string  `json:"materialUnit"`
	MaterialCostID  uint    `json:"materialCostID"`
	MaterialCostM19 string  `json:"materialCostM19"`
	LocationType    string  `json:"locationType"`
	LocationName    string  `json:"locationName"`
	LocationID      uint    `json:"locationID"`
	Amount          float64 `json:"amount"`
}

type ReportWriteOffBalanceFilter struct {
	WriteOffType       string `json:"writeOffType"`
	LocationID uint   `json:"locationID"`
}
