package usecase

import (
	"context"
	"errors"

	"backend-v2/internal/apperr"
	"backend-v2/internal/db"
	"backend-v2/internal/dto"
	"backend-v2/model"

	"github.com/jackc/pgx/v5"
)

type objectUsecase struct {
	q *db.Queries
}

func NewObjectUsecase(q *db.Queries) IObjectUsecase {
	return &objectUsecase{q: q}
}

type IObjectUsecase interface {
	GetAll(projectID uint) ([]model.Object, error)
	GetPaginated(page, limit int, data model.Object) ([]dto.ObjectPaginated, error)
	GetByID(id uint) (model.Object, error)
	Create(data dto.ObjectCreate) (model.Object, error)
	Update(data model.Object) (model.Object, error)
	Delete(id uint) error
	Count() (int64, error)
	GetTeamsByObjectID(objectID uint) ([]dto.TeamDataForSelect, error)
}

func (u *objectUsecase) GetAll(projectID uint) ([]model.Object, error) {
	rows, err := u.q.ListObjectsByProject(context.Background(), pgInt8(projectID))
	if err != nil {
		return nil, err
	}
	out := make([]model.Object, len(rows))
	for i, r := range rows {
		out[i] = toModelObject(r)
	}
	return out, nil
}

func (u *objectUsecase) GetPaginated(page, limit int, filter model.Object) ([]dto.ObjectPaginated, error) {
	rows, err := u.q.ListObjectsPaginatedFiltered(context.Background(), db.ListObjectsPaginatedFilteredParams{
		Column1: int64(filter.ObjectDetailedID),
		Column2: filter.Type,
		Column3: filter.Name,
		Column4: filter.Status,
		Limit:   int32(limit),
		Offset:  int32((page - 1) * limit),
	})
	if err != nil {
		return []dto.ObjectPaginated{}, err
	}

	result := []dto.ObjectPaginated{}
	latestEntry := dto.ObjectPaginated{}
	for index, object := range rows {
		if index == 0 {
			latestEntry = dto.ObjectPaginated{
				ID:          uint(object.ID),
				Type:        object.ObjectType,
				Name:        object.ObjectName,
				Status:      object.ObjectStatus,
				Supervisors: []string{},
			}
		}

		if latestEntry.ID == uint(object.ID) {
			latestEntry.Supervisors = append(latestEntry.Supervisors, object.SupervisorName)
		} else {
			result = append(result, latestEntry)
			latestEntry = dto.ObjectPaginated{
				ID:     uint(object.ID),
				Type:   object.ObjectType,
				Name:   object.ObjectName,
				Status: object.ObjectStatus,
				Supervisors: []string{
					object.SupervisorName,
				},
			}
		}
	}

	if len(rows) != 0 {
		result = append(result, latestEntry)
	}

	return result, nil
}

func (u *objectUsecase) GetByID(id uint) (model.Object, error) {
	row, err := u.q.GetObject(context.Background(), int64(id))
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Object{}, apperr.NotFound("Запись не найдена", nil)
	}
	if err != nil {
		return model.Object{}, apperr.FromDB(err)
	}
	return toModelObject(row), nil
}

func (u *objectUsecase) Create(data dto.ObjectCreate) (model.Object, error) {
	row, err := u.q.CreateObject(context.Background(), db.CreateObjectParams{
		ObjectDetailedID: pgInt8(0),
		Type:             pgText(""),
		Name:             pgText(data.Name),
		Status:           pgText(data.Status),
		ProjectID:        pgInt8(data.ProjectID),
	})
	if err != nil {
		return model.Object{}, err
	}
	return toModelObject(row), nil
}

func (u *objectUsecase) Update(data model.Object) (model.Object, error) {
	row, err := u.q.UpdateObject(context.Background(), db.UpdateObjectParams{
		ID:               int64(data.ID),
		ObjectDetailedID: pgInt8(data.ObjectDetailedID),
		Type:             pgText(data.Type),
		Name:             pgText(data.Name),
		Status:           pgText(data.Status),
		ProjectID:        pgInt8(data.ProjectID),
	})
	if err != nil {
		return model.Object{}, err
	}
	return toModelObject(row), nil
}

func (u *objectUsecase) Delete(id uint) error {
	return u.q.DeleteObject(context.Background(), int64(id))
}

func (u *objectUsecase) Count() (int64, error) {
	return u.q.CountObjects(context.Background())
}

func (u *objectUsecase) GetTeamsByObjectID(objectID uint) ([]dto.TeamDataForSelect, error) {
	rows, err := u.q.ListTeamsForSelectByObjectID(context.Background(), pgInt8(objectID))
	if err != nil {
		return nil, err
	}
	out := make([]dto.TeamDataForSelect, len(rows))
	for i, r := range rows {
		out[i] = dto.TeamDataForSelect{
			ID:             uint(r.ID),
			TeamNumber:     r.TeamNumber,
			TeamLeaderName: r.TeamLeaderName,
		}
	}
	return out, nil
}

func toModelObject(o db.Object) model.Object {
	return model.Object{
		ID:               uint(o.ID),
		ObjectDetailedID: uintFromPgInt8(o.ObjectDetailedID),
		Type:             o.Type.String,
		Name:             o.Name.String,
		Status:           o.Status.String,
		ProjectID:        uintFromPgInt8(o.ProjectID),
	}
}
