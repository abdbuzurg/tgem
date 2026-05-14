package model

type District struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	Name      string `json:"name"`
	ProjectID uint   `json:"projectID"`

	InvoiceOutput []InvoiceOutput `json:"-" gorm:"foreignKey:DistrictID"`
	InvoiceObject []InvoiceObject `json:"-" gorm:"foreignKey:DistrictID"`
}
