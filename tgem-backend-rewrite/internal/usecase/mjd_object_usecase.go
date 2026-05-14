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

type mjdObjectUsecase struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

func NewMJDObjectUsecase(pool *pgxpool.Pool) IMJDObjectUsecase {
	return &mjdObjectUsecase{
		pool: pool,
		q:    db.New(pool),
	}
}

type IMJDObjectUsecase interface {
	GetPaginated(page, limit int, filter dto.MJDObjectSearchParameters) ([]dto.MJDObjectPaginated, error)
	Count(filter dto.MJDObjectSearchParameters) (int64, error)
	Create(data dto.MJDObjectCreate) (model.MJD_Object, error)
	Update(data dto.MJDObjectCreate) (model.MJD_Object, error)
	Delete(id, projectID uint) error
	TemplateFile(filepath string, projectID uint) (string, error)
	Import(projectID uint, filepath string) error
	Export(projectID uint) (string, error)
	GetObjectNamesForSearch(projectID uint) ([]dto.DataForSelect[string], error)
}

func (u *mjdObjectUsecase) GetPaginated(page, limit int, filter dto.MJDObjectSearchParameters) ([]dto.MJDObjectPaginated, error) {
	ctx := context.Background()
	rows, err := u.q.ListMJDObjectsPaginated(ctx, db.ListMJDObjectsPaginatedParams{
		ProjectID: pgInt8(filter.ProjectID),
		Column2:   filter.ObjectName,
		Column3:   int64(filter.TeamID),
		Column4:   int64(filter.SupervisorWorkerID),
		Column5:   int64(filter.TPObjectID),
		Limit:     int32(limit),
		Offset:    int32((page - 1) * limit),
	})
	if err != nil {
		return []dto.MJDObjectPaginated{}, err
	}

	result := []dto.MJDObjectPaginated{}
	for _, row := range rows {
		supervisorNames, err := u.q.ListSupervisorNamesByObjectID(ctx, row.ObjectID)
		if err != nil {
			return []dto.MJDObjectPaginated{}, err
		}

		teamNumbers, err := u.q.ListTeamNumbersByObjectID(ctx, row.ObjectID)
		if err != nil {
			return []dto.MJDObjectPaginated{}, err
		}

		tpNames, err := u.q.ListTPNourashesObjectNamesByTarget(ctx, db.ListTPNourashesObjectNamesByTargetParams{
			TargetID:   pgInt8(uint(row.ObjectID)),
			TargetType: pgText("mjd_objects"),
		})
		if err != nil {
			return []dto.MJDObjectPaginated{}, err
		}

		result = append(result, dto.MJDObjectPaginated{
			ObjectID:         uint(row.ObjectID),
			ObjectDetailedID: uint(row.ObjectDetailedID),
			Name:             row.Name,
			Status:           row.Status,
			Model:            row.Model,
			AmountStores:     uint(row.AmountStores),
			AmountEntrances:  uint(row.AmountEntrances),
			HasBasement:      row.HasBasement,
			Supervisors:      supervisorNames,
			Teams:            teamNumbers,
			TPNames:          tpNames,
		})
	}
	return result, nil
}

func (u *mjdObjectUsecase) Count(filter dto.MJDObjectSearchParameters) (int64, error) {
	return u.q.CountMJDObjectsFiltered(context.Background(), db.CountMJDObjectsFilteredParams{
		ProjectID: pgInt8(filter.ProjectID),
		Column2:   filter.ObjectName,
		Column3:   int64(filter.TeamID),
		Column4:   int64(filter.SupervisorWorkerID),
		Column5:   int64(filter.TPObjectID),
	})
}

