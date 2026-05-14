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

type kl04kvObjectUsecase struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

func NewKL04KVObjectUsecase(pool *pgxpool.Pool) IKL04KVObjectUsecase {
	return &kl04kvObjectUsecase{
		pool: pool,
		q:    db.New(pool),
	}
}

type IKL04KVObjectUsecase interface {
	GetPaginated(page, limit int, filter dto.KL04KVObjectSearchParameters) ([]dto.KL04KVObjectPaginated, error)
	Count(filter dto.KL04KVObjectSearchParameters) (int64, error)
	Create(data dto.KL04KVObjectCreate) (model.KL04KV_Object, error)
	Delete(projectID, id uint) error
	Update(data dto.KL04KVObjectCreate) (model.KL04KV_Object, error)
	TemplateFile(filepath string, projectID uint) (string, error)
	Import(projectID uint, filepath string) error
	Export(projectID uint) (string, error)
	GetObjectNamesForSearch(projectID uint) ([]dto.DataForSelect[string], error)
}

func (u *kl04kvObjectUsecase) GetPaginated(page, limit int, filter dto.KL04KVObjectSearchParameters) ([]dto.KL04KVObjectPaginated, error) {
	ctx := context.Background()
	rows, err := u.q.ListKL04KVObjectsPaginated(ctx, db.ListKL04KVObjectsPaginatedParams{
		ProjectID: pgInt8(filter.ProjectID),
		Column2:   filter.ObjectName,
		Column3:   int64(filter.TeamID),
		Column4:   int64(filter.SupervisorWorkerID),
		Column5:   int64(filter.TPObjectID),
		Limit:     int32(limit),
		Offset:    int32((page - 1) * limit),
	})
	if err != nil {
		return []dto.KL04KVObjectPaginated{}, err
	}

	result := []dto.KL04KVObjectPaginated{}
	for _, row := range rows {
		supervisorNames, err := u.q.ListSupervisorNamesByObjectID(ctx, row.ObjectID)
		if err != nil {
			return []dto.KL04KVObjectPaginated{}, err
		}

		teamNumbers, err := u.q.ListTeamNumbersByObjectID(ctx, row.ObjectID)
		if err != nil {
			return []dto.KL04KVObjectPaginated{}, err
		}

		tpNames, err := u.q.ListTPNourashesObjectNamesByTarget(ctx, db.ListTPNourashesObjectNamesByTargetParams{
			TargetID:   pgInt8(uint(row.ObjectID)),
			TargetType: pgText("kl04kv_objects"),
		})
		if err != nil {
			return []dto.KL04KVObjectPaginated{}, err
		}

		result = append(result, dto.KL04KVObjectPaginated{
			ObjectID:         uint(row.ObjectID),
			ObjectDetailedID: uint(row.ObjectDetailedID),
			Name:             row.Name,
			Status:           row.Status,
			Nourashes:        row.Nourashes,
			Length:           float64FromPgNumeric(row.Length),
			Supervisors:      supervisorNames,
			Teams:            teamNumbers,
			TPNames:          tpNames,
		})
	}

	return result, nil
}

func (u *kl04kvObjectUsecase) Count(filter dto.KL04KVObjectSearchParameters) (int64, error) {
	return u.q.CountKL04KVObjectsFiltered(context.Background(), db.CountKL04KVObjectsFilteredParams{
		ProjectID: pgInt8(filter.ProjectID),
		Column2:   filter.ObjectName,
		Column3:   int64(filter.TeamID),
		Column4:   int64(filter.SupervisorWorkerID),
		Column5:   int64(filter.TPObjectID),
	})
}

