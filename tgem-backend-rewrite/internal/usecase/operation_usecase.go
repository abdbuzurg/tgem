package usecase

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"backend-v2/internal/db"
	"backend-v2/internal/dto"
	"backend-v2/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"github.com/xuri/excelize/v2"
)

type operationUsecase struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

func NewOperationUsecase(pool *pgxpool.Pool) IOperationUsecase {
	return &operationUsecase{
		pool: pool,
		q:    db.New(pool),
	}
}

type IOperationUsecase interface {
	GetAll(projectID uint) ([]dto.OperationPaginated, error)
	GetPaginated(page, limit int, filter dto.OperationSearchParameters) ([]dto.OperationPaginated, error)
	GetByID(id uint) (model.Operation, error)
	GetByName(name string, projectID uint) (model.Operation, error)
	Create(data dto.Operation) (model.Operation, error)
	Update(data dto.Operation) (model.Operation, error)
	Delete(id uint) error
	Count(filter dto.OperationSearchParameters) (int64, error)
	Import(projectID uint, filepath string) error
	TemplateFile(filepath string, projectID uint) (string, error)
}

func (u *operationUsecase) GetAll(projectID uint) ([]dto.OperationPaginated, error) {
	rows, err := u.q.ListOperationsByProject(context.Background(), pgInt8(projectID))
	if err != nil {
		return nil, err
	}
	out := make([]dto.OperationPaginated, len(rows))
	for i, r := range rows {
		out[i] = dto.OperationPaginated{
			ID:               uintFromPgInt8(r.ID),
			Name:             r.Name,
			Code:             r.Code,
			CostPrime:        decimalFromPgNumeric(r.CostPrime),
			CostWithCustomer: decimalFromPgNumeric(r.CostWithCustomer),
			MaterialID:       uint(r.MaterialID),
			MaterialName:     r.MaterialName,
		}
	}
	return out, nil
}

func (u *operationUsecase) GetPaginated(page, limit int, filter dto.OperationSearchParameters) ([]dto.OperationPaginated, error) {
	rows, err := u.q.ListOperationsPaginated(context.Background(), db.ListOperationsPaginatedParams{
		ProjectID: pgInt8(filter.ProjectID),
		Column2:   filter.Name,
		Column3:   filter.Code,
		Column4:   int64(filter.MaterialID),
		Limit:     int32(limit),
		Offset:    int32((page - 1) * limit),
	})
	if err != nil {
		return nil, err
	}
	out := make([]dto.OperationPaginated, len(rows))
	for i, r := range rows {
		out[i] = dto.OperationPaginated{
			ID:                        uintFromPgInt8(r.ID),
			Name:                      r.Name,
			Code:                      r.Code,
			CostPrime:                 decimalFromPgNumeric(r.CostPrime),
			CostWithCustomer:          decimalFromPgNumeric(r.CostWithCustomer),
			PlannedAmountForProject:   r.PlannedAmountForProject,
			ShowPlannedAmountInReport: r.ShowPlannedAmountInReport,
			MaterialID:                uint(r.MaterialID),
			MaterialName:              r.MaterialName,
		}
	}
	return out, nil
}

func (u *operationUsecase) GetByID(id uint) (model.Operation, error) {
	row, err := u.q.GetOperation(context.Background(), int64(id))
	if err != nil {
		return model.Operation{}, err
	}
	return toModelOperation(row), nil
}

