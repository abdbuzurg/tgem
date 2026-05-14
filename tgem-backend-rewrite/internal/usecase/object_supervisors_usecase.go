package usecase

import (
	"context"

	"backend-v2/internal/db"
	"backend-v2/model"
)

type objectSupervisorsUsecase struct {
	q *db.Queries
}

func NewObjectSupervisorsUsecase(q *db.Queries) IObjectSupervisorsUsecase {
	return &objectSupervisorsUsecase{q: q}
}

type IObjectSupervisorsUsecase interface {
	GetByObjectID(objectID uint) ([]model.ObjectSupervisors, error)
	GetBySupervisorWorkerID(workerID uint) ([]model.ObjectSupervisors, error)
	CreateBatch(data []model.ObjectSupervisors) ([]model.ObjectSupervisors, error)
}

func (u *objectSupervisorsUsecase) GetByObjectID(objectID uint) ([]model.ObjectSupervisors, error) {
	rows, err := u.q.ListObjectSupervisorsByObjectID(context.Background(), pgInt8(objectID))
	if err != nil {
		return nil, err
	}
	out := make([]model.ObjectSupervisors, len(rows))
	for i, r := range rows {
		out[i] = model.ObjectSupervisors{
			ID:                 uint(r.ID),
			SupervisorWorkerID: uintFromPgInt8(r.SupervisorWorkerID),
			ObjectID:           uintFromPgInt8(r.ObjectID),
		}
	}
	return out, nil
}

func (u *objectSupervisorsUsecase) GetBySupervisorWorkerID(workerID uint) ([]model.ObjectSupervisors, error) {
	rows, err := u.q.ListObjectSupervisorsByWorkerID(context.Background(), pgInt8(workerID))
	if err != nil {
		return nil, err
	}
	out := make([]model.ObjectSupervisors, len(rows))
	for i, r := range rows {
		out[i] = model.ObjectSupervisors{
			ID:                 uint(r.ID),
			SupervisorWorkerID: uintFromPgInt8(r.SupervisorWorkerID),
			ObjectID:           uintFromPgInt8(r.ObjectID),
		}
	}
	return out, nil
}

func (u *objectSupervisorsUsecase) CreateBatch(data []model.ObjectSupervisors) ([]model.ObjectSupervisors, error) {
	batch := make([]db.CreateObjectSupervisorsBatchParams, len(data))
	for i, d := range data {
		batch[i] = db.CreateObjectSupervisorsBatchParams{
			SupervisorWorkerID: pgInt8(d.SupervisorWorkerID),
			ObjectID:           pgInt8(d.ObjectID),
		}
	}
	_, err := u.q.CreateObjectSupervisorsBatch(context.Background(), batch)
	return data, err
}
