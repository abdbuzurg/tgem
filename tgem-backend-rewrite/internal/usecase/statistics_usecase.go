package usecase

import (
	"context"
	"fmt"

	"backend-v2/internal/db"
	"backend-v2/internal/dto"
)

type statisticsUsecase struct {
	q *db.Queries
}

type IStatisticsUsecase interface {
	InvoiceCountStat(projectID uint) ([]dto.PieChartData, error)
	InvoiceInputCreatorStat(projectID uint) ([]dto.PieChartData, error)
	InvoiceOutputCreatorStat(projectID uint) ([]dto.PieChartData, error)
	CountMaterialInInvoices(materialID uint) ([]dto.PieChartData, error)
	LocationMaterial(materialID uint) ([]dto.PieChartData, error)
}

func NewStatisticsUsecase(q *db.Queries) IStatisticsUsecase {
	return &statisticsUsecase{q: q}
}

func (u *statisticsUsecase) InvoiceCountStat(projectID uint) ([]dto.PieChartData, error) {
	ctx := context.Background()
	pid := pgInt8(projectID)

	in, err := u.q.CountInvoiceInputs(ctx, pid)
	if err != nil {
		return []dto.PieChartData{}, err
	}
	out, err := u.q.CountInvoiceOutputs(ctx, pid)
	if err != nil {
		return []dto.PieChartData{}, err
	}
	ret, err := u.q.CountInvoiceReturns(ctx, pid)
	if err != nil {
		return []dto.PieChartData{}, err
	}
	wo, err := u.q.CountInvoiceWriteOffs(ctx, pid)
	if err != nil {
		return []dto.PieChartData{}, err
	}

	return []dto.PieChartData{
		{ID: 0, Value: float64(in), Label: "Приход"},
		{ID: 1, Value: float64(out), Label: "Отпуск"},
		{ID: 2, Value: float64(ret), Label: "Возврат"},
		{ID: 3, Value: float64(wo), Label: "Списание"},
	}, nil
}

func (u *statisticsUsecase) InvoiceInputCreatorStat(projectID uint) ([]dto.PieChartData, error) {
	ctx := context.Background()
	pid := pgInt8(projectID)

	creators, err := u.q.ListInvoiceInputUniqueCreators(ctx, pid)
	if err != nil {
		return []dto.PieChartData{}, err
	}

	result := []dto.PieChartData{}
	fmt.Println(creators)
	for index, workerID := range creators {
		count, err := u.q.CountInvoiceInputCreatorInvoices(ctx, db.CountInvoiceInputCreatorInvoicesParams{
			ProjectID:        pid,
			ReleasedWorkerID: pgInt8(uint(workerID)),
		})
		if err != nil {
			return []dto.PieChartData{}, err
		}

		worker, err := u.q.GetWorker(ctx, workerID)
		if err != nil {
			return []dto.PieChartData{}, err
		}

		result = append(result, dto.PieChartData{
			ID:    uint(index),
			Value: float64(count),
			Label: worker.Name.String,
		})
	}

	return result, nil
}

func (u *statisticsUsecase) InvoiceOutputCreatorStat(projectID uint) ([]dto.PieChartData, error) {
	ctx := context.Background()
	pid := pgInt8(projectID)

	creators, err := u.q.ListInvoiceOutputUniqueCreators(ctx, pid)
	if err != nil {
		return []dto.PieChartData{}, err
	}

	result := []dto.PieChartData{}
	for index, workerID := range creators {
		count, err := u.q.CountInvoiceOutputCreatorInvoices(ctx, db.CountInvoiceOutputCreatorInvoicesParams{
			ProjectID:        pid,
			ReleasedWorkerID: pgInt8(uint(workerID)),
		})
		if err != nil {
			return []dto.PieChartData{}, err
		}

		worker, err := u.q.GetWorker(ctx, workerID)
		if err != nil {
			return []dto.PieChartData{}, err
		}

		result = append(result, dto.PieChartData{
			ID:    uint(index),
			Value: float64(count),
			Label: worker.Name.String,
		})
	}

	return result, nil
}

func (u *statisticsUsecase) CountMaterialInInvoices(materialID uint) ([]dto.PieChartData, error) {
	result := []dto.PieChartData{
		{ID: 0, Value: 0, Label: "Приход"},
		{ID: 1, Value: 0, Label: "Отпуск"},
		{ID: 2, Value: 0, Label: "Возврат"},
		{ID: 3, Value: 0, Label: "В процессе корректировки"},
		{ID: 4, Value: 0, Label: "Прошел корректировку"},
		{ID: 5, Value: 0, Label: "Списание"},
		{ID: 6, Value: 0, Label: "Отпуск вне проекта"},
	}

	rows, err := u.q.CountMaterialInInvoices(context.Background(), int64(materialID))
	if err != nil {
		return []dto.PieChartData{}, err
	}

	for _, r := range rows {
		switch r.InvoiceType {
		case "input":
			result[0].Value += r.Amount
		case "output":
			result[1].Value += r.Amount
		case "return":
			result[2].Value += r.Amount
		case "object":
			result[3].Value += r.Amount
		case "object-correction":
			result[4].Value += r.Amount
		case "writeoff":
			result[5].Value += r.Amount
		case "output-out-of-project":
			result[6].Value += r.Amount
		default:
			fmt.Println("Unknown InvoiceType")
		}
	}

	return result, nil
}

func (u *statisticsUsecase) LocationMaterial(materialID uint) ([]dto.PieChartData, error) {
	result := []dto.PieChartData{
		{ID: 0, Value: 0, Label: "Склад"},
		{ID: 1, Value: 0, Label: "Бригада"},
		{ID: 2, Value: 0, Label: "Объект"},
		{ID: 3, Value: 0, Label: "Списание Склада"},
		{ID: 4, Value: 0, Label: "Потеря Склада"},
		{ID: 5, Value: 0, Label: "Потеря Бригады"},
		{ID: 6, Value: 0, Label: "Списание Объекта"},
		{ID: 7, Value: 0, Label: "Потеря Объекта"},
		{ID: 8, Value: 0, Label: "Вышло из проекта"},
	}

	rows, err := u.q.CountMaterialInLocations(context.Background(), int64(materialID))
	if err != nil {
		return []dto.PieChartData{}, err
	}

	for _, r := range rows {
		switch r.LocationType {
		case "warehouse":
			result[0].Value += r.Amount
		case "team":
			result[1].Value += r.Amount
		case "object":
			result[2].Value += r.Amount
		case "writeoff-warehouse":
			result[3].Value += r.Amount
		case "loss-warehouse":
			result[4].Value += r.Amount
		case "loss-team":
			result[5].Value += r.Amount
		case "writeoff-object":
			result[6].Value += r.Amount
		case "loss-object":
			result[7].Value += r.Amount
		case "out-of-project":
			result[8].Value += r.Amount
		default:
			fmt.Println("Unknown Storage")
		}
	}

	return result, nil
}
