package usecase

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"backend-v2/internal/db"
	"backend-v2/internal/dto"
	"backend-v2/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xuri/excelize/v2"
)

type substationCellObjectUsecase struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

func NewSubstationCellObjectUsecase(pool *pgxpool.Pool) ISubstationCellObjectUsecase {
	return &substationCellObjectUsecase{
		pool: pool,
		q:    db.New(pool),
	}
}

type ISubstationCellObjectUsecase interface {
	GetPaginated(int, int, dto.SubstationCellObjectSearchParameters) ([]dto.SubstationCellObjectPaginated, error)
	Count(dto.SubstationCellObjectSearchParameters) (int64, error)
	Create(dto.SubstationCellObjectCreate) (model.SubstationCellObject, error)
	Update(dto.SubstationCellObjectCreate) (model.SubstationCellObject, error)
	Delete(id, projectID uint) error
	TemplateFile(filePath string, projectID uint) (string, error)
	Import(projectID uint, filepath string) error
	Export(projectID uint) (string, error)
	GetObjectNamesForSearch(projectID uint) ([]dto.DataForSelect[string], error)
}

func (u *substationCellObjectUsecase) GetPaginated(page, limit int, filter dto.SubstationCellObjectSearchParameters) ([]dto.SubstationCellObjectPaginated, error) {
	ctx := context.Background()
	rows, err := u.q.ListSubstationCellObjectsPaginated(ctx, db.ListSubstationCellObjectsPaginatedParams{
		ProjectID: pgInt8(filter.ProjectID),
		Column2:   filter.ObjectName,
		Column3:   int64(filter.TeamID),
		Column4:   int64(filter.SupervisorWorkerID),
		Column5:   int64(filter.SubstationObjectID),
		Limit:     int32(limit),
		Offset:    int32((page - 1) * limit),
	})
	if err != nil {
		return []dto.SubstationCellObjectPaginated{}, err
	}

	result := []dto.SubstationCellObjectPaginated{}
	for _, row := range rows {
		supervisorNames, err := u.q.ListSupervisorNamesByObjectID(ctx, row.ObjectID)
		if err != nil {
			return []dto.SubstationCellObjectPaginated{}, err
		}

		teamNumbers, err := u.q.ListTeamNumbersByObjectID(ctx, row.ObjectID)
		if err != nil {
			return []dto.SubstationCellObjectPaginated{}, err
		}

		substationName, err := u.q.GetSubstationNameByCellObjectID(ctx, pgInt8(uint(row.ObjectID)))
		if errors.Is(err, pgx.ErrNoRows) {
			substationName = ""
		} else if err != nil {
			return []dto.SubstationCellObjectPaginated{}, err
		}

		result = append(result, dto.SubstationCellObjectPaginated{
			ObjectID:         uint(row.ObjectID),
			ObjectDetailedID: uint(row.ObjectDetailedID),
			Name:             row.Name,
			Status:           row.Status,
			Supervisors:      supervisorNames,
			Teams:            teamNumbers,
			SubstationName:   substationName,
		})
	}

	return result, nil
}

func (u *substationCellObjectUsecase) Count(filter dto.SubstationCellObjectSearchParameters) (int64, error) {
	return u.q.CountSubstationCellObjectsFiltered(context.Background(), db.CountSubstationCellObjectsFilteredParams{
		ProjectID: pgInt8(filter.ProjectID),
		Column2:   filter.ObjectName,
		Column3:   int64(filter.TeamID),
		Column4:   int64(filter.SupervisorWorkerID),
		Column5:   int64(filter.SubstationObjectID),
	})
}

