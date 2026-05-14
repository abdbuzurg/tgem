package model

type Team struct {
	ID           uint   `json:"id" gorm:"primaryKey"`
	ProjectID    uint   `json:"projectID"`
	Number       string `json:"number" gorm:"tinyText"`
	MobileNumber string `json:"mobileNumber" gorm:"tinyText"`
	Company      string `json:"company" gorm:"tinyText"`

	TeamLeaderss     []TeamLeaders     `json:"-" gorm:"foreignKey:TeamID"`
	InvoiceOutputs   []InvoiceOutput   `json:"-" gorm:"foreignKey:TeamID"`
	InvoiceObject    []InvoiceObject   `json:"-" gorm:"foreignKey:TeamID"`
	ObjectTeams      []ObjectTeams     `json:"-" gorm:"foreignKey:TeamID"`
}
