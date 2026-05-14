package usecase

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"backend-v2/internal/db"
	"backend-v2/internal/dto"
	"backend-v2/internal/utils"
	"backend-v2/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xuri/excelize/v2"
)

type invoiceInputUsecase struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

func NewInvoiceInputUsecase(pool *pgxpool.Pool) IInvoiceInputUsecase {
	return &invoiceInputUsecase{
		pool: pool,
		q:    db.New(pool),
	}
}

type IInvoiceInputUsecase interface {
	GetAll() ([]model.InvoiceInput, error)
	GetPaginated(page, limit int, filter dto.InvoiceInputSearchParameters) ([]dto.InvoiceInputPaginated, error)
	GetByID(id uint) (model.InvoiceInput, error)
	GetInvoiceMaterialsWithoutSerialNumbers(id uint) ([]dto.InvoiceMaterialsWithoutSerialNumberView, error)
	GetInvoiceMaterialsWithSerialNumbers(id uint) ([]dto.InvoiceMaterialsWithSerialNumberView, error)
	Create(data dto.InvoiceInput) (model.InvoiceInput, error)
	Update(data dto.InvoiceInput) (model.InvoiceInput, error)
	Delete(id uint) error
	Count(filter dto.InvoiceInputSearchParameters) (int64, error)
	Confirmation(id, projectID uint) error
	UniqueCode(projectID uint) ([]dto.DataForSelect[string], error)
	UniqueWarehouseManager(projectID uint) ([]dto.DataForSelect[uint], error)
	UniqueReleased(projectID uint) ([]dto.DataForSelect[uint], error)
	Report(filter dto.InvoiceInputReportFilterRequest) (string, error)
	NewMaterialCost(data model.MaterialCost) error
	NewMaterialAndItsCost(data dto.NewMaterialDataFromInvoiceInput) error
	GetMaterialsForEdit(id uint) ([]dto.InvoiceInputMaterialForEdit, error)
	Import(filePath string, projectID uint, workerID uint) error
	GetParametersForSearch(projectID uint) (dto.InvoiceInputParametersForSearch, error)
}

func (u *invoiceInputUsecase) GetAll() ([]model.InvoiceInput, error) {
	rows, err := u.q.ListInvoiceInputs(context.Background())
	if err != nil {
		return nil, err
	}
	out := make([]model.InvoiceInput, len(rows))
	for i, r := range rows {
		out[i] = toModelInvoiceInput(r)
	}
	return out, nil
}

func (u *invoiceInputUsecase) GetPaginated(page, limit int, filter dto.InvoiceInputSearchParameters) ([]dto.InvoiceInputPaginated, error) {
	ctx := context.Background()
	if len(filter.Materials) != 0 {
		materialIDs := make([]int64, len(filter.Materials))
		for i, id := range filter.Materials {
			materialIDs[i] = int64(id)
		}
		rows, err := u.q.ListInvoiceInputsPaginatedByMaterials(ctx, db.ListInvoiceInputsPaginatedByMaterialsParams{
			ProjectID:   pgInt8(filter.ProjectID),
			MaterialIds: materialIDs,
			Limit:       int32(limit),
			Offset:      int32((page - 1) * limit),
		})
		if err != nil {
			return nil, err
		}
		return invoiceInputPaginatedRowsToDTO(rows), nil
	}

	rows, err := u.q.ListInvoiceInputsPaginatedFiltered(ctx, db.ListInvoiceInputsPaginatedFilteredParams{
		ProjectID: pgInt8(filter.ProjectID),
		Column2:   int64(filter.WarehouseManagerWorkerID),
		Column3:   int64(filter.ReleasedWorkerID),
		Column4:   filter.DeliveryCode,
		Column5:   pgTimestamptz(filter.DateFrom),
		Column6:   pgTimestamptz(filter.DateTo),
		Limit:     int32(limit),
		Offset:    int32((page - 1) * limit),
	})
	if err != nil {
		return nil, err
	}
	return invoiceInputPaginatedFilteredRowsToDTO(rows), nil
}

func invoiceInputPaginatedRowsToDTO(rows []db.ListInvoiceInputsPaginatedByMaterialsRow) []dto.InvoiceInputPaginated {
	out := make([]dto.InvoiceInputPaginated, len(rows))
	for i, r := range rows {
		out[i] = dto.InvoiceInputPaginated{
			ID:                   uint(r.ID),
			Confirmation:         r.Confirmation,
			DeliveryCode:         r.DeliveryCode,
			WarehouseManagerName: r.WarehouseManagerName,
			ReleasedName:         r.ReleasedName,
			DateOfInvoice:        timeFromPgTimestamptz(r.DateOfInvoice),
		}
	}
	return out
}

