package usecase

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"backend-v2/internal/db"
	"backend-v2/model"

	"github.com/xuri/excelize/v2"
)

type materialUsecase struct {
	q *db.Queries
}

func NewMaterialUsecase(q *db.Queries) IMaterialUsecase {
	return &materialUsecase{q: q}
}

type IMaterialUsecase interface {
	GetAll(projectID uint) ([]model.Material, error)
	GetPaginated(page, limit int, data model.Material) ([]model.Material, error)
	GetByID(id uint) (model.Material, error)
	Create(data model.Material) (model.Material, error)
	Update(data model.Material) (model.Material, error)
	Delete(id uint) error
	Count(filter model.Material) (int64, error)
	Import(projectID uint, filepath string) error
	Export(projectID uint) (string, error)
}

func (u *materialUsecase) GetAll(projectID uint) ([]model.Material, error) {
	rows, err := u.q.ListMaterialsByProject(context.Background(), pgInt8(projectID))
	if err != nil {
		return nil, err
	}
	out := make([]model.Material, len(rows))
	for i, r := range rows {
		out[i] = toModelMaterial(r)
	}
	return out, nil
}

func (u *materialUsecase) GetPaginated(page, limit int, data model.Material) ([]model.Material, error) {
	rows, err := u.q.ListMaterialsPaginatedFiltered(context.Background(), db.ListMaterialsPaginatedFilteredParams{
		ProjectID: pgInt8(data.ProjectID),
		Column2:   data.Category,
		Column3:   data.Code,
		Column4:   data.Name,
		Column5:   data.Unit,
		Limit:     int32(limit),
		Offset:    int32((page - 1) * limit),
	})
	if err != nil {
		return nil, err
	}
	out := make([]model.Material, len(rows))
	for i, r := range rows {
		out[i] = toModelMaterial(r)
	}
	return out, nil
}

func (u *materialUsecase) GetByID(id uint) (model.Material, error) {
	row, err := u.q.GetMaterial(context.Background(), int64(id))
	if err != nil {
		return model.Material{}, err
	}
	return toModelMaterial(row), nil
}

func (u *materialUsecase) Create(data model.Material) (model.Material, error) {
	row, err := u.q.CreateMaterial(context.Background(), createMaterialParamsFromModel(data))
	if err != nil {
		return model.Material{}, err
	}
	return toModelMaterial(row), nil
}

func (u *materialUsecase) Update(data model.Material) (model.Material, error) {
	row, err := u.q.UpdateMaterial(context.Background(), db.UpdateMaterialParams{
		ID:                        int64(data.ID),
		Category:                  pgText(data.Category),
		Code:                      pgText(data.Code),
		Name:                      pgText(data.Name),
		Unit:                      pgText(data.Unit),
		Notes:                     pgText(data.Notes),
		HasSerialNumber:           pgBool(data.HasSerialNumber),
		Article:                   pgText(data.Article),
		ProjectID:                 pgInt8(data.ProjectID),
		PlannedAmountForProject:   pgNumericFromFloat64(data.PlannedAmountForProject),
		ShowPlannedAmountInReport: pgBool(data.ShowPlannedAmountInReport),
	})
	if err != nil {
		return model.Material{}, err
	}
	return toModelMaterial(row), nil
}

func (u *materialUsecase) Delete(id uint) error {
	return u.q.DeleteMaterial(context.Background(), int64(id))
}

func (u *materialUsecase) Count(filter model.Material) (int64, error) {
	return u.q.CountMaterialsFiltered(context.Background(), db.CountMaterialsFilteredParams{
		ProjectID: pgInt8(filter.ProjectID),
		Column2:   filter.Category,
		Column3:   filter.Code,
		Column4:   filter.Name,
		Column5:   filter.Unit,
	})
}

