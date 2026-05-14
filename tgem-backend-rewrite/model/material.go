package model

type Material struct {
	ID                        uint    `json:"id" gorm:"primaryKey"`
	Category                  string  `json:"category" gorm:"tinyText"`
	Code                      string  `json:"code" gorm:"tinyText"`
	Name                      string  `json:"name"`
	Unit                      string  `json:"unit" gorm:"tinyText"`
	Notes                     string  `json:"notes"`
	HasSerialNumber           bool    `json:"hasSerialNumber"`
	Article                   string  `json:"article"`
	ProjectID                 uint    `json:"projectID"`
	PlannedAmountForProject   float64 `json:"plannedAmountForProject"`
	ShowPlannedAmountInReport bool    `json:"showPlannedAmountInReport"`

	MaterialCosts      []MaterialCost      `json:"-" gorm:"foreignKey:MaterialID"`
	OperationMaterials []OperationMaterial `json:"-" gorm:"foreignKey:MaterialID"`
}
