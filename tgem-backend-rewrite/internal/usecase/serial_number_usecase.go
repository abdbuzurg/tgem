package usecase

import (
	"context"

	"backend-v2/internal/db"
	"backend-v2/model"
)

type serialNumberUsecase struct {
	q *db.Queries
}

func NewSerialNumberUsecase(q *db.Queries) ISerialNumberUsecase {
	return &serialNumberUsecase{q: q}
}

type ISerialNumberUsecase interface {
	GetAll() ([]model.SerialNumber, error)
	GetCodesByMaterialID(materialID uint) ([]string, error)
	Create(data model.SerialNumber) (model.SerialNumber, error)
	Update(data model.SerialNumber) (model.SerialNumber, error)
	Delete(id uint) error
}

func (u *serialNumberUsecase) GetAll() ([]model.SerialNumber, error) {
	rows, err := u.q.ListSerialNumbers(context.Background())
	if err != nil {
		return nil, err
	}
	out := make([]model.SerialNumber, len(rows))
	for i, r := range rows {
		out[i] = toModelSerialNumber(r)
	}
	return out, nil
}

func (u *serialNumberUsecase) GetCodesByMaterialID(materialID uint) ([]string, error) {
	ctx := context.Background()
	materialCosts, err := u.q.ListMaterialCostsByMaterialID(ctx, pgInt8(materialID))
	if err != nil {
		return nil, err
	}

	var codes []string
	for _, mc := range materialCosts {
		serialNumbers, err := u.q.ListSerialNumbersByMaterialCostID(ctx, pgInt8(uint(mc.ID)))
		if err != nil {
			return codes, err
		}
		for _, sn := range serialNumbers {
			codes = append(codes, sn.Code.String)
		}
	}
	return codes, nil
}

func (u *serialNumberUsecase) Create(data model.SerialNumber) (model.SerialNumber, error) {
	row, err := u.q.CreateSerialNumber(context.Background(), db.CreateSerialNumberParams{
		ProjectID:      pgInt8(data.ProjectID),
		MaterialCostID: pgInt8(data.MaterialCostID),
		Code:           pgText(data.Code),
	})
	if err != nil {
		return model.SerialNumber{}, err
	}
	return toModelSerialNumber(row), nil
}

func (u *serialNumberUsecase) Update(data model.SerialNumber) (model.SerialNumber, error) {
	row, err := u.q.UpdateSerialNumber(context.Background(), db.UpdateSerialNumberParams{
		ID:             int64(data.ID),
		ProjectID:      pgInt8(data.ProjectID),
		MaterialCostID: pgInt8(data.MaterialCostID),
		Code:           pgText(data.Code),
	})
	if err != nil {
		return model.SerialNumber{}, err
	}
	return toModelSerialNumber(row), nil
}

func (u *serialNumberUsecase) Delete(id uint) error {
	return u.q.DeleteSerialNumber(context.Background(), int64(id))
}

func toModelSerialNumber(s db.SerialNumber) model.SerialNumber {
	return model.SerialNumber{
		ID:             uint(s.ID),
		ProjectID:      uintFromPgInt8(s.ProjectID),
		MaterialCostID: uintFromPgInt8(s.MaterialCostID),
		Code:           s.Code.String,
	}
}
