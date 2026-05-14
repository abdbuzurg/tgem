package usecase

import (
	"context"

	"backend-v2/internal/db"
	"backend-v2/model"
)

type resourceUsecase struct {
	q *db.Queries
}

func NewResourceUsecase(q *db.Queries) IResourceUsecase {
	return &resourceUsecase{q: q}
}

type IResourceUsecase interface {
	GetAll() ([]model.Resource, error)
}

func (u *resourceUsecase) GetAll() ([]model.Resource, error) {
	rows, err := u.q.ListResources(context.Background())
	if err != nil {
		return nil, err
	}
	out := make([]model.Resource, len(rows))
	for i, r := range rows {
		out[i] = toModelResource(r)
	}
	return out, nil
}

func toModelResource(r db.Resource) model.Resource {
	return model.Resource{
		ID:       uint(r.ID),
		Category: r.Category.String,
		Name:     r.Name.String,
		Url:      r.Url.String,
	}
}
