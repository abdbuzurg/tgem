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

	"github.com/shopspring/decimal"
	"github.com/xuri/excelize/v2"
)

type materialCostUsecase struct {
	q *db.Queries
}

func NewMaterialCostUsecase(q *db.Queries) IMaterialCostUsecase {
	return &materialCostUsecase{q: q}
}

type IMaterialCostUsecase interface {
	GetAll() ([]model.MaterialCost, error)
	GetPaginated(page, limit int, filter dto.MaterialCostSearchFilter) ([]dto.MaterialCostView, error)
	GetByID(id uint) (model.MaterialCost, error)
	Create(data model.MaterialCost) (model.MaterialCost, error)
	Update(data model.MaterialCost) (model.MaterialCost, error)
	Delete(id uint) error
	Count(filter dto.MaterialCostSearchFilter) (int64, error)
	GetByMaterialID(materialID uint) ([]model.MaterialCost, error)
	Import(projectID uint, filePath string) error
	ImportTemplateFile(projectID uint) (string, error)
	Export(projectID uint) (string, error)
}

func (u *materialCostUsecase) GetAll() ([]model.MaterialCost, error) {
	rows, err := u.q.ListMaterialCosts(context.Background())
	if err != nil {
		return nil, err
	}
	out := make([]model.MaterialCost, len(rows))
	for i, r := range rows {
		out[i] = toModelMaterialCost(r)
	}
	return out, nil
}

func (u *materialCostUsecase) GetPaginated(page, limit int, filter dto.MaterialCostSearchFilter) ([]dto.MaterialCostView, error) {
	rows, err := u.q.ListMaterialCostsViewFiltered(context.Background(), db.ListMaterialCostsViewFilteredParams{
		ProjectID: pgInt8(filter.ProjectID),
		Column2:   filter.MaterialName,
		Limit:     int32(limit),
		Offset:    int32((page - 1) * limit),
	})
	if err != nil {
		return nil, err
	}
	out := make([]dto.MaterialCostView, len(rows))
	for i, r := range rows {
		out[i] = dto.MaterialCostView{
			ID:               uint(r.ID),
			CostPrime:        decimalFromPgNumeric(r.CostPrime),
			CostM19:          decimalFromPgNumeric(r.CostM19),
			CostWithCustomer: decimalFromPgNumeric(r.CostWithCustomer),
			MaterialName:     r.MaterialName,
		}
	}
	return out, nil
}

func (u *materialCostUsecase) GetByID(id uint) (model.MaterialCost, error) {
	row, err := u.q.GetMaterialCost(context.Background(), int64(id))
	if err != nil {
		return model.MaterialCost{}, err
	}
	return toModelMaterialCost(row), nil
}

func (u *materialCostUsecase) Create(data model.MaterialCost) (model.MaterialCost, error) {
	row, err := u.q.CreateMaterialCost(context.Background(), db.CreateMaterialCostParams{
		MaterialID:       pgInt8(data.MaterialID),
		CostPrime:        pgNumericFromDecimal(data.CostPrime),
		CostM19:          pgNumericFromDecimal(data.CostM19),
		CostWithCustomer: pgNumericFromDecimal(data.CostWithCustomer),
	})
	if err != nil {
		return model.MaterialCost{}, err
	}
	return toModelMaterialCost(row), nil
}

func (u *materialCostUsecase) Update(data model.MaterialCost) (model.MaterialCost, error) {
	row, err := u.q.UpdateMaterialCost(context.Background(), db.UpdateMaterialCostParams{
		ID:               int64(data.ID),
		MaterialID:       pgInt8(data.MaterialID),
		CostPrime:        pgNumericFromDecimal(data.CostPrime),
		CostM19:          pgNumericFromDecimal(data.CostM19),
		CostWithCustomer: pgNumericFromDecimal(data.CostWithCustomer),
	})
	if err != nil {
		return model.MaterialCost{}, err
	}
	return toModelMaterialCost(row), nil
}

