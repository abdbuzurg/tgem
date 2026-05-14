package model

type Role struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`

	Permissions []Permission `json:"-" gorm:"foreignKey:RoleID"`
	Users       []User       `json:"-" gorm:"foreignKey:RoleID"`
}