func invoiceInputPaginatedFilteredRowsToDTO(rows []db.ListInvoiceInputsPaginatedFilteredRow) []dto.InvoiceInputPaginated {
	out := make([]dto.InvoiceInputPaginated, len(rows))
	for i, r := range rows {
		out[i] = dto.InvoiceInputPaginated{
			ID:                   uint(r.ID),
			Confirmation:         r.Confirmation,
			DeliveryCode:         r.DeliveryCode,
			WarehouseManagerName: r.WarehouseManagerName,
			ReleasedName:         r.ReleasedName,
			DateOfInvoice:        timeFromPgTimestamptz(r.DateOfInvoice),
		}
	}
	return out
}

func (u *invoiceInputUsecase) GetByID(id uint) (model.InvoiceInput, error) {
	row, err := u.q.GetInvoiceInput(context.Background(), int64(id))
	if err != nil {
		return model.InvoiceInput{}, err
	}
	return toModelInvoiceInput(row), nil
}

func (u *invoiceInputUsecase) GetInvoiceMaterialsWithoutSerialNumbers(id uint) ([]dto.InvoiceMaterialsWithoutSerialNumberView, error) {
	rows, err := u.q.ListInvoiceMaterialsWithoutSerialNumbers(context.Background(), db.ListInvoiceMaterialsWithoutSerialNumbersParams{
		InvoiceType: pgText("input"),
		InvoiceID:   pgInt8(id),
	})
	if err != nil {
		return nil, err
	}
	out := make([]dto.InvoiceMaterialsWithoutSerialNumberView, len(rows))
	for i, r := range rows {
		out[i] = dto.InvoiceMaterialsWithoutSerialNumberView{
			ID:           uint(r.ID),
			MaterialName: r.MaterialName,
			MaterialUnit: r.MaterialUnit,
			IsDefected:   r.IsDefected,
			CostM19:      decimalFromPgNumeric(r.CostM19),
			Amount:       float64FromPgNumeric(r.Amount),
			Notes:        r.Notes,
		}
	}
	return out, nil
}

func (u *invoiceInputUsecase) GetInvoiceMaterialsWithSerialNumbers(id uint) ([]dto.InvoiceMaterialsWithSerialNumberView, error) {
	rows, err := u.q.ListInvoiceMaterialsWithSerialNumbers(context.Background(), db.ListInvoiceMaterialsWithSerialNumbersParams{
		InvoiceType: pgText("input"),
		InvoiceID:   pgInt8(id),
	})
	if err != nil {
		return nil, err
	}

	queryData := make([]dto.InvoiceMaterialsWithSerialNumberQuery, len(rows))
	for i, r := range rows {
		queryData[i] = dto.InvoiceMaterialsWithSerialNumberQuery{
			ID:           uint(r.ID),
			MaterialName: r.MaterialName,
			MaterialUnit: r.MaterialUnit,
			IsDefected:   r.IsDefected,
			CostM19:      decimalFromPgNumeric(r.CostM19),
			SerialNumber: r.SerialNumber,
			Amount:       float64FromPgNumeric(r.Amount),
			Notes:        r.Notes,
		}
	}

	result := []dto.InvoiceMaterialsWithSerialNumberView{}
	current := dto.InvoiceMaterialsWithSerialNumberView{}
	for index, materialInfo := range queryData {
		if index == 0 {
			current = dto.InvoiceMaterialsWithSerialNumberView{
				ID:            materialInfo.ID,
				MaterialName:  materialInfo.MaterialName,
				MaterialUnit:  materialInfo.MaterialUnit,
				SerialNumbers: []string{},
				Amount:        materialInfo.Amount,
				CostM19:       materialInfo.CostM19,
				Notes:         materialInfo.Notes,
			}
		}

		if current.MaterialName == materialInfo.MaterialName && current.CostM19.Equal(materialInfo.CostM19) {
			if len(current.SerialNumbers) == 0 {
				current.SerialNumbers = append(current.SerialNumbers, materialInfo.SerialNumber)
				continue
			}

			if current.SerialNumbers[len(current.SerialNumbers)-1] != materialInfo.SerialNumber {
				current.SerialNumbers = append(current.SerialNumbers, materialInfo.SerialNumber)
			}
		} else {
			result = append(result, current)
			current = dto.InvoiceMaterialsWithSerialNumberView{
				ID:            materialInfo.ID,
				MaterialName:  materialInfo.MaterialName,
				MaterialUnit:  materialInfo.MaterialUnit,
				SerialNumbers: []string{materialInfo.SerialNumber},
				Amount:        materialInfo.Amount,
				CostM19:       materialInfo.CostM19,
				Notes:         materialInfo.Notes,
			}
		}
	}

	if len(queryData) != 0 {
		result = append(result, current)
	}

	return result, nil
}

