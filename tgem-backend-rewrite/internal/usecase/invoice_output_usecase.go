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

type invoiceOutputUsecase struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

func NewInvoiceOutputUsecase(pool *pgxpool.Pool) IInvoiceOutputUsecase {
	return &invoiceOutputUsecase{
		pool: pool,
		q:    db.New(pool),
	}
}

type IInvoiceOutputUsecase interface {
	GetAll() ([]model.InvoiceOutput, error)
	GetPaginated(page, limit int, data model.InvoiceOutput) ([]dto.InvoiceOutputPaginated, error)
	GetByID(id uint) (model.InvoiceOutput, error)
	GetDocument(deliveryCode string, projectID uint) (string, error)
	GetInvoiceMaterialsWithoutSerialNumbers(id, projectID uint) ([]dto.InvoiceMaterialsWithoutSerialNumberView, error)
	GetInvoiceMaterialsWithSerialNumbers(id, projectID uint) ([]dto.InvoiceMaterialsWithSerialNumberView, error)
	Create(data dto.InvoiceOutput) (model.InvoiceOutput, error)
	Update(data dto.InvoiceOutput) (model.InvoiceOutput, error)
	Delete(id uint) error
	Count(projectID uint) (int64, error)
	Confirmation(id uint) error
	UniqueCode(projectID uint) ([]dto.DataForSelect[string], error)
	UniqueWarehouseManager(projectID uint) ([]dto.DataForSelect[uint], error)
	UniqueRecieved(projectID uint) ([]dto.DataForSelect[uint], error)
	UniqueDistrict(projectID uint) ([]dto.DataForSelect[uint], error)
	UniqueTeam(projectID uint) ([]dto.DataForSelect[uint], error)
	Report(filter dto.InvoiceOutputReportFilterRequest) (string, error)
	GetTotalMaterialAmount(projectID, materialID uint) (float64, error)
	GetSerialNumbersByMaterial(projectID, materialID uint) ([]string, error)
	GetAvailableMaterialsInWarehouse(projectID uint) ([]dto.AvailableMaterialsInWarehouse, error)
	GetMaterialsForEdit(id, projectID uint) ([]dto.InvoiceOutputMaterialsForEdit, error)
	Import(filePath string, projectID uint, workerID uint) error
}

func (u *invoiceOutputUsecase) GetAll() ([]model.InvoiceOutput, error) {
	rows, err := u.q.ListInvoiceOutputs(context.Background())
	if err != nil {
		return nil, err
	}
	out := make([]model.InvoiceOutput, len(rows))
	for i, r := range rows {
		out[i] = toModelInvoiceOutput(r)
	}
	return out, nil
}

func (u *invoiceOutputUsecase) GetByID(id uint) (model.InvoiceOutput, error) {
	row, err := u.q.GetInvoiceOutput(context.Background(), int64(id))
	if err != nil {
		return model.InvoiceOutput{}, err
	}
	return toModelInvoiceOutput(row), nil
}

func (u *invoiceOutputUsecase) GetPaginated(page, limit int, data model.InvoiceOutput) ([]dto.InvoiceOutputPaginated, error) {
	rows, err := u.q.ListInvoiceOutputsPaginatedFiltered(context.Background(), db.ListInvoiceOutputsPaginatedFilteredParams{
		ProjectID: pgInt8(data.ProjectID),
		Column2:   int64(data.DistrictID),
		Column3:   int64(data.WarehouseManagerWorkerID),
		Column4:   int64(data.ReleasedWorkerID),
		Column5:   int64(data.RecipientWorkerID),
		Column6:   int64(data.TeamID),
		Column7:   data.DeliveryCode,
		Limit:     int32(limit),
		Offset:    int32((page - 1) * limit),
	})
	if err != nil {
		return nil, err
	}

	out := make([]dto.InvoiceOutputPaginated, len(rows))
	for i, r := range rows {
		out[i] = dto.InvoiceOutputPaginated{
			ID:                   uint(r.ID),
			DeliveryCode:         r.DeliveryCode,
			DistrictID:           uint(r.DistrictID),
			DistrictName:         r.DistrictName,
			TeamID:               uint(r.TeamID),
			TeamName:             r.TeamName,
			WarehouseManagerID:   uint(r.WarehouseManagerID),
			WarehouseManagerName: r.WarehouseManagerName,
			ReleasedName:         r.ReleasedName,
			RecipientID:          uint(r.RecipientID),
			RecipientName:        r.RecipientName,
			DateOfInvoice:        timeFromPgTimestamptz(r.DateOfInvoice),
			Confirmation:         r.Confirmation,
			Notes:                r.Notes,
		}
	}
	return out, nil
}

