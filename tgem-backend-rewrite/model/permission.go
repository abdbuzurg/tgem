package model

type Permission struct {
	ID         uint `gorm:"primaryKey" json:"id"`
	RoleID     uint `json:"roleID"`
	ResourceID uint `json:"resourceID"`

	R          bool `json:"r"`
	W          bool `json:"w"`
	U          bool `json:"u"`
	D          bool `json:"d"`
}
