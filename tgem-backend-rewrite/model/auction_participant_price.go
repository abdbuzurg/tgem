package model

import "github.com/shopspring/decimal"

type AuctionParticipantPrice struct {
	ID            uint `json:"id" gorm:"primaryKey"`
	AuctionItemID uint
	UserID        uint
	UnitPrice     decimal.Decimal `json:"unitPrice"`
	Comments      string          `json:"comments"`
}