func (u *invoiceOutputUsecase) GetInvoiceMaterialsWithoutSerialNumbers(id, projectID uint) ([]dto.InvoiceMaterialsWithoutSerialNumberView, error) {
	rows, err := u.q.ListInvoiceMaterialsWithoutSerialNumbers(context.Background(), db.ListInvoiceMaterialsWithoutSerialNumbersParams{
		InvoiceType: pgText("output"),
		InvoiceID:   pgInt8(id),
		ProjectID:   pgInt8(projectID),
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

func (u *invoiceOutputUsecase) GetInvoiceMaterialsWithSerialNumbers(id, projectID uint) ([]dto.InvoiceMaterialsWithSerialNumberView, error) {
	rows, err := u.q.ListInvoiceMaterialsWithSerialNumbers(context.Background(), db.ListInvoiceMaterialsWithSerialNumbersParams{
		InvoiceType: pgText("output"),
		InvoiceID:   pgInt8(id),
		ProjectID:   pgInt8(projectID),
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

func (u *invoiceOutputUsecase) Create(data dto.InvoiceOutput) (model.InvoiceOutput, error) {
	ctx := context.Background()

	count, err := u.q.GetInvoiceCount(ctx, db.GetInvoiceCountParams{
		InvoiceType: pgText("output"),
		ProjectID:   pgInt8(data.Details.ProjectID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		count = 0
	} else if err != nil {
		return model.InvoiceOutput{}, err
	}

	data.Details.DeliveryCode = utils.UniqueCodeGeneration("О", count+1, data.Details.ProjectID)

	invoiceMaterialForCreate, serialNumberMovements, err := u.buildInvoiceOutputItems(ctx, data)
	if err != nil {
		return model.InvoiceOutput{}, err
	}

	if err := u.GenerateExcelFile(data); err != nil {
		return model.InvoiceOutput{}, err
	}

	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return model.InvoiceOutput{}, err
	}
	defer tx.Rollback(ctx)
	qtx := u.q.WithTx(tx)

	invoiceRow, err := qtx.CreateInvoiceOutput(ctx, db.CreateInvoiceOutputParams{
		DistrictID:               pgInt8(data.Details.DistrictID),
		ProjectID:                pgInt8(data.Details.ProjectID),
		WarehouseManagerWorkerID: pgInt8(data.Details.WarehouseManagerWorkerID),
		ReleasedWorkerID:         pgInt8(data.Details.ReleasedWorkerID),
		RecipientWorkerID:        pgInt8(data.Details.RecipientWorkerID),
		TeamID:                   pgInt8(data.Details.TeamID),
		DeliveryCode:             pgText(data.Details.DeliveryCode),
		DateOfInvoice:            pgTimestamptz(data.Details.DateOfInvoice),
		Notes:                    pgText(data.Details.Notes),
		Confirmation:             pgBool(data.Details.Confirmation),
	})
	if err != nil {
		return model.InvoiceOutput{}, err
	}

	if err := writeInvoiceOutputItems(ctx, qtx, uint(invoiceRow.ID), invoiceMaterialForCreate, serialNumberMovements); err != nil {
		return model.InvoiceOutput{}, err
	}

	if err := qtx.IncrementInvoiceCount(ctx, db.IncrementInvoiceCountParams{
		InvoiceType: pgText("output"),
		ProjectID:   invoiceRow.ProjectID,
	}); err != nil {
		return model.InvoiceOutput{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return model.InvoiceOutput{}, err
	}
	return toModelInvoiceOutput(invoiceRow), nil
}

func (u *invoiceOutputUsecase) Update(data dto.InvoiceOutput) (model.InvoiceOutput, error) {
	ctx := context.Background()
	invoiceMaterialForCreate, serialNumberMovements, err := u.buildInvoiceOutputItems(ctx, data)
	if err != nil {
		return model.InvoiceOutput{}, err
	}

	excelFilePath := filepath.Join("./storage/import_excel/output/", data.Details.DeliveryCode+".xlsx")
	if err := os.Remove(excelFilePath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return model.InvoiceOutput{}, err
	}

	if err := u.GenerateExcelFile(data); err != nil {
		return model.InvoiceOutput{}, err
	}

	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return model.InvoiceOutput{}, err
	}
	defer tx.Rollback(ctx)
	qtx := u.q.WithTx(tx)

	invoiceRow, err := qtx.UpdateInvoiceOutput(ctx, db.UpdateInvoiceOutputParams{
		ID:                       int64(data.Details.ID),
		DistrictID:               pgInt8(data.Details.DistrictID),
		ProjectID:                pgInt8(data.Details.ProjectID),
		WarehouseManagerWorkerID: pgInt8(data.Details.WarehouseManagerWorkerID),
		ReleasedWorkerID:         pgInt8(data.Details.ReleasedWorkerID),
		RecipientWorkerID:        pgInt8(data.Details.RecipientWorkerID),
		TeamID:                   pgInt8(data.Details.TeamID),
		DeliveryCode:             pgText(data.Details.DeliveryCode),
		DateOfInvoice:            pgTimestamptz(data.Details.DateOfInvoice),
		Notes:                    pgText(data.Details.Notes),
		Confirmation:             pgBool(data.Details.Confirmation),
	})
	if err != nil {
		return model.InvoiceOutput{}, err
	}

	if err := qtx.DeleteInvoiceMaterialsByInvoice(ctx, db.DeleteInvoiceMaterialsByInvoiceParams{
		InvoiceType: pgText("output"),
		InvoiceID:   pgInt8(uint(invoiceRow.ID)),
	}); err != nil {
		return model.InvoiceOutput{}, err
	}

	if err := qtx.DeleteSerialNumberMovementsByInvoice(ctx, db.DeleteSerialNumberMovementsByInvoiceParams{
		InvoiceType: pgText("output"),
		InvoiceID:   pgInt8(uint(invoiceRow.ID)),
	}); err != nil {
		return model.InvoiceOutput{}, err
	}

	if err := writeInvoiceOutputItems(ctx, qtx, uint(invoiceRow.ID), invoiceMaterialForCreate, serialNumberMovements); err != nil {
		return model.InvoiceOutput{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return model.InvoiceOutput{}, err
	}
	return toModelInvoiceOutput(invoiceRow), nil
}

// buildInvoiceOutputItems mirrors the GORM-era logic that splits an output
// invoice's logical items into per-cost invoice_materials rows and the
// matching serial_number_movements rows. Reads only — no writes — so the
// caller can run them inside its own transaction.
func (u *invoiceOutputUsecase) buildInvoiceOutputItems(ctx context.Context, data dto.InvoiceOutput) ([]model.InvoiceMaterials, []model.SerialNumberMovement, error) {
	invoiceMaterialForCreate := []model.InvoiceMaterials{}
	serialNumberMovements := []model.SerialNumberMovement{}

	for _, item := range data.Items {
		if len(item.SerialNumbers) == 0 {
			rows, err := u.q.ListMaterialAmountSortedByCostM19InLocation(ctx, db.ListMaterialAmountSortedByCostM19InLocationParams{
				ProjectID:    pgInt8(data.Details.ProjectID),
				LocationType: pgText("warehouse"),
				LocationID:   pgInt8(0),
				ID:           int64(item.MaterialID),
			})
			if err != nil {
				return nil, nil, err
			}

			index := 0
			amountLeft := item.Amount
			for amountLeft > 0 {
				materialAmount := float64FromPgNumeric(rows[index].MaterialAmount)
				invoiceMaterialCreate := model.InvoiceMaterials{
					ProjectID:      data.Details.ProjectID,
					MaterialCostID: uint(rows[index].MaterialCostID),
					InvoiceType:    "output",
					Notes:          item.Notes,
				}
				if materialAmount <= amountLeft {
					invoiceMaterialCreate.Amount = materialAmount
					amountLeft -= materialAmount
				} else {
					invoiceMaterialCreate.Amount = amountLeft
					amountLeft = 0
				}
				invoiceMaterialForCreate = append(invoiceMaterialForCreate, invoiceMaterialCreate)
				index++
			}
			continue
		}

		mcSnRows, err := u.q.ListMaterialCostIDAndSerialNumberIDByCodes(ctx, db.ListMaterialCostIDAndSerialNumberIDByCodesParams{
			ID:           int64(item.MaterialID),
			LocationType: pgText("warehouse"),
			LocationID:   pgInt8(0),
			Codes:        item.SerialNumbers,
		})
		if err != nil {
			return nil, nil, err
		}

		var invoiceMaterialCreate model.InvoiceMaterials
		for index, oneEntry := range mcSnRows {
			serialNumberMovements = append(serialNumberMovements, model.SerialNumberMovement{
				SerialNumberID: uint(oneEntry.SerialNumberID),
				ProjectID:      data.Details.ProjectID,
				InvoiceType:    "output",
			})

			if index == 0 {
				invoiceMaterialCreate = model.InvoiceMaterials{
					ProjectID:      data.Details.ProjectID,
					MaterialCostID: uint(oneEntry.MaterialCostID),
					InvoiceType:    "output",
					Notes:          item.Notes,
				}
			}

			if uint(oneEntry.MaterialCostID) == invoiceMaterialCreate.MaterialCostID {
				invoiceMaterialCreate.Amount++
			} else {
				invoiceMaterialForCreate = append(invoiceMaterialForCreate, invoiceMaterialCreate)
				invoiceMaterialCreate = model.InvoiceMaterials{
					ProjectID:      data.Details.ProjectID,
					MaterialCostID: uint(oneEntry.MaterialCostID),
					InvoiceType:    "output",
					Notes:          item.Notes,
				}
			}
		}
		invoiceMaterialForCreate = append(invoiceMaterialForCreate, invoiceMaterialCreate)
	}

	// GORM-era de-dup: ensure each MaterialCostID appears at most once.
	correct := []model.InvoiceMaterials{}
	for _, entry := range invoiceMaterialForCreate {
		if len(correct) == 0 {
			correct = append(correct, entry)
			continue
		}
		duplicate := false
		for _, c := range correct {
			if entry.MaterialCostID == c.MaterialCostID {
				duplicate = true
				break
			}
		}
		if !duplicate {
			correct = append(correct, entry)
		}
	}
	return correct, serialNumberMovements, nil
}

func writeInvoiceOutputItems(ctx context.Context, qtx *db.Queries, invoiceID uint, materials []model.InvoiceMaterials, movements []model.SerialNumberMovement) error {
	if len(materials) > 0 {
		batch := make([]db.CreateInvoiceMaterialsBatchParams, len(materials))
		for i, m := range materials {
			batch[i] = db.CreateInvoiceMaterialsBatchParams{
				ProjectID:      pgInt8(m.ProjectID),
				MaterialCostID: pgInt8(m.MaterialCostID),
				InvoiceID:      pgInt8(invoiceID),
				InvoiceType:    pgText(m.InvoiceType),
				IsDefected:     pgBool(m.IsDefected),
				Amount:         pgNumericFromFloat64(m.Amount),
				Notes:          pgText(m.Notes),
			}
		}
		if _, err := qtx.CreateInvoiceMaterialsBatch(ctx, batch); err != nil {
			return err
		}
	}

	if len(movements) > 0 {
		batch := make([]db.CreateSerialNumberMovementsBatchParams, len(movements))
		for i, m := range movements {
			batch[i] = db.CreateSerialNumberMovementsBatchParams{
				SerialNumberID: pgInt8(m.SerialNumberID),
				ProjectID:      pgInt8(m.ProjectID),
				InvoiceID:      pgInt8(invoiceID),
				InvoiceType:    pgText(m.InvoiceType),
				IsDefected:     pgBool(m.IsDefected),
				Confirmation:   pgBool(m.Confirmation),
			}
		}
		if _, err := qtx.CreateSerialNumberMovementsBatch(ctx, batch); err != nil {
			return err
		}
	}
	return nil
}

func (u *invoiceOutputUsecase) Delete(id uint) error {
	ctx := context.Background()
	invoiceOutput, err := u.q.GetInvoiceOutput(ctx, int64(id))
	if err != nil {
		return err
	}

	excelFilePath := filepath.Join("./storage/import_excel/output/", invoiceOutput.DeliveryCode.String+".xlsx")
	if err := os.Remove(excelFilePath); err != nil {
		return err
	}

	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	qtx := u.q.WithTx(tx)
	if err := qtx.DeleteInvoiceOutput(ctx, int64(id)); err != nil {
		return err
	}
	if err := qtx.DeleteInvoiceMaterialsByInvoice(ctx, db.DeleteInvoiceMaterialsByInvoiceParams{
		InvoiceType: pgText("output"),
		InvoiceID:   pgInt8(id),
	}); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (u *invoiceOutputUsecase) Count(projectID uint) (int64, error) {
	return u.q.CountInvoiceOutputsByProject(context.Background(), pgInt8(projectID))
}

func (u *invoiceOutputUsecase) Confirmation(id uint) error {
	ctx := context.Background()

	invoiceOutput, err := u.q.GetInvoiceOutput(ctx, int64(id))
	if err != nil {
		return err
	}

	invoiceMaterials, err := u.q.ListInvoiceMaterialsByInvoice(ctx, db.ListInvoiceMaterialsByInvoiceParams{
		InvoiceID:   pgInt8(id),
		InvoiceType: pgText("output"),
		ProjectID:   invoiceOutput.ProjectID,
	})
	if err != nil {
		return err
	}

	materialsInWarehouse, err := u.q.ListMaterialLocationsForInvoiceConfirmation(ctx, db.ListMaterialLocationsForInvoiceConfirmationParams{
		LocationType: pgText("warehouse"),
		LocationID:   pgInt8(0),
		InvoiceType:  pgText("output"),
		InvoiceID:    pgInt8(id),
	})
	if err != nil {
		return err
	}

	materialsInTeam, err := u.q.ListMaterialLocationsForInvoiceConfirmation(ctx, db.ListMaterialLocationsForInvoiceConfirmationParams{
		LocationType: pgText("team"),
		LocationID:   invoiceOutput.TeamID,
		InvoiceType:  pgText("output"),
		InvoiceID:    pgInt8(id),
	})
	if err != nil {
		return err
	}

	type warehouseLoc struct {
		row    db.MaterialLocation
		amount float64
	}
	warehouseSlice := make([]warehouseLoc, len(materialsInWarehouse))
	for i, ml := range materialsInWarehouse {
		warehouseSlice[i] = warehouseLoc{row: ml, amount: float64FromPgNumeric(ml.Amount)}
	}
	type teamLoc struct {
		row    *db.MaterialLocation
		params *db.CreateMaterialLocationParams
		amount float64
	}
	teamSlice := make([]*teamLoc, 0, len(materialsInTeam))
	for i := range materialsInTeam {
		ml := materialsInTeam[i]
		teamSlice = append(teamSlice, &teamLoc{row: &ml, amount: float64FromPgNumeric(ml.Amount)})
	}

	for _, im := range invoiceMaterials {
		whIndex := -1
		for index, w := range warehouseSlice {
			if uintFromPgInt8(w.row.MaterialCostID) == uintFromPgInt8(im.MaterialCostID) {
				whIndex = index
				break
			}
		}
		if whIndex == -1 {
			return fmt.Errorf("Mатериал отсутствует на складе для подтверждения накладной")
		}
		imAmount := float64FromPgNumeric(im.Amount)
		if warehouseSlice[whIndex].amount < imAmount {
			mcID := int64(uintFromPgInt8(warehouseSlice[whIndex].row.MaterialCostID))
			material, err := u.q.GetMaterialByMaterialCostID(ctx, mcID)
			if err != nil {
				return fmt.Errorf("Ошибка при подсчете материала, система не смогла разпознать материал: %v", err)
			}

			materialCost, err := u.q.GetMaterialCost(ctx, mcID)
			if err != nil {
				return fmt.Errorf("Ошибка при подсчете материала, система не смогла разпознать ценник материалa: %v", err)
			}

			return fmt.Errorf("Mатериал <<%v>> c ценником %v - указано больше чем имеется на складе. Измените количество",
				material.Name.String, decimalFromPgNumeric(materialCost.CostM19))
		}
		warehouseSlice[whIndex].amount -= imAmount

		teamIndex := -1
		for index, t := range teamSlice {
			if t.row != nil && uintFromPgInt8(t.row.MaterialCostID) == uintFromPgInt8(im.MaterialCostID) {
				teamIndex = index
				break
			}
			if t.params != nil && uintFromPgInt8(t.params.MaterialCostID) == uintFromPgInt8(im.MaterialCostID) {
				teamIndex = index
				break
			}
		}
		if teamIndex != -1 {
			teamSlice[teamIndex].amount += imAmount
		} else {
			teamSlice = append(teamSlice, &teamLoc{
				params: &db.CreateMaterialLocationParams{
					ProjectID:      invoiceOutput.ProjectID,
					MaterialCostID: im.MaterialCostID,
					LocationType:   pgText("team"),
					LocationID:     invoiceOutput.TeamID,
					Amount:         im.Amount,
				},
				amount: imAmount,
			})
		}
	}

	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	qtx := u.q.WithTx(tx)

	if err := qtx.ConfirmInvoiceOutput(ctx, int64(id)); err != nil {
		return err
	}

	for _, w := range warehouseSlice {
		if err := qtx.UpsertMaterialLocationByID(ctx, db.UpsertMaterialLocationByIDParams{
			Amount: pgNumericFromFloat64(w.amount),
			ID:     w.row.ID,
		}); err != nil {
			return err
		}
	}

	for _, t := range teamSlice {
		if t.row != nil {
			if err := qtx.UpsertMaterialLocationByID(ctx, db.UpsertMaterialLocationByIDParams{
				Amount: pgNumericFromFloat64(t.amount),
				ID:     t.row.ID,
			}); err != nil {
				return err
			}
		} else {
			params := *t.params
			params.Amount = pgNumericFromFloat64(t.amount)
			if _, err := qtx.CreateMaterialLocation(ctx, params); err != nil {
				return err
			}
		}
	}

	if err := qtx.ConfirmSerialNumberMovementsByInvoice(ctx, db.ConfirmSerialNumberMovementsByInvoiceParams{
		InvoiceID:   pgInt8(id),
		InvoiceType: pgText("output"),
	}); err != nil {
		return err
	}

	if err := qtx.ConfirmSerialNumberLocationsByOutputInvoice(ctx, db.ConfirmSerialNumberLocationsByOutputInvoiceParams{
		TeamID:    int64(uintFromPgInt8(invoiceOutput.TeamID)),
		InvoiceID: int64(id),
	}); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (u *invoiceOutputUsecase) UniqueCode(projectID uint) ([]dto.DataForSelect[string], error) {
	rows, err := u.q.ListInvoiceOutputUniqueCodes(context.Background(), pgInt8(projectID))
	if err != nil {
		return nil, err
	}
	out := make([]dto.DataForSelect[string], len(rows))
	for i, r := range rows {
		out[i] = dto.DataForSelect[string]{Label: r.Label, Value: r.Value}
	}
	return out, nil
}

func (u *invoiceOutputUsecase) UniqueWarehouseManager(projectID uint) ([]dto.DataForSelect[uint], error) {
	rows, err := u.q.ListInvoiceOutputUniqueWarehouseManagers(context.Background(), pgInt8(projectID))
	if err != nil {
		return nil, err
	}
	out := make([]dto.DataForSelect[uint], len(rows))
	for i, r := range rows {
		out[i] = dto.DataForSelect[uint]{Label: r.Label, Value: uint(r.Value)}
	}
	return out, nil
}

func (u *invoiceOutputUsecase) UniqueRecieved(projectID uint) ([]dto.DataForSelect[uint], error) {
	rows, err := u.q.ListInvoiceOutputUniqueRecieved(context.Background(), pgInt8(projectID))
	if err != nil {
		return nil, err
	}
	out := make([]dto.DataForSelect[uint], len(rows))
	for i, r := range rows {
		out[i] = dto.DataForSelect[uint]{Label: r.Label, Value: uint(r.Value)}
	}
	return out, nil
}

func (u *invoiceOutputUsecase) UniqueDistrict(projectID uint) ([]dto.DataForSelect[uint], error) {
	rows, err := u.q.ListInvoiceOutputUniqueDistricts(context.Background(), pgInt8(projectID))
	if err != nil {
		return nil, err
	}
	out := make([]dto.DataForSelect[uint], len(rows))
	for i, r := range rows {
		out[i] = dto.DataForSelect[uint]{Label: r.Label, Value: uint(r.Value)}
	}
	return out, nil
}

func (u *invoiceOutputUsecase) UniqueTeam(projectID uint) ([]dto.DataForSelect[uint], error) {
	rows, err := u.q.ListInvoiceOutputUniqueTeams(context.Background(), pgInt8(projectID))
	if err != nil {
		return nil, err
	}
	out := make([]dto.DataForSelect[uint], len(rows))
	for i, r := range rows {
		out[i] = dto.DataForSelect[uint]{Label: r.Label, Value: uint(r.Value)}
	}
	return out, nil
}

func (u *invoiceOutputUsecase) Report(filter dto.InvoiceOutputReportFilterRequest) (string, error) {
	ctx := context.Background()
	invoices, err := u.q.ListInvoiceOutputReportFilterData(ctx, db.ListInvoiceOutputReportFilterDataParams{
		ProjectID: pgInt8(filter.ProjectID),
		Column2:   filter.Code,
		Column3:   int64(filter.ReceivedID),
		Column4:   int64(filter.WarehouseManagerID),
		Column5:   int64(filter.TeamID),
		Column6:   pgTimestamptz(filter.DateFrom),
		Column7:   pgTimestamptz(filter.DateTo),
	})
	if err != nil {
		return "", err
	}

	templateFilePath := filepath.Join("./internal/templates/", "Invoice Output Report.xlsx")
	f, err := excelize.OpenFile(templateFilePath)
	if err != nil {
		return "", err
	}
	sheetName := "Sheet1"
	f.SetCellStr(sheetName, "L1", "ID материала")
	f.SetCellStr(sheetName, "M1", "Код материала")

	rowCount := 2
	for _, invoice := range invoices {
		invoiceMaterials, err := u.q.ListInvoiceOutputMaterialDataForReport(ctx, pgInt8(uint(invoice.ID)))
		if err != nil {
			return "", err
		}

		for _, im := range invoiceMaterials {
			f.SetCellStr(sheetName, "A"+fmt.Sprint(rowCount), invoice.DeliveryCode)
			f.SetCellStr(sheetName, "B"+fmt.Sprint(rowCount), invoice.WarehouseManagerName)
			f.SetCellStr(sheetName, "C"+fmt.Sprint(rowCount), invoice.RecipientName)
			f.SetCellStr(sheetName, "D"+fmt.Sprint(rowCount), invoice.TeamNumber)
			f.SetCellStr(sheetName, "E"+fmt.Sprint(rowCount), invoice.TeamLeaderName)

			dateOfInvoice := timeFromPgTimestamptz(invoice.DateOfInvoice).String()
			if len(dateOfInvoice) > 10 {
				dateOfInvoice = dateOfInvoice[:len(dateOfInvoice)-10]
			}
			f.SetCellValue(sheetName, "F"+fmt.Sprint(rowCount), dateOfInvoice)

			f.SetCellStr(sheetName, "G"+fmt.Sprint(rowCount), im.MaterialName)
			f.SetCellStr(sheetName, "H"+fmt.Sprint(rowCount), im.MaterialUnit)
			f.SetCellFloat(sheetName, "I"+fmt.Sprint(rowCount), float64FromPgNumeric(im.Amount), 2, 64)

			materialCost, _ := decimalFromPgNumeric(im.MaterialCostM19).Float64()
			f.SetCellFloat(sheetName, "J"+fmt.Sprint(rowCount), materialCost, 2, 64)
			f.SetCellValue(sheetName, "K"+fmt.Sprint(rowCount), im.Notes)
			f.SetCellInt(sheetName, "L"+fmt.Sprint(rowCount), int(im.MaterialID))
			f.SetCellStr(sheetName, "M"+fmt.Sprint(rowCount), im.MaterialCode)
			rowCount++
		}
	}

	currentTime := time.Now()
	fileName := fmt.Sprintf(
		"Отсчет накладной отпуск - %s.xlsx",
		currentTime.Format("02-01-2006"),
	)

	tempFilePath := filepath.Join("./storage/import_excel/temp/", fileName)
	f.SaveAs(tempFilePath)
	if err := f.Close(); err != nil {
		fmt.Println(err)
	}

	return fileName, nil
}

func (u *invoiceOutputUsecase) GetTotalMaterialAmount(projectID, materialID uint) (float64, error) {
	amount, err := u.q.GetTotalAmountInWarehouse(context.Background(), db.GetTotalAmountInWarehouseParams{
		ProjectID: pgInt8(projectID),
		ID:        int64(materialID),
	})
	if err != nil {
		return 0, err
	}
	return float64FromPgNumeric(amount), nil
}

func (u *invoiceOutputUsecase) GetSerialNumbersByMaterial(projectID, materialID uint) ([]string, error) {
	return u.q.ListSerialNumberCodesByMaterialID(context.Background(), db.ListSerialNumberCodesByMaterialIDParams{
		ProjectID:    pgInt8(projectID),
		ID:           int64(materialID),
		LocationType: pgText("warehouse"),
	})
}

func (u *invoiceOutputUsecase) GetAvailableMaterialsInWarehouse(projectID uint) ([]dto.AvailableMaterialsInWarehouse, error) {
	rows, err := u.q.ListInvoiceOutputAvailableMaterialsInWarehouse(context.Background(), pgInt8(projectID))
	if err != nil {
		return nil, err
	}

	data := make([]dto.AvailableMaterialsInWarehouse, len(rows))
	for i, r := range rows {
		data[i] = dto.AvailableMaterialsInWarehouse{
			ID:              uint(r.ID),
			Name:            r.Name,
			Unit:            r.Unit,
			HasSerialNumber: r.HasSerialNumber,
			Amount:          float64FromPgNumeric(r.Amount),
		}
	}

	result := []dto.AvailableMaterialsInWarehouse{}
	currentMaterial := dto.AvailableMaterialsInWarehouse{}
	for index, oneEntry := range data {
		if currentMaterial.ID == oneEntry.ID {
			currentMaterial.Amount += oneEntry.Amount
		} else {
			if index != 0 {
				result = append(result, currentMaterial)
			}
			currentMaterial = oneEntry
		}
	}
	if len(data) != 0 {
		result = append(result, currentMaterial)
	}
	return result, nil
}

func (u *invoiceOutputUsecase) GetMaterialsForEdit(id, projectID uint) ([]dto.InvoiceOutputMaterialsForEdit, error) {
	rows, err := u.q.ListInvoiceOutputMaterialsForEdit(context.Background(), db.ListInvoiceOutputMaterialsForEditParams{
		InvoiceID: pgInt8(id),
		ProjectID: pgInt8(projectID),
	})
	if err != nil {
		return []dto.InvoiceOutputMaterialsForEdit{}, nil
	}

	data := make([]dto.InvoiceOutputMaterialsForEdit, len(rows))
	for i, r := range rows {
		data[i] = dto.InvoiceOutputMaterialsForEdit{
			MaterialID:      uint(r.MaterialID),
			MaterialName:    r.MaterialName,
			Unit:            r.MaterialUnit,
			WarehouseAmount: float64FromPgNumeric(r.WarehouseAmount),
			Amount:          float64FromPgNumeric(r.Amount),
			Notes:           r.Notes,
			HasSerialNumber: r.HasSerialNumber,
		}
	}

	var result []dto.InvoiceOutputMaterialsForEdit
	for index, entry := range data {
		if index == 0 {
			result = append(result, entry)
			continue
		}

		lastItemIndex := len(result) - 1
		if result[lastItemIndex].MaterialID == entry.MaterialID {
			result[lastItemIndex].Amount += entry.Amount
			result[lastItemIndex].WarehouseAmount += entry.WarehouseAmount
		} else {
			result = append(result, entry)
		}
	}

	return result, nil
}

func (u *invoiceOutputUsecase) Import(filePath string, projectID uint, workerID uint) error {
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
		InvoiceType: pgText("output"),
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
	importData := []dto.InvoiceOutputImportData{}
	currentInvoiceOutput := model.InvoiceOutput{}
	currentInvoiceMaterials := []model.InvoiceMaterials{}
	for len(rows) > index {
		excelInvoiceOutput := model.InvoiceOutput{
			ID:               0,
			ProjectID:        projectID,
			ReleasedWorkerID: workerID,
			Confirmation:     false,
			Notes:            "",
		}

		districtName, err := f.GetCellValue(sheetName, "L"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filePath)
			return fmt.Errorf("Нету данных в ячейке L%v: %v", index+1, err)
		}

		district, err := u.q.GetDistrictByName(ctx, pgText(districtName))
		if err != nil {
			f.Close()
			os.Remove(filePath)
			return fmt.Errorf("Имя %v в ячейке L%v не найдено в базе: %v", districtName, index+1, err)
		}
		excelInvoiceOutput.DistrictID = uint(district.ID)

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
		excelInvoiceOutput.WarehouseManagerWorkerID = uint(warehouseManager.ID)

		recipientName, err := f.GetCellValue(sheetName, "C"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filePath)
			return fmt.Errorf("Нету данных в ячейке C%v: %v", index+1, err)
		}

		recipient, err := u.q.GetWorkerByName(ctx, pgText(recipientName))
		if err != nil {
			f.Close()
			os.Remove(filePath)
			return fmt.Errorf("Имя %v в ячейке C%v не найдено в базе: %v", recipientName, index+1, err)
		}
		excelInvoiceOutput.RecipientWorkerID = uint(recipient.ID)

		teamNumber, err := f.GetCellValue(sheetName, "D"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filePath)
			return fmt.Errorf("Нету данных в ячейке D%v: %v", index+1, err)
		}

		team, err := u.q.GetTeamByNumber(ctx, pgText(teamNumber))
		if err != nil {
			f.Close()
			os.Remove(filePath)
			return fmt.Errorf("Имя %v в ячейке D%v не найдено в базе: %v", teamNumber, index+1, err)
		}
		excelInvoiceOutput.TeamID = uint(team.ID)

		dateOfInvoiceInExcel, err := f.GetCellValue(sheetName, "F"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filePath)
			return fmt.Errorf("Нету данных в ячейке F%v: %v", index+1, err)
		}

		dateLayout := "2006/01/02"
		dateOfInvoice, err := time.Parse(dateLayout, dateOfInvoiceInExcel)
		if err != nil {
			f.Close()
			os.Remove(filePath)
			return fmt.Errorf("Неправильные данные в ячейке F%v: %v", index+1, err)
		}
		excelInvoiceOutput.DateOfInvoice = dateOfInvoice
		if index == 1 {
			currentInvoiceOutput = excelInvoiceOutput
		}

		excelInvoiceMaterial := model.InvoiceMaterials{
			InvoiceType: "output",
			IsDefected:  false,
			ProjectID:   projectID,
		}

		materialName, err := f.GetCellValue(sheetName, "G"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filePath)
			return fmt.Errorf("Нету данных в ячейке G%v: %v", index+1, err)
		}

		material, err := u.q.GetMaterialByProjectAndName(ctx, db.GetMaterialByProjectAndNameParams{
			ProjectID: pgInt8(projectID),
			Name:      pgText(materialName),
		})
		if err != nil {
			f.Close()
			os.Remove(filePath)
			return fmt.Errorf("Материал %v в ячейке G%v не найдено в базе: %v", materialName, index+1, err)
		}

		materialCosts, err := u.q.ListMaterialCostsByMaterialID(ctx, pgInt8(uint(material.ID)))
		if err != nil || len(materialCosts) == 0 {
			f.Close()
			os.Remove(filePath)
			return fmt.Errorf("Цена Материала %v в ячейке G%v не найдено в базе: %v", materialName, index+1, err)
		}
		excelInvoiceMaterial.MaterialCostID = uint(materialCosts[0].ID)

		amountExcel, err := f.GetCellValue(sheetName, "I"+fmt.Sprint(index+1))
		if err != nil {
			f.Close()
			os.Remove(filePath)
			return fmt.Errorf("Нету данных в ячейке I%v: %v", index+1, err)
		}

		amount, err := strconv.ParseFloat(amountExcel, 64)
		if err != nil {
			f.Close()
			os.Remove(filePath)
			return fmt.Errorf("Нету данных в ячейке I%v: %v", index+1, err)
		}
		excelInvoiceMaterial.Amount = amount

		if currentInvoiceOutput.DateOfInvoice.Equal(excelInvoiceOutput.DateOfInvoice) {
			currentInvoiceMaterials = append(currentInvoiceMaterials, excelInvoiceMaterial)
		} else {
			count++
			currentInvoiceOutput.DeliveryCode = utils.UniqueCodeGeneration("O", count, projectID)
			importData = append(importData, dto.InvoiceOutputImportData{
				Details: currentInvoiceOutput,
				Items:   currentInvoiceMaterials,
			})

			currentInvoiceOutput = excelInvoiceOutput
			currentInvoiceMaterials = []model.InvoiceMaterials{excelInvoiceMaterial}
		}

		index++
	}

	currentInvoiceOutput.DeliveryCode = utils.UniqueCodeGeneration("O", count, projectID)
	importData = append(importData, dto.InvoiceOutputImportData{
		Details: currentInvoiceOutput,
		Items:   currentInvoiceMaterials,
	})

	return u.importInBatches(importData, projectID)
}

func (u *invoiceOutputUsecase) importInBatches(data []dto.InvoiceOutputImportData, projectID uint) error {
	ctx := context.Background()
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	qtx := u.q.WithTx(tx)

	for _, invoice := range data {
		invoiceRow, err := qtx.CreateInvoiceOutput(ctx, db.CreateInvoiceOutputParams{
			DistrictID:               pgInt8(invoice.Details.DistrictID),
			ProjectID:                pgInt8(invoice.Details.ProjectID),
			WarehouseManagerWorkerID: pgInt8(invoice.Details.WarehouseManagerWorkerID),
			ReleasedWorkerID:         pgInt8(invoice.Details.ReleasedWorkerID),
			RecipientWorkerID:        pgInt8(invoice.Details.RecipientWorkerID),
			TeamID:                   pgInt8(invoice.Details.TeamID),
			DeliveryCode:             pgText(invoice.Details.DeliveryCode),
			DateOfInvoice:            pgTimestamptz(invoice.Details.DateOfInvoice),
			Notes:                    pgText(invoice.Details.Notes),
			Confirmation:             pgBool(invoice.Details.Confirmation),
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
		InvoiceType: pgText("output"),
		ProjectID:   pgInt8(projectID),
		Amount:      int64(len(data)),
	}); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (u *invoiceOutputUsecase) GenerateExcelFile(data dto.InvoiceOutput) error {
	ctx := context.Background()
	templateFilePath := filepath.Join("./internal/templates/output.xlsx")
	f, err := excelize.OpenFile(templateFilePath)
	if err != nil {
		return err
	}

	sheetName := "Отпуск"
	startingRow := 5
	f.InsertRows(sheetName, startingRow, len(data.Items))

	defaultStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Size: 8, VertAlign: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		},
		Alignment: &excelize.Alignment{Horizontal: "left", WrapText: true, Vertical: "center"},
	})

	materialNamingStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Size: 8, VertAlign: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center", WrapText: true},
	})

	workerNamingStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 9, VertAlign: "center"},
		Alignment: &excelize.Alignment{Horizontal: "left", WrapText: true, Vertical: "center"},
	})

	f.SetCellValue(sheetName, "C1", fmt.Sprintf(`НАКЛАДНАЯ
№ %s
от %s года
на отпуск материала
`, data.Details.DeliveryCode, utils.DateConverter(data.Details.DateOfInvoice)))

	project, err := u.q.GetProject(ctx, int64(data.Details.ProjectID))
	if err != nil {
		return err
	}

	f.SetCellValue(sheetName, "C3", fmt.Sprintf("Отпуск разрешил: %s", project.ProjectManager.String))

	district, err := u.q.GetDistrict(ctx, int64(data.Details.DistrictID))
	if err != nil {
		return err
	}
	f.MergeCell(sheetName, "D1", "F1")
	f.SetCellStr(sheetName, "D1", fmt.Sprintf(`%s
Регион: %s `, project.Name.String, district.Name.String))

	f.SetCellStyle(sheetName, "G4", "G4", defaultStyle)
	f.SetCellStr(sheetName, "G4", "ID материала")
	for index, oneEntry := range data.Items {
		material, err := u.q.GetMaterial(ctx, int64(oneEntry.MaterialID))
		if err != nil {
			return err
		}
		f.SetCellStyle(sheetName, "A"+fmt.Sprint(startingRow+index), "G"+fmt.Sprint(startingRow+index), defaultStyle)
		f.SetCellStyle(sheetName, "B"+fmt.Sprint(startingRow+index), "B"+fmt.Sprint(startingRow+index), materialNamingStyle)

		f.SetCellInt(sheetName, "A"+fmt.Sprint(startingRow+index), index+1)
		f.SetCellStr(sheetName, "B"+fmt.Sprint(startingRow+index), material.Code.String)
		f.SetCellStr(sheetName, "C"+fmt.Sprint(startingRow+index), material.Name.String)
		f.SetCellStr(sheetName, "D"+fmt.Sprint(startingRow+index), material.Unit.String)
		f.SetCellFloat(sheetName, "E"+fmt.Sprint(startingRow+index), oneEntry.Amount, 3, 64)
		f.SetCellStr(sheetName, "F"+fmt.Sprint(startingRow+index), oneEntry.Notes)
		f.SetCellInt(sheetName, "G"+fmt.Sprint(startingRow+index), int(material.ID))
	}

	warehouseManager, err := u.q.GetWorker(ctx, int64(data.Details.WarehouseManagerWorkerID))
	if err != nil {
		return err
	}
	f.SetCellStyle(sheetName, "C"+fmt.Sprint(6+len(data.Items)), "C"+fmt.Sprint(6+len(data.Items)), workerNamingStyle)
	f.SetCellStr(sheetName, "C"+fmt.Sprint(6+len(data.Items)), warehouseManager.Name.String)

	released, err := u.q.GetWorker(ctx, int64(data.Details.ReleasedWorkerID))
	if err != nil {
		return err
	}
	f.SetCellStyle(sheetName, "C"+fmt.Sprint(8+len(data.Items)), "C"+fmt.Sprint(8+len(data.Items)), workerNamingStyle)
	f.SetCellStr(sheetName, "C"+fmt.Sprint(8+len(data.Items)), released.Name.String)

	teamData, err := u.q.ListTeamNumberAndLeadersByID(ctx, db.ListTeamNumberAndLeadersByIDParams{
		ProjectID: pgInt8(data.Details.ProjectID),
		ID:        int64(data.Details.TeamID),
	})
	if err != nil {
		return err
	}
	f.SetCellStyle(sheetName, "C"+fmt.Sprint(10+len(data.Items)), "C"+fmt.Sprint(10+len(data.Items)), workerNamingStyle)
	if len(teamData) > 0 {
		f.SetCellStr(sheetName, "C"+fmt.Sprint(10+len(data.Items)), teamData[0].TeamLeaderName)
	}

	recipient, err := u.q.GetWorker(ctx, int64(data.Details.RecipientWorkerID))
	if err != nil {
		return err
	}
	f.SetCellStyle(sheetName, "C"+fmt.Sprint(12+len(data.Items)), "C"+fmt.Sprint(12+len(data.Items)), workerNamingStyle)
	f.SetCellStr(sheetName, "C"+fmt.Sprint(12+len(data.Items)), recipient.Name.String)

	excelFilePath := filepath.Join("./storage/import_excel/output/", data.Details.DeliveryCode+".xlsx")
	if err := f.SaveAs(excelFilePath); err != nil {
		return err
	}

	return nil
}

