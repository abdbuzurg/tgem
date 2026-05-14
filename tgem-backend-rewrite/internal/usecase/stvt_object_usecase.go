package usecase

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"backend-v2/internal/db"
	"backend-v2/internal/dto"
	"backend-v2/model"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xuri/excelize/v2"
)

type stvtObjectUsecase struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

func NewSTVTObjectUsecase(pool *pgxpool.Pool) ISTVTObjectUsecase {
	return &stvtObjectUsecase{
		pool: pool,
		q:    db.New(pool),
	}
}

type ISTVTObjectUsecase interface {
	GetPaginated(page, limit int, filter dto.STVTObjectSearchParameters) ([]dto.STVTObjectPaginated, error)
	Count(filter dto.STVTObjectSearchParameters) (int64, error)
	Create(data dto.STVTObjectCreate) (model.STVT_Object, error)
	Update(data dto.STVTObjectCreate) (model.STVT_Object, error)
	Delete(id, projectID uint) error
	TemplateFile(filePath string, projectID uint) (string, error)
	Import(projectID uint, filePath string) error
	GetObjectNamesForSearch(projectID uint) ([]dto.DataForSelect[string], error)
	Export(projectID uint) (string, error)
}

func (u *stvtObjectUsecase) GetPaginated(page, limit int, filter dto.STVTObjectSearchParameters) ([]dto.STVTObjectPaginated, error) {
	ctx := context.Background()
	rows, err := u.q.ListSTVTObjectsPaginated(ctx, db.ListSTVTObjectsPaginatedParams{
		ProjectID: pgInt8(filter.ProjectID),
		Column2:   filter.ObjectName,
		Column3:   int64(filter.TeamID),
		Column4:   int64(filter.SupervisorWorkerID),
		Limit:     int32(limit),
		Offset:    int32((page - 1) * limit),
	})
	if err != nil {
		return []dto.STVTObjectPaginated{}, err
	}

	result := []dto.STVTObjectPaginated{}
	for _, row := range rows {
		supervisorNames, err := u.q.ListSupervisorNamesByObjectID(ctx, row.ObjectID)
		if err != nil {
			return []dto.STVTObjectPaginated{}, err
		}

		teamNumbers, err := u.q.ListTeamNumbersByObjectID(ctx, row.ObjectID)
		if err != nil {
			return []dto.STVTObjectPaginated{}, err
		}

		result = append(result, dto.STVTObjectPaginated{
			ObjectID:         uint(row.ObjectID),
			ObjectDetailedID: uint(row.ObjectDetailedID),
			Name:             row.Name,
			Status:           row.Status,
			VoltageClass:     row.VoltageClass,
			TTCoefficient:    row.TtCoefficient,
			Supervisors:      supervisorNames,
			Teams:            teamNumbers,
		})
	}

	return result, nil
}

func (u *stvtObjectUsecase) Count(filter dto.STVTObjectSearchParameters) (int64, error) {
	return u.q.CountSTVTObjectsFiltered(context.Background(), db.CountSTVTObjectsFilteredParams{
		ProjectID: pgInt8(filter.ProjectID),
		Column2:   filter.ObjectName,
		Column3:   int64(filter.TeamID),
		Column4:   int64(filter.SupervisorWorkerID),
	})
}

