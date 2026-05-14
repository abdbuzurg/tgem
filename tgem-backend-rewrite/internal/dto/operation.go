package dto

import "github.com/shopspring/decimal"

type Operation struct {
	ID                        uint            `json:"id" gorm:"primaryKey"`
	ProjectID                 uint            `json:"projectID"`
	MaterialID                uint            `json:"materialID"`
	Name                      string          `json:"name"`
	Code                      string          `json:"code" gorm:"tinyText"`
	CostPrime                 decimal.Decimal `json:"costPrime" gorm:"type:decimal(20,4)"`
	CostWithCustomer          decimal.Decimal `json:"costWithCustomer" gorm:"type:decimal(20,4)"`
	PlannedAmountForProject   float64         `json:"plannedAmountForProject"`
	ShowPlannedAmountInReport bool            `json:"showPlannedAmountInReport"`
}

type OperationSearchParameters struct {
	ProjectID  uint
	Code       string
	Name       string
	MaterialID uint
}

type OperationPaginated struct {
	ID                        uint            `json:"id"`
	Name                      string          `json:"name"`
	Code                      string          `json:"code"`
	CostPrime                 decimal.Decimal `json:"costPrime"`
	CostWithCustomer          decimal.Decimal `json:"costWithCustomer"`
	MaterialName              string          `json:"materialName"`
	MaterialID                uint            `json:"materialID"`
	ShowPlannedAmountInReport bool            `json:"showPlannedAmountInReport"`
	PlannedAmountForProject   float64         `json:"plannedAmountForProject"`
}

type OperationImportDataForInsert struct {
	ID                        uint `json:"id"`
	ProjectID                 uint
	Name                      string          `json:"name"`
	Code                      string          `json:"code"`
	CostPrime                 decimal.Decimal `json:"costPrime"`
	CostWithCustomer          decimal.Decimal `json:"costWithCustomer"`
	PlannedAmountForProject   float64         `json:"plannedAmountForProject"`
	ShowPlannedAmountInReport bool            `json:"showPlannedAmountInReport"`
	MaterialID                uint
}
