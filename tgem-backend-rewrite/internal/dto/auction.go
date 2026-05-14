package dto

import "github.com/shopspring/decimal"

type AuctionDataForPublicQueryResult struct {
	PackageID        uint
	PackageName      string
	ItemName         string
	ItemDescription  string
	ItemUnit         string
	ItemQuantity     float64
	ItemNote         string
	ParticipantPrice decimal.Decimal
	ParticipantTitle string
}

type AuctionItemDataForPublic struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Unit        string  `json:"unit"`
	Quantity    float64 `json:"quantity"`
	Note        string  `json:"note"`
}

type AuctionParticipantTotalForPackageForPublic struct {
	ParticipantTitle string  `json:"participantTitle"`
	TotalPrice       float64 `json:"totalPrice"`
}

type AuctionDataForPublic struct {
	PackageName                     string                                       `json:"packageName"`
	PackageItems                    []AuctionItemDataForPublic                   `json:"packageItems"`
	ParticipantTotalPriceForPackage []AuctionParticipantTotalForPackageForPublic `json:"participantTotalPriceForPackage"`
}

type AuctionDataForPrivateQueryResult struct {
	PackageID          uint
	PackageName        string
	ItemID             uint
	ItemName           string
	ItemDescription    string
	ItemUnit           string
	ItemQuantity       float64
	ItemNote           string
	ParticipantComment string
	ParticipantUserID  uint
	ParticipantPrice   decimal.Decimal
	ParticipantTitle   string
}

type AuctionItemDataForPrivate struct {
	ID            uint    `json:"itemID"`
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	Unit          string  `json:"unit"`
	Quantity      float64 `json:"quantity"`
	Note          string  `json:"note"`
	UserUnitPrice float64 `json:"userUnitPrice"`
	Comment       string  `json:"comment"`
}

type AuctionParticipantTotalForPackageForPrivate struct {
	ParticipantTitle string  `json:"participantTitle"`
	TotalPrice       float64 `json:"participantPrice"`
	IsCurrentUser    bool    `json:"isCurrentUser"`
}

type AuctionDataForPrivate struct {
	PackageName                     string                                        `json:"packageName"`
	PackageItems                    []AuctionItemDataForPrivate                   `json:"packageItems"`
	ParticipantTotalPriceForPackage []AuctionParticipantTotalForPackageForPrivate `json:"-"`
	MinimumPackagePrice             float64                                       `json:"minimumPackagePrice"`
}

type ParticipantDataForSave struct {
	ItemID    uint            `json:"itemID"`
	Comment   string          `json:"comment"`
	UnitPrice decimal.Decimal `json:"unitPrice"`
}