func (u *materialUsecase) Import(projectID uint, filepath string) error {
	f, err := excelize.OpenFile(filepath)
	if err != nil {
		f.Close()
		os.Remove(filepath)
		return fmt.Errorf("Не смог открыть файл: %v", err)
	}

	defer func() {
		f.Close()
		os.Remove(filepath)
	}()

	sheetName := "Материалы"
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return fmt.Errorf("Не смог найти таблицу 'Импорт': %v", err)
	}

	if len(rows) == 1 {
		return fmt.Errorf("Файл не имеет данных")
	}

	materials := []model.Material{}
	index := 1
	for len(rows) > index {
		material := model.Material{
			ProjectID: projectID,
		}

		material.Name, err = f.GetCellValue(sheetName, "A"+fmt.Sprint(index+1))
		if err != nil {
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке А%d: %v", index+1, err)
		}

		material.Code, err = f.GetCellValue(sheetName, "B"+fmt.Sprint(index+1))
		if err != nil {
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке B%d: %v", index+1, err)
		}

		material.Category, err = f.GetCellValue(sheetName, "C"+fmt.Sprint(index+1))
		if err != nil {
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке C%d: %v", index+1, err)
		}

		material.Unit, err = f.GetCellValue(sheetName, "D"+fmt.Sprint(index+1))
		if err != nil {
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке D%d: %v", index+1, err)
		}

		material.Article, err = f.GetCellValue(sheetName, "E"+fmt.Sprint(index+1))
		if err != nil {
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке E%d: %v", index+1, err)
		}

		serialNumberStatus, err := f.GetCellValue(sheetName, "F"+fmt.Sprint(index+1))
		if err != nil {
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке F%d: %v", index+1, err)
		}

		serialNumberStatus = strings.ToLower(serialNumberStatus)
		if serialNumberStatus == "да" {
			material.HasSerialNumber = true
		} else {
			material.HasSerialNumber = false
		}

		showInReport, err := f.GetCellValue(sheetName, "G"+fmt.Sprint(index+1))
		if err != nil {
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке G%d: %v", index+1, err)
		}

		showInReport = strings.ToLower(showInReport)
		if showInReport == "да" {
			material.ShowPlannedAmountInReport = true
		} else {
			material.ShowPlannedAmountInReport = false
		}

		plannedAmountForProject, err := f.GetCellValue(sheetName, "H"+fmt.Sprint(index+1))
		if err != nil {
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке H%d: %v", index+1, err)
		}

		if material.PlannedAmountForProject, err = strconv.ParseFloat(plannedAmountForProject, 64); err != nil {
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке H%d: %v", index+1, err)
		}

		material.Notes, err = f.GetCellValue(sheetName, "I"+fmt.Sprint(index+1))
		if err != nil {
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке I%d: %v", index+1, err)
		}

		materials = append(materials, material)
		index++
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("Не удалось закрыть Excel файл: %v", err)
	}

	if err := os.Remove(filepath); err != nil {
		return fmt.Errorf("Не удалось удалить импортированный файл после сохранения данных: %v", err)
	}

	batch := make([]db.CreateMaterialsBatchParams, len(materials))
	for i, m := range materials {
		batch[i] = db.CreateMaterialsBatchParams{
			Category:                  pgText(m.Category),
			Code:                      pgText(m.Code),
			Name:                      pgText(m.Name),
			Unit:                      pgText(m.Unit),
			Notes:                     pgText(m.Notes),
			HasSerialNumber:           pgBool(m.HasSerialNumber),
			Article:                   pgText(m.Article),
			ProjectID:                 pgInt8(m.ProjectID),
			PlannedAmountForProject:   pgNumericFromFloat64(m.PlannedAmountForProject),
			ShowPlannedAmountInReport: pgBool(m.ShowPlannedAmountInReport),
		}
	}
	if _, err := u.q.CreateMaterialsBatch(context.Background(), batch); err != nil {
		return fmt.Errorf("Не удалось сохранить данные: %v", err)
	}

	return nil
}

func (u *materialUsecase) Export(projectID uint) (string, error) {
	materialTempalteFilePath := filepath.Join("./internal/templates", "Шаблон для импорта Материалов.xlsx")
	f, err := excelize.OpenFile(materialTempalteFilePath)
	if err != nil {
		return "", fmt.Errorf("Не смог открыть файл: %v", err)
	}
	sheetName := "Материалы"
	startingRow := 2

	materialCount, err := u.Count(model.Material{ProjectID: projectID})
	if err != nil {
		return "", err
	}

	limit := 100
	page := 1
	for materialCount > 0 {
		rows, err := u.q.ListMaterialsByProjectPaginated(context.Background(), db.ListMaterialsByProjectPaginatedParams{
			ProjectID: pgInt8(projectID),
			Limit:     int32(limit),
			Offset:    int32((page - 1) * limit),
		})
		if err != nil {
			return "", err
		}

		for index, r := range rows {
			material := toModelMaterial(r)
			f.SetCellStr(sheetName, "A"+fmt.Sprint(startingRow+index), material.Name)
			f.SetCellStr(sheetName, "B"+fmt.Sprint(startingRow+index), material.Code)
			f.SetCellStr(sheetName, "C"+fmt.Sprint(startingRow+index), material.Category)
			f.SetCellStr(sheetName, "D"+fmt.Sprint(startingRow+index), material.Unit)
			f.SetCellStr(sheetName, "E"+fmt.Sprint(startingRow+index), material.Article)

			serialNumberStatus := "Нет"
			if material.HasSerialNumber {
				serialNumberStatus = "Да"
			}

			f.SetCellStr(sheetName, "F"+fmt.Sprint(startingRow+index), serialNumberStatus)
			f.SetCellStr(sheetName, "G"+fmt.Sprint(startingRow+index), material.Notes)
		}

		startingRow = page*limit + 2
		page++
		materialCount -= int64(limit)
	}

	exportFileName := "Экспорт Материалов.xlsx"
	exportFilePath := filepath.Join("./storage/import_excel/temp/", exportFileName)
	if err := f.SaveAs(exportFilePath); err != nil {
		return "", err
	}

	return exportFileName, nil
}

func createMaterialParamsFromModel(m model.Material) db.CreateMaterialParams {
	return db.CreateMaterialParams{
		Category:                  pgText(m.Category),
		Code:                      pgText(m.Code),
		Name:                      pgText(m.Name),
		Unit:                      pgText(m.Unit),
		Notes:                     pgText(m.Notes),
		HasSerialNumber:           pgBool(m.HasSerialNumber),
		Article:                   pgText(m.Article),
		ProjectID:                 pgInt8(m.ProjectID),
		PlannedAmountForProject:   pgNumericFromFloat64(m.PlannedAmountForProject),
		ShowPlannedAmountInReport: pgBool(m.ShowPlannedAmountInReport),
	}
}

func toModelMaterial(m db.Material) model.Material {
	planned := 0.0
	if m.PlannedAmountForProject.Valid {
		f, err := m.PlannedAmountForProject.Float64Value()
		if err == nil && f.Valid {
			planned = f.Float64
		}
	}
	return model.Material{
		ID:                        uint(m.ID),
		Category:                  m.Category.String,
		Code:                      m.Code.String,
		Name:                      m.Name.String,
		Unit:                      m.Unit.String,
		Notes:                     m.Notes.String,
		HasSerialNumber:           m.HasSerialNumber.Bool,
		Article:                   m.Article.String,
		ProjectID:                 uintFromPgInt8(m.ProjectID),
		PlannedAmountForProject:   planned,
		ShowPlannedAmountInReport: m.ShowPlannedAmountInReport.Bool,
	}
}