// GetByName preserves the GORM-era quirk: ErrRecordNotFound was swallowed
// (returned (zero-value, nil)), so callers receive an empty Operation
// rather than an error on misses. Folded into ErrNoRows here.
func (u *operationUsecase) GetByName(name string, projectID uint) (model.Operation, error) {
	row, err := u.q.GetOperationByName(context.Background(), db.GetOperationByNameParams{
		Name:      pgText(name),
		ProjectID: pgInt8(projectID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Operation{}, nil
	}
	if err != nil {
		return model.Operation{}, err
	}
	return toModelOperation(row), nil
}

func (u *operationUsecase) Create(data dto.Operation) (model.Operation, error) {
	ctx := context.Background()
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return model.Operation{}, err
	}
	defer tx.Rollback(ctx)

	qtx := u.q.WithTx(tx)
	row, err := qtx.CreateOperation(ctx, db.CreateOperationParams{
		ProjectID:                 pgInt8(data.ProjectID),
		Name:                      pgText(data.Name),
		Code:                      pgText(data.Code),
		CostPrime:                 pgNumericFromDecimal(data.CostPrime),
		CostWithCustomer:          pgNumericFromDecimal(data.CostWithCustomer),
		PlannedAmountForProject:   pgNumericFromFloat64(data.PlannedAmountForProject),
		ShowPlannedAmountInReport: pgBool(data.ShowPlannedAmountInReport),
	})
	if err != nil {
		return model.Operation{}, err
	}

	if data.MaterialID != 0 {
		if err := qtx.CreateOperationMaterial(ctx, db.CreateOperationMaterialParams{
			OperationID: pgInt8(uint(row.ID)),
			MaterialID:  pgInt8(data.MaterialID),
		}); err != nil {
			return model.Operation{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return model.Operation{}, err
	}
	return toModelOperation(row), nil
}

func (u *operationUsecase) Update(data dto.Operation) (model.Operation, error) {
	ctx := context.Background()
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return model.Operation{}, err
	}
	defer tx.Rollback(ctx)

	qtx := u.q.WithTx(tx)
	if err := qtx.UpdateOperation(ctx, db.UpdateOperationParams{
		ID:                        int64(data.ID),
		ProjectID:                 pgInt8(data.ProjectID),
		Name:                      pgText(data.Name),
		Code:                      pgText(data.Code),
		CostPrime:                 pgNumericFromDecimal(data.CostPrime),
		CostWithCustomer:          pgNumericFromDecimal(data.CostWithCustomer),
		PlannedAmountForProject:   pgNumericFromFloat64(data.PlannedAmountForProject),
		ShowPlannedAmountInReport: pgBool(data.ShowPlannedAmountInReport),
	}); err != nil {
		return model.Operation{}, err
	}

	if err := qtx.DeleteOperationMaterialsByOperationID(ctx, pgInt8(data.ID)); err != nil {
		return model.Operation{}, err
	}

	if data.MaterialID != 0 {
		if err := qtx.CreateOperationMaterial(ctx, db.CreateOperationMaterialParams{
			OperationID: pgInt8(data.ID),
			MaterialID:  pgInt8(data.MaterialID),
		}); err != nil {
			return model.Operation{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return model.Operation{}, err
	}

	// Return the same shape the GORM repo did: only the fields that
	// were set in `result` (Update did not re-read the row).
	return model.Operation{
		ID:                        data.ID,
		Name:                      data.Name,
		Code:                      data.Code,
		ProjectID:                 data.ProjectID,
		CostPrime:                 data.CostPrime,
		CostWithCustomer:          data.CostWithCustomer,
		ShowPlannedAmountInReport: data.ShowPlannedAmountInReport,
		PlannedAmountForProject:   data.PlannedAmountForProject,
	}, nil
}

func (u *operationUsecase) Delete(id uint) error {
	ctx := context.Background()
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := u.q.WithTx(tx)
	if err := qtx.DeleteOperationMaterialsByOperationID(ctx, pgInt8(id)); err != nil {
		return err
	}
	if err := qtx.DeleteOperation(ctx, int64(id)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (u *operationUsecase) Count(filter dto.OperationSearchParameters) (int64, error) {
	return u.q.CountOperationsFiltered(context.Background(), db.CountOperationsFilteredParams{
		ProjectID: pgInt8(filter.ProjectID),
		Column2:   filter.Name,
		Column3:   filter.Code,
		Column4:   int64(filter.MaterialID),
	})
}

func (u *operationUsecase) Import(projectID uint, filepath string) error {
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

	sheetName := "Услуги"
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return fmt.Errorf("Не смог найти таблицу 'Импорт': %v", err)
	}

	if len(rows) == 1 {
		return fmt.Errorf("Файл не имеет данных")
	}

	operations := []dto.OperationImportDataForInsert{}
	index := 1
	for len(rows) > index {
		operation := dto.OperationImportDataForInsert{
			ProjectID: projectID,
		}

		operation.Code, err = f.GetCellValue(sheetName, "A"+fmt.Sprint(index+1))
		if err != nil {
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке А%d: %v", index+1, err)
		}

		operation.Name, err = f.GetCellValue(sheetName, "B"+fmt.Sprint(index+1))
		if err != nil {
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке B%d: %v", index+1, err)
		}

		costPrimeStr, err := f.GetCellValue(sheetName, "C"+fmt.Sprint(index+1))
		if err != nil {
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке C%d: %v", index+1, err)
		}

		costPrimeFloat64, err := strconv.ParseFloat(costPrimeStr, 64)
		if err != nil {
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке C%d: %v", index+1, err)
		}
		operation.CostPrime = decimal.NewFromFloat(costPrimeFloat64)

		costWithCustomerStr, err := f.GetCellValue(sheetName, "D"+fmt.Sprint(index+1))
		if err != nil {
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке D%d: %v", index+1, err)
		}

		costWithCustomerFloat64, err := strconv.ParseFloat(costWithCustomerStr, 64)
		if err != nil {
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке D%d: %v", index+1, err)
		}
		operation.CostWithCustomer = decimal.NewFromFloat(costWithCustomerFloat64)

		showInReport, err := f.GetCellValue(sheetName, "E"+fmt.Sprint(index+1))
		if err != nil {
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке E%d: %v", index+1, err)
		}

		showInReport = strings.ToLower(showInReport)
		if showInReport == "да" {
			operation.ShowPlannedAmountInReport = true
		} else {
			operation.ShowPlannedAmountInReport = false
		}

		plannedAmountForProject, err := f.GetCellValue(sheetName, "E"+fmt.Sprint(index+1))
		if err != nil {
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке E%d: %v", index+1, err)
		}

		if operation.PlannedAmountForProject, err = strconv.ParseFloat(plannedAmountForProject, 64); err != nil {
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке E%d: %v", index+1, err)
		}

		materialName, err := f.GetCellValue(sheetName, "G"+fmt.Sprint(index+1))
		if err != nil {
			return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке G%d: %v", index+1, err)
		}

		if materialName != "" {
			material, err := u.q.GetMaterialByProjectAndName(context.Background(), db.GetMaterialByProjectAndNameParams{
				ProjectID: pgInt8(projectID),
				Name:      pgText(materialName),
			})
			if err != nil {
				return fmt.Errorf("Ошибка в файле, неправильный формат данных в ячейке G%d: %v", index+1, err)
			}

			operation.MaterialID = uint(material.ID)
		}

		operations = append(operations, operation)
		index++
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("Не удалось закрыть Excel файл: %v", err)
	}

	if err := os.Remove(filepath); err != nil {
		return fmt.Errorf("Не удалось удалить импортированный файл после сохранения данных: %v", err)
	}

	if err := u.createOperationsInBatch(operations); err != nil {
		return fmt.Errorf("Не удалось сохранить данные: %v", err)
	}

	return nil
}

func (u *operationUsecase) createOperationsInBatch(operations []dto.OperationImportDataForInsert) error {
	ctx := context.Background()
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := u.q.WithTx(tx)
	for _, o := range operations {
		row, err := qtx.CreateOperation(ctx, db.CreateOperationParams{
			ProjectID:                 pgInt8(o.ProjectID),
			Name:                      pgText(o.Name),
			Code:                      pgText(o.Code),
			CostPrime:                 pgNumericFromDecimal(o.CostPrime),
			CostWithCustomer:          pgNumericFromDecimal(o.CostWithCustomer),
			PlannedAmountForProject:   pgNumericFromFloat64(o.PlannedAmountForProject),
			ShowPlannedAmountInReport: pgBool(o.ShowPlannedAmountInReport),
		})
		if err != nil {
			return err
		}

		if o.MaterialID != 0 {
			if err := qtx.CreateOperationMaterial(ctx, db.CreateOperationMaterialParams{
				OperationID: pgInt8(uint(row.ID)),
				MaterialID:  pgInt8(o.MaterialID),
			}); err != nil {
				return err
			}
		}
	}
	return tx.Commit(ctx)
}

func (u *operationUsecase) TemplateFile(filePath string, projectID uint) (string, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		f.Close()
		return "", fmt.Errorf("Не смог открыть шаблонный файл: %v", err)
	}

	materialsSheet := "Материалы"
	allMaterials, err := u.q.ListMaterialsByProject(context.Background(), pgInt8(projectID))
	if err != nil {
		f.Close()
		return "", fmt.Errorf("Данные материалов недоступны: %v", err)
	}

	for index, materials := range allMaterials {
		f.SetCellStr(materialsSheet, "A"+fmt.Sprint(index+2), materials.Name.String)
	}

	currentTime := time.Now()
	temporaryFileName := fmt.Sprintf(
		"Шаблон для импорта Услуг - %s.xlsx",
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

func toModelOperation(o db.Operation) model.Operation {
	planned := 0.0
	if o.PlannedAmountForProject.Valid {
		f, err := o.PlannedAmountForProject.Float64Value()
		if err == nil && f.Valid {
			planned = f.Float64
		}
	}
	return model.Operation{
		ID:                        uint(o.ID),
		ProjectID:                 uintFromPgInt8(o.ProjectID),
		Name:                      o.Name.String,
		Code:                      o.Code.String,
		CostPrime:                 decimalFromPgNumeric(o.CostPrime),
		CostWithCustomer:          decimalFromPgNumeric(o.CostWithCustomer),
		PlannedAmountForProject:   planned,
		ShowPlannedAmountInReport: o.ShowPlannedAmountInReport.Bool,
	}
}
