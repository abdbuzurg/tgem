package dto

type ObjectPaginated struct {
	ID          uint     `json:"id"`
	Name        string   `json:"name"`
	Type        string   `json:"type" gorm:"tinyText"`
	Status      string   `json:"status"`
	Supervisors []string `json:"supervisors"`
}

type ObjectPaginatedQuery struct {
	ID             uint   `json:"id"`
	ObjectName     string `json:"name"`
	ObjectType     string `json:"type"`
	ObjectStatus   string `json:"status"`
	SupervisorName string `json:"supervisors"`
}

type ObjectCreate struct {
	Type      string `json:"type" gorm:"tinyText"`
	Name      string `json:"name" gorm:"tinyText"`
	Status    string `json:"status" gorm:"tinyText"`
	ProjectID uint   `json:"projectID"`

	Model          string  `json:"model" gorm:"tinyText"`
	AmountStores   uint    `json:"amountStores"`
	AmountEntraces uint    `json:"amountEntraces"`
	HasBasement    bool    `json:"hasBasement"`
	VoltageClass   string  `json:"voltageClass" gorm:"tinyText"`
	Nourashes      string  `json:"nourashes" gorm:"tinyText"`
	TTCoefficient  string  `json:"ttCoefficient" gorm:"tinyText"`
	AmountFeeders  uint    `json:"amountFeeders"`
	Length         float64 `json:"length"`

	Supervisors []uint `json:"supervisors"`
}

type TeamDataForSelect struct {
	ID             uint   `json:"id"`
	TeamNumber     string `json:"teamNumber"`
	TeamLeaderName string `json:"teamLeaderName"`
}

type ObjectDataForSelect struct {
  ID uint `json:"id"`
  ObjectName string `json:"objectName"`
  ObjectType string `json:"objectType"`
}