func (u *substationCellObjectUsecase) Create(data dto.SubstationCellObjectCreate) (model.SubstationCellObject, error) {
	ctx := context.Background()
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return model.SubstationCellObject{}, err
	}
	defer tx.Rollback(ctx)

	qtx := u.q.WithTx(tx)

	detailID, err := qtx.CreateSubstationCellObjectDetail(ctx)
	if err != nil {
		return model.SubstationCellObject{}, err
	}

	objectRow, err := qtx.CreateObject(ctx, db.CreateObjectParams{
		ObjectDetailedID: pgInt8(uint(detailID)),
		Type:             pgText("substation_cell_objects"),
		Name:             pgText(data.BaseInfo.Name),
		Status:           pgText(data.BaseInfo.Status),
		ProjectID:        pgInt8(data.BaseInfo.ProjectID),
	})
	if err != nil {
		return model.SubstationCellObject{}, err
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
			return model.SubstationCellObject{}, err
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
			return model.SubstationCellObject{}, err
		}
	}

	if data.SubstationObjectID != 0 {
		if err := qtx.CreateSubstationCellNourash(ctx, db.CreateSubstationCellNourashParams{
			SubstationObjectID:     pgInt8(data.SubstationObjectID),
			SubstationCellObjectID: pgInt8(uint(objectRow.ID)),
		}); err != nil {
			return model.SubstationCellObject{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return model.SubstationCellObject{}, err
	}

	return model.SubstationCellObject{ID: uint(detailID)}, nil
}

func (u *substationCellObjectUsecase) Update(data dto.SubstationCellObjectCreate) (model.SubstationCellObject, error) {
	ctx := context.Background()
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return model.SubstationCellObject{}, err
	}
	defer tx.Rollback(ctx)

	qtx := u.q.WithTx(tx)

	// substation_cell_objects has only the id column — nothing on the
	// detail row to update. The GORM-era code's Updates(&substationCell)
	// against an empty struct was a no-op too. Skipped here.

	if _, err := qtx.UpdateObject(ctx, db.UpdateObjectParams{
		ID:               int64(data.BaseInfo.ID),
		ObjectDetailedID: pgInt8(data.BaseInfo.ObjectDetailedID),
		Type:             pgText("substation_cell_objects"),
		Name:             pgText(data.BaseInfo.Name),
		Status:           pgText(data.BaseInfo.Status),
		ProjectID:        pgInt8(data.BaseInfo.ProjectID),
	}); err != nil {
		return model.SubstationCellObject{}, err
	}

	if err := qtx.DeleteObjectSupervisorsByObjectID(ctx, pgInt8(data.BaseInfo.ID)); err != nil {
		return model.SubstationCellObject{}, err
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
			return model.SubstationCellObject{}, err
		}
	}

	if err := qtx.DeleteObjectTeamsByObjectID(ctx, pgInt8(data.BaseInfo.ID)); err != nil {
		return model.SubstationCellObject{}, err
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
			return model.SubstationCellObject{}, err
		}
	}

	if err := qtx.DeleteSubstationCellNourashesByCellObjectID(ctx, pgInt8(data.BaseInfo.ID)); err != nil {
		return model.SubstationCellObject{}, err
	}

	if data.SubstationObjectID != 0 {
		if err := qtx.CreateSubstationCellNourash(ctx, db.CreateSubstationCellNourashParams{
			SubstationObjectID:     pgInt8(data.SubstationObjectID),
			SubstationCellObjectID: pgInt8(data.BaseInfo.ID),
		}); err != nil {
			return model.SubstationCellObject{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return model.SubstationCellObject{}, err
	}

	return model.SubstationCellObject{ID: data.BaseInfo.ObjectDetailedID}, nil
}

func (u *substationCellObjectUsecase) Delete(id, projectID uint) error {
	ctx := context.Background()
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := u.q.WithTx(tx)
	if err := qtx.DeleteSubstationCellObjectSupervisorsCascade(ctx, db.DeleteSubstationCellObjectSupervisorsCascadeParams{
		ProjectID: pgInt8(projectID),
		ID:        int64(id),
	}); err != nil {
		return err
	}
	if err := qtx.DeleteSubstationCellObjectTeamsCascade(ctx, db.DeleteSubstationCellObjectTeamsCascadeParams{
		ProjectID: pgInt8(projectID),
		ID:        int64(id),
	}); err != nil {
		return err
	}
	if err := qtx.DeleteSubstationCellObjectNourashesCascade(ctx, db.DeleteSubstationCellObjectNourashesCascadeParams{
		ProjectID: pgInt8(projectID),
		ID:        int64(id),
	}); err != nil {
		return err
	}
	if err := qtx.DeleteSubstationCellObjectDetail(ctx, int64(id)); err != nil {
		return err
	}
	if err := qtx.DeleteObjectBySubstationCellDetailedID(ctx, pgInt8(id)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (u *substationCellObjectUsecase) TemplateFile(filePath string, projectID uint) (string, error) {
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

	substationSheetName := "Подстанции"
	allSubstationNames, err := u.q.ListSubstationObjectNamesByProject(ctx, pgInt8(projectID))
	if err != nil {
		f.Close()
		return "", fmt.Errorf("Данны подстанций не доступны: %v", err)
	}

	for index, name := range allSubstationNames {
		f.SetCellStr(substationSheetName, "A"+fmt.Sprint(index+2), name)
	}

	currentTime := time.Now()
	temporaryFileName := fmt.Sprintf(
		"Шаблон для импорт Ячеек Подстанции - %s.xlsx",
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

func (u *substationCellObjectUsecase) Import(projectID uint, importFilePath string) error {
	ctx := context.Background()
	f, err := excelize.OpenFile(importFilePath)
	if err != nil {
		f.Close()
		os.Remove(importFilePath)
		return fmt.Errorf("Не смог открыть файл: %v", err)
	}

	sheetName := "Ячейка Подстанции"
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

	substationCells := []dto.SubstationCellObjectImportData{}
	index := 1
	for len(rows) > index {
		object := model.Object{
			ProjectID: projectID,
			Type:      "substation_cell_objects",
		}

		substationCell := model.SubstationCellObject{}

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

		supervisorName, err := f.GetCellValue(sheetName, "C"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(importFilePath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке C%d: %v", index+1, err)
		}

		var supervisorWorkerID uint
		if supervisorName != "" {
			worker, err := u.q.GetWorkerByName(ctx, pgText(supervisorName))
			if err != nil {
				f.Close()
				os.Remove(importFilePath)
				return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке C%d: %v", index+1, err)
			}
			supervisorWorkerID = uint(worker.ID)
		}

		teamNumber, err := f.GetCellValue(sheetName, "D"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(importFilePath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке D%d: %v", index+1, err)
		}
		var teamID uint
		if teamNumber != "" {
			team, err := u.q.GetTeamByNumber(ctx, pgText(teamNumber))
			if err != nil {
				f.Close()
				os.Remove(importFilePath)
				return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке D%d: %v", index+1, err)
			}
			teamID = uint(team.ID)
		}

		substationName, err := f.GetCellValue(sheetName, "E"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(importFilePath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке E%d: %v", index+1, err)
		}
		var substationObjectID uint
		if substationName != "" {
			obj, err := u.q.GetSubstationObjectByName(ctx, pgText(substationName))
			if err != nil {
				f.Close()
				os.Remove(importFilePath)
				return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке D%d: %v", index+1, err)
			}
			substationObjectID = uint(obj.ID)
		}

		substationCells = append(substationCells, dto.SubstationCellObjectImportData{
			Object:         object,
			SubstationCell: substationCell,
			ObjectTeam: model.ObjectTeams{
				TeamID: teamID,
			},
			ObjectSupervisors: model.ObjectSupervisors{
				SupervisorWorkerID: supervisorWorkerID,
			},
			Nourashes: model.SubstationCellNourashesSubstationObject{
				SubstationObjectID: substationObjectID,
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

	return u.createInBatches(substationCells)
}

func (u *substationCellObjectUsecase) createInBatches(data []dto.SubstationCellObjectImportData) error {
	ctx := context.Background()
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := u.q.WithTx(tx)
	for index, row := range data {
		detailID, err := qtx.CreateSubstationCellObjectDetail(ctx)
		if err != nil {
			return err
		}

		objectRow, err := qtx.CreateObject(ctx, db.CreateObjectParams{
			ObjectDetailedID: pgInt8(uint(detailID)),
			Type:             pgText(row.Object.Type),
			Name:             pgText(row.Object.Name),
			Status:           pgText(row.Object.Status),
			ProjectID:        pgInt8(row.Object.ProjectID),
		})
		if err != nil {
			return err
		}

		data[index].SubstationCell.ID = uint(detailID)
		data[index].Object.ObjectDetailedID = uint(detailID)
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

		if row.Nourashes.SubstationObjectID != 0 {
			if err := qtx.CreateSubstationCellNourash(ctx, db.CreateSubstationCellNourashParams{
				SubstationObjectID:     pgInt8(row.Nourashes.SubstationObjectID),
				SubstationCellObjectID: pgInt8(uint(objectRow.ID)),
			}); err != nil {
				return err
			}
		}
	}
	return tx.Commit(ctx)
}

func (u *substationCellObjectUsecase) Export(projectID uint) (string, error) {
	ctx := context.Background()
	substationCellTempalteFilePath := filepath.Join("./internal/templates/", "Шаблон для импорт Ячеек Подстанции.xlsx")
	f, err := excelize.OpenFile(substationCellTempalteFilePath)
	if err != nil {
		f.Close()
		return "", fmt.Errorf("Не смог открыть файл: %v", err)
	}
	sheetName := "Ячейка Подстанции"
	startingRow := 2

	substationCellCount, err := u.Count(dto.SubstationCellObjectSearchParameters{ProjectID: projectID})
	if err != nil {
		return "", err
	}
	limit := 100
	page := 1

	for substationCellCount > 0 {
		substationCells, err := u.GetPaginated(page, limit, dto.SubstationCellObjectSearchParameters{ProjectID: projectID})
		if err != nil {
			return "", err
		}

		for index, substationCell := range substationCells {
			f.SetCellStr(sheetName, "A"+fmt.Sprint(startingRow+index), substationCell.Name)
			f.SetCellStr(sheetName, "B"+fmt.Sprint(startingRow+index), substationCell.Status)

			supervisorsCombined := ""
			for index, supervisor := range substationCell.Supervisors {
				if index == 0 {
					supervisorsCombined += supervisor
					continue
				}

				supervisorsCombined += ", " + supervisor
			}
			f.SetCellStr(sheetName, "C"+fmt.Sprint(startingRow+index), supervisorsCombined)

			teamNumbersCombined := ""
			for index, teamNumber := range substationCell.Teams {
				if index == 0 {
					teamNumbersCombined += teamNumber
					continue
				}

				teamNumbersCombined += ", " + teamNumber
			}
			f.SetCellStr(sheetName, "D"+fmt.Sprint(startingRow+index), teamNumbersCombined)

			substationName, err := u.q.GetSubstationNameByCellObjectID(ctx, pgInt8(substationCell.ObjectID))
			if errors.Is(err, pgx.ErrNoRows) {
				substationName = ""
			} else if err != nil {
				return "", err
			}
			f.SetCellValue(sheetName, "E"+fmt.Sprint(startingRow+index), substationName)
		}

		startingRow = page*limit + 2
		page++
		substationCellCount -= int64(limit)
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

func (u *substationCellObjectUsecase) GetObjectNamesForSearch(projectID uint) ([]dto.DataForSelect[string], error) {
	rows, err := u.q.ListSubstationCellObjectNamesForSearch(context.Background(), pgInt8(projectID))
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
