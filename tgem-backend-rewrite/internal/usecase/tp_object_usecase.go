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

type tpObjectUsecase struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

func NewTPObjectUsecase(pool *pgxpool.Pool) ITPObjectUsecase {
	return &tpObjectUsecase{
		pool: pool,
		q:    db.New(pool),
	}
}

type ITPObjectUsecase interface {
	GetAllOnlyObjects(projectID uint) ([]model.Object, error)
	GetPaginated(page, limit int, filter dto.TPObjectSearchParameters) ([]dto.TPObjectPaginated, error)
	Count(filter dto.TPObjectSearchParameters) (int64, error)
	Create(data dto.TPObjectCreate) (model.TP_Object, error)
	Update(data dto.TPObjectCreate) (model.TP_Object, error)
	Delete(id, projectID uint) error
	TemplateFile(filePath string, projectID uint) (string, error)
	Import(projectID uint, filepath string) error
	GetObjectNamesForSearch(projectID uint) ([]dto.DataForSelect[string], error)
	Export(projectID uint) (string, error)
}

func (u *tpObjectUsecase) GetAllOnlyObjects(projectID uint) ([]model.Object, error) {
	rows, err := u.q.ListObjectsByProjectAndType(context.Background(), db.ListObjectsByProjectAndTypeParams{
		ProjectID: pgInt8(projectID),
		Type:      pgText("tp_objects"),
	})
	if err != nil {
		return nil, err
	}
	out := make([]model.Object, len(rows))
	for i, r := range rows {
		out[i] = toModelObject(r)
	}
	return out, nil
}

func (u *tpObjectUsecase) GetPaginated(page, limit int, filter dto.TPObjectSearchParameters) ([]dto.TPObjectPaginated, error) {
	ctx := context.Background()
	rows, err := u.q.ListTPObjectsPaginated(ctx, db.ListTPObjectsPaginatedParams{
		ProjectID: pgInt8(filter.ProjectID),
		Column2:   filter.ObjectName,
		Column3:   int64(filter.TeamID),
		Column4:   int64(filter.SupervisorWorkerID),
		Limit:     int32(limit),
		Offset:    int32((page - 1) * limit),
	})
	if err != nil {
		return []dto.TPObjectPaginated{}, err
	}

	result := []dto.TPObjectPaginated{}
	for _, object := range rows {
		supervisorNames, err := u.q.ListSupervisorNamesByObjectID(ctx, object.ObjectID)
		if err != nil {
			return []dto.TPObjectPaginated{}, err
		}

		teamNumbers, err := u.q.ListTeamNumbersByObjectID(ctx, object.ObjectID)
		if err != nil {
			return []dto.TPObjectPaginated{}, err
		}

		result = append(result, dto.TPObjectPaginated{
			ObjectID:         uint(object.ObjectID),
			ObjectDetailedID: uint(object.ObjectDetailedID),
			Name:             object.Name,
			Status:           object.Status,
			Model:            object.Model,
			VoltageClass:     object.VoltageClass,
			Supervisors:      supervisorNames,
			Teams:            teamNumbers,
		})
	}

	return result, nil
}

func (u *tpObjectUsecase) Count(filter dto.TPObjectSearchParameters) (int64, error) {
	return u.q.CountTPObjectsFiltered(context.Background(), db.CountTPObjectsFilteredParams{
		ProjectID: pgInt8(filter.ProjectID),
		Column2:   filter.ObjectName,
		Column3:   int64(filter.TeamID),
		Column4:   int64(filter.SupervisorWorkerID),
	})
}