func (u *invoiceOutputUsecase) GetDocument(deliveryCode string, projectID uint) (string, error) {
	invoiceOutput, err := u.q.GetInvoiceOutputByDeliveryCode(context.Background(), db.GetInvoiceOutputByDeliveryCodeParams{
		DeliveryCode: pgText(deliveryCode),
		ProjectID:    pgInt8(projectID),
	})
	if err != nil {
		return "", err
	}

	if invoiceOutput.Confirmation.Bool {
		return ".pdf", nil
	}
	return ".xlsx", nil
}

func toModelInvoiceOutput(r db.InvoiceOutput) model.InvoiceOutput {
	return model.InvoiceOutput{
		ID:                       uint(r.ID),
		DistrictID:               uintFromPgInt8(r.DistrictID),
		ProjectID:                uintFromPgInt8(r.ProjectID),
		WarehouseManagerWorkerID: uintFromPgInt8(r.WarehouseManagerWorkerID),
		ReleasedWorkerID:         uintFromPgInt8(r.ReleasedWorkerID),
		RecipientWorkerID:        uintFromPgInt8(r.RecipientWorkerID),
		TeamID:                   uintFromPgInt8(r.TeamID),
		DeliveryCode:             r.DeliveryCode.String,
		DateOfInvoice:            timeFromPgTimestamptz(r.DateOfInvoice),
		Notes:                    r.Notes.String,
		Confirmation:             r.Confirmation.Bool,
	}
}
