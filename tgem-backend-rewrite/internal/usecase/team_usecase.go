package usecase

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"backend-v2/internal/db"
	"backend-v2/internal/dto"
	"backend-v2/internal/utils"
	"backend-v2/model"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xuri/excelize/v2"
)

type teamUsecase struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

func NewTeamUsecase(pool *pgxpool.Pool) ITeamUsecase {
	return &teamUsecase{
		pool: pool,
		q:    db.New(pool),
	}
}

type ITeamUsecase interface {
	GetAll(projectID uint) ([]model.Team, error)
	GetPaginated(page, limit int, searchParameters dto.TeamSearchParameters) ([]dto.TeamPaginated, error)
	GetByID(id uint) (model.Team, error)
	Create(data dto.TeamMutation) (model.Team, error)
	Update(data dto.TeamMutation) (model.Team, error)
	Delete(id uint) error
	Count(searchParameters dto.TeamSearchParameters) (int64, error)
	TemplateFile(projectID uint, filepath string) (string, error)
	Import(projectID uint, filepath string) error
	DoesTeamNumberAlreadyExistForCreate(teamNumber string, projectID uint) (bool, error)
	DoesTeamNumberAlreadyExistForUpdate(teamNumber string, id uint, projectID uint) (bool, error)
	GetAllForSelect(projectID uint) ([]dto.TeamDataForSelect, error)
	GetAllUniqueTeamNumbers(projectID uint) ([]string, error)
	GetAllUniqueMobileNumber(projectID uint) ([]string, error)
	GetAllUniqueCompanies(projectID uint) ([]string, error)
	Export(projectID uint) (string, error)
}

func (u *teamUsecase) GetAll(projectID uint) ([]model.Team, error) {
	rows, err := u.q.ListTeamsByProject(context.Background(), pgInt8(projectID))
	if err != nil {
		return nil, err
	}
	out := make([]model.Team, len(rows))
	for i, r := range rows {
		out[i] = toModelTeam(r)
	}
	return out, nil
}

func (u *teamUsecase) GetPaginated(page, limit int, searchParameters dto.TeamSearchParameters) ([]dto.TeamPaginated, error) {
	rows, err := u.q.ListTeamsPaginated(context.Background(), db.ListTeamsPaginatedParams{
		ProjectID: pgInt8(searchParameters.ProjectID),
		Column2:   searchParameters.Number,
		Column3:   searchParameters.MobileNumber,
		Column4:   searchParameters.Company,
		Column5:   int64(searchParameters.TeamLeaderID),
		Limit:     int32(limit),
		Offset:    int32((page - 1) * limit),
	})
	if err != nil {
		return []dto.TeamPaginated{}, err
	}

	result := []dto.TeamPaginated{}
	latestEntry := dto.TeamPaginated{}
	for index, team := range rows {
		if latestEntry.ID == uint(team.ID) {
			if !utils.DoesExist(latestEntry.LeaderNames, team.LeaderName) {
				latestEntry.LeaderNames = append(latestEntry.LeaderNames, team.LeaderName)
			}
		} else {
			if index != 0 {
				result = append(result, latestEntry)
			}
			latestEntry = dto.TeamPaginated{
				ID:           uint(team.ID),
				Number:       team.TeamNumber,
				MobileNumber: team.TeamMobileNumber,
				Company:      team.TeamCompany,
				LeaderNames:  []string{team.LeaderName},
			}
		}
	}

	if len(rows) > 0 {
		result = append(result, latestEntry)
	}

	return result, nil
}

func (u *teamUsecase) GetByID(id uint) (model.Team, error) {
	row, err := u.q.GetTeam(context.Background(), int64(id))
	if err != nil {
		return model.Team{}, err
	}
	return toModelTeam(row), nil
}

func (u *teamUsecase) Create(data dto.TeamMutation) (model.Team, error) {
	ctx := context.Background()
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return model.Team{}, err
	}
	defer tx.Rollback(ctx)

	qtx := u.q.WithTx(tx)
	row, err := qtx.CreateTeam(ctx, db.CreateTeamParams{
		ProjectID:    pgInt8(data.ProjectID),
		Number:       pgText(data.Number),
		MobileNumber: pgText(data.MobileNumber),
		Company:      pgText(data.Company),
	})
	if err != nil {
		return model.Team{}, err
	}

	if err := u.insertTeamLeaders(ctx, qtx, uint(row.ID), data.LeaderWorkerIDs); err != nil {
		return model.Team{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return model.Team{}, err
	}
	return toModelTeam(row), nil
}