func (u *tpObjectUsecase) Create(data dto.TPObjectCreate) (model.TP_Object, error) {
	ctx := context.Background()
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return model.TP_Object{}, err
	}
	defer tx.Rollback(ctx)

	qtx := u.q.WithTx(tx)

	tpRow, err := qtx.CreateTPObjectDetail(ctx, db.CreateTPObjectDetailParams{
		Model:        pgText(data.DetailedInfo.Model),
		VoltageClass: pgText(data.DetailedInfo.VoltageClass),
	})
	if err != nil {
		return model.TP_Object{}, err
	}

	objectRow, err := qtx.CreateObject(ctx, db.CreateObjectParams{
		ObjectDetailedID: pgInt8(uint(tpRow.ID)),
		Type:             pgText("tp_objects"),
		Name:             pgText(data.BaseInfo.Name),
		Status:           pgText(data.BaseInfo.Status),
		ProjectID:        pgInt8(data.BaseInfo.ProjectID),
	})
	if err != nil {
		return model.TP_Object{}, err
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
			return model.TP_Object{}, err
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
			return model.TP_Object{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return model.TP_Object{}, err
	}

	return model.TP_Object{
		ID:           uint(tpRow.ID),
		Model:        tpRow.Model.String,
		VoltageClass: tpRow.VoltageClass.String,
	}, nil
}

func (u *tpObjectUsecase) Update(data dto.TPObjectCreate) (model.TP_Object, error) {
	ctx := context.Background()
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return model.TP_Object{}, err
	}
	defer tx.Rollback(ctx)

	qtx := u.q.WithTx(tx)

	if err := qtx.UpdateTPObjectDetail(ctx, db.UpdateTPObjectDetailParams{
		ID:           int64(data.BaseInfo.ObjectDetailedID),
		Model:        pgText(data.DetailedInfo.Model),
		VoltageClass: pgText(data.DetailedInfo.VoltageClass),
	}); err != nil {
		return model.TP_Object{}, err
	}

	if _, err := qtx.UpdateObject(ctx, db.UpdateObjectParams{
		ID:               int64(data.BaseInfo.ID),
		ObjectDetailedID: pgInt8(data.BaseInfo.ObjectDetailedID),
		Type:             pgText("tp_objects"),
		Name:             pgText(data.BaseInfo.Name),
		Status:           pgText(data.BaseInfo.Status),
		ProjectID:        pgInt8(data.BaseInfo.ProjectID),
	}); err != nil {
		return model.TP_Object{}, err
	}

	if err := qtx.DeleteObjectSupervisorsByObjectID(ctx, pgInt8(data.BaseInfo.ID)); err != nil {
		return model.TP_Object{}, err
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
			return model.TP_Object{}, err
		}
	}

	if err := qtx.DeleteObjectTeamsByObjectID(ctx, pgInt8(data.BaseInfo.ID)); err != nil {
		return model.TP_Object{}, err
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
			return model.TP_Object{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return model.TP_Object{}, err
	}

	return model.TP_Object{
		ID:           data.BaseInfo.ObjectDetailedID,
		Model:        data.DetailedInfo.Model,
		VoltageClass: data.DetailedInfo.VoltageClass,
	}, nil
}