func (u *kl04kvObjectUsecase) Create(data dto.KL04KVObjectCreate) (model.KL04KV_Object, error) {
	ctx := context.Background()
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return model.KL04KV_Object{}, err
	}
	defer tx.Rollback(ctx)

	qtx := u.q.WithTx(tx)

	detailRow, err := qtx.CreateKL04KVObjectDetail(ctx, db.CreateKL04KVObjectDetailParams{
		Length:    pgNumericFromFloat64(data.DetailedInfo.Length),
		Nourashes: pgText(data.DetailedInfo.Nourashes),
	})
	if err != nil {
		return model.KL04KV_Object{}, err
	}

	objectRow, err := qtx.CreateObject(ctx, db.CreateObjectParams{
		ObjectDetailedID: pgInt8(uint(detailRow.ID)),
		Type:             pgText(data.BaseInfo.Type),
		Name:             pgText(data.BaseInfo.Name),
		Status:           pgText(data.BaseInfo.Status),
		ProjectID:        pgInt8(data.BaseInfo.ProjectID),
	})
	if err != nil {
		return model.KL04KV_Object{}, err
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
			return model.KL04KV_Object{}, err
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
			return model.KL04KV_Object{}, err
		}
	}

	if len(data.NourashedByTPObjectID) != 0 {
		batch := make([]db.CreateTPNourashesObjectsBatchParams, len(data.NourashedByTPObjectID))
		for i, tpObjectID := range data.NourashedByTPObjectID {
			batch[i] = db.CreateTPNourashesObjectsBatchParams{
				TpObjectID: pgInt8(tpObjectID),
				TargetID:   pgInt8(uint(objectRow.ID)),
				TargetType: pgText("kl04kv_objects"),
			}
		}
		if _, err := qtx.CreateTPNourashesObjectsBatch(ctx, batch); err != nil {
			return model.KL04KV_Object{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return model.KL04KV_Object{}, err
	}

	return model.KL04KV_Object{
		ID:        uint(detailRow.ID),
		Length:    float64FromPgNumeric(detailRow.Length),
		Nourashes: detailRow.Nourashes.String,
	}, nil
}

func (u *kl04kvObjectUsecase) Update(data dto.KL04KVObjectCreate) (model.KL04KV_Object, error) {
	ctx := context.Background()
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return model.KL04KV_Object{}, err
	}
	defer tx.Rollback(ctx)

	qtx := u.q.WithTx(tx)

	if err := qtx.UpdateKL04KVObjectDetail(ctx, db.UpdateKL04KVObjectDetailParams{
		ID:        int64(data.BaseInfo.ObjectDetailedID),
		Length:    pgNumericFromFloat64(data.DetailedInfo.Length),
		Nourashes: pgText(data.DetailedInfo.Nourashes),
	}); err != nil {
		return model.KL04KV_Object{}, err
	}

	if _, err := qtx.UpdateObject(ctx, db.UpdateObjectParams{
		ID:               int64(data.BaseInfo.ID),
		ObjectDetailedID: pgInt8(data.BaseInfo.ObjectDetailedID),
		Type:             pgText(data.BaseInfo.Type),
		Name:             pgText(data.BaseInfo.Name),
		Status:           pgText(data.BaseInfo.Status),
		ProjectID:        pgInt8(data.BaseInfo.ProjectID),
	}); err != nil {
		return model.KL04KV_Object{}, err
	}

	if err := qtx.DeleteObjectSupervisorsByObjectID(ctx, pgInt8(data.BaseInfo.ID)); err != nil {
		return model.KL04KV_Object{}, err
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
			return model.KL04KV_Object{}, err
		}
	}

	if err := qtx.DeleteObjectTeamsByObjectID(ctx, pgInt8(data.BaseInfo.ID)); err != nil {
		return model.KL04KV_Object{}, err
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
			return model.KL04KV_Object{}, err
		}
	}

	if err := qtx.DeleteTPNourashesObjectsByTarget(ctx, db.DeleteTPNourashesObjectsByTargetParams{
		TargetID:   pgInt8(data.BaseInfo.ID),
		TargetType: pgText("kl04kv_objects"),
	}); err != nil {
		return model.KL04KV_Object{}, err
	}

	if len(data.NourashedByTPObjectID) != 0 {
		batch := make([]db.CreateTPNourashesObjectsBatchParams, len(data.NourashedByTPObjectID))
		for i, tpObjectID := range data.NourashedByTPObjectID {
			batch[i] = db.CreateTPNourashesObjectsBatchParams{
				TpObjectID: pgInt8(tpObjectID),
				TargetID:   pgInt8(data.BaseInfo.ID),
				TargetType: pgText("kl04kv_objects"),
			}
		}
		if _, err := qtx.CreateTPNourashesObjectsBatch(ctx, batch); err != nil {
			return model.KL04KV_Object{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return model.KL04KV_Object{}, err
	}

	return model.KL04KV_Object{
		ID:        data.BaseInfo.ObjectDetailedID,
		Length:    data.DetailedInfo.Length,
		Nourashes: data.DetailedInfo.Nourashes,
	}, nil
}

func (u *kl04kvObjectUsecase) Delete(projectID, id uint) error {
	ctx := context.Background()
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := u.q.WithTx(tx)
	if err := qtx.DeleteKL04KVObjectSupervisorsCascade(ctx, db.DeleteKL04KVObjectSupervisorsCascadeParams{
		ProjectID: pgInt8(projectID),
		ID:        int64(id),
	}); err != nil {
		return err
	}
	if err := qtx.DeleteKL04KVObjectTeamsCascade(ctx, db.DeleteKL04KVObjectTeamsCascadeParams{
		ProjectID: pgInt8(projectID),
		ID:        int64(id),
	}); err != nil {
		return err
	}
	if err := qtx.DeleteKL04KVObjectTPNourashesCascade(ctx, db.DeleteKL04KVObjectTPNourashesCascadeParams{
		ProjectID: pgInt8(projectID),
		ID:        int64(id),
	}); err != nil {
		return err
	}
	if err := qtx.DeleteKL04KVObjectDetail(ctx, int64(id)); err != nil {
		return err
	}
	if err := qtx.DeleteObjectByKL04KVDetailedID(ctx, pgInt8(id)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (u *kl04kvObjectUsecase) TemplateFile(filePath string, projectID uint) (string, error) {
	ctx := context.Background()
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		f.Close()
		return "", fmt.Errorf("Не смог открыть шаблонный файл: %v", err)
	}

	supervisorSheetName := "Супервайзеры"
	allSupervisors, err := u.q.ListWorkersByJobTitleInProject(ctx, db.ListWorkersByJobTitleInProjectParams{
		JobTitleInProject: pgText("Супервайзер"),
		ProjectID:         pgInt8(projectID),
	})
	if err != nil {
		f.Close()
		return "", fmt.Errorf("Данные супервайзеров недоступны: %v", err)
	}

	for index, supervisor := range allSupervisors {
		f.SetCellStr(supervisorSheetName, "A"+fmt.Sprint(index+2), supervisor.Name.String)
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
		"Шаблон импорта КЛ04КВ - %s.xlsx",
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

func (u *kl04kvObjectUsecase) Import(projectID uint, importFilePath string) error {
	ctx := context.Background()
	f, err := excelize.OpenFile(importFilePath)
	if err != nil {
		f.Close()
		os.Remove(importFilePath)
		return fmt.Errorf("Не смог открыть файл: %v", err)
	}

	sheetName := "КЛ 04 КВ"
	rows, err := f.GetRows(sheetName)
	if err != nil {
		f.Close()
		os.Remove(importFilePath)
		return fmt.Errorf("Не смог найти таблицу 'КЛ 04 КВ': %v", err)
	}

	if len(rows) == 1 {
		f.Close()
		os.Remove(importFilePath)
		return fmt.Errorf("Файл не имеет данных")
	}

	kl04kvs := []dto.KL04KVObjectImportData{}
	index := 1
	for len(rows) > index {
		object := model.Object{
			ProjectID: projectID,
			Type:      "kl04kv_objects",
		}

		kl04kv := model.KL04KV_Object{}

		object.Name, err = f.GetCellValue(sheetName, "A"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(importFilePath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке А%d: %v", index+1, err)
		}
		if object.Name == "" {
			f.Close()
			os.Remove(importFilePath)
			return fmt.Errorf("Ошибка в файле, ячейкa А%d должна иметь данные: %v", index+1, err)
		}

		object.Status, err = f.GetCellValue(sheetName, "B"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(importFilePath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке B%d: %v", index+1, err)
		}
		if object.Status == "" {
			f.Close()
			os.Remove(importFilePath)
			return fmt.Errorf("Ошибка в файле, ячейкa B%d должна иметь данные: %v", index+1, err)
		}

		kl04kv.Nourashes, err = f.GetCellValue(sheetName, "C"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(importFilePath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке C%d: %v", index+1, err)
		}
		if kl04kv.Nourashes == "" {
			f.Close()
			os.Remove(importFilePath)
			return fmt.Errorf("Ошибка в файле, ячейкa C%d должна иметь данные: %v", index+1, err)
		}

		lengthSTR, err := f.GetCellValue(sheetName, "D"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(importFilePath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке D%d: %v", index+1, err)
		}
		if lengthSTR == "" {
			f.Close()
			os.Remove(importFilePath)
			return fmt.Errorf("Ошибка в файле, ячейкa D%d должна иметь данные: %v", index+1, err)
		}

		kl04kv.Length, err = strconv.ParseFloat(lengthSTR, 64)
		if err != nil {
			f.Close()
			os.Remove(importFilePath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке D%d: %v", index+1, err)
		}

		supervisorName, err := f.GetCellValue(sheetName, "E"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(importFilePath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке E%d: %v", index+1, err)
		}
		var supervisorWorkerID uint
		if supervisorName != "" {
			worker, err := u.q.GetWorkerByName(ctx, pgText(supervisorName))
			if err != nil {
				f.Close()
				os.Remove(importFilePath)
				return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке E%d: %v", index+1, err)
			}
			supervisorWorkerID = uint(worker.ID)
		}

		teamNumber, err := f.GetCellValue(sheetName, "F"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(importFilePath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке F%d: %v", index+1, err)
		}
		var teamID uint
		if teamNumber != "" {
			team, err := u.q.GetTeamByNumber(ctx, pgText(teamNumber))
			if err != nil {
				f.Close()
				os.Remove(importFilePath)
				return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке F%d: %v", index+1, err)
			}
			teamID = uint(team.ID)
		}

		tpName, err := f.GetCellValue(sheetName, "G"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(importFilePath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке G%d: %v", index+1, err)
		}
		var tpObjectID uint
		if tpName != "" {
			tpObject, err := u.q.GetObjectByName(ctx, pgText(tpName))
			if err != nil {
				f.Close()
				os.Remove(importFilePath)
				return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке G%d: %v", index+1, err)
			}
			tpObjectID = uint(tpObject.ID)
		}

		kl04kvs = append(kl04kvs, dto.KL04KVObjectImportData{
			Object: object,
			Kl04KV: kl04kv,
			ObjectSupervisors: model.ObjectSupervisors{
				SupervisorWorkerID: supervisorWorkerID,
			},
			ObjectTeam: model.ObjectTeams{
				TeamID: teamID,
			},
			NourashedByTP: model.TPNourashesObjects{
				TP_ObjectID: tpObjectID,
				TargetType:  "kl04kv_objects",
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

	return u.createInBatches(kl04kvs)
}

func (u *kl04kvObjectUsecase) createInBatches(data []dto.KL04KVObjectImportData) error {
	ctx := context.Background()
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := u.q.WithTx(tx)
	for index, row := range data {
		detailRow, err := qtx.CreateKL04KVObjectDetail(ctx, db.CreateKL04KVObjectDetailParams{
			Length:    pgNumericFromFloat64(row.Kl04KV.Length),
			Nourashes: pgText(row.Kl04KV.Nourashes),
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

		data[index].Kl04KV.ID = uint(detailRow.ID)
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

func (u *kl04kvObjectUsecase) Export(projectID uint) (string, error) {
	kl04kvTempalteFilePath := filepath.Join("./internal/templates", "Шаблон для импорта КЛ 04 КВ.xlsx")
	f, err := excelize.OpenFile(kl04kvTempalteFilePath)
	if err != nil {
		f.Close()
		return "", fmt.Errorf("Не смог открыть файл: %v", err)
	}
	sheetName := "КЛ 04 КВ"
	startingRow := 2

	kl04kvCount, err := u.Count(dto.KL04KVObjectSearchParameters{ProjectID: projectID})
	if err != nil {
		return "", err
	}

	limit := 100
	page := 1

	for kl04kvCount > 0 {
		kl04kvs, err := u.GetPaginated(page, limit, dto.KL04KVObjectSearchParameters{ProjectID: projectID})
		if err != nil {
			return "", err
		}

		for index, kl04kv := range kl04kvs {
			f.SetCellStr(sheetName, "A"+fmt.Sprint(startingRow+index), kl04kv.Name)
			f.SetCellStr(sheetName, "B"+fmt.Sprint(startingRow+index), kl04kv.Status)
			f.SetCellStr(sheetName, "C"+fmt.Sprint(startingRow+index), kl04kv.Nourashes)
			f.SetCellFloat(sheetName, "D"+fmt.Sprint(startingRow+index), kl04kv.Length, 2, 64)

			supervisorsCombined := ""
			for index, supervisor := range kl04kv.Supervisors {
				if index == 0 {
					supervisorsCombined += supervisor
					continue
				}

				supervisorsCombined += ", " + supervisor
			}
			f.SetCellStr(sheetName, "E"+fmt.Sprint(startingRow+index), supervisorsCombined)

			teamNumbersCombined := ""
			for index, teamNumber := range kl04kv.Teams {
				if index == 0 {
					teamNumbersCombined += teamNumber
					continue
				}

				teamNumbersCombined += ", " + teamNumber
			}
			f.SetCellStr(sheetName, "F"+fmt.Sprint(startingRow+index), teamNumbersCombined)

			tpNamesCombined := ""
			for index, tpName := range kl04kv.TPNames {
				if index == 0 {
					tpNamesCombined += tpName
					continue
				}

				tpNamesCombined += ", " + tpName
			}
			f.SetCellStr(sheetName, "G"+fmt.Sprint(startingRow+index), tpNamesCombined)
		}

		startingRow = page*limit + 2
		page++
		kl04kvCount -= int64(limit)
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

func (u *kl04kvObjectUsecase) GetObjectNamesForSearch(projectID uint) ([]dto.DataForSelect[string], error) {
	rows, err := u.q.ListKL04KVObjectNamesForSearch(context.Background(), pgInt8(projectID))
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