func (u *teamUsecase) Update(data dto.TeamMutation) (model.Team, error) {
	ctx := context.Background()
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return model.Team{}, err
	}
	defer tx.Rollback(ctx)

	qtx := u.q.WithTx(tx)
	if err := qtx.UpdateTeam(ctx, db.UpdateTeamParams{
		ID:           int64(data.ID),
		Number:       pgText(data.Number),
		MobileNumber: pgText(data.MobileNumber),
		Company:      pgText(data.Company),
	}); err != nil {
		return model.Team{}, err
	}

	if err := qtx.DeleteTeamLeadersByTeamID(ctx, pgInt8(data.ID)); err != nil {
		return model.Team{}, err
	}

	if err := u.insertTeamLeaders(ctx, qtx, data.ID, data.LeaderWorkerIDs); err != nil {
		return model.Team{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return model.Team{}, err
	}

	// Mirror GORM's Update: returns the partial Team built from the
	// mutation input (no row re-read).
	return model.Team{
		ID:           data.ID,
		Number:       data.Number,
		MobileNumber: data.MobileNumber,
		Company:      data.Company,
	}, nil
}

func (u *teamUsecase) Delete(id uint) error {
	ctx := context.Background()
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := u.q.WithTx(tx)
	if err := qtx.DeleteTeamLeadersByTeamID(ctx, pgInt8(id)); err != nil {
		return err
	}
	if err := qtx.DeleteTeam(ctx, int64(id)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (u *teamUsecase) Count(searchParameters dto.TeamSearchParameters) (int64, error) {
	return u.q.CountTeamsFiltered(context.Background(), db.CountTeamsFilteredParams{
		ProjectID: pgInt8(searchParameters.ProjectID),
		Column2:   searchParameters.Number,
		Column3:   searchParameters.MobileNumber,
		Column4:   searchParameters.Company,
		Column5:   int64(searchParameters.TeamLeaderID),
	})
}

func (u *teamUsecase) TemplateFile(projectID uint, filePath string) (string, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		f.Close()
		return "", fmt.Errorf("Не смог открыть шаблонный файл: %v", err)
	}

	teamLeaderSheetName := "Бригадиры"
	teamLeaders, err := u.q.ListWorkersByProject(context.Background(), pgInt8(projectID))
	if err != nil {
		f.Close()
		return "", fmt.Errorf("Данные бригадиров недоступны: %v", err)
	}

	for index, teamLeader := range teamLeaders {
		f.SetCellValue(teamLeaderSheetName, "A"+fmt.Sprint(index+2), teamLeader.Name.String)
	}

	currentTime := time.Now()
	temporaryFileName := fmt.Sprintf(
		"Шаблон для импорта Бригад - %s.xlsx",
		currentTime.Format("02-01-2006"),
	)
	temporaryFilePath := filepath.Join("./storage/import_excel/temp/", temporaryFileName)
	if err := f.SaveAs(temporaryFilePath); err != nil {
		return "", fmt.Errorf("Не удалось обновить шаблон с новыми данными: %v", err)
	}

	f.Close()

	return temporaryFilePath, nil
}

func (u *teamUsecase) Import(projectID uint, filepathStr string) error {
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

	mutationData := []dto.TeamMutation{}
	index := 1
	for len(rows) > index {
		oneEntry := dto.TeamMutation{
			ProjectID: projectID,
		}

		oneEntry.Number, err = f.GetCellValue(sheetName, "A"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepathStr)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке А%d: %v", index+1, err)
		}

		teamLeader, err := f.GetCellValue(sheetName, "B"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepathStr)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке B%d: %v", index+1, err)
		}
		teamLeaderDataFromDB, err := u.q.GetWorkerByName(context.Background(), pgText(teamLeader))
		if err != nil {
			f.Close()
			os.Remove(filepathStr)
			return fmt.Errorf("Ошибка в файле, заданный бригадир в ячейке B%d отсутствует в базе: %v", index+1, err)
		}
		oneEntry.LeaderWorkerIDs = append(oneEntry.LeaderWorkerIDs, uint(teamLeaderDataFromDB.ID))

		oneEntry.MobileNumber, err = f.GetCellValue(sheetName, "C"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepathStr)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке C%d: %v", index+1, err)
		}

		oneEntry.Company, err = f.GetCellValue(sheetName, "D"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepathStr)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке D%d: %v", index+1, err)
		}

		mutationData = append(mutationData, oneEntry)
		index++
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("Ошибка при закрытии файла: %v", err)
	}

	if err := os.Remove(filepathStr); err != nil {
		return fmt.Errorf("Ошибка при удалении временного файла: %v", err)
	}

	return u.createTeamsBatch(mutationData)
}

