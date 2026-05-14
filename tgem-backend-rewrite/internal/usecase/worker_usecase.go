package usecase

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"backend-v2/internal/db"
	"backend-v2/internal/dto"
	"backend-v2/model"

	"github.com/xuri/excelize/v2"
)

type workerUsecase struct {
	q *db.Queries
}

func NewWorkerUsecase(q *db.Queries) IWorkerUsecase {
	return &workerUsecase{q: q}
}

type IWorkerUsecase interface {
	GetAll(projectID uint) ([]model.Worker, error)
	GetPaginated(page, limit int, filter dto.WorkerSearchParameters) ([]model.Worker, error)
	GetByID(id uint) (model.Worker, error)
	GetByJobTitleInProject(jobTitleInProject string, projectID uint) ([]model.Worker, error)
	Create(data model.Worker) (model.Worker, error)
	Update(data model.Worker) (model.Worker, error)
	Delete(id uint) error
	Count(filter dto.WorkerSearchParameters) (int64, error)
	Import(filepath string, projectID uint) error
	GetWorkerInformationForSearch(projectID uint) (dto.WorkerInformationForSearch, error)
	Export(projectID uint) (string, error)
}

func (u *workerUsecase) GetAll(projectID uint) ([]model.Worker, error) {
	rows, err := u.q.ListWorkersByProject(context.Background(), pgInt8(projectID))
	if err != nil {
		return nil, err
	}
	out := make([]model.Worker, len(rows))
	for i, r := range rows {
		out[i] = toModelWorker(r)
	}
	return out, nil
}

func (u *workerUsecase) GetPaginated(page, limit int, filter dto.WorkerSearchParameters) ([]model.Worker, error) {
	rows, err := u.q.ListWorkersPaginatedFiltered(context.Background(), db.ListWorkersPaginatedFilteredParams{
		ProjectID: pgInt8(filter.ProjectID),
		Column2:   filter.Name,
		Column3:   filter.MobileNumber,
		Column4:   filter.JobTitleInCompany,
		Column5:   filter.JobTitleInProject,
		Column6:   filter.CompanyWorkerID,
		Limit:     int32(limit),
		Offset:    int32((page - 1) * limit),
	})
	if err != nil {
		return nil, err
	}
	out := make([]model.Worker, len(rows))
	for i, r := range rows {
		out[i] = toModelWorker(r)
	}
	return out, nil
}

func (u *workerUsecase) GetByID(id uint) (model.Worker, error) {
	row, err := u.q.GetWorker(context.Background(), int64(id))
	if err != nil {
		return model.Worker{}, err
	}
	return toModelWorker(row), nil
}

func (u *workerUsecase) GetByJobTitleInProject(jobTitleInProject string, projectID uint) ([]model.Worker, error) {
	rows, err := u.q.ListWorkersByJobTitleInProject(context.Background(), db.ListWorkersByJobTitleInProjectParams{
		JobTitleInProject: pgText(jobTitleInProject),
		ProjectID:         pgInt8(projectID),
	})
	if err != nil {
		return nil, err
	}
	out := make([]model.Worker, len(rows))
	for i, r := range rows {
		out[i] = toModelWorker(r)
	}
	return out, nil
}

func (u *workerUsecase) Create(data model.Worker) (model.Worker, error) {
	row, err := u.q.CreateWorker(context.Background(), db.CreateWorkerParams{
		ProjectID:         pgInt8(data.ProjectID),
		Name:              pgText(data.Name),
		CompanyWorkerID:   pgText(data.CompanyWorkerID),
		JobTitleInCompany: pgText(data.JobTitleInCompany),
		JobTitleInProject: pgText(data.JobTitleInProject),
		MobileNumber:      pgText(data.MobileNumber),
	})
	if err != nil {
		return model.Worker{}, err
	}
	return toModelWorker(row), nil
}

func (u *workerUsecase) Update(data model.Worker) (model.Worker, error) {
	row, err := u.q.UpdateWorker(context.Background(), db.UpdateWorkerParams{
		ID:                int64(data.ID),
		ProjectID:         pgInt8(data.ProjectID),
		Name:              pgText(data.Name),
		CompanyWorkerID:   pgText(data.CompanyWorkerID),
		JobTitleInCompany: pgText(data.JobTitleInCompany),
		JobTitleInProject: pgText(data.JobTitleInProject),
		MobileNumber:      pgText(data.MobileNumber),
	})
	if err != nil {
		return model.Worker{}, err
	}
	return toModelWorker(row), nil
}

func (u *workerUsecase) Delete(id uint) error {
	return u.q.DeleteWorker(context.Background(), int64(id))
}

func (u *workerUsecase) Count(filter dto.WorkerSearchParameters) (int64, error) {
	return u.q.CountWorkersFiltered(context.Background(), db.CountWorkersFilteredParams{
		ProjectID: pgInt8(filter.ProjectID),
		Column2:   filter.Name,
		Column3:   filter.MobileNumber,
		Column4:   filter.JobTitleInCompany,
		Column5:   filter.JobTitleInProject,
	})
}

