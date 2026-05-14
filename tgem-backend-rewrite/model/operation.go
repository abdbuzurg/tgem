package model

import "github.com/shopspring/decimal"

type Operation struct {
	ID                        uint            `json:"id" gorm:"primaryKey"`
	ProjectID                 uint            `json:"projectID"`
	Name                      string          `json:"name"`
	Code                      string          `json:"code" gorm:"tinyText"`
	CostPrime                 decimal.Decimal `json:"costPrime" gorm:"type:decimal(20,4)"`
	CostWithCustomer          decimal.Decimal `json:"costWithCustomer" gorm:"type:decimal(20,4)"`
	PlannedAmountForProject   float64         `json:"plannedAmountForProject"`
	ShowPlannedAmountInReport bool            `json:"showPlannedAmountInReport"`

	InvoiceOperations         []InvoiceOperations         `json:"-" gorm:"foreignKey:OperationID"`
	OperationMaterials        []OperationMaterial         `json:"-" gorm:"foreignKey:OperationID"`
	ProjectProgressOperations []ProjectProgressOperations `json:"-" gorm:"foreignKey:OperationID"`
}
