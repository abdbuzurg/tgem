package dto

import "github.com/shopspring/decimal"

type MaterialCostView struct {
	ID               uint            `json:"id"`
	CostPrime        decimal.Decimal `json:"costPrime"`
	CostM19          decimal.Decimal `json:"costM19"`
	CostWithCustomer decimal.Decimal `json:"costWithCustomer"`
	MaterialName     string          `json:"materialName"`
}

type MaterialCostSearchFilter struct {
	ProjectID    uint   `json:"projectID"`
	MaterialName string `json:"materialName"`
}
