package model

type Resource struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	Category string `json:"category"`
	Name     string `json:"name"`
	Url      string `json:"url"`

	Permissions []Permission `json:"-" gorm:"foreignKey:ResourceID"`
}
