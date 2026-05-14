package model

type AuctionItem struct {
	ID               uint    `json:"id" gorm:"primaryKey"`
	AuctionPackageID uint    `json:"auctionPackageID"`
	Name             string  `json:"name"`
	Description      string  `json:"description"`
	Unit             string  `json:"unit"`
	Quantity         float64 `json:"quantity"`
	Note             string  `json:"note"`
}
