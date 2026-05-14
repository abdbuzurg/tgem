package model

type Auction struct {
	ID   uint   `json:"id" gorm:"primaryKey"`
	Name string `json:"name"`

	AuctionPackages []AuctionPackage `gorm:"foreignKey:AuctionID"`
}
