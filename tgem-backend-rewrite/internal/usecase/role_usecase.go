package usecase

import (
	"context"

	"backend-v2/internal/db"
	"backend-v2/model"
)

type roleUsecase struct {
	q *db.Queries
}

func NewRoleUsecase(q *db.Queries) IRoleUsecase {
	return &roleUsecase{q: q}
}

type IRoleUsecase interface {
	GetAll() ([]model.Role, error)
	Create(data model.Role) (model.Role, error)
	Update(data model.Role) (model.Role, error)
	Delete(id uint) error
}

func (u *roleUsecase) GetAll() ([]model.Role, error) {
	rows, err := u.q.ListRoles(context.Background())
	if err != nil {
		return nil, err
	}
	out := make([]model.Role, len(rows))
	for i, r := range rows {
		out[i] = toModelRole(r)
	}
	return out, nil
}

func (u *roleUsecase) Create(data model.Role) (model.Role, error) {
	row, err := u.q.CreateRole(context.Background(), db.CreateRoleParams{
		Name:        pgText(data.Name),
		Description: pgText(data.Description),
	})
	if err != nil {
		return model.Role{}, err
	}
	return toModelRole(row), nil
}

func (u *roleUsecase) Update(data model.Role) (model.Role, error) {
	row, err := u.q.UpdateRole(context.Background(), db.UpdateRoleParams{
		ID:          int64(data.ID),
		Name:        pgText(data.Name),
		Description: pgText(data.Description),
	})
	if err != nil {
		return model.Role{}, err
	}
	return toModelRole(row), nil
}

func (u *roleUsecase) Delete(id uint) error {
	return u.q.DeleteRole(context.Background(), int64(id))
}

func toModelRole(r db.Role) model.Role {
	return model.Role{
		ID:          uint(r.ID),
		Name:        r.Name.String,
		Description: r.Description.String,
	}
}
