package model

import "github.com/shopspring/decimal"

type MaterialCost struct {
	ID               uint            `json:"id" gorm:"primaryKey"`
	MaterialID       uint            `json:"materialID"`
	CostPrime        decimal.Decimal `json:"costPrime" gorm:"type:decimal(20,4)"`
	CostM19          decimal.Decimal `json:"costM19" gorm:"type:decimal(20,4)"`
	CostWithCustomer decimal.Decimal `json:"costWithCustomer" gorm:"type:decimal(20,4)"`

	InvoiceMaterials           []InvoiceMaterials         `json:"-" gorm:"foreignKey:MaterialCostID"`
	MaterialLocations          []MaterialLocation         `json:"-" gorm:"foreignKey:MaterialCostID"`
	ProjectProgressesMaterials []ProjectProgressMaterials `json:"-" gorm:"foreignKey:MaterialCostID"`
}
