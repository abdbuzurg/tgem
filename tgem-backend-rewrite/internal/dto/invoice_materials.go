package dto

import "github.com/shopspring/decimal"

type InvoiceMaterial struct {
	ID            uint            `json:"id" gorm:"primaryKey"`
	MaterialName  string          `json:"materialName"`
	MaterialPrice decimal.Decimal `json:"materialPrice"`
	InvoiceID     uint            `json:"invoiceID"`
	InvoiceType   string          `json:"invoiceType" gorm:"tinyText"`
	Amount        float64         `json:"amount"`
	Notes         string          `json:"notes"`
	Unit          string          `json:"unit"`
}

type InvoiceMaterialsWithoutSerialNumberView struct {
	ID           uint            `json:"id"`
	MaterialName string          `json:"materialName"`
	MaterialUnit string          `json:"materialUnit"`
	IsDefected   bool            `json:"isDefected"`
	CostM19      decimal.Decimal `json:"costM19"`
	Amount       float64         `json:"amount"`
	Notes        string          `json:"notes"`
}

type InvoiceMaterialsWithSerialNumberQuery struct {
	ID           uint            `json:"id"`
	MaterialName string          `json:"materialName"`
	MaterialUnit string          `json:"materialUnit"`
	IsDefected   bool            `json:"isDefected"`
	SerialNumber string          `json:"serialNumber"`
	CostM19      decimal.Decimal `json:"costM19"`
	Amount       float64         `json:"amount"`
	Notes        string          `json:"notes"`
}

type InvoiceMaterialsWithSerialNumberView struct {
	ID            uint            `json:"id"`
	MaterialName  string          `json:"materialName"`
	MaterialUnit  string          `json:"materialUnit"`
	SerialNumbers []string        `json:"serialNumbers"`
	IsDefected    bool            `json:"isDefected"`
	CostM19       decimal.Decimal `json:"costM19"`
	Amount        float64         `json:"amount"`
	Notes         string          `json:"notes"`
}

type InvoiceMaterialsDataForReport struct {
	InvoiceMaterialID        uint
	MaterialName             string
	MaterialUnit             string
	MaterialCategory         string
	MaterialCostPrime        decimal.Decimal
	MaterialCostM19          decimal.Decimal
	MaterialCostWithCustomer decimal.Decimal
	InvoiceMaterialAmount    float64
	InvoiceMaterialNotes     string
}