func (u *mjdObjectUsecase) Create(data dto.MJDObjectCreate) (model.MJD_Object, error) {
	ctx := context.Background()
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return model.MJD_Object{}, err
	}
	defer tx.Rollback(ctx)

	qtx := u.q.WithTx(tx)

	detailRow, err := qtx.CreateMJDObjectDetail(ctx, db.CreateMJDObjectDetailParams{
		Model:           pgText(data.DetailedInfo.Model),
		AmountStores:    pgInt8(data.DetailedInfo.AmountStores),
		AmountEntrances: pgInt8(data.DetailedInfo.AmountEntrances),
		HasBasement:     pgBool(data.DetailedInfo.HasBasement),
	})
	if err != nil {
		return model.MJD_Object{}, err
	}

	objectRow, err := qtx.CreateObject(ctx, db.CreateObjectParams{
		ObjectDetailedID: pgInt8(uint(detailRow.ID)),
		Type:             pgText("mjd_objects"),
		Name:             pgText(data.BaseInfo.Name),
		Status:           pgText(data.BaseInfo.Status),
		ProjectID:        pgInt8(data.BaseInfo.ProjectID),
	})
	if err != nil {
		return model.MJD_Object{}, err
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
			return model.MJD_Object{}, err
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
			return model.MJD_Object{}, err
		}
	}

	if len(data.NourashedByTPObjectID) != 0 {
		batch := make([]db.CreateTPNourashesObjectsBatchParams, len(data.NourashedByTPObjectID))
		for i, tpObjectID := range data.NourashedByTPObjectID {
			batch[i] = db.CreateTPNourashesObjectsBatchParams{
				TpObjectID: pgInt8(tpObjectID),
				TargetID:   pgInt8(uint(objectRow.ID)),
				TargetType: pgText("mjd_objects"),
			}
		}
		if _, err := qtx.CreateTPNourashesObjectsBatch(ctx, batch); err != nil {
			return model.MJD_Object{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return model.MJD_Object{}, err
	}

	return model.MJD_Object{
		ID:              uint(detailRow.ID),
		Model:           detailRow.Model.String,
		AmountStores:    uint(detailRow.AmountStores.Int64),
		AmountEntrances: uint(detailRow.AmountEntrances.Int64),
		HasBasement:     detailRow.HasBasement.Bool,
	}, nil
}

func (u *mjdObjectUsecase) Update(data dto.MJDObjectCreate) (model.MJD_Object, error) {
	ctx := context.Background()
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return model.MJD_Object{}, err
	}
	defer tx.Rollback(ctx)

	qtx := u.q.WithTx(tx)

	if err := qtx.UpdateMJDObjectDetail(ctx, db.UpdateMJDObjectDetailParams{
		ID:              int64(data.BaseInfo.ObjectDetailedID),
		Model:           pgText(data.DetailedInfo.Model),
		AmountStores:    pgInt8(data.DetailedInfo.AmountStores),
		AmountEntrances: pgInt8(data.DetailedInfo.AmountEntrances),
		HasBasement:     pgBool(data.DetailedInfo.HasBasement),
	}); err != nil {
		return model.MJD_Object{}, err
	}

	if _, err := qtx.UpdateObject(ctx, db.UpdateObjectParams{
		ID:               int64(data.BaseInfo.ID),
		ObjectDetailedID: pgInt8(data.BaseInfo.ObjectDetailedID),
		Type:             pgText(data.BaseInfo.Type),
		Name:             pgText(data.BaseInfo.Name),
		Status:           pgText(data.BaseInfo.Status),
		ProjectID:        pgInt8(data.BaseInfo.ProjectID),
	}); err != nil {
		return model.MJD_Object{}, err
	}

	if err := qtx.DeleteObjectSupervisorsByObjectID(ctx, pgInt8(data.BaseInfo.ID)); err != nil {
		return model.MJD_Object{}, err
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
			return model.MJD_Object{}, err
		}
	}

	if err := qtx.DeleteObjectTeamsByObjectID(ctx, pgInt8(data.BaseInfo.ID)); err != nil {
		return model.MJD_Object{}, err
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
			return model.MJD_Object{}, err
		}
	}

	if err := qtx.DeleteTPNourashesObjectsByTarget(ctx, db.DeleteTPNourashesObjectsByTargetParams{
		TargetID:   pgInt8(data.BaseInfo.ID),
		TargetType: pgText("mjd_objects"),
	}); err != nil {
		return model.MJD_Object{}, err
	}

	if len(data.NourashedByTPObjectID) != 0 {
		batch := make([]db.CreateTPNourashesObjectsBatchParams, len(data.NourashedByTPObjectID))
		for i, tpObjectID := range data.NourashedByTPObjectID {
			batch[i] = db.CreateTPNourashesObjectsBatchParams{
				TpObjectID: pgInt8(tpObjectID),
				TargetID:   pgInt8(data.BaseInfo.ID),
				TargetType: pgText("mjd_objects"),
			}
		}
		if _, err := qtx.CreateTPNourashesObjectsBatch(ctx, batch); err != nil {
			return model.MJD_Object{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return model.MJD_Object{}, err
	}

	return model.MJD_Object{
		ID:              data.BaseInfo.ObjectDetailedID,
		Model:           data.DetailedInfo.Model,
		AmountStores:    data.DetailedInfo.AmountStores,
		AmountEntrances: data.DetailedInfo.AmountEntrances,
		HasBasement:     data.DetailedInfo.HasBasement,
	}, nil
}

func (u *mjdObjectUsecase) Delete(id, projectID uint) error {
	ctx := context.Background()
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := u.q.WithTx(tx)
	if err := qtx.DeleteMJDObjectSupervisorsCascade(ctx, db.DeleteMJDObjectSupervisorsCascadeParams{
		ProjectID: pgInt8(projectID),
		ID:        int64(id),
	}); err != nil {
		return err
	}
	if err := qtx.DeleteMJDObjectTeamsCascade(ctx, db.DeleteMJDObjectTeamsCascadeParams{
		ProjectID: pgInt8(projectID),
		ID:        int64(id),
	}); err != nil {
		return err
	}
	if err := qtx.DeleteMJDObjectTPNourashesCascade(ctx, db.DeleteMJDObjectTPNourashesCascadeParams{
		ProjectID: pgInt8(projectID),
		ID:        int64(id),
	}); err != nil {
		return err
	}
	if err := qtx.DeleteMJDObjectDetail(ctx, int64(id)); err != nil {
		return err
	}
	if err := qtx.DeleteObjectByMJDDetailedID(ctx, pgInt8(id)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (u *mjdObjectUsecase) TemplateFile(filePath string, projectID uint) (string, error) {
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

	allTPNames, err := u.q.ListTPObjectNamesByProject(ctx, pgInt8(projectID))
	if err != nil {
		f.Close()
		return "", fmt.Errorf("Данны бригад не доступны: %v", err)
	}

	tpObjectSheetName := "ТП"
	for index, tpName := range allTPNames {
		f.SetCellStr(tpObjectSheetName, "A"+fmt.Sprint(index+2), tpName)
	}

	currentTime := time.Now()
	temporaryFileName := fmt.Sprintf(
		"Шаблон импорта МЖД - %s.xlsx",
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

func (u *mjdObjectUsecase) Import(projectID uint, importFilePath string) error {
	ctx := context.Background()
	f, err := excelize.OpenFile(importFilePath)
	if err != nil {
		f.Close()
		os.Remove(importFilePath)
		return fmt.Errorf("Не смог открыть файл: %v", err)
	}

	sheetName := "МЖД"
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

	mjds := []dto.MJDObjectImportData{}
	index := 1
	for len(rows) > index {
		object := model.Object{
			ProjectID: projectID,
			Type:      "mjd_objects",
		}
		mjd := model.MJD_Object{}

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

		mjd.Model, err = f.GetCellValue(sheetName, "C"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(importFilePath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке C%d: %v", index+1, err)
		}

		amountEntrancesSTR, err := f.GetCellValue(sheetName, "D"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(importFilePath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке D%d: %v", index+1, err)
		}

		amountEntrancesUINT64, err := strconv.ParseUint(amountEntrancesSTR, 10, 64)
		if err != nil {
			f.Close()
			os.Remove(importFilePath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке D%d: %v", index+1, err)
		}
		mjd.AmountEntrances = uint(amountEntrancesUINT64)

		amountStoresSTR, err := f.GetCellValue(sheetName, "E"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(importFilePath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке E%d: %v", index+1, err)
		}

		amountStoresUINT64, err := strconv.ParseUint(amountStoresSTR, 10, 64)
		if err != nil {
			f.Close()
			os.Remove(importFilePath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке E%d: %v", index+1, err)
		}
		mjd.AmountStores = uint(amountStoresUINT64)

		hasBasement, err := f.GetCellValue(sheetName, "F"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(importFilePath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке F%d: %v", index+1, err)
		}

		if hasBasement == "Да" {
			mjd.HasBasement = true
		} else {
			mjd.HasBasement = false
		}

		supervisorName, err := f.GetCellValue(sheetName, "G"+fmt.Sprint(index+1))
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

		teamNumber, err := f.GetCellValue(sheetName, "H"+fmt.Sprint(index+1))
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

		tpName, err := f.GetCellValue(sheetName, "I"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(importFilePath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке I%d: %v", index+1, err)
		}
		var tpObjectID uint
		if tpName != "" {
			tpObject, err := u.q.GetObjectByName(ctx, pgText(tpName))
			if err != nil {
				f.Close()
				os.Remove(importFilePath)
				return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке I%d: %v", index+1, err)
			}
			tpObjectID = uint(tpObject.ID)
		}

		mjds = append(mjds, dto.MJDObjectImportData{
			Object: object,
			MJD:    mjd,
			ObjectSupervisors: model.ObjectSupervisors{
				SupervisorWorkerID: supervisorWorkerID,
			},
			ObjectTeam: model.ObjectTeams{
				TeamID: teamID,
			},
			NourashedByTP: model.TPNourashesObjects{
				TP_ObjectID: tpObjectID,
				TargetType:  "mjd_objects",
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

	return u.createInBatches(mjds)
}

func (u *mjdObjectUsecase) createInBatches(data []dto.MJDObjectImportData) error {
	ctx := context.Background()
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := u.q.WithTx(tx)
	for index, row := range data {
		detailRow, err := qtx.CreateMJDObjectDetail(ctx, db.CreateMJDObjectDetailParams{
			Model:           pgText(row.MJD.Model),
			AmountStores:    pgInt8(row.MJD.AmountStores),
			AmountEntrances: pgInt8(row.MJD.AmountEntrances),
			HasBasement:     pgBool(row.MJD.HasBasement),
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

		data[index].MJD.ID = uint(detailRow.ID)
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

		if row.NourashedByTP.TP_ObjectID != 0 {
			if _, err := qtx.CreateTPNourashesObjectsBatch(ctx, []db.CreateTPNourashesObjectsBatchParams{{
				TpObjectID: pgInt8(row.NourashedByTP.TP_ObjectID),
				TargetID:   pgInt8(uint(objectRow.ID)),
				TargetType: pgText(row.NourashedByTP.TargetType),
			}}); err != nil {
				return err
			}
		}
	}
	return tx.Commit(ctx)
}

func (u *mjdObjectUsecase) Export(projectID uint) (string, error) {
	mjdTempalteFilePath := filepath.Join("./internal/templates/", "Шаблон для импорта МЖД.xlsx")
	f, err := excelize.OpenFile(mjdTempalteFilePath)
	if err != nil {
		f.Close()
		return "", fmt.Errorf("Не смог открыть файл: %v", err)
	}
	sheetName := "МЖД"
	startingRow := 2

	mjdCount, err := u.Count(dto.MJDObjectSearchParameters{ProjectID: projectID})
	if err != nil {
		return "", err
	}
	limit := 100
	page := 1

	for mjdCount > 0 {
		mjds, err := u.GetPaginated(page, limit, dto.MJDObjectSearchParameters{ProjectID: projectID})
		if err != nil {
			return "", err
		}

		for index, mjd := range mjds {
			f.SetCellStr(sheetName, "A"+fmt.Sprint(startingRow+index), mjd.Name)
			f.SetCellStr(sheetName, "B"+fmt.Sprint(startingRow+index), mjd.Status)
			f.SetCellStr(sheetName, "C"+fmt.Sprint(startingRow+index), mjd.Model)
			f.SetCellInt(sheetName, "D"+fmt.Sprint(startingRow+index), int(mjd.AmountEntrances))
			f.SetCellInt(sheetName, "E"+fmt.Sprint(startingRow+index), int(mjd.AmountStores))

			hasBasement := "Да"
			if !mjd.HasBasement {
				hasBasement = "Нет"
			}
			f.SetCellStr(sheetName, "F"+fmt.Sprint(startingRow+index), hasBasement)

			supervisorsCombined := ""
			for index, supervisor := range mjd.Supervisors {
				if index == 0 {
					supervisorsCombined += supervisor
					continue
				}

				supervisorsCombined += ", " + supervisor
			}
			f.SetCellStr(sheetName, "G"+fmt.Sprint(startingRow+index), supervisorsCombined)

			teamNumbersCombined := ""
			for index, teamNumber := range mjd.Teams {
				if index == 0 {
					teamNumbersCombined += teamNumber
					continue
				}

				teamNumbersCombined += ", " + teamNumber
			}
			f.SetCellStr(sheetName, "H"+fmt.Sprint(startingRow+index), teamNumbersCombined)

			tpNamesCombined := ""
			for index, tpName := range mjd.TPNames {
				if index == 0 {
					tpNamesCombined += tpName
					continue
				}

				tpNamesCombined += ", " + tpName
			}
			f.SetCellStr(sheetName, "I"+fmt.Sprint(startingRow+index), tpNamesCombined)
		}

		startingRow = page*limit + 2
		page++
		mjdCount -= int64(limit)
	}

	currentTime := time.Now()
	exportFileName := fmt.Sprintf(
		"Экспорт СИП - %s.xlsx",
		currentTime.Format("02-01-2006"),
	)
	exportFilePath := filepath.Join("./storage/import_excel/temp/", exportFileName)
	if err := f.SaveAs(exportFilePath); err != nil {
		return "", err
	}

	return exportFileName, nil
}

func (u *mjdObjectUsecase) GetObjectNamesForSearch(projectID uint) ([]dto.DataForSelect[string], error) {
	rows, err := u.q.ListMJDObjectNamesForSearch(context.Background(), pgInt8(projectID))
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