func (u *tpObjectUsecase) Delete(id, projectID uint) error {
	ctx := context.Background()
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := u.q.WithTx(tx)
	if err := qtx.DeleteTPObjectSupervisorsCascade(ctx, db.DeleteTPObjectSupervisorsCascadeParams{
		ProjectID: pgInt8(projectID),
		ID:        int64(id),
	}); err != nil {
		return err
	}
	if err := qtx.DeleteTPObjectTeamsCascade(ctx, db.DeleteTPObjectTeamsCascadeParams{
		ProjectID: pgInt8(projectID),
		ID:        int64(id),
	}); err != nil {
		return err
	}
	if err := qtx.DeleteTPObjectDetail(ctx, int64(id)); err != nil {
		return err
	}
	if err := qtx.DeleteObjectByTPObjectDetailedID(ctx, pgInt8(id)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (u *tpObjectUsecase) TemplateFile(filePath string, projectID uint) (string, error) {
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
		"Шаблон импорта ТП - %s.xlsx",
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

func (u *tpObjectUsecase) Import(projectID uint, importFilePath string) error {
	ctx := context.Background()
	f, err := excelize.OpenFile(importFilePath)
	if err != nil {
		f.Close()
		os.Remove(importFilePath)
		return fmt.Errorf("Не смог открыть файл: %v", err)
	}

	sheetName := "ТП"
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

	tps := []dto.TPObjectImportData{}
	index := 1
	for len(rows) > index {
		object := model.Object{
			ProjectID: projectID,
			Type:      "tp_objects",
		}

		tp := model.TP_Object{}

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

		tp.Model, err = f.GetCellValue(sheetName, "C"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(importFilePath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке D%d: %v", index+1, err)
		}

		tp.VoltageClass, err = f.GetCellValue(sheetName, "D"+fmt.Sprint(index+1))
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

		tps = append(tps, dto.TPObjectImportData{
			Object: object,
			TP:     tp,
			ObjectSupervisors: model.ObjectSupervisors{
				SupervisorWorkerID: supervisorWorkerID,
			},
			ObjectTeam: model.ObjectTeams{
				TeamID: teamID,
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

	return u.createInBatches(tps)
}

func (u *tpObjectUsecase) createInBatches(data []dto.TPObjectImportData) error {
	ctx := context.Background()
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := u.q.WithTx(tx)
	for index, row := range data {
		tpRow, err := qtx.CreateTPObjectDetail(ctx, db.CreateTPObjectDetailParams{
			Model:        pgText(row.TP.Model),
			VoltageClass: pgText(row.TP.VoltageClass),
		})
		if err != nil {
			return err
		}

		objectRow, err := qtx.CreateObject(ctx, db.CreateObjectParams{
			ObjectDetailedID: pgInt8(uint(tpRow.ID)),
			Type:             pgText(row.Object.Type),
			Name:             pgText(row.Object.Name),
			Status:           pgText(row.Object.Status),
			ProjectID:        pgInt8(row.Object.ProjectID),
		})
		if err != nil {
			return err
		}

		data[index].TP.ID = uint(tpRow.ID)
		data[index].Object.ObjectDetailedID = uint(tpRow.ID)
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

func (u *tpObjectUsecase) GetObjectNamesForSearch(projectID uint) ([]dto.DataForSelect[string], error) {
	rows, err := u.q.ListTPObjectNamesForSearch(context.Background(), pgInt8(projectID))
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

func (u *tpObjectUsecase) Export(projectID uint) (string, error) {
	ctx := context.Background()
	tpTempalteFilePath := filepath.Join("./internal/templates/", "Шаблон для импорта ТП.xlsx")
	f, err := excelize.OpenFile(tpTempalteFilePath)
	if err != nil {
		f.Close()
		return "", fmt.Errorf("Не смог открыть файл: %v", err)
	}
	sheetName := "ТП"
	startingRow := 2

	tpCount, err := u.Count(dto.TPObjectSearchParameters{ProjectID: projectID})
	if err != nil {
		return "", err
	}
	limit := 100
	page := 1

	for tpCount > 0 {
		tps, err := u.GetPaginated(page, limit, dto.TPObjectSearchParameters{ProjectID: projectID})
		if err != nil {
			return "", err
		}

		for index, tp := range tps {
			supervisorNames, err := u.q.ListSupervisorNamesByObjectID(ctx, int64(tp.ObjectID))
			if err != nil {
				return "", err
			}

			teamNumbers, err := u.q.ListTeamNumbersByObjectID(ctx, int64(tp.ObjectID))
			if err != nil {
				return "", err
			}

			f.SetCellStr(sheetName, "A"+fmt.Sprint(startingRow+index), tp.Name)
			f.SetCellStr(sheetName, "B"+fmt.Sprint(startingRow+index), tp.Status)
			f.SetCellStr(sheetName, "C"+fmt.Sprint(startingRow+index), tp.Model)
			f.SetCellStr(sheetName, "D"+fmt.Sprint(startingRow+index), tp.VoltageClass)

			supervisorsCombined := ""
			for index, supervisor := range supervisorNames {
				if index == 0 {
					supervisorsCombined += supervisor
					continue
				}

				supervisorsCombined += ", " + supervisor
			}
			f.SetCellStr(sheetName, "E"+fmt.Sprint(startingRow+index), supervisorsCombined)

			teamNumbersCombined := ""
			for index, teamNumber := range teamNumbers {
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
		tpCount -= int64(limit)
	}

	currentTime := time.Now()
	exportFileName := fmt.Sprintf(
		"Экспорт ТП - %s.xlsx",
		currentTime.Format("02-01-2006"),
	)
	exportFilePath := filepath.Join("./storage/import_excel/temp/", exportFileName)
	if err := f.SaveAs(exportFilePath); err != nil {
		return "", err
	}

	return exportFileName, nil
}