func (u *workerUsecase) Import(filepathStr string, projectID uint) error {
	f, err := excelize.OpenFile(filepathStr)
	if err != nil {
		f.Close()
		os.Remove(filepathStr)
		return fmt.Errorf("Не смог открыть файл: %v", err)
	}

	sheetName := "Импорт"
	rows, err := f.GetRows(sheetName)
	if err != nil {
		f.Close()
		os.Remove(filepathStr)
		return fmt.Errorf("Не смог найти таблицу 'Импорт': %v", err)
	}

	if len(rows) == 1 {
		f.Close()
		os.Remove(filepathStr)
		return fmt.Errorf("Файл не имеет данных")
	}

	workers := []model.Worker{}
	index := 1
	for len(rows) > index {
		worker := model.Worker{
			ProjectID: projectID,
		}

		worker.Name, err = f.GetCellValue(sheetName, "A"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepathStr)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке А%d: %v", index+1, err)
		}
		worker.JobTitleInProject, err = f.GetCellValue(sheetName, "B"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepathStr)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке B%d: %v", index+1, err)
		}

		worker.MobileNumber, err = f.GetCellValue(sheetName, "C"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepathStr)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке C%d: %v", index+1, err)
		}

		worker.JobTitleInCompany, err = f.GetCellValue(sheetName, "D"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepathStr)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке B%d: %v", index+1, err)
		}

		worker.CompanyWorkerID, err = f.GetCellValue(sheetName, "E"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepathStr)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке C%d: %v", index+1, err)
		}

		workers = append(workers, worker)
		index++
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("Ошибка при закрытии файла: %v", err)
	}

	if err := os.Remove(filepathStr); err != nil {
		return fmt.Errorf("Ошибка при удалении временного файла: %v", err)
	}

	batch := make([]db.CreateWorkersBatchParams, len(workers))
	for i, w := range workers {
		batch[i] = db.CreateWorkersBatchParams{
			ProjectID:         pgInt8(w.ProjectID),
			Name:              pgText(w.Name),
			CompanyWorkerID:   pgText(w.CompanyWorkerID),
			JobTitleInCompany: pgText(w.JobTitleInCompany),
			JobTitleInProject: pgText(w.JobTitleInProject),
			MobileNumber:      pgText(w.MobileNumber),
		}
	}
	if _, err := u.q.CreateWorkersBatch(context.Background(), batch); err != nil {
		return err
	}

	return nil
}

func (u *workerUsecase) GetWorkerInformationForSearch(projectID uint) (dto.WorkerInformationForSearch, error) {
	ctx := context.Background()
	pid := pgInt8(projectID)

	names, err := u.q.ListDistinctWorkerNames(ctx, pid)
	if err != nil {
		return dto.WorkerInformationForSearch{}, err
	}
	jtCompany, err := u.q.ListDistinctWorkerJobTitlesInCompany(ctx, pid)
	if err != nil {
		return dto.WorkerInformationForSearch{}, err
	}
	jtProject, err := u.q.ListDistinctWorkerJobTitlesInProject(ctx, pid)
	if err != nil {
		return dto.WorkerInformationForSearch{}, err
	}
	companyIDs, err := u.q.ListDistinctWorkerCompanyIDs(ctx, pid)
	if err != nil {
		return dto.WorkerInformationForSearch{}, err
	}
	mobiles, err := u.q.ListDistinctWorkerMobileNumbers(ctx, pid)
	if err != nil {
		return dto.WorkerInformationForSearch{}, err
	}

	return dto.WorkerInformationForSearch{
		Name:              names,
		JobTitleInCompany: jtCompany,
		JobTitleInProject: jtProject,
		CompanyWorkerID:   companyIDs,
		MobileNumber:      mobiles,
	}, nil
}

func (u *workerUsecase) Export(projectID uint) (string, error) {
	workers, err := u.GetAll(projectID)
	if err != nil {
		return "", err
	}

	materialTempalteFilePath := filepath.Join("./internal/templates", "Шаблон для импорта Рабочего Персонала.xlsx")
	f, err := excelize.OpenFile(materialTempalteFilePath)
	if err != nil {
		f.Close()
		return "", fmt.Errorf("Не смог открыть файл: %v", err)
	}
	sheetName := "Импорт"
	startingRow := 2

	for index, worker := range workers {
		f.SetCellStr(sheetName, "A"+fmt.Sprint(startingRow+index), worker.Name)
		f.SetCellStr(sheetName, "B"+fmt.Sprint(startingRow+index), worker.JobTitleInProject)
		f.SetCellStr(sheetName, "C"+fmt.Sprint(startingRow+index), worker.MobileNumber)
		f.SetCellStr(sheetName, "D"+fmt.Sprint(startingRow+index), worker.JobTitleInCompany)
		f.SetCellStr(sheetName, "E"+fmt.Sprint(startingRow+index), worker.CompanyWorkerID)
	}

	exportFileName := "Экспорт Рабочего Персонала.xlsx"
	exportFilePath := filepath.Join("./storage/import_excel/temp/", exportFileName)
	if err := f.SaveAs(exportFilePath); err != nil {
		return "", err
	}

	return exportFileName, nil
}

func toModelWorker(w db.Worker) model.Worker {
	return model.Worker{
		ID:                uint(w.ID),
		ProjectID:         uintFromPgInt8(w.ProjectID),
		Name:              w.Name.String,
		CompanyWorkerID:   w.CompanyWorkerID.String,
		JobTitleInCompany: w.JobTitleInCompany.String,
		JobTitleInProject: w.JobTitleInProject.String,
		MobileNumber:      w.MobileNumber.String,
	}
}