func (u *invoiceInputUsecase) Create(data dto.InvoiceInput) (model.InvoiceInput, error) {
	ctx := context.Background()

	// GORM-era CountInvoice returned 0 silently when no invoice_counts row
	// existed for the (input, project) pair. Preserve that by folding
	// pgx.ErrNoRows into a 0 count.
	count, err := u.q.GetInvoiceCount(ctx, db.GetInvoiceCountParams{
		InvoiceType: pgText("input"),
		ProjectID:   pgInt8(data.Details.ProjectID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		count = 0
	} else if err != nil {
		return model.InvoiceInput{}, err
	}

	code := utils.UniqueCodeGeneration("П", count+1, data.Details.ProjectID)
	data.Details.DeliveryCode = code

	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return model.InvoiceInput{}, err
	}
	defer tx.Rollback(ctx)
	qtx := u.q.WithTx(tx)

	invoiceRow, err := qtx.CreateInvoiceInput(ctx, db.CreateInvoiceInputParams{
		ProjectID:                pgInt8(data.Details.ProjectID),
		WarehouseManagerWorkerID: pgInt8(data.Details.WarehouseManagerWorkerID),
		ReleasedWorkerID:         pgInt8(data.Details.ReleasedWorkerID),
		DeliveryCode:             pgText(data.Details.DeliveryCode),
		Notes:                    pgText(data.Details.Notes),
		DateOfInvoice:            pgTimestamptz(data.Details.DateOfInvoice),
		Confirmed:                pgBool(data.Details.Confirmed),
	})
	if err != nil {
		return model.InvoiceInput{}, err
	}

	if err := writeInvoiceInputItems(ctx, qtx, uint(invoiceRow.ID), data); err != nil {
		return model.InvoiceInput{}, err
	}

	if err := qtx.IncrementInvoiceCount(ctx, db.IncrementInvoiceCountParams{
		InvoiceType: pgText("input"),
		ProjectID:   pgInt8(uintFromPgInt8(invoiceRow.ProjectID)),
	}); err != nil {
		return model.InvoiceInput{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return model.InvoiceInput{}, err
	}

	return toModelInvoiceInput(invoiceRow), nil
}

func (u *invoiceInputUsecase) Update(data dto.InvoiceInput) (model.InvoiceInput, error) {
	ctx := context.Background()
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return model.InvoiceInput{}, err
	}
	defer tx.Rollback(ctx)
	qtx := u.q.WithTx(tx)

	invoiceRow, err := qtx.UpdateInvoiceInput(ctx, db.UpdateInvoiceInputParams{
		ID:                       int64(data.Details.ID),
		ProjectID:                pgInt8(data.Details.ProjectID),
		WarehouseManagerWorkerID: pgInt8(data.Details.WarehouseManagerWorkerID),
		ReleasedWorkerID:         pgInt8(data.Details.ReleasedWorkerID),
		DeliveryCode:             pgText(data.Details.DeliveryCode),
		Notes:                    pgText(data.Details.Notes),
		DateOfInvoice:            pgTimestamptz(data.Details.DateOfInvoice),
		Confirmed:                pgBool(data.Details.Confirmed),
	})
	if err != nil {
		return model.InvoiceInput{}, err
	}

	if err := qtx.DeleteInvoiceMaterialsByInvoice(ctx, db.DeleteInvoiceMaterialsByInvoiceParams{
		InvoiceType: pgText("input"),
		InvoiceID:   pgInt8(uint(invoiceRow.ID)),
	}); err != nil {
		return model.InvoiceInput{}, err
	}

	if err := qtx.DeleteSerialNumberMovementsByInvoice(ctx, db.DeleteSerialNumberMovementsByInvoiceParams{
		InvoiceType: pgText("input"),
		InvoiceID:   pgInt8(uint(invoiceRow.ID)),
	}); err != nil {
		return model.InvoiceInput{}, err
	}

	if err := writeInvoiceInputItems(ctx, qtx, uint(invoiceRow.ID), data); err != nil {
		return model.InvoiceInput{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return model.InvoiceInput{}, err
	}

	return toModelInvoiceInput(invoiceRow), nil
}

// writeInvoiceInputItems creates the invoice_materials, serial_numbers and
// serial_number_movements rows for an invoice. Used by both Create and the
// post-delete path of Update. The serial_number rows are inserted one at a
// time so the freshly-assigned ids can be used for the matching
// serial_number_movements rows.
func writeInvoiceInputItems(ctx context.Context, qtx *db.Queries, invoiceID uint, data dto.InvoiceInput) error {
	if len(data.Items) == 0 {
		return nil
	}

	imBatch := make([]db.CreateInvoiceMaterialsBatchParams, 0, len(data.Items))
	for _, item := range data.Items {
		imBatch = append(imBatch, db.CreateInvoiceMaterialsBatchParams{
			ProjectID:      pgInt8(data.Details.ProjectID),
			MaterialCostID: pgInt8(item.MaterialData.MaterialCostID),
			InvoiceID:      pgInt8(invoiceID),
			InvoiceType:    pgText("input"),
			IsDefected:     pgBool(item.MaterialData.IsDefected),
			Amount:         pgNumericFromFloat64(item.MaterialData.Amount),
			Notes:          pgText(item.MaterialData.Notes),
		})
	}
	if _, err := qtx.CreateInvoiceMaterialsBatch(ctx, imBatch); err != nil {
		return err
	}

	for _, item := range data.Items {
		if len(item.SerialNumbers) == 0 {
			continue
		}
		for _, code := range item.SerialNumbers {
			snRow, err := qtx.CreateSerialNumber(ctx, db.CreateSerialNumberParams{
				ProjectID:      pgInt8(data.Details.ProjectID),
				MaterialCostID: pgInt8(item.MaterialData.MaterialCostID),
				Code:           pgText(code),
			})
			if err != nil {
				return err
			}
			if _, err := qtx.CreateSerialNumberMovementsBatch(ctx, []db.CreateSerialNumberMovementsBatchParams{{
				SerialNumberID: pgInt8(uint(snRow.ID)),
				ProjectID:      pgInt8(data.Details.ProjectID),
				InvoiceID:      pgInt8(invoiceID),
				InvoiceType:    pgText("input"),
				IsDefected:     pgBool(false),
				Confirmation:   pgBool(false),
			}}); err != nil {
				return err
			}
		}
	}
	return nil
}

func (u *invoiceInputUsecase) Delete(id uint) error {
	ctx := context.Background()
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	qtx := u.q.WithTx(tx)

	if err := qtx.DeleteInvoiceInput(ctx, int64(id)); err != nil {
		return err
	}
	if err := qtx.DeleteInvoiceMaterialsByInvoice(ctx, db.DeleteInvoiceMaterialsByInvoiceParams{
		InvoiceType: pgText("input"),
		InvoiceID:   pgInt8(id),
	}); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (u *invoiceInputUsecase) Count(filter dto.InvoiceInputSearchParameters) (int64, error) {
	ctx := context.Background()
	if len(filter.Materials) != 0 {
		materialIDs := make([]int64, len(filter.Materials))
		for i, id := range filter.Materials {
			materialIDs[i] = int64(id)
		}
		return u.q.CountInvoiceInputsByMaterials(ctx, db.CountInvoiceInputsByMaterialsParams{
			ProjectID:   pgInt8(filter.ProjectID),
			MaterialIds: materialIDs,
		})
	}

	return u.q.CountInvoiceInputsFiltered(ctx, db.CountInvoiceInputsFilteredParams{
		ProjectID: pgInt8(filter.ProjectID),
		Column2:   int64(filter.WarehouseManagerWorkerID),
		Column3:   int64(filter.ReleasedWorkerID),
		Column4:   filter.DeliveryCode,
		Column5:   pgTimestamptz(filter.DateFrom),
		Column6:   pgTimestamptz(filter.DateTo),
	})
}

func (u *invoiceInputUsecase) Confirmation(id, projectID uint) error {
	ctx := context.Background()
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	qtx := u.q.WithTx(tx)

	invoiceMaterials, err := qtx.ListInvoiceMaterialsByInvoice(ctx, db.ListInvoiceMaterialsByInvoiceParams{
		InvoiceID:   pgInt8(id),
		InvoiceType: pgText("input"),
		ProjectID:   pgInt8(projectID),
	})
	if err != nil {
		return err
	}

	// material_locations rows for this invoice's costs already in the warehouse.
	// The GORM repo had `materialLocationRepo.GetMaterialsInLocationBasedOnInvoiceID`
	// — preserved in the GORM repo, used here via raw query equivalent below.
	existing, err := qtx.ListMaterialLocationsForInvoiceConfirmation(ctx, db.ListMaterialLocationsForInvoiceConfirmationParams{
		LocationType: pgText("warehouse"),
		LocationID:   pgInt8(0),
		InvoiceType:  pgText("input"),
		InvoiceID:    pgInt8(id),
	})
	if err != nil {
		return err
	}

	for _, im := range invoiceMaterials {
		matched := false
		for _, ml := range existing {
			if uintFromPgInt8(ml.MaterialCostID) == uintFromPgInt8(im.MaterialCostID) {
				newAmount := float64FromPgNumeric(ml.Amount) + float64FromPgNumeric(im.Amount)
				if err := qtx.UpsertMaterialLocationByID(ctx, db.UpsertMaterialLocationByIDParams{
					Amount: pgNumericFromFloat64(newAmount),
					ID:     ml.ID,
				}); err != nil {
					return err
				}
				matched = true
				break
			}
		}
		if !matched {
			if _, err := qtx.CreateMaterialLocation(ctx, db.CreateMaterialLocationParams{
				ProjectID:      pgInt8(projectID),
				MaterialCostID: im.MaterialCostID,
				LocationID:     pgInt8(0),
				LocationType:   pgText("warehouse"),
				Amount:         im.Amount,
			}); err != nil {
				return err
			}
		}
	}

	movements, err := qtx.ListSerialNumberMovementsByInvoice(ctx, db.ListSerialNumberMovementsByInvoiceParams{
		InvoiceID:   pgInt8(id),
		InvoiceType: pgText("input"),
	})
	if err != nil {
		return err
	}

	if err := qtx.ConfirmInvoiceInput(ctx, int64(id)); err != nil {
		return err
	}

	if err := qtx.ConfirmSerialNumberMovementsByInvoice(ctx, db.ConfirmSerialNumberMovementsByInvoiceParams{
		InvoiceID:   pgInt8(id),
		InvoiceType: pgText("input"),
	}); err != nil {
		return err
	}

	if len(movements) > 0 {
		batch := make([]db.CreateSerialNumberLocationsBatchParams, len(movements))
		for i, m := range movements {
			batch[i] = db.CreateSerialNumberLocationsBatchParams{
				SerialNumberID: m.SerialNumberID,
				ProjectID:      pgInt8(projectID),
				LocationID:     pgInt8(0),
				LocationType:   pgText("warehouse"),
			}
		}
		if _, err := qtx.CreateSerialNumberLocationsBatch(ctx, batch); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (u *invoiceInputUsecase) UniqueCode(projectID uint) ([]dto.DataForSelect[string], error) {
	rows, err := u.q.ListInvoiceInputUniqueDeliveryCodes(context.Background(), pgInt8(projectID))
	if err != nil {
		return nil, err
	}
	out := make([]dto.DataForSelect[string], len(rows))
	for i, r := range rows {
		out[i] = dto.DataForSelect[string]{Label: r.Label, Value: r.Value}
	}
	return out, nil
}

func (u *invoiceInputUsecase) UniqueWarehouseManager(projectID uint) ([]dto.DataForSelect[uint], error) {
	rows, err := u.q.ListInvoiceInputUniqueWarehouseManagers(context.Background(), pgInt8(projectID))
	if err != nil {
		return nil, err
	}
	out := make([]dto.DataForSelect[uint], len(rows))
	for i, r := range rows {
		out[i] = dto.DataForSelect[uint]{Label: r.Label, Value: uint(r.Value)}
	}
	return out, nil
}

func (u *invoiceInputUsecase) UniqueReleased(projectID uint) ([]dto.DataForSelect[uint], error) {
	rows, err := u.q.ListInvoiceInputUniqueReleasedWorkers(context.Background(), pgInt8(projectID))
	if err != nil {
		return nil, err
	}
	out := make([]dto.DataForSelect[uint], len(rows))
	for i, r := range rows {
		out[i] = dto.DataForSelect[uint]{Label: r.Label, Value: uint(r.Value)}
	}
	return out, nil
}

func (u *invoiceInputUsecase) Report(filter dto.InvoiceInputReportFilterRequest) (string, error) {
	ctx := context.Background()
	invoices, err := u.q.ListInvoiceInputReportFilterData(ctx, db.ListInvoiceInputReportFilterDataParams{
		ProjectID: pgInt8(filter.ProjectID),
		Column2:   filter.Code,
		Column3:   int64(filter.ReleasedID),
		Column4:   int64(filter.WarehouseManagerID),
		Column5:   pgTimestamptz(filter.DateFrom),
		Column6:   pgTimestamptz(filter.DateTo),
	})
	if err != nil {
		return "", err
	}

	templateFilePath := filepath.Join("./internal/templates/", "Invoice Input Report.xlsx")
	f, err := excelize.OpenFile(templateFilePath)
	if err != nil {
		return "", err
	}
	sheetName := "Sheet1"

	f.SetCellStr(sheetName, "J1", "ID материала")

	rowCount := 2
	for _, invoice := range invoices {
		invoiceMaterials, err := u.q.ListInvoiceMaterialsDataForReport(ctx, db.ListInvoiceMaterialsDataForReportParams{
			InvoiceType: pgText("input"),
			InvoiceID:   pgInt8(uint(invoice.ID)),
		})
		if err != nil {
			return "", err
		}

		for _, im := range invoiceMaterials {
			f.SetCellStr(sheetName, "A"+fmt.Sprint(rowCount), invoice.DeliveryCode)
			f.SetCellStr(sheetName, "B"+fmt.Sprint(rowCount), invoice.WarehouseManagerName)
			f.SetCellStr(sheetName, "C"+fmt.Sprint(rowCount), invoice.ReleasedName)

			dateOfInvoice := timeFromPgTimestamptz(invoice.DateOfInvoice).String()
			if len(dateOfInvoice) > 10 {
				dateOfInvoice = dateOfInvoice[:len(dateOfInvoice)-10]
			}
			f.SetCellStr(sheetName, "D"+fmt.Sprint(rowCount), dateOfInvoice)

			f.SetCellValue(sheetName, "E"+fmt.Sprint(rowCount), im.MaterialName)
			f.SetCellValue(sheetName, "F"+fmt.Sprint(rowCount), im.MaterialUnit)
			f.SetCellFloat(sheetName, "G"+fmt.Sprint(rowCount), float64FromPgNumeric(im.InvoiceMaterialAmount), 2, 64)

			costM19, _ := decimalFromPgNumeric(im.MaterialCostM19).Float64()
			f.SetCellFloat(sheetName, "H"+fmt.Sprint(rowCount), costM19, 2, 64)
			f.SetCellValue(sheetName, "I"+fmt.Sprint(rowCount), im.InvoiceMaterialNotes)
			f.SetCellInt(sheetName, "J"+fmt.Sprint(rowCount), int(im.MaterialID))
			rowCount++
		}
	}

	currentTime := time.Now()
	fileName := fmt.Sprintf(
		"Отсчет накладной приход - %s.xlsx",
		currentTime.Format("02-01-2006"),
	)

	tempFilePath := filepath.Join("./storage/import_excel/temp/", fileName)
	f.SaveAs(tempFilePath)
	if err := f.Close(); err != nil {
		fmt.Println(err)
	}

	return fileName, nil
}

func (u *invoiceInputUsecase) NewMaterialCost(data model.MaterialCost) error {
	_, err := u.q.CreateMaterialCost(context.Background(), db.CreateMaterialCostParams{
		MaterialID:       pgInt8(data.MaterialID),
		CostPrime:        pgNumericFromDecimal(data.CostPrime),
		CostM19:          pgNumericFromDecimal(data.CostM19),
		CostWithCustomer: pgNumericFromDecimal(data.CostWithCustomer),
	})
	return err
}

func (u *invoiceInputUsecase) NewMaterialAndItsCost(data dto.NewMaterialDataFromInvoiceInput) error {
	ctx := context.Background()
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	qtx := u.q.WithTx(tx)

	materialRow, err := qtx.CreateMaterial(ctx, db.CreateMaterialParams{
		Category:                  pgText(data.Category),
		Code:                      pgText(data.Code),
		Name:                      pgText(data.Name),
		Unit:                      pgText(data.Unit),
		Notes:                     pgText(data.Notes),
		HasSerialNumber:           pgBool(data.HasSerialNumber),
		Article:                   pgText(data.Article),
		ProjectID:                 pgInt8(data.ProjectID),
		PlannedAmountForProject:   pgNumericFromFloat64(0),
		ShowPlannedAmountInReport: pgBool(false),
	})
	if err != nil {
		return err
	}

	if _, err := qtx.CreateMaterialCost(ctx, db.CreateMaterialCostParams{
		MaterialID:       pgInt8(uint(materialRow.ID)),
		CostPrime:        pgNumericFromDecimal(data.CostPrime),
		CostM19:          pgNumericFromDecimal(data.CostM19),
		CostWithCustomer: pgNumericFromDecimal(data.CostWithCustomer),
	}); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (u *invoiceInputUsecase) GetMaterialsForEdit(id uint) ([]dto.InvoiceInputMaterialForEdit, error) {
	rows, err := u.q.ListInvoiceInputMaterialsForEdit(context.Background(), pgInt8(id))
	if err != nil {
		return nil, err
	}
	out := make([]dto.InvoiceInputMaterialForEdit, len(rows))
	for i, r := range rows {
		costM19, _ := decimalFromPgNumeric(r.MaterialCost).Float64()
		out[i] = dto.InvoiceInputMaterialForEdit{
			MaterialID:      uint(r.MaterialID),
			MaterialName:    r.MaterialName,
			Unit:            r.Unit,
			Amount:          float64FromPgNumeric(r.Amount),
			MaterialCostID:  uint(r.MaterialCostID),
			MaterialCost:    costM19,
			Notes:           r.Notes,
			HasSerialNumber: r.HasSerialNumber,
		}
	}
	return out, nil
}

func (u *invoiceInputUsecase) Import(filePath string, projectID uint, workerID uint) error {
	ctx := context.Background()
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		f.Close()
		os.Remove(filePath)
		return fmt.Errorf("Не смог открыть файл: %v", err)
	}

	sheetName := "Sheet1"
	rows, err := f.GetRows(sheetName)
	if err != nil {
		f.Close()
		os.Remove(filePath)
		return fmt.Errorf("Не смог найти таблицу 'Импорт': %v", err)
	}

	if len(rows) == 1 {
		f.Close()
		os.Remove(filePath)
		return fmt.Errorf("Файл не имеет данных")
	}

	count, err := u.q.GetInvoiceCount(ctx, db.GetInvoiceCountParams{
		InvoiceType: pgText("input"),
		ProjectID:   pgInt8(projectID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		count = 0
	} else if err != nil {
		f.Close()
		os.Remove(filePath)
		return fmt.Errorf("Файл не имеет данных")
	}

	index := 1
	importData := []dto.InvoiceInputImportData{}
	currentInvoiceInput := model.InvoiceInput{}
	currentInvoiceMaterials := []model.InvoiceMaterials{}
	for len(rows) > index {
		excelInvoiceInput := model.InvoiceInput{
			ID:               0,
			ProjectID:        projectID,
			ReleasedWorkerID: workerID,
			Confirmed:        false,
			Notes:            "",
		}

		warehouseManagerName, err := f.GetCellValue(sheetName, "B"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filePath)
			return fmt.Errorf("Нету данных в ячейке B%v: %v", index+1, err)
		}

		warehouseManager, err := u.q.GetWorkerByName(ctx, pgText(warehouseManagerName))
		if err != nil {
			f.Close()
			os.Remove(filePath)
			return fmt.Errorf("Имя %v в ячейке B%v не найдено в базе: %v", warehouseManagerName, index+1, err)
		}
		excelInvoiceInput.WarehouseManagerWorkerID = uint(warehouseManager.ID)

		dateOfInvoiceInExcel, err := f.GetCellValue(sheetName, "D"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filePath)
			return fmt.Errorf("Нету данных в ячейке D%v: %v", index+1, err)
		}

		dateLayout := "2006/01/02"
		dateOfInvoice, err := time.Parse(dateLayout, dateOfInvoiceInExcel)
		if err != nil {
			f.Close()
			os.Remove(filePath)
			return fmt.Errorf("Неправильные данные в ячейке D%v: %v", index+1, err)
		}

		excelInvoiceInput.DateOfInvoice = dateOfInvoice
		if index == 1 {
			currentInvoiceInput = excelInvoiceInput
		}

		excelInvoiceMaterial := model.InvoiceMaterials{
			InvoiceType: "input",
			IsDefected:  false,
			ProjectID:   projectID,
		}

		materialName, err := f.GetCellValue(sheetName, "E"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filePath)
			return fmt.Errorf("Нету данных в ячейке E%v: %v", index+1, err)
		}

		material, err := u.q.GetMaterialByProjectAndName(ctx, db.GetMaterialByProjectAndNameParams{
			ProjectID: pgInt8(projectID),
			Name:      pgText(materialName),
		})
		if err != nil {
			f.Close()
			os.Remove(filePath)
			return fmt.Errorf("Материал %v в ячейке E%v не найдено в базе: %v", warehouseManagerName, index+1, err)
		}

		materialCosts, err := u.q.ListMaterialCostsByMaterialID(ctx, pgInt8(uint(material.ID)))
		if err != nil || len(materialCosts) == 0 {
			f.Close()
			os.Remove(filePath)
			return fmt.Errorf("Цена Материала %v в ячейке E%v не найдено в базе: %v", warehouseManagerName, index+1, err)
		}

		excelInvoiceMaterial.MaterialCostID = uint(materialCosts[0].ID)

		amountExcel, err := f.GetCellValue(sheetName, "G"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filePath)
			return fmt.Errorf("Нету данных в ячейке G%v: %v", index+1, err)
		}

		amount, err := strconv.ParseFloat(amountExcel, 64)
		if err != nil {
			f.Close()
			os.Remove(filePath)
			return fmt.Errorf("Нету данных в ячейке G%v: %v", index+1, err)
		}

		excelInvoiceMaterial.Amount = amount

		if currentInvoiceInput.DateOfInvoice.Equal(excelInvoiceInput.DateOfInvoice) {
			currentInvoiceMaterials = append(currentInvoiceMaterials, excelInvoiceMaterial)
		} else {
			currentInvoiceInput.DeliveryCode = utils.UniqueCodeGeneration("П", count, projectID)
			count++
			importData = append(importData, dto.InvoiceInputImportData{
				Details: currentInvoiceInput,
				Items:   currentInvoiceMaterials,
			})

			currentInvoiceInput = excelInvoiceInput
			currentInvoiceMaterials = []model.InvoiceMaterials{excelInvoiceMaterial}
		}

		index++
	}

	currentInvoiceInput.DeliveryCode = utils.UniqueCodeGeneration("П", count, projectID)
	importData = append(importData, dto.InvoiceInputImportData{
		Details: currentInvoiceInput,
		Items:   currentInvoiceMaterials,
	})

	return u.importInBatches(importData, projectID)
}

func (u *invoiceInputUsecase) importInBatches(data []dto.InvoiceInputImportData, projectID uint) error {
	ctx := context.Background()
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	qtx := u.q.WithTx(tx)

	for _, invoice := range data {
		invoiceRow, err := qtx.CreateInvoiceInput(ctx, db.CreateInvoiceInputParams{
			ProjectID:                pgInt8(invoice.Details.ProjectID),
			WarehouseManagerWorkerID: pgInt8(invoice.Details.WarehouseManagerWorkerID),
			ReleasedWorkerID:         pgInt8(invoice.Details.ReleasedWorkerID),
			DeliveryCode:             pgText(invoice.Details.DeliveryCode),
			Notes:                    pgText(invoice.Details.Notes),
			DateOfInvoice:            pgTimestamptz(invoice.Details.DateOfInvoice),
			Confirmed:                pgBool(invoice.Details.Confirmed),
		})
		if err != nil {
			return err
		}

		batch := make([]db.CreateInvoiceMaterialsBatchParams, len(invoice.Items))
		for i, item := range invoice.Items {
			batch[i] = db.CreateInvoiceMaterialsBatchParams{
				ProjectID:      pgInt8(item.ProjectID),
				MaterialCostID: pgInt8(item.MaterialCostID),
				InvoiceID:      pgInt8(uint(invoiceRow.ID)),
				InvoiceType:    pgText(item.InvoiceType),
				IsDefected:     pgBool(item.IsDefected),
				Amount:         pgNumericFromFloat64(item.Amount),
				Notes:          pgText(item.Notes),
			}
		}
		if _, err := qtx.CreateInvoiceMaterialsBatch(ctx, batch); err != nil {
			return err
		}
	}

	if err := qtx.IncrementInvoiceCountBy(ctx, db.IncrementInvoiceCountByParams{
		InvoiceType: pgText("input"),
		ProjectID:   pgInt8(projectID),
		Amount:      int64(len(data)),
	}); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (u *invoiceInputUsecase) GetParametersForSearch(projectID uint) (dto.InvoiceInputParametersForSearch, error) {
	ctx := context.Background()
	deliveryCodes, err := u.q.ListInvoiceInputAllDeliveryCodes(ctx, pgInt8(projectID))
	if err != nil {
		return dto.InvoiceInputParametersForSearch{}, err
	}

	whRows, err := u.q.ListInvoiceInputAllWarehouseManagers(ctx, pgInt8(projectID))
	if err != nil {
		return dto.InvoiceInputParametersForSearch{}, err
	}
	warehouseManagers := make([]dto.DataForSelect[uint], len(whRows))
	for i, r := range whRows {
		warehouseManagers[i] = dto.DataForSelect[uint]{Label: r.Label, Value: uint(r.Value)}
	}

	relRows, err := u.q.ListInvoiceInputAllReleasedWorkers(ctx, pgInt8(projectID))
	if err != nil {
		return dto.InvoiceInputParametersForSearch{}, err
	}
	releaseds := make([]dto.DataForSelect[uint], len(relRows))
	for i, r := range relRows {
		releaseds[i] = dto.DataForSelect[uint]{Label: r.Label, Value: uint(r.Value)}
	}

	matRows, err := u.q.ListInvoiceInputAllMaterialsThatExist(ctx, pgInt8(projectID))
	if err != nil {
		return dto.InvoiceInputParametersForSearch{}, err
	}
	materials := make([]dto.DataForSelect[uint], len(matRows))
	for i, r := range matRows {
		materials[i] = dto.DataForSelect[uint]{Label: r.Label, Value: uint(r.Value)}
	}

	return dto.InvoiceInputParametersForSearch{
		DeliveryCodes:     deliveryCodes,
		WarehouseManagers: warehouseManagers,
		Releaseds:         releaseds,
		Materials:         materials,
	}, nil
}

func toModelInvoiceInput(r db.InvoiceInput) model.InvoiceInput {
	return model.InvoiceInput{
		ID:                       uint(r.ID),
		ProjectID:                uintFromPgInt8(r.ProjectID),
		WarehouseManagerWorkerID: uintFromPgInt8(r.WarehouseManagerWorkerID),
		ReleasedWorkerID:         uintFromPgInt8(r.ReleasedWorkerID),
		DeliveryCode:             r.DeliveryCode.String,
		Notes:                    r.Notes.String,
		DateOfInvoice:            timeFromPgTimestamptz(r.DateOfInvoice),
		Confirmed:                r.Confirmed.Bool,
	}
}