func (u *materialCostUsecase) Delete(id uint) error {
	return u.q.DeleteMaterialCost(context.Background(), int64(id))
}

func (u *materialCostUsecase) Count(filter dto.MaterialCostSearchFilter) (int64, error) {
	return u.q.CountMaterialCostsFiltered(context.Background(), db.CountMaterialCostsFilteredParams{
		ProjectID: pgInt8(filter.ProjectID),
		Column2:   filter.MaterialName,
	})
}

func (u *materialCostUsecase) GetByMaterialID(materialID uint) ([]model.MaterialCost, error) {
	rows, err := u.q.ListMaterialCostsByMaterialIDSorted(context.Background(), pgInt8(materialID))
	if err != nil {
		return nil, err
	}
	out := make([]model.MaterialCost, len(rows))
	for i, r := range rows {
		out[i] = toModelMaterialCost(r)
	}
	return out, nil
}

func (u *materialCostUsecase) ImportTemplateFile(projectID uint) (string, error) {
	materialCostTemplateFilePath := filepath.Join("./internal/templates/", "Шаблон импорта ценников для материалов.xlsx")
	f, err := excelize.OpenFile(materialCostTemplateFilePath)
	if err != nil {
		f.Close()
		return "", fmt.Errorf("Не смог открыть файл: %v", err)
	}

	sheetName := "Материалы"
	materials, err := u.q.ListMaterialsByProject(context.Background(), pgInt8(projectID))
	if err != nil {
		f.Close()
		return "", err
	}

	startingRow := 2

	for index, material := range materials {
		f.SetCellStr(sheetName, "A"+fmt.Sprint(startingRow+index), material.Name.String)
	}

	currentTime := time.Now()
	temporaryFileName := fmt.Sprintf(
		"Шаблон импорта Ценник Материал - %s.xlsx",
		currentTime.Format("02-01-2006"),
	)
	temporaryFilePath := filepath.Join("./storage/import_excel/temp/", temporaryFileName)
	if err := f.SaveAs(temporaryFilePath); err != nil {
		return "", fmt.Errorf("Не удалось обновить шаблон с новыми данными: %v", err)
	}

	f.Close()

	return temporaryFilePath, nil
}

func (u *materialCostUsecase) Import(projectID uint, filepathStr string) error {
	f, err := excelize.OpenFile(filepathStr)
	if err != nil {
		f.Close()
		os.Remove(filepathStr)
		return fmt.Errorf("Не смог открыть файл: %v", err)
	}

	sheetName := "Ценники Материалов"
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

	index := 1
	materialCosts := []model.MaterialCost{}
	for len(rows) > index {
		materialCost := model.MaterialCost{}

		materialName, err := f.GetCellValue(sheetName, "A"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filepathStr)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке А%d: %v", index+1, err)
		}

		material, err := u.q.GetMaterialByProjectAndName(context.Background(), db.GetMaterialByProjectAndNameParams{
			ProjectID: pgInt8(projectID),
			Name:      pgText(materialName),
		})
		if err != nil {
			f.Close()
			os.Remove(filepathStr)
			return fmt.Errorf("Ошибка в файле: наименование на строке %v не найдено в базе: %v", index+1, err)
		}

		materialCost.MaterialID = uint(material.ID)

		costPrime, err := f.GetCellValue(sheetName, "B"+fmt.Sprint(index+1))
		if err != nil || costPrime == "" {
			f.Close()
			os.Remove(filepathStr)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке B%d: %v", index+1, err)
		}

		materialCost.CostPrime, err = decimal.NewFromString(costPrime)
		if err != nil {
			f.Close()
			os.Remove(filepathStr)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке B%d: %v", index+1, err)
		}

		costM19, err := f.GetCellValue(sheetName, "C"+fmt.Sprint(index+1))
		if err != nil || costM19 == "" {
			f.Close()
			os.Remove(filepathStr)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке C%d: %v", index+1, err)
		}

		materialCost.CostM19, err = decimal.NewFromString(costM19)
		if err != nil {
			f.Close()
			os.Remove(filepathStr)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке C%d: %v", index+1, err)
		}

		costWithCustomer, err := f.GetCellValue(sheetName, "D"+fmt.Sprint(index+1))
		if err != nil || costWithCustomer == "" {
			f.Close()
			os.Remove(filepathStr)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке D%d: %v", index+1, err)
		}

		materialCost.CostWithCustomer, err = decimal.NewFromString(costWithCustomer)
		if err != nil {
			f.Close()
			os.Remove(filepathStr)
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке D%d: %v", index+1, err)
		}

		materialCosts = append(materialCosts, materialCost)
		index++
	}

	f.Close()
	os.Remove(filepathStr)

	batch := make([]db.CreateMaterialCostsBatchParams, len(materialCosts))
	for i, mc := range materialCosts {
		batch[i] = db.CreateMaterialCostsBatchParams{
			MaterialID:       pgInt8(mc.MaterialID),
			CostPrime:        pgNumericFromDecimal(mc.CostPrime),
			CostM19:          pgNumericFromDecimal(mc.CostM19),
			CostWithCustomer: pgNumericFromDecimal(mc.CostWithCustomer),
		}
	}
	if _, err := u.q.CreateMaterialCostsBatch(context.Background(), batch); err != nil {
		return fmt.Errorf("Ошибка при сохранение данных: %v", err)
	}

	return nil
}

