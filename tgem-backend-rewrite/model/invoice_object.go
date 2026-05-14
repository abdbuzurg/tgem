package model

import "time"

type InvoiceObject struct {
	ID                  uint      `gorm:"primaryKey" json:"id"`
	DistrictID          uint      `json:"districtID"`
	DeliveryCode        string    `json:"deliveryCode"`
	ProjectID           uint      `json:"projectID"`
	SupervisorWorkerID  uint      `json:"supervisorWorkerID"`
	ObjectID            uint      `json:"objectID"`
	TeamID              uint      `json:"teamID"`
	DateOfInvoice       time.Time `json:"dateOfInvoice"`
	ConfirmedByOperator bool      `json:"confirmedByOperator"`
	DateOfCorrection    time.Time `json:"dateOfCorrection"`

	InvoiceObjectOperators []InvoiceObjectOperator `json:"-" gorm:"foreignKey:InvoiceObjectID"`
}
