package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

type ProgressReportData struct {
	MaterialID                            uint
	MaterialCode                          string
	MaterialName                          string
	MaterialUnit                          string
	MaterialAmountPlannedForProject       float64
	MaterialAmountRecieved                float64
	MaterialAmountWaitingToBeRecieved     float64
	MaterialAmountInstalled               float64
	MaterialAmountWaitingToBeInstalled    float64
	MaterialAmountInWarehouse             float64
	MaterialAmountInTeams                 float64
	MaterialAmountInObjects               float64
	MaterialAmountInAllWriteOffs          float64
	BudgetOfRecievedMaterials             decimal.Decimal
	BudgetOfInstalledMaterials            decimal.Decimal
	BudgetOfMaterialsWaitingToBeInstalled decimal.Decimal
}

type MaterialDataForProgressReportQueryResult struct {
	ID                      uint
	Code                    string
	Name                    string
	Unit                    string
	PlannedAmountForProject float64
	LocationAmount          float64
	LocationType            string
}

type InvoiceMaterialDataForProgressReportQueryResult struct {
	MaterialID       uint
	Amount           float64
	InvoiceType      string
	CostWithCustomer decimal.Decimal
}

type InvoiceOperationDataForProgressReportQueryResult struct {
	ID                      uint
	Code                    string
	Name                    string
	CostWithCustomer        decimal.Decimal
	PlannedAmountForProject float64
	AmountInInvoice         float64
}

type MaterialDataForProgressReportInGivenDateQueryResult struct {
	ID                      uint
	Code                    string
	Name                    string
	Unit                    string
	AmountPlannedForProject float64
	AmountReceived          float64
	AmountInstalled         float64
	AmountInWarehouse       float64
	AmountInTeams           float64
	AmountInObjects         float64
	AmountWriteOff          float64
	CostWithCustomer        decimal.Decimal
}

type InvoiceOperationDataForProgressReportInGivenDataQueryResult struct {
	ID                      uint
	Code                    string
	Name                    string
	CostWithCustomer        decimal.Decimal
	AmountPlannedForProject float64
	AmountInstalled         float64
}

type MaterialDataForRemainingMaterialAnalysisQueryResult struct {
	ID                      uint
	Code                    string
	Name                    string
	Unit                    string
	PlannedAmountForProject float64
	LocationAmount          float64
	LocationType            string
}

type MaterialsInstalledOnObjectForRemainingMaterialAnalysisQueryResult struct {
	ID               uint
	Amount           float64
	DateOfCorrection time.Time
}
