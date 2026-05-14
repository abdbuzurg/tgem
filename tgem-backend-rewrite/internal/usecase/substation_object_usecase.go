package usecase

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"backend-v2/internal/db"
	"backend-v2/internal/dto"
	"backend-v2/model"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xuri/excelize/v2"
)

type substationObjectUsecase struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

func NewSubstationObjectUsecase(pool *pgxpool.Pool) ISubstationObjectUsecase {
	return &substationObjectUsecase{
		pool: pool,
		q:    db.New(pool),
	}
}

type ISubstationObjectUsecase interface {
	GetPaginated(page, limit int, filter dto.SubstationObjectSearchParameters) ([]dto.SubstationObjectPaginated, error)
	Count(filter dto.SubstationObjectSearchParameters) (int64, error)
	Create(data dto.SubstationObjectCreate) (model.Substation_Object, error)
	Update(data dto.SubstationObjectCreate) (model.Substation_Object, error)
	Delete(id, projectID uint) error
	TemplateFile(filepath string, projectID uint) (string, error)
	Import(projectID uint, filepath string) error
	GetObjectNamesForSearch(projectID uint) ([]dto.DataForSelect[string], error)
	Export(projectID uint) (string, error)
	GetAll(projectID uint) ([]model.Object, error)
}

func (u *substationObjectUsecase) GetAll(projectID uint) ([]model.Object, error) {
	rows, err := u.q.ListSubstationObjectsAll(context.Background(), pgInt8(projectID))
	if err != nil {
		return nil, err
	}
	out := make([]model.Object, len(rows))
	for i, r := range rows {
		out[i] = toModelObject(r)
	}
	return out, nil
}

func (u *substationObjectUsecase) GetPaginated(page, limit int, filter dto.SubstationObjectSearchParameters) ([]dto.SubstationObjectPaginated, error) {
	ctx := context.Background()
	rows, err := u.q.ListSubstationObjectsPaginated(ctx, db.ListSubstationObjectsPaginatedParams{
		ProjectID: pgInt8(filter.ProjectID),
		Column2:   filter.ObjectName,
		Column3:   int64(filter.TeamID),
		Column4:   int64(filter.SupervisorWorkerID),
		Limit:     int32(limit),
		Offset:    int32((page - 1) * limit),
	})
	if err != nil {
		return []dto.SubstationObjectPaginated{}, err
	}

	result := []dto.SubstationObjectPaginated{}
	for _, row := range rows {
		supervisorNames, err := u.q.ListSupervisorNamesByObjectID(ctx, row.ObjectID)
		if err != nil {
			return []dto.SubstationObjectPaginated{}, err
		}

		teamNumbers, err := u.q.ListTeamNumbersByObjectID(ctx, row.ObjectID)
		if err != nil {
			return []dto.SubstationObjectPaginated{}, err
		}

		result = append(result, dto.SubstationObjectPaginated{
			ObjectID:             uint(row.ObjectID),
			ObjectDetailedID:     uint(row.ObjectDetailedID),
			Name:                 row.Name,
			Status:               row.Status,
			VoltageClass:         row.VoltageClass,
			NumberOfTransformers: strconv.FormatInt(row.NumberOfTransformers, 10),
			Supervisors:          supervisorNames,
			Teams:                teamNumbers,
		})
	}

	return result, nil
}

func (u *substationObjectUsecase) Count(filter dto.SubstationObjectSearchParameters) (int64, error) {
	return u.q.CountSubstationObjectsFiltered(context.Background(), db.CountSubstationObjectsFilteredParams{
		ProjectID: pgInt8(filter.ProjectID),
		Column2:   filter.ObjectName,
		Column3:   int64(filter.TeamID),
		Column4:   int64(filter.SupervisorWorkerID),
	})
}

