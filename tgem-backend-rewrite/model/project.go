package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type Project struct {
	ID                   uint            `json:"id" gorm:"primaryKey"`
	Name                 string          `json:"name" gorm:"tinyText"`
	Client               string          `json:"client" gorm:"tinyText"`
	Budget               decimal.Decimal `json:"budget" gorm:"type:decimal(20,2)"`
	BudgetCurrency       string          `json:"budgetCurrency"`
	Description          string          `json:"description"`
	SignedDateOfContract time.Time       `json:"signedDateOfContract"`
	DateStart            time.Time       `json:"dateStart"`
	DateEnd              time.Time       `json:"dateEnd"`
	ProjectManager       string          `json:"projectManager"`

	Districts                  []District                  `json:"-" gorm:"foreignKey:ProjectID"`
	UserActions                []UserAction                `json:"-" gorm:"foreignKey:ProjectID"`
	UserInProjects             []UserInProject             `json:"-" gorm:"foreignKey:ProjectID"`
	Materials                  []Material                  `json:"-" gorm:"foreignKey:ProjectID"`
	MaterialLocations          []MaterialLocation          `json:"-" gorm:"foreignKey:ProjectID"`
	SerialNumbers              []SerialNumber              `json:"-" gorm:"foreignKey:ProjectID"`
	SerialNumberMovements      []SerialNumberMovement      `json:"-" gorm:"foreignKey:ProjectID"`
	SerialNumberLocations      []SerialNumberLocation      `json:"-" gorm:"foreignKey:ProjectID"`
	Objects                    []Object                    `json:"-" gorm:"foreignKey:ProjectID"`
	Teams                      []Team                      `json:"-" gorm:"foreignKey:ProjectID"`
	InvoiceInputs              []InvoiceInput              `json:"-" gorm:"foreignKey:ProjectID"`
	InvoiceOutputs             []InvoiceOutput             `json:"-" gorm:"foreignKey:ProjectID"`
	InvoiceReturns             []InvoiceReturn             `json:"-" gorm:"foreignKey:ProjectID"`
	InvoiceObject              []InvoiceObject             `json:"-" gorm:"foreignKey:ProjectID"`
	InvoiceWriteOffs           []InvoiceWriteOff           `json:"-" gorm:"foreignKey:ProjectID"`
	InvoiceOperations          []InvoiceOperations         `json:"-" gorm:"foreignKey:ProjectID"`
	InvoiceMaterials           []InvoiceMaterials          `json:"-" gorm:"foreignKey:ProjectID"`
	InvoiceOutputOutOfProjects []InvoiceOutputOutOfProject `json:"-" gorm:"foreignKey:ProjectID"`
	InvoiceCounts              []InvoiceCount              `json:"-" gorm:"foreignKey:ProjectID"`
	Operations                 []Operation                 `json:"-" gorm:"foreignKey:ProjectID"`
	WorkerAttendances          []WorkerAttendance          `json:"-" gorm:"foreignKey:ProjectID"`
	ProjectProgressMaterials   []ProjectProgressMaterials  `json:"-" gorm:"foreignKey:ProjectID"`
	ProjectProgressOperations  []ProjectProgressOperations `json:"-" gorm:"foreignKey:ProjectID"`
}
