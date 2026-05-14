package model

type User struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	WorkerID uint   `json:"workerID"`
	Username string `json:"username" gorm:"tinyText"`
	Password string `json:"password"`
	RoleID   uint   `json:"roleID"`

	UserActions    []UserAction    `json:"-" gorm:"foreignKey:UserID"`
	UserInProjects []UserInProject `json:"-" gorm:"foreignKey:UserID"`
  AuctionParticipantPrices []AuctionParticipantPrice `json:"-" gorm:"foreignKey:UserID"`
}