func (u *substationObjectUsecase) Create(data dto.SubstationObjectCreate) (model.Substation_Object, error) {
	ctx := context.Background()
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return model.Substation_Object{}, err
	}
	defer tx.Rollback(ctx)

	qtx := u.q.WithTx(tx)

	detailRow, err := qtx.CreateSubstationObjectDetail(ctx, db.CreateSubstationObjectDetailParams{
		VoltageClass:         pgText(data.DetailedInfo.VoltageClass),
		NumberOfTransformers: pgInt8(data.DetailedInfo.NumberOfTransformers),
	})
	if err != nil {
		return model.Substation_Object{}, err
	}

	objectRow, err := qtx.CreateObject(ctx, db.CreateObjectParams{
		ObjectDetailedID: pgInt8(uint(detailRow.ID)),
		Type:             pgText("substation_objects"),
		Name:             pgText(data.BaseInfo.Name),
		Status:           pgText(data.BaseInfo.Status),
		ProjectID:        pgInt8(data.BaseInfo.ProjectID),
	})
	if err != nil {
		return model.Substation_Object{}, err
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
			return model.Substation_Object{}, err
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
			return model.Substation_Object{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return model.Substation_Object{}, err
	}

	return model.Substation_Object{
		ID:                   uint(detailRow.ID),
		VoltageClass:         detailRow.VoltageClass.String,
		NumberOfTransformers: uint(detailRow.NumberOfTransformers.Int64),
	}, nil
}

func (u *substationObjectUsecase) Update(data dto.SubstationObjectCreate) (model.Substation_Object, error) {
	ctx := context.Background()
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return model.Substation_Object{}, err
	}
	defer tx.Rollback(ctx)

	qtx := u.q.WithTx(tx)

	if err := qtx.UpdateSubstationObjectDetail(ctx, db.UpdateSubstationObjectDetailParams{
		ID:                   int64(data.BaseInfo.ObjectDetailedID),
		VoltageClass:         pgText(data.DetailedInfo.VoltageClass),
		NumberOfTransformers: pgInt8(data.DetailedInfo.NumberOfTransformers),
	}); err != nil {
		return model.Substation_Object{}, err
	}

	if _, err := qtx.UpdateObject(ctx, db.UpdateObjectParams{
		ID:               int64(data.BaseInfo.ID),
		ObjectDetailedID: pgInt8(data.BaseInfo.ObjectDetailedID),
		Type:             pgText("substation_objects"),
		Name:             pgText(data.BaseInfo.Name),
		Status:           pgText(data.BaseInfo.Status),
		ProjectID:        pgInt8(data.BaseInfo.ProjectID),
	}); err != nil {
		return model.Substation_Object{}, err
	}

	if err := qtx.DeleteObjectSupervisorsByObjectID(ctx, pgInt8(data.BaseInfo.ID)); err != nil {
		return model.Substation_Object{}, err
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
			return model.Substation_Object{}, err
		}
	}

	if err := qtx.DeleteObjectTeamsByObjectID(ctx, pgInt8(data.BaseInfo.ID)); err != nil {
		return model.Substation_Object{}, err
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
			return model.Substation_Object{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return model.Substation_Object{}, err
	}

	return model.Substation_Object{
		ID:                   data.BaseInfo.ObjectDetailedID,
		VoltageClass:         data.DetailedInfo.VoltageClass,
		NumberOfTransformers: data.DetailedInfo.NumberOfTransformers,
	}, nil
}