func (u *materialCostUsecase) Export(projectID uint) (string, error) {
	materialTempalteFilePath := filepath.Join("./internal/templates", "Шаблон импорта ценников для материалов.xlsx")
	f, err := excelize.OpenFile(materialTempalteFilePath)
	if err != nil {
		f.Close()
		return "", fmt.Errorf("Не смог открыть файл: %v", err)
	}
	sheetName := "Ценники Материалов"
	startingRow := 2

	materialCostCount, err := u.Count(dto.MaterialCostSearchFilter{ProjectID: projectID})
	if err != nil {
		return "", err
	}

	limit := 100
	page := 1
	for materialCostCount > 0 {
		rows, err := u.q.ListMaterialCostsViewByProject(context.Background(), db.ListMaterialCostsViewByProjectParams{
			ProjectID: pgInt8(projectID),
			Limit:     int32(limit),
			Offset:    int32((page - 1) * limit),
		})
		if err != nil {
			return "", err
		}

		for index, r := range rows {
			f.SetCellStr(sheetName, "A"+fmt.Sprint(startingRow+index), r.MaterialName)
			costPrime, _ := decimalFromPgNumeric(r.CostPrime).Float64()
			f.SetCellFloat(sheetName, "B"+fmt.Sprint(startingRow+index), costPrime, 2, 64)
			costM19, _ := decimalFromPgNumeric(r.CostM19).Float64()
			f.SetCellFloat(sheetName, "C"+fmt.Sprint(startingRow+index), costM19, 2, 64)
			costWithCustomer, _ := decimalFromPgNumeric(r.CostWithCustomer).Float64()
			f.SetCellFloat(sheetName, "D"+fmt.Sprint(startingRow+index), costWithCustomer, 2, 64)
		}

		startingRow = page*limit + 2
		page++
		materialCostCount -= int64(limit)
	}

	exportFileName := "Экспорт Ценников для Материалов.xlsx"
	exportFilePath := filepath.Join("./storage/import_excel/temp/", exportFileName)
	if err := f.SaveAs(exportFilePath); err != nil {
		return "", err
	}

	return exportFileName, nil
}

func toModelMaterialCost(mc db.MaterialCost) model.MaterialCost {
	return model.MaterialCost{
		ID:               uint(mc.ID),
		MaterialID:       uintFromPgInt8(mc.MaterialID),
		CostPrime:        decimalFromPgNumeric(mc.CostPrime),
		CostM19:          decimalFromPgNumeric(mc.CostM19),
		CostWithCustomer: decimalFromPgNumeric(mc.CostWithCustomer),
	}
}
