package model

type SerialNumber struct {
	ID             uint   `json:"id" gorm:"primaryKey"`
	ProjectID      uint   `json:"projectID"`
	MaterialCostID uint   `json:"materialCostID"`
	Code           string `json:"code" gorm:"text"`

  SerialNumberLocations []SerialNumberLocation `json:"-" gorm:"foreignKey:SerialNumberID"`
  SerialNumberMovements []SerialNumberMovement `json:"-" gorm:"foreignKey:SerialNumberID"`
}
