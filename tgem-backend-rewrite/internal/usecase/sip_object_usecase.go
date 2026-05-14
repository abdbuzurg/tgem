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

type sipObjectUsecase struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

func NewSIPObjectUsecase(pool *pgxpool.Pool) ISIPObjectUsecase {
	return &sipObjectUsecase{
		pool: pool,
		q:    db.New(pool),
	}
}

type ISIPObjectUsecase interface {
	GetPaginated(page, limit int, filter dto.SIPObjectSearchParameters) ([]dto.SIPObjectPaginated, error)
	Count(filter dto.SIPObjectSearchParameters) (int64, error)
	Create(data dto.SIPObjectCreate) (model.SIP_Object, error)
	Update(data dto.SIPObjectCreate) (model.SIP_Object, error)
	Delete(id, projectID uint) error
	TemplateFile(filepath string, projectID uint) (string, error)
	Import(projectID uint, filepath string) error
	GetTPNames(projectID uint) ([]string, error)
	GetObjectNamesForSearch(projectID uint) ([]dto.DataForSelect[string], error)
	Export(projectID uint) (string, error)
}

func (u *sipObjectUsecase) GetPaginated(page, limit int, filter dto.SIPObjectSearchParameters) ([]dto.SIPObjectPaginated, error) {
	ctx := context.Background()
	rows, err := u.q.ListSIPObjectsPaginated(ctx, db.ListSIPObjectsPaginatedParams{
		ProjectID: pgInt8(filter.ProjectID),
		Column2:   filter.ObjectName,
		Column3:   int64(filter.TeamID),
		Column4:   int64(filter.SupervisorWorkerID),
		Limit:     int32(limit),
		Offset:    int32((page - 1) * limit),
	})
	if err != nil {
		return []dto.SIPObjectPaginated{}, err
	}

	result := []dto.SIPObjectPaginated{}
	for _, row := range rows {
		// Preserves GORM-era de-duplication: if the previous row in the
		// FULL JOIN result has the same ObjectID, skip it. The DISTINCT
		// already handles distinct columns at the SQL level, but the
		// GORM-era usecase guards against repeats here too.
		if len(result) != 0 {
			if uint(row.ObjectID) == result[len(result)-1].ObjectID {
				continue
			}
		}

		supervisorNames, err := u.q.ListSupervisorNamesByObjectID(ctx, row.ObjectID)
		if err != nil {
			return []dto.SIPObjectPaginated{}, err
		}

		teamNumbers, err := u.q.ListTeamNumbersByObjectID(ctx, row.ObjectID)
		if err != nil {
			return []dto.SIPObjectPaginated{}, err
		}

		tpNames, err := u.q.ListTPNourashesObjectNamesByTarget(ctx, db.ListTPNourashesObjectNamesByTargetParams{
			TargetID:   pgInt8(uint(row.ObjectID)),
			TargetType: pgText("sip_objects"),
		})
		if err != nil {
			return []dto.SIPObjectPaginated{}, err
		}

		result = append(result, dto.SIPObjectPaginated{
			ObjectID:         uint(row.ObjectID),
			ObjectDetailedID: uint(row.ObjectDetailedID),
			Name:             row.Name,
			Status:           row.Status,
			AmountFeeders:    uint(row.AmountFeeders),
			Supervisors:      supervisorNames,
			Teams:            teamNumbers,
			TPNames:          tpNames,
		})
	}

	return result, nil
}

func (u *sipObjectUsecase) Count(filter dto.SIPObjectSearchParameters) (int64, error) {
	return u.q.CountSIPObjectsFiltered(context.Background(), db.CountSIPObjectsFilteredParams{
		ProjectID: pgInt8(filter.ProjectID),
		Column2:   filter.ObjectName,
		Column3:   int64(filter.TeamID),
		Column4:   int64(filter.SupervisorWorkerID),
	})
}