func (u *stvtObjectUsecase) Create(data dto.STVTObjectCreate) (model.STVT_Object, error) {
	ctx := context.Background()
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return model.STVT_Object{}, err
	}
	defer tx.Rollback(ctx)

	qtx := u.q.WithTx(tx)

	detailRow, err := qtx.CreateSTVTObjectDetail(ctx, db.CreateSTVTObjectDetailParams{
		VoltageClass:  pgText(data.DetailedInfo.VoltageClass),
		TtCoefficient: pgText(data.DetailedInfo.TTCoefficient),
	})
	if err != nil {
		return model.STVT_Object{}, err
	}

	objectRow, err := qtx.CreateObject(ctx, db.CreateObjectParams{
		ObjectDetailedID: pgInt8(uint(detailRow.ID)),
		Type:             pgText("stvt_objects"),
		Name:             pgText(data.BaseInfo.Name),
		Status:           pgText(data.BaseInfo.Status),
		ProjectID:        pgInt8(data.BaseInfo.ProjectID),
	})
	if err != nil {
		return model.STVT_Object{}, err
	}

	if len(data.Supervisors) != 0 {
		batch := make([]db.CreateObjectSupervisorsBatchParams, len(data.Supervisors))
		for i, supervisorWorkerID := range data.Supervisors {
			batch[i] = db.CreateObjectSupervisorsBatchParams{
				ObjectID:           pgInt8(uint(objectRow.ID)),
				SupervisorWorkerID: pgInt8(supervisorWorkerID),
			}
		}
		if _, err := qtx.CreateObjectSupervisorsBatch(ctx, batch); err != nil {
			return model.STVT_Object{}, err
		}
	}

	if len(data.Teams) != 0 {
		batch := make([]db.CreateObjectTeamsBatchParams, len(data.Teams))
		for i, teamID := range data.Teams {
			batch[i] = db.CreateObjectTeamsBatchParams{
				ObjectID: pgInt8(uint(objectRow.ID)),
				TeamID:   pgInt8(teamID),
			}
		}
		if _, err := qtx.CreateObjectTeamsBatch(ctx, batch); err != nil {
			return model.STVT_Object{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return model.STVT_Object{}, err
	}

	return model.STVT_Object{
		ID:            uint(detailRow.ID),
		VoltageClass:  detailRow.VoltageClass.String,
		TTCoefficient: detailRow.TtCoefficient.String,
	}, nil
}

func (u *stvtObjectUsecase) Update(data dto.STVTObjectCreate) (model.STVT_Object, error) {
	ctx := context.Background()
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return model.STVT_Object{}, err
	}
	defer tx.Rollback(ctx)

	qtx := u.q.WithTx(tx)

	if err := qtx.UpdateSTVTObjectDetail(ctx, db.UpdateSTVTObjectDetailParams{
		ID:            int64(data.BaseInfo.ObjectDetailedID),
		VoltageClass:  pgText(data.DetailedInfo.VoltageClass),
		TtCoefficient: pgText(data.DetailedInfo.TTCoefficient),
	}); err != nil {
		return model.STVT_Object{}, err
	}

	if _, err := qtx.UpdateObject(ctx, db.UpdateObjectParams{
		ID:               int64(data.BaseInfo.ID),
		ObjectDetailedID: pgInt8(data.BaseInfo.ObjectDetailedID),
		Type:             pgText("stvt_objects"),
		Name:             pgText(data.BaseInfo.Name),
		Status:           pgText(data.BaseInfo.Status),
		ProjectID:        pgInt8(data.BaseInfo.ProjectID),
	}); err != nil {
		return model.STVT_Object{}, err
	}

	if err := qtx.DeleteObjectSupervisorsByObjectID(ctx, pgInt8(data.BaseInfo.ID)); err != nil {
		return model.STVT_Object{}, err
	}

	if len(data.Supervisors) != 0 {
		batch := make([]db.CreateObjectSupervisorsBatchParams, len(data.Supervisors))
		for i, supervisorWorkerID := range data.Supervisors {
			batch[i] = db.CreateObjectSupervisorsBatchParams{
				ObjectID:           pgInt8(data.BaseInfo.ID),
				SupervisorWorkerID: pgInt8(supervisorWorkerID),
			}
		}
		if _, err := qtx.CreateObjectSupervisorsBatch(ctx, batch); err != nil {
			return model.STVT_Object{}, err
		}
	}

	if err := qtx.DeleteObjectTeamsByObjectID(ctx, pgInt8(data.BaseInfo.ID)); err != nil {
		return model.STVT_Object{}, err
	}

	if len(data.Teams) != 0 {
		batch := make([]db.CreateObjectTeamsBatchParams, len(data.Teams))
		for i, teamID := range data.Teams {
			batch[i] = db.CreateObjectTeamsBatchParams{
				ObjectID: pgInt8(data.BaseInfo.ID),
				TeamID:   pgInt8(teamID),
			}
		}
		if _, err := qtx.CreateObjectTeamsBatch(ctx, batch); err != nil {
			return model.STVT_Object{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return model.STVT_Object{}, err
	}

	return model.STVT_Object{
		ID:            data.BaseInfo.ObjectDetailedID,
		VoltageClass:  data.DetailedInfo.VoltageClass,
		TTCoefficient: data.DetailedInfo.TTCoefficient,
	}, nil
}

func (u *stvtObjectUsecase) Delete(id, projectID uint) error {
	ctx := context.Background()
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := u.q.WithTx(tx)
	if err := qtx.DeleteSTVTObjectSupervisorsCascade(ctx, db.DeleteSTVTObjectSupervisorsCascadeParams{
		ProjectID: pgInt8(projectID),
		ID:        int64(id),
	}); err != nil {
		return err
	}
	if err := qtx.DeleteSTVTObjectTeamsCascade(ctx, db.DeleteSTVTObjectTeamsCascadeParams{
		ProjectID: pgInt8(projectID),
		ID:        int64(id),
	}); err != nil {
		return err
	}
	if err := qtx.DeleteSTVTObjectDetail(ctx, int64(id)); err != nil {
		return err
	}
	if err := qtx.DeleteObjectBySTVTDetailedID(ctx, pgInt8(id)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (u *stvtObjectUsecase) TemplateFile(filePath string, projectID uint) (string, error) {
	ctx := context.Background()
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		f.Close()
		return "", fmt.Errorf("Не смог открыть шаблонный файл: %v", err)
	}

	sheetName := "Супервайзеры"
	allSupervisors, err := u.q.ListWorkersByJobTitleInProject(ctx, db.ListWorkersByJobTitleInProjectParams{
		JobTitleInProject: pgText("Супервайзер"),
		ProjectID:         pgInt8(projectID),
	})
	if err != nil {
		f.Close()
		return "", fmt.Errorf("Данные супервайзеров недоступны: %v", err)
	}

	for index, supervisor := range allSupervisors {
		f.SetCellValue(sheetName, "A"+fmt.Sprint(index+2), supervisor.Name.String)
	}

	allTeams, err := u.q.ListTeamsByProject(ctx, pgInt8(projectID))
	if err != nil {
		f.Close()
		return "", fmt.Errorf("Данны бригад не доступны: %v", err)
	}

	teamSheetName := "Бригады"
	for index, team := range allTeams {
		f.SetCellStr(teamSheetName, "A"+fmt.Sprint(index+2), team.Number.String)
	}

	currentTime := time.Now()
	temporaryFileName := fmt.Sprintf(
		"Шаблон импорта СТВТ - %s.xlsx",
		currentTime.Format("02-01-2006"),
	)

	temporaryFilePath := filepath.Join("./storage/import_excel/temp/", temporaryFileName)
	if err := f.SaveAs(temporaryFilePath); err != nil {
		return "", fmt.Errorf("Не удалось обновить шаблон с новыми данными: %v", err)
	}

	if err := f.Close(); err != nil {
		return "", err
	}

	return temporaryFilePath, nil
}

func (u *stvtObjectUsecase) Import(projectID uint, importFilePath string) error {
	ctx := context.Background()
	f, err := excelize.OpenFile(importFilePath)
	if err != nil {
		f.Close()
		os.Remove(importFilePath)
		return fmt.Errorf("Не смог открыть файл: %v", err)
	}

	sheetName := "СТВТ"
	rows, err := f.GetRows(sheetName)
	if err != nil {
		f.Close()
		os.Remove(importFilePath)
		return fmt.Errorf("Не смог найти таблицу 'Импорт': %v", err)
	}

	if len(rows) == 1 {
		f.Close()
		os.Remove(importFilePath)
		return fmt.Errorf("Файл не имеет данных")
	}

	stvts := []dto.STVTObjectImportData{}
	index := 1
	for len(rows) > index {
		object := model.Object{
			ProjectID: projectID,
			Type:      "stvt_objects",
		}

		stvt := model.STVT_Object{}

		object.Name, err = f.GetCellValue(sheetName, "A"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(importFilePath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке А%d: %v", index+1, err)
		}

		object.Status, err = f.GetCellValue(sheetName, "B"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(importFilePath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке B%d: %v", index+1, err)
		}

		stvt.VoltageClass, err = f.GetCellValue(sheetName, "C"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(importFilePath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке D%d: %v", index+1, err)
		}

		stvt.TTCoefficient, err = f.GetCellValue(sheetName, "D"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(importFilePath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке E%d: %v", index+1, err)
		}

		supervisorName, err := f.GetCellValue(sheetName, "E"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(importFilePath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке G%d: %v", index+1, err)
		}
		var supervisorWorkerID uint
		if supervisorName != "" {
			worker, err := u.q.GetWorkerByName(ctx, pgText(supervisorName))
			if err != nil {
				f.Close()
				os.Remove(importFilePath)
				return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке G%d: %v", index+1, err)
			}
			supervisorWorkerID = uint(worker.ID)
		}

		teamNumber, err := f.GetCellValue(sheetName, "F"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(importFilePath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке H%d: %v", index+1, err)
		}
		var teamID uint
		if teamNumber != "" {
			team, err := u.q.GetTeamByNumber(ctx, pgText(teamNumber))
			if err != nil {
				f.Close()
				os.Remove(importFilePath)
				return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке H%d: %v", index+1, err)
			}
			teamID = uint(team.ID)
		}

		stvts = append(stvts, dto.STVTObjectImportData{
			Object: object,
			STVT:   stvt,
			ObjectTeam: model.ObjectTeams{
				TeamID: teamID,
			},
			ObjectSupervisors: model.ObjectSupervisors{
				SupervisorWorkerID: supervisorWorkerID,
			},
		})
		index++
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("Ошибка при закрытии файла: %v", err)
	}

	if err := os.Remove(importFilePath); err != nil {
		return fmt.Errorf("Ошибка при удалении временного файла: %v", err)
	}

	return u.createInBatches(stvts)
}

func (u *stvtObjectUsecase) createInBatches(data []dto.STVTObjectImportData) error {
	ctx := context.Background()
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := u.q.WithTx(tx)
	for index, row := range data {
		detailRow, err := qtx.CreateSTVTObjectDetail(ctx, db.CreateSTVTObjectDetailParams{
			VoltageClass:  pgText(row.STVT.VoltageClass),
			TtCoefficient: pgText(row.STVT.TTCoefficient),
		})
		if err != nil {
			return err
		}

		objectRow, err := qtx.CreateObject(ctx, db.CreateObjectParams{
			ObjectDetailedID: pgInt8(uint(detailRow.ID)),
			Type:             pgText(row.Object.Type),
			Name:             pgText(row.Object.Name),
			Status:           pgText(row.Object.Status),
			ProjectID:        pgInt8(row.Object.ProjectID),
		})
		if err != nil {
			return err
		}

		data[index].STVT.ID = uint(detailRow.ID)
		data[index].Object.ObjectDetailedID = uint(detailRow.ID)
		data[index].Object.ID = uint(objectRow.ID)

		if row.ObjectSupervisors.SupervisorWorkerID != 0 {
			if _, err := qtx.CreateObjectSupervisorsBatch(ctx, []db.CreateObjectSupervisorsBatchParams{{
				ObjectID:           pgInt8(uint(objectRow.ID)),
				SupervisorWorkerID: pgInt8(row.ObjectSupervisors.SupervisorWorkerID),
			}}); err != nil {
				return err
			}
		}

		if row.ObjectTeam.TeamID != 0 {
			if err := qtx.CreateObjectTeam(ctx, db.CreateObjectTeamParams{
				ObjectID: pgInt8(uint(objectRow.ID)),
				TeamID:   pgInt8(row.ObjectTeam.TeamID),
			}); err != nil {
				return err
			}
		}
	}
	return tx.Commit(ctx)
}

func (u *stvtObjectUsecase) GetObjectNamesForSearch(projectID uint) ([]dto.DataForSelect[string], error) {
	rows, err := u.q.ListSTVTObjectNamesForSearch(context.Background(), pgInt8(projectID))
	if err != nil {
		return nil, err
	}
	out := make([]dto.DataForSelect[string], len(rows))
	for i, r := range rows {
		out[i] = dto.DataForSelect[string]{
			Label: r.Label,
			Value: r.Value,
		}
	}
	return out, nil
}

func (u *stvtObjectUsecase) Export(projectID uint) (string, error) {
	stvtTempalteFilePath := filepath.Join("./internal/templates/", "Шаблон для импорта СТВТ.xlsx")
	f, err := excelize.OpenFile(stvtTempalteFilePath)
	if err != nil {
		f.Close()
		return "", fmt.Errorf("Не смог открыть файл: %v", err)
	}
	sheetName := "СТВТ"
	startingRow := 2

	stvtCount, err := u.Count(dto.STVTObjectSearchParameters{ProjectID: projectID})
	if err != nil {
		return "", err
	}
	limit := 100
	page := 1

	for stvtCount > 0 {
		stvts, err := u.GetPaginated(page, limit, dto.STVTObjectSearchParameters{ProjectID: projectID})
		if err != nil {
			return "", err
		}

		for index, stvt := range stvts {
			f.SetCellStr(sheetName, "A"+fmt.Sprint(startingRow+index), stvt.Name)
			f.SetCellStr(sheetName, "B"+fmt.Sprint(startingRow+index), stvt.Status)
			f.SetCellStr(sheetName, "C"+fmt.Sprint(startingRow+index), stvt.VoltageClass)
			f.SetCellStr(sheetName, "D"+fmt.Sprint(startingRow+index), stvt.TTCoefficient)

			supervisorsCombined := ""
			for index, supervisor := range stvt.Supervisors {
				if index == 0 {
					supervisorsCombined += supervisor
					continue
				}

				supervisorsCombined += ", " + supervisor
			}
			f.SetCellStr(sheetName, "E"+fmt.Sprint(startingRow+index), supervisorsCombined)

			teamNumbersCombined := ""
			for index, teamNumber := range stvt.Teams {
				if index == 0 {
					teamNumbersCombined += teamNumber
					continue
				}

				teamNumbersCombined += ", " + teamNumber
			}
			f.SetCellStr(sheetName, "F"+fmt.Sprint(startingRow+index), teamNumbersCombined)
		}

		startingRow = page*limit + 2
		page++
		stvtCount -= int64(limit)
	}

	currentTime := time.Now()
	exportFileName := fmt.Sprintf(
		"Экспорт СТВТ - %s.xlsx",
		currentTime.Format("02-01-2006"),
	)
	exportFilePath := filepath.Join("./storage/import_excel/temp/", exportFileName)
	if err := f.SaveAs(exportFilePath); err != nil {
		return "", err
	}
	return exportFileName, nil
}
