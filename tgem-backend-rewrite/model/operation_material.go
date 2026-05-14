package model

type OperationMaterial struct {
	ID          uint `json:"id" gorm:"primaryKey"`
	OperationID uint `json:"operationID"`
	MaterialID  uint `json:"materialID"`
}
