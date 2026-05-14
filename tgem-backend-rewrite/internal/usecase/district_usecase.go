package usecase

import (
	"context"

	"backend-v2/internal/db"
	"backend-v2/model"
)

type districtUsecase struct {
	q *db.Queries
}

func NewDistrictUsecase(q *db.Queries) IDistrictUsecase {
	return &districtUsecase{q: q}
}

type IDistrictUsecase interface {
	GetAll(projectID uint) ([]model.District, error)
	GetPaginated(page, limit int, projectID uint) ([]model.District, error)
	GetByID(id uint) (model.District, error)
	Create(data model.District) (model.District, error)
	Update(data model.District) (model.District, error)
	Delete(id uint) error
	Count(projectID uint) (int64, error)
}

func (u *districtUsecase) GetAll(projectID uint) ([]model.District, error) {
	rows, err := u.q.ListDistricts(context.Background(), pgInt8(projectID))
	if err != nil {
		return nil, err
	}
	out := make([]model.District, len(rows))
	for i, r := range rows {
		out[i] = toModelDistrict(r)
	}
	return out, nil
}

func (u *districtUsecase) GetPaginated(page, limit int, projectID uint) ([]model.District, error) {
	rows, err := u.q.ListDistrictsPaginated(context.Background(), db.ListDistrictsPaginatedParams{
		ProjectID: pgInt8(projectID),
		Limit:     int32(limit),
		Offset:    int32((page - 1) * limit),
	})
	if err != nil {
		return nil, err
	}
	out := make([]model.District, len(rows))
	for i, r := range rows {
		out[i] = toModelDistrict(r)
	}
	return out, nil
}

func (u *districtUsecase) GetByID(id uint) (model.District, error) {
	row, err := u.q.GetDistrict(context.Background(), int64(id))
	if err != nil {
		return model.District{}, err
	}
	return toModelDistrict(row), nil
}

func (u *districtUsecase) Create(data model.District) (model.District, error) {
	row, err := u.q.CreateDistrict(context.Background(), db.CreateDistrictParams{
		Name:      pgText(data.Name),
		ProjectID: pgInt8(data.ProjectID),
	})
	if err != nil {
		return model.District{}, err
	}
	return toModelDistrict(row), nil
}

func (u *districtUsecase) Update(data model.District) (model.District, error) {
	row, err := u.q.UpdateDistrict(context.Background(), db.UpdateDistrictParams{
		ID:        int64(data.ID),
		Name:      pgText(data.Name),
		ProjectID: pgInt8(data.ProjectID),
	})
	if err != nil {
		return model.District{}, err
	}
	return toModelDistrict(row), nil
}

func (u *districtUsecase) Delete(id uint) error {
	return u.q.DeleteDistrict(context.Background(), int64(id))
}

func (u *districtUsecase) Count(projectID uint) (int64, error) {
	return u.q.CountDistricts(context.Background(), pgInt8(projectID))
}

func toModelDistrict(d db.District) model.District {
	return model.District{
		ID:        uint(d.ID),
		Name:      d.Name.String,
		ProjectID: uintFromPgInt8(d.ProjectID),
	}
}