func (u *sipObjectUsecase) Create(data dto.SIPObjectCreate) (model.SIP_Object, error) {
	ctx := context.Background()
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return model.SIP_Object{}, err
	}
	defer tx.Rollback(ctx)

	qtx := u.q.WithTx(tx)

	detailRow, err := qtx.CreateSIPObjectDetail(ctx, pgInt8(data.DetailedInfo.AmountFeeders))
	if err != nil {
		return model.SIP_Object{}, err
	}

	objectRow, err := qtx.CreateObject(ctx, db.CreateObjectParams{
		ObjectDetailedID: pgInt8(uint(detailRow.ID)),
		Type:             pgText("sip_objects"),
		Name:             pgText(data.BaseInfo.Name),
		Status:           pgText(data.BaseInfo.Status),
		ProjectID:        pgInt8(data.BaseInfo.ProjectID),
	})
	if err != nil {
		return model.SIP_Object{}, err
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
			return model.SIP_Object{}, err
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
			return model.SIP_Object{}, err
		}
	}

	if len(data.NourashedByTPObjectID) != 0 {
		batch := make([]db.CreateTPNourashesObjectsBatchParams, len(data.NourashedByTPObjectID))
		for i, tpObjectID := range data.NourashedByTPObjectID {
			batch[i] = db.CreateTPNourashesObjectsBatchParams{
				TpObjectID: pgInt8(tpObjectID),
				TargetID:   pgInt8(uint(objectRow.ID)),
				TargetType: pgText("sip_objects"),
			}
		}
		if _, err := qtx.CreateTPNourashesObjectsBatch(ctx, batch); err != nil {
			return model.SIP_Object{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return model.SIP_Object{}, err
	}

	return model.SIP_Object{
		ID:            uint(detailRow.ID),
		AmountFeeders: uint(detailRow.AmountFeeders.Int64),
	}, nil
}

func (u *sipObjectUsecase) Update(data dto.SIPObjectCreate) (model.SIP_Object, error) {
	ctx := context.Background()
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return model.SIP_Object{}, err
	}
	defer tx.Rollback(ctx)

	qtx := u.q.WithTx(tx)

	if err := qtx.UpdateSIPObjectDetail(ctx, db.UpdateSIPObjectDetailParams{
		ID:            int64(data.BaseInfo.ObjectDetailedID),
		AmountFeeders: pgInt8(data.DetailedInfo.AmountFeeders),
	}); err != nil {
		return model.SIP_Object{}, err
	}

	if _, err := qtx.UpdateObject(ctx, db.UpdateObjectParams{
		ID:               int64(data.BaseInfo.ID),
		ObjectDetailedID: pgInt8(data.BaseInfo.ObjectDetailedID),
		Type:             pgText(data.BaseInfo.Type),
		Name:             pgText(data.BaseInfo.Name),
		Status:           pgText(data.BaseInfo.Status),
		ProjectID:        pgInt8(data.BaseInfo.ProjectID),
	}); err != nil {
		return model.SIP_Object{}, err
	}

	if err := qtx.DeleteObjectSupervisorsByObjectID(ctx, pgInt8(data.BaseInfo.ID)); err != nil {
		return model.SIP_Object{}, err
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
			return model.SIP_Object{}, err
		}
	}

	if err := qtx.DeleteObjectTeamsByObjectID(ctx, pgInt8(data.BaseInfo.ID)); err != nil {
		return model.SIP_Object{}, err
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
			return model.SIP_Object{}, err
		}
	}

	if err := qtx.DeleteTPNourashesObjectsByTarget(ctx, db.DeleteTPNourashesObjectsByTargetParams{
		TargetID:   pgInt8(data.BaseInfo.ID),
		TargetType: pgText("sip_objects"),
	}); err != nil {
		return model.SIP_Object{}, err
	}

	if len(data.NourashedByTPObjectID) != 0 {
		batch := make([]db.CreateTPNourashesObjectsBatchParams, len(data.NourashedByTPObjectID))
		for i, tpObjectID := range data.NourashedByTPObjectID {
			batch[i] = db.CreateTPNourashesObjectsBatchParams{
				TpObjectID: pgInt8(tpObjectID),
				TargetID:   pgInt8(data.BaseInfo.ID),
				TargetType: pgText("sip_objects"),
			}
		}
		if _, err := qtx.CreateTPNourashesObjectsBatch(ctx, batch); err != nil {
			return model.SIP_Object{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return model.SIP_Object{}, err
	}

	return model.SIP_Object{
		ID:            data.BaseInfo.ObjectDetailedID,
		AmountFeeders: data.DetailedInfo.AmountFeeders,
	}, nil
}

func (u *sipObjectUsecase) Delete(id, projectID uint) error {
	ctx := context.Background()
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := u.q.WithTx(tx)
	if err := qtx.DeleteSIPObjectSupervisorsCascade(ctx, db.DeleteSIPObjectSupervisorsCascadeParams{
		ProjectID: pgInt8(projectID),
		ID:        int64(id),
	}); err != nil {
		return err
	}
	if err := qtx.DeleteSIPObjectTeamsCascade(ctx, db.DeleteSIPObjectTeamsCascadeParams{
		ProjectID: pgInt8(projectID),
		ID:        int64(id),
	}); err != nil {
		return err
	}
	if err := qtx.DeleteTPNourashesObjectsByTarget(ctx, db.DeleteTPNourashesObjectsByTargetParams{
		TargetID:   pgInt8(id),
		TargetType: pgText("sip_objects"),
	}); err != nil {
		return err
	}
	if err := qtx.DeleteSIPObjectDetail(ctx, int64(id)); err != nil {
		return err
	}
	if err := qtx.DeleteObjectBySIPDetailedID(ctx, pgInt8(id)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (u *sipObjectUsecase) TemplateFile(filePath string, projectID uint) (string, error) {
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
	fileName := fmt.Sprintf(
		"Шаблон для Импорта СИП - %s.xlsx",
		currentTime.Format("02-01-2006"),
	)
	tmpFilePath := filepath.Join("./storage/import_excel/temp/", fileName)
	if err := f.SaveAs(tmpFilePath); err != nil {
		return "", fmt.Errorf("Не удалось обновить шаблон с новыми данными: %v", err)
	}

	if err := f.Close(); err != nil {
		return "", err
	}

	return tmpFilePath, nil
}

func (u *sipObjectUsecase) Import(projectID uint, importFilePath string) error {
	ctx := context.Background()
	f, err := excelize.OpenFile(importFilePath)
	if err != nil {
		f.Close()
		os.Remove(importFilePath)
		return fmt.Errorf("Не смог открыть файл: %v", err)
	}

	sheetName := "СИП"
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

	sips := []dto.SIPObjectImportData{}
	index := 1
	for len(rows) > index {
		object := model.Object{
			ProjectID: projectID,
			Type:      "sip_objects",
		}

		sip := model.SIP_Object{}

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

		amountFeedersSTR, err := f.GetCellValue(sheetName, "C"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(importFilePath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке C%d: %v", index+1, err)
		}

		amountFeedersUINT64, err := strconv.ParseUint(amountFeedersSTR, 10, 64)
		if err != nil {
			f.Close()
			os.Remove(importFilePath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке C%d: %v", index+1, err)
		}
		sip.AmountFeeders = uint(amountFeedersUINT64)

		supervisorName, err := f.GetCellValue(sheetName, "D"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(importFilePath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке D%d: %v", index+1, err)
		}

		var supervisorWorkerID uint
		if supervisorName != "" {
			worker, err := u.q.GetWorkerByName(ctx, pgText(supervisorName))
			if err != nil {
				f.Close()
				os.Remove(importFilePath)
				return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке D%d: %v", index+1, err)
			}
			supervisorWorkerID = uint(worker.ID)
		}

		teamNumber, err := f.GetCellValue(sheetName, "E"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(importFilePath)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке E%d: %v", index+1, err)
		}

		var teamID uint
		if teamNumber != "" {
			team, err := u.q.GetTeamByNumber(ctx, pgText(teamNumber))
			if err != nil {
				f.Close()
				os.Remove(importFilePath)
				return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке E%d: %v", index+1, err)
			}
			teamID = uint(team.ID)
		}

		tpName, err := f.GetCellValue(sheetName, "F"+fmt.Sprint(index+1))
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

		sips = append(sips, dto.SIPObjectImportData{
			Object: object,
			SIP:    sip,
			ObjectSupervisors: model.ObjectSupervisors{
				SupervisorWorkerID: supervisorWorkerID,
			},
			ObjectTeam: model.ObjectTeams{
				TeamID: teamID,
			},
			NourashedByTP: model.TPNourashesObjects{
				TP_ObjectID: tpObjectID,
				TargetType:  "sip_objects",
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

	return u.createInBatches(sips)
}

func (u *sipObjectUsecase) createInBatches(data []dto.SIPObjectImportData) error {
	ctx := context.Background()
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := u.q.WithTx(tx)
	for index, row := range data {
		detailRow, err := qtx.CreateSIPObjectDetail(ctx, pgInt8(row.SIP.AmountFeeders))
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

		data[index].SIP.ID = uint(detailRow.ID)
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

func (u *sipObjectUsecase) GetTPNames(projectID uint) ([]string, error) {
	return u.q.ListTPObjectNamesByProject(context.Background(), pgInt8(projectID))
}

func (u *sipObjectUsecase) GetObjectNamesForSearch(projectID uint) ([]dto.DataForSelect[string], error) {
	rows, err := u.q.ListSIPObjectNamesForSearch(context.Background(), pgInt8(projectID))
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

func (u *sipObjectUsecase) Export(projectID uint) (string, error) {
	sipTempalteFilePath := filepath.Join("./internal/templates/", "Шаблон для импорта СИП.xlsx")
	f, err := excelize.OpenFile(sipTempalteFilePath)
	if err != nil {
		f.Close()
		return "", fmt.Errorf("Не смог открыть файл: %v", err)
	}
	sheetName := "СИП"
	startingRow := 2

	sipCount, err := u.Count(dto.SIPObjectSearchParameters{ProjectID: projectID})
	if err != nil {
		return "", err
	}
	limit := 100
	page := 1

	for sipCount > 0 {
		sips, err := u.GetPaginated(page, limit, dto.SIPObjectSearchParameters{ProjectID: projectID})
		if err != nil {
			return "", err
		}

		for index, sip := range sips {
			f.SetCellStr(sheetName, "A"+fmt.Sprint(startingRow+index), sip.Name)
			f.SetCellStr(sheetName, "B"+fmt.Sprint(startingRow+index), sip.Status)
			f.SetCellUint(sheetName, "C"+fmt.Sprint(startingRow+index), uint64(sip.AmountFeeders))

			supervisorsCombined := ""
			for index, supervisor := range sip.Supervisors {
				if index == 0 {
					supervisorsCombined += supervisor
					continue
				}
				supervisorsCombined += ", " + supervisor
			}
			f.SetCellStr(sheetName, "D"+fmt.Sprint(startingRow+index), supervisorsCombined)

			teamNumbersCombined := ""
			for index, teamNumber := range sip.Teams {
				if index == 0 {
					teamNumbersCombined += teamNumber
					continue
				}

				teamNumbersCombined += ", " + teamNumber
			}
			f.SetCellStr(sheetName, "E"+fmt.Sprint(startingRow+index), teamNumbersCombined)
		}

		startingRow = page*limit + 2
		page++
		sipCount -= int64(limit)
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

	if err := f.Close(); err != nil {
		return "", err
	}

	return exportFileName, nil
}
