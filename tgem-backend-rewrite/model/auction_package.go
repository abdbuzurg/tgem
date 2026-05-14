package model

type AuctionPackage struct {
	ID        uint `json:"id" gorm:"primaryKey"`
	AuctionID uint
	Name      string `json:"name"`

	AuctionItems []AuctionItem `gorm:"foreignKey:AuctionPackageID"`
}
