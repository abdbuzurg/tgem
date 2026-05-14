package usecase

import (
	"context"

	"backend-v2/internal/db"
	"backend-v2/internal/dto"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

type auctionUsecase struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

func NewAuctionUsecase(pool *pgxpool.Pool) IAuctionUsecase {
	return &auctionUsecase{
		pool: pool,
		q:    db.New(pool),
	}
}

type IAuctionUsecase interface {
	GetAuctionDataForPublic(auctionID uint) ([]dto.AuctionDataForPublic, error)
	GetAuctionDataForPrivate(auctionID, userID uint) ([]dto.AuctionDataForPrivate, error)
	SaveParticipantChanges(userID uint, participantData []dto.ParticipantDataForSave) error
}

func (u *auctionUsecase) GetAuctionDataForPublic(auctionID uint) ([]dto.AuctionDataForPublic, error) {
	rows, err := u.q.GetAuctionDataForPublic(context.Background(), int64(auctionID))
	if err != nil {
		// Preserves the GORM-era quirk: errors here are swallowed and an
		// empty slice is returned. Documented as a deviation in PROGRESS.md.
		return []dto.AuctionDataForPublic{}, nil
	}

	publicAuctionDataRaw := make([]dto.AuctionDataForPublicQueryResult, len(rows))
	for i, r := range rows {
		publicAuctionDataRaw[i] = dto.AuctionDataForPublicQueryResult{
			PackageID:        uint(r.PackageID),
			PackageName:      r.PackageName,
			ItemName:         r.ItemName,
			ItemDescription:  r.ItemDescription,
			ItemUnit:         r.ItemUnit,
			ItemQuantity:     r.ItemQuantity,
			ItemNote:         r.ItemNote,
			ParticipantPrice: decimalFromString(r.ParticipantPrice),
			ParticipantTitle: r.ParticipantTitle,
		}
	}

	result := []dto.AuctionDataForPublic{}
	for index, raw := range publicAuctionDataRaw {
		totalPriceDecimal := raw.ParticipantPrice.Mul(decimal.NewFromFloat(raw.ItemQuantity))
		totalPrice, exact := totalPriceDecimal.Float64()
		if !exact {
			return []dto.AuctionDataForPublic{}, nil
		}

		if index == 0 {
			entry := dto.AuctionDataForPublic{
				PackageName: raw.PackageName,
				PackageItems: []dto.AuctionItemDataForPublic{{
					Name:        raw.ItemName,
					Description: raw.ItemDescription,
					Unit:        raw.ItemUnit,
					Quantity:    raw.ItemQuantity,
					Note:        raw.ItemNote,
				}},
				ParticipantTotalPriceForPackage: []dto.AuctionParticipantTotalForPackageForPublic{},
			}
			if raw.ParticipantTitle != "" && totalPrice != 0 {
				entry.ParticipantTotalPriceForPackage = append(entry.ParticipantTotalPriceForPackage, dto.AuctionParticipantTotalForPackageForPublic{
					ParticipantTitle: raw.ParticipantTitle,
					TotalPrice:       totalPrice,
				})
			}

			result = append(result, entry)
		}

		lastPackageIndex := len(result) - 1
		if raw.PackageName == result[lastPackageIndex].PackageName {
			itemExists := false
			for _, item := range result[lastPackageIndex].PackageItems {
				if raw.ItemName == item.Name {
					itemExists = true
					break
				}
			}
			if !itemExists {
				result[lastPackageIndex].PackageItems = append(result[lastPackageIndex].PackageItems, dto.AuctionItemDataForPublic{
					Name:        raw.ItemName,
					Description: raw.ItemDescription,
					Unit:        raw.ItemUnit,
					Quantity:    raw.ItemQuantity,
					Note:        raw.ItemNote,
				})
			}

			participantIndex := -1
			for index, participant := range result[lastPackageIndex].ParticipantTotalPriceForPackage {
				if raw.ParticipantTitle == participant.ParticipantTitle {
					participantIndex = index
					break
				}
			}

			if participantIndex != -1 {
				result[lastPackageIndex].ParticipantTotalPriceForPackage[participantIndex].TotalPrice += totalPrice
			} else {
				result[lastPackageIndex].ParticipantTotalPriceForPackage = append(result[lastPackageIndex].ParticipantTotalPriceForPackage, dto.AuctionParticipantTotalForPackageForPublic{
					ParticipantTitle: raw.ParticipantTitle,
					TotalPrice:       totalPrice,
				})
			}
		} else {
			entry := dto.AuctionDataForPublic{
				PackageName: raw.PackageName,
				PackageItems: []dto.AuctionItemDataForPublic{{
					Name:        raw.ItemName,
					Description: raw.ItemDescription,
					Unit:        raw.ItemUnit,
					Quantity:    raw.ItemQuantity,
					Note:        raw.ItemNote,
				}},
				ParticipantTotalPriceForPackage: []dto.AuctionParticipantTotalForPackageForPublic{},
			}
			if raw.ParticipantTitle != "" && totalPrice != 0 {
				entry.ParticipantTotalPriceForPackage = append(entry.ParticipantTotalPriceForPackage, dto.AuctionParticipantTotalForPackageForPublic{
					ParticipantTitle: raw.ParticipantTitle,
					TotalPrice:       totalPrice,
				})
			}

			result = append(result, entry)
		}
	}

	return result, nil
}

func (u *auctionUsecase) GetAuctionDataForPrivate(auctionID, userID uint) ([]dto.AuctionDataForPrivate, error) {
	rows, err := u.q.GetAuctionDataForPrivate(context.Background(), int64(auctionID))
	if err != nil {
		return []dto.AuctionDataForPrivate{}, err
	}

	privateDataForAuction := make([]dto.AuctionDataForPrivateQueryResult, len(rows))
	for i, r := range rows {
		privateDataForAuction[i] = dto.AuctionDataForPrivateQueryResult{
			PackageID:          uint(r.PackageID),
			PackageName:        r.PackageName,
			ItemID:             uint(r.ItemID),
			ItemName:           r.ItemName,
			ItemDescription:    r.ItemDescription,
			ItemUnit:           r.ItemUnit,
			ItemQuantity:       r.ItemQuantity,
			ItemNote:           r.ItemNote,
			ParticipantComment: r.ParticipantComment,
			ParticipantUserID:  uint(r.ParticipantUserID),
			ParticipantPrice:   decimalFromString(r.ParticipantPrice),
			ParticipantTitle:   r.ParticipantTitle,
		}
	}

	result := []dto.AuctionDataForPrivate{}
	for index, raw := range privateDataForAuction {
		totalPriceDecimal := raw.ParticipantPrice.Mul(decimal.NewFromFloat(raw.ItemQuantity))
		totalPrice, exact := totalPriceDecimal.Float64()
		if !exact {
			return []dto.AuctionDataForPrivate{}, nil
		}

		if index == 0 {
			entry := dto.AuctionDataForPrivate{
				PackageName: raw.PackageName,
				PackageItems: []dto.AuctionItemDataForPrivate{{
					ID:          raw.ItemID,
					Name:        raw.ItemName,
					Description: raw.ItemDescription,
					Unit:        raw.ItemUnit,
					Quantity:    raw.ItemQuantity,
					Note:        raw.ItemNote,
				}},
				ParticipantTotalPriceForPackage: []dto.AuctionParticipantTotalForPackageForPrivate{},
			}
			if raw.ParticipantUserID == userID {
				pricePerUnitFloat, exact := raw.ParticipantPrice.Float64()
				if !exact {
					return []dto.AuctionDataForPrivate{}, nil
				}
				entry.PackageItems[0].UserUnitPrice = pricePerUnitFloat
				entry.PackageItems[0].Comment = raw.ParticipantComment
			}
			if raw.ParticipantTitle != "" && totalPrice != 0 {
				entry.ParticipantTotalPriceForPackage = append(entry.ParticipantTotalPriceForPackage, dto.AuctionParticipantTotalForPackageForPrivate{
					ParticipantTitle: raw.ParticipantTitle,
					TotalPrice:       totalPrice,
					IsCurrentUser:    false,
				})

				if raw.ParticipantUserID == userID {
					entry.ParticipantTotalPriceForPackage[0].IsCurrentUser = true
				}
			}

			result = append(result, entry)
			continue
		}

		lastPackageIndex := len(result) - 1
		if raw.PackageName == result[lastPackageIndex].PackageName {
			itemIndex := -1
			for subIndex, item := range result[lastPackageIndex].PackageItems {
				if raw.ItemName == item.Name {
					itemIndex = subIndex
					break
				}
			}
			if itemIndex == -1 {
				result[lastPackageIndex].PackageItems = append(result[lastPackageIndex].PackageItems, dto.AuctionItemDataForPrivate{
					ID:          raw.ItemID,
					Name:        raw.ItemName,
					Description: raw.ItemDescription,
					Unit:        raw.ItemUnit,
					Quantity:    raw.ItemQuantity,
					Note:        raw.ItemNote,
				})

				itemIndex = len(result[lastPackageIndex].PackageItems) - 1
			}

			if raw.ParticipantUserID == userID {
				pricePerUnitFloat, exact := raw.ParticipantPrice.Float64()
				if !exact {
					return []dto.AuctionDataForPrivate{}, nil
				}

				result[lastPackageIndex].PackageItems[itemIndex].UserUnitPrice = pricePerUnitFloat
				result[lastPackageIndex].PackageItems[itemIndex].Comment = raw.ParticipantComment
			}

			participantIndex := -1
			for subIndex, participant := range result[lastPackageIndex].ParticipantTotalPriceForPackage {
				if raw.ParticipantTitle == participant.ParticipantTitle {
					participantIndex = subIndex
					break
				}
			}

			if participantIndex != -1 {
				result[lastPackageIndex].ParticipantTotalPriceForPackage[participantIndex].TotalPrice += totalPrice
			} else {
				result[lastPackageIndex].ParticipantTotalPriceForPackage = append(result[lastPackageIndex].ParticipantTotalPriceForPackage, dto.AuctionParticipantTotalForPackageForPrivate{
					ParticipantTitle: raw.ParticipantTitle,
					TotalPrice:       totalPrice,
					IsCurrentUser:    raw.ParticipantUserID == userID,
				})
			}
		} else {
			entry := dto.AuctionDataForPrivate{
				PackageName: raw.PackageName,
				PackageItems: []dto.AuctionItemDataForPrivate{{
					ID:          raw.ItemID,
					Name:        raw.ItemName,
					Description: raw.ItemDescription,
					Unit:        raw.ItemUnit,
					Quantity:    raw.ItemQuantity,
					Note:        raw.ItemNote,
				}},
				ParticipantTotalPriceForPackage: []dto.AuctionParticipantTotalForPackageForPrivate{},
			}
			if raw.ParticipantUserID == userID {
				pricePerUnitFloat, exact := raw.ParticipantPrice.Float64()
				if !exact {
					return []dto.AuctionDataForPrivate{}, nil
				}
				entry.PackageItems[0].UserUnitPrice = pricePerUnitFloat
				entry.PackageItems[0].Comment = raw.ParticipantComment
			}

			if raw.ParticipantTitle != "" && totalPrice != 0 {
				entry.ParticipantTotalPriceForPackage = append(entry.ParticipantTotalPriceForPackage, dto.AuctionParticipantTotalForPackageForPrivate{
					ParticipantTitle: raw.ParticipantTitle,
					TotalPrice:       totalPrice,
					IsCurrentUser:    false,
				})

				if raw.ParticipantUserID == userID {
					entry.ParticipantTotalPriceForPackage[0].IsCurrentUser = true
				}
			}

			result = append(result, entry)
		}
	}

	for index, auctionPackage := range result {
		for _, totalPrice := range auctionPackage.ParticipantTotalPriceForPackage {
			if result[index].MinimumPackagePrice == 0 {
				result[index].MinimumPackagePrice = totalPrice.TotalPrice
				continue
			}

			if result[index].MinimumPackagePrice > totalPrice.TotalPrice {
				result[index].MinimumPackagePrice = totalPrice.TotalPrice
			}
		}
	}

	return result, nil
}

// SaveParticipantChanges atomically upserts each participant price
// using the unique index on (auction_item_id, user_id) added by
// migration 00002. ON CONFLICT makes this safe under concurrent writes
// — the GORM-era FirstOrCreate-in-a-loop race is gone.
func (u *auctionUsecase) SaveParticipantChanges(userID uint, participantData []dto.ParticipantDataForSave) error {
	ctx := context.Background()
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := u.q.WithTx(tx)
	for _, entry := range participantData {
		if err := qtx.UpsertAuctionParticipantPrice(ctx, db.UpsertAuctionParticipantPriceParams{
			AuctionItemID: pgInt8(entry.ItemID),
			UserID:        pgInt8(userID),
			UnitPrice:     pgText(entry.UnitPrice.String()),
			Comments:      pgText(entry.Comment),
		}); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