func (u *substationObjectUsecase) Delete(id, projectID uint) error {
	ctx := context.Background()
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := u.q.WithTx(tx)
	if err := qtx.DeleteSubstationObjectSupervisorsCascade(ctx, db.DeleteSubstationObjectSupervisorsCascadeParams{
		ProjectID: pgInt8(projectID),
		ID:        int64(id),
	}); err != nil {
		return err
	}
	if err := qtx.DeleteSubstationObjectTeamsCascade(ctx, db.DeleteSubstationObjectTeamsCascadeParams{
		ProjectID: pgInt8(projectID),
		ID:        int64(id),
	}); err != nil {
		return err
	}
	if err := qtx.DeleteSubstationObjectDetail(ctx, int64(id)); err != nil {
		return err
	}
	if err := qtx.DeleteObjectBySubstationDetailedID(ctx, pgInt8(id)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (u *substationObjectUsecase) TemplateFile(filePath string, projectID uint) (string, error) {
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
		"Шаблон импорта Подстанция - %s.xlsx",
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

func (u *substationObjectUsecase) Import(projectID uint, importFilePath string) error {
	ctx := context.Background()
	f, err := excelize.OpenFile(importFilePath)
	if err != nil {
		f.Close()
		os.Remove(importFilePath)
		return fmt.Errorf("Не смог открыть файл: %v", err)
	}

	sheetName := "Подстанция"
	rows, err := f.GetRows(sheetName)
	if err != nil {
		f.Close()
		os.Remove(importFilePath)
		return fmt.Errorf("Не смог найти таблицу 'Подстанция': %v", err)
	}

	if len(rows) == 1 {
		f.Close()
		os.Remove(importFilePath)
		return fmt.Errorf("Файл не имеет данных")
	}

	substations := []dto.SubstationObjectImportData{}
	index := 1
	for len(rows) > index {
		object := model.Object{
			ProjectID: projectID,
			Type:      "substation_objects",
		}

		substation := model.Substation_Object{}

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

		substation.VoltageClass, err = f.GetCellValue(sheetName, "C"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(importFilePath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке D%d: %v", index+1, err)
		}

		numberOfTransformersStr, err := f.GetCellValue(sheetName, "D"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(importFilePath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке E%d: %v", index+1, err)
		}

		numberOfTransformersUINT64, err := strconv.ParseUint(numberOfTransformersStr, 10, 64)
		if err != nil {
			f.Close()
			os.Remove(importFilePath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке E%d: %v", index+1, err)
		}
		substation.NumberOfTransformers = uint(numberOfTransformersUINT64)

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

		substations = append(substations, dto.SubstationObjectImportData{
			Object:     object,
			Substation: substation,
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

	return u.createInBatches(substations)
}

func (u *substationObjectUsecase) createInBatches(data []dto.SubstationObjectImportData) error {
	ctx := context.Background()
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := u.q.WithTx(tx)
	for index, row := range data {
		detailRow, err := qtx.CreateSubstationObjectDetail(ctx, db.CreateSubstationObjectDetailParams{
			VoltageClass:         pgText(row.Substation.VoltageClass),
			NumberOfTransformers: pgInt8(row.Substation.NumberOfTransformers),
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

		data[index].Substation.ID = uint(detailRow.ID)
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

func (u *substationObjectUsecase) GetObjectNamesForSearch(projectID uint) ([]dto.DataForSelect[string], error) {
	rows, err := u.q.ListSubstationObjectNamesForSearch(context.Background(), pgInt8(projectID))
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

func (u *substationObjectUsecase) Export(projectID uint) (string, error) {
	substationTempalteFilePath := filepath.Join("./internal/templates/", "Шаблон для импорта Подстанции.xlsx")
	f, err := excelize.OpenFile(substationTempalteFilePath)
	if err != nil {
		f.Close()
		return "", fmt.Errorf("Не смог открыть файл: %v", err)
	}
	sheetName := "Подстанция"
	startingRow := 2

	substationCount, err := u.Count(dto.SubstationObjectSearchParameters{ProjectID: projectID})
	if err != nil {
		return "", err
	}
	limit := 100
	page := 1

	for substationCount > 0 {
		substations, err := u.GetPaginated(page, limit, dto.SubstationObjectSearchParameters{ProjectID: projectID})
		if err != nil {
			return "", err
		}

		for index, substation := range substations {
			f.SetCellStr(sheetName, "A"+fmt.Sprint(startingRow+index), substation.Name)
			f.SetCellStr(sheetName, "B"+fmt.Sprint(startingRow+index), substation.Status)
			f.SetCellStr(sheetName, "C"+fmt.Sprint(startingRow+index), substation.VoltageClass)
			f.SetCellStr(sheetName, "D"+fmt.Sprint(startingRow+index), substation.NumberOfTransformers)

			supervisorsCombined := ""
			for index, supervisor := range substation.Supervisors {
				if index == 0 {
					supervisorsCombined += supervisor
					continue
				}

				supervisorsCombined += ", " + supervisor
			}
			f.SetCellStr(sheetName, "E"+fmt.Sprint(startingRow+index), supervisorsCombined)

			teamNumbersCombined := ""
			for index, teamNumber := range substation.Teams {
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
		substationCount -= int64(limit)
	}

	currentTime := time.Now()
	exportFileName := fmt.Sprintf(
		"Экспорт Подстанция - %s.xlsx",
		currentTime.Format("02-01-2006"),
	)
	exportFilePath := filepath.Join("./storage/import_excel/temp/", exportFileName)
	if err := f.SaveAs(exportFilePath); err != nil {
		return "", err
	}

	return exportFileName, nil
}
