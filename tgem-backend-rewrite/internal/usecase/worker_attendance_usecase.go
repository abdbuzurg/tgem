package usecase

import (
	"context"
	"fmt"
	"os"
	"time"

	"backend-v2/internal/db"
	"backend-v2/internal/dto"
	"backend-v2/model"

	"github.com/xuri/excelize/v2"
)

type workerAttendanceUsecase struct {
	q *db.Queries
}

type IWorkerAttendanceUsecase interface {
	Import(projectID uint, filePath string) error
	GetPaginated(projectID uint) ([]dto.WorkerAttendancePaginated, error)
	Count(projectID uint) (int64, error)
}

func NewWorkerAttendanceUsecase(q *db.Queries) IWorkerAttendanceUsecase {
	return &workerAttendanceUsecase{q: q}
}

func (u *workerAttendanceUsecase) Import(projectID uint, filepath string) error {
	f, err := excelize.OpenFile(filepath)
	if err != nil {
		f.Close()
		os.Remove(filepath)
		return fmt.Errorf("Не смог открыть файл: %v", err)
	}

	sheetName := "морфо"
	rows, err := f.GetRows(sheetName)
	if err != nil {
		f.Close()
		os.Remove(filepath)
		return fmt.Errorf("Не смог найти таблицу 'Импорт': %v", err)
	}

	if len(rows) == 1 {
		f.Close()
		os.Remove(filepath)
		return fmt.Errorf("Файл не имеет данных")
	}

	type workerAttendanceRaw struct {
		CompanyWorkerID string
		Date            time.Time
	}

	excelData := []workerAttendanceRaw{}
	index := 1
	dateTimeLayout := "01-02-06 3:04:05 PM"
	for len(rows) > index {
		dateExcel, err := f.GetCellValue(sheetName, "A"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке А%d: %v", index+1, err)
		}

		timeExcel, err := f.GetCellValue(sheetName, "B"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке B%d: %v", index+1, err)
		}

		companyWorkerID, err := f.GetCellValue(sheetName, "E"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке B%d: %v", index+1, err)
		}

		dateTime, err := time.Parse(dateTimeLayout, dateExcel+" "+timeExcel)
		if err != nil {
			f.Close()
			os.Remove(filepath)
			return err
		}

		excelData = append(excelData, workerAttendanceRaw{
			CompanyWorkerID: companyWorkerID,
			Date:            dateTime,
		})

		index++
	}
	f.Close()
	os.Remove(filepath)

	ctx := context.Background()
	workerAttendance := []model.WorkerAttendance{}
	for _, entry := range excelData {
		worker, err := u.q.GetWorkerByCompanyID(ctx, pgText(entry.CompanyWorkerID))
		if err != nil {
			return err
		}

		isNewWorkerAttendance := true
		for index, attendance := range workerAttendance {
			if attendance.WorkerID == uint(worker.ID) {
				if attendance.Start.Day() == entry.Date.Day() {
					workerAttendance[index].End = entry.Date
					isNewWorkerAttendance = false
					break
				}
			}
		}

		if isNewWorkerAttendance {
			workerAttendance = append(workerAttendance, model.WorkerAttendance{
				WorkerID:  uint(worker.ID),
				ProjectID: projectID,
				Start:     entry.Date,
			})
		}
	}

	fmt.Println(workerAttendance)

	batch := make([]db.CreateWorkerAttendancesBatchParams, len(workerAttendance))
	for i, wa := range workerAttendance {
		batch[i] = db.CreateWorkerAttendancesBatchParams{
			ProjectID: pgInt8(wa.ProjectID),
			WorkerID:  pgInt8(wa.WorkerID),
			Start:     pgTimestamptz(wa.Start),
			End:       pgTimestamptz(wa.End),
		}
	}
	if _, err := u.q.CreateWorkerAttendancesBatch(ctx, batch); err != nil {
		return err
	}
	return nil
}

func (u *workerAttendanceUsecase) GetPaginated(projectID uint) ([]dto.WorkerAttendancePaginated, error) {
	rows, err := u.q.ListWorkerAttendancesPaginated(context.Background(), pgInt8(projectID))
	if err != nil {
		return []dto.WorkerAttendancePaginated{}, err
	}

	location, _ := time.LoadLocation("UTC")
	result := make([]dto.WorkerAttendancePaginated, len(rows))
	for i, r := range rows {
		result[i] = dto.WorkerAttendancePaginated{
			ID:              uint(r.ID),
			WorkerName:      r.WorkerName,
			CompanyWorkerID: r.CompanyWorkerID,
			Start:           timeFromPgTimestamptz(r.Start).In(location),
			End:             timeFromPgTimestamptz(r.End).In(location),
		}
	}

	return result, nil
}

func (u *workerAttendanceUsecase) Count(projectID uint) (int64, error) {
	return u.q.CountWorkerAttendances(context.Background(), pgInt8(projectID))
}