func (u *teamUsecase) createTeamsBatch(data []dto.TeamMutation) error {
	ctx := context.Background()
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := u.q.WithTx(tx)
	for _, oneEntry := range data {
		row, err := qtx.CreateTeam(ctx, db.CreateTeamParams{
			ProjectID:    pgInt8(oneEntry.ProjectID),
			Number:       pgText(oneEntry.Number),
			MobileNumber: pgText(oneEntry.MobileNumber),
			Company:      pgText(oneEntry.Company),
		})
		if err != nil {
			return err
		}
		if err := u.insertTeamLeaders(ctx, qtx, uint(row.ID), oneEntry.LeaderWorkerIDs); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (u *teamUsecase) insertTeamLeaders(ctx context.Context, qtx *db.Queries, teamID uint, leaderIDs []uint) error {
	if len(leaderIDs) == 0 {
		return nil
	}
	batch := make([]db.CreateTeamLeadersBatchParams, len(leaderIDs))
	for i, leaderID := range leaderIDs {
		batch[i] = db.CreateTeamLeadersBatchParams{
			TeamID:         pgInt8(teamID),
			LeaderWorkerID: pgInt8(leaderID),
		}
	}
	_, err := qtx.CreateTeamLeadersBatch(ctx, batch)
	return err
}

func (u *teamUsecase) DoesTeamNumberAlreadyExistForCreate(teamNumber string, projectID uint) (bool, error) {
	return u.q.TeamNumberExistsForCreate(context.Background(), db.TeamNumberExistsForCreateParams{
		Number:    pgText(teamNumber),
		ProjectID: pgInt8(projectID),
	})
}

func (u *teamUsecase) DoesTeamNumberAlreadyExistForUpdate(teamNumber string, id uint, projectID uint) (bool, error) {
	return u.q.TeamNumberExistsForUpdate(context.Background(), db.TeamNumberExistsForUpdateParams{
		Number:    pgText(teamNumber),
		ID:        int64(id),
		ProjectID: pgInt8(projectID),
	})
}

func (u *teamUsecase) GetAllForSelect(projectID uint) ([]dto.TeamDataForSelect, error) {
	rows, err := u.q.ListTeamsForSelect(context.Background(), pgInt8(projectID))
	if err != nil {
		return nil, err
	}
	out := make([]dto.TeamDataForSelect, len(rows))
	for i, r := range rows {
		out[i] = dto.TeamDataForSelect{
			ID:             uint(r.ID),
			TeamNumber:     r.TeamNumber,
			TeamLeaderName: r.TeamLeaderName,
		}
	}
	return out, nil
}

func (u *teamUsecase) GetAllUniqueTeamNumbers(projectID uint) ([]string, error) {
	return u.q.ListDistinctTeamNumbers(context.Background(), pgInt8(projectID))
}

func (u *teamUsecase) GetAllUniqueMobileNumber(projectID uint) ([]string, error) {
	return u.q.ListDistinctTeamMobileNumbers(context.Background(), pgInt8(projectID))
}

func (u *teamUsecase) GetAllUniqueCompanies(projectID uint) ([]string, error) {
	return u.q.ListDistinctTeamCompanies(context.Background(), pgInt8(projectID))
}

func (u *teamUsecase) Export(projectID uint) (string, error) {
	teamCount, err := u.Count(dto.TeamSearchParameters{ProjectID: projectID})
	if err != nil {
		return "", err
	}

	limit := 100
	page := 1
	teamDataForExport := []dto.TeamPaginated{}
	for teamCount > 0 {
		teams, err := u.GetPaginated(page, limit, dto.TeamSearchParameters{ProjectID: projectID})
		if err != nil {
			return "", err
		}
		teamDataForExport = append(teamDataForExport, teams...)

		page++
		teamCount -= int64(limit)
	}

	materialTempalteFilePath := filepath.Join("./internal/templates", "Шаблон для импорта Бригады.xlsx")
	f, err := excelize.OpenFile(materialTempalteFilePath)
	if err != nil {
		f.Close()
		return "", fmt.Errorf("Не смог открыть файл: %v", err)
	}
	sheetName := "Импорт"
	startingRow := 2

	for index, team := range teamDataForExport {
		f.SetCellStr(sheetName, "A"+fmt.Sprint(startingRow+index), team.Number)
		if len(team.LeaderNames) > 0 {
			f.SetCellStr(sheetName, "B"+fmt.Sprint(startingRow+index), team.LeaderNames[0])
		}
		f.SetCellStr(sheetName, "C"+fmt.Sprint(startingRow+index), team.MobileNumber)
		f.SetCellStr(sheetName, "D"+fmt.Sprint(startingRow+index), team.Company)
	}

	exportFileName := "Экспорт Бригад.xlsx"
	exportFilePath := filepath.Join("./storage/import_excel/temp/", exportFileName)
	if err := f.SaveAs(exportFilePath); err != nil {
		return "", err
	}

	return exportFileName, nil
}

func toModelTeam(t db.Team) model.Team {
	return model.Team{
		ID:           uint(t.ID),
		ProjectID:    uintFromPgInt8(t.ProjectID),
		Number:       t.Number.String,
		MobileNumber: t.MobileNumber.String,
		Company:      t.Company.String,
	}
}
