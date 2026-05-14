package usecase

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"backend-v2/internal/db"
	"backend-v2/internal/dto"
	"backend-v2/internal/utils"
	"backend-v2/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xuri/excelize/v2"
)

type invoiceWriteOffUsecase struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

func NewInvoiceWriteOffUsecase(pool *pgxpool.Pool) IInvoiceWriteOffUsecase {
	return &invoiceWriteOffUsecase{
		pool: pool,
		q:    db.New(pool),
	}
}

type IInvoiceWriteOffUsecase interface {
	GetAll() ([]model.InvoiceWriteOff, error)
	GetPaginated(page, limit int, data dto.InvoiceWriteOffSearchParameters) ([]dto.InvoiceWriteOffPaginated, error)
	GetInvoiceMaterialsWithoutSerialNumbers(id uint) ([]dto.InvoiceMaterialsWithoutSerialNumberView, error)
	GetByID(id uint) (model.InvoiceWriteOff, error)
	Create(data dto.InvoiceWriteOff) (model.InvoiceWriteOff, error)
	Update(data dto.InvoiceWriteOff) (model.InvoiceWriteOff, error)
	Delete(id uint) error
	Count(filter dto.InvoiceWriteOffSearchParameters) (int64, error)
	GetMaterialsForEdit(id uint, locationType string, locationID uint) ([]dto.InvoiceWriteOffMaterialsForEdit, error)
	Confirmation(id, projectID uint) error
	Report(parameters dto.InvoiceWriteOffReportParameters) (string, error)
	GetMaterialsInLocation(projectID, locationID uint, locationType string) ([]dto.InvoiceReturnMaterialForSelect, error)
}

func writeOffLocationOf(writeOffType string) (string, error) {
	switch writeOffType {
	case "loss-warehouse", "writeoff-warehouse":
		return "warehouse", nil
	case "loss-team":
		return "team", nil
	case "loss-object", "writeoff-object":
		return "object", nil
	}
	return "", fmt.Errorf("Неправильный вид списание обнаружен")
}

func (u *invoiceWriteOffUsecase) GetAll() ([]model.InvoiceWriteOff, error) {
	rows, err := u.q.ListInvoiceWriteOffs(context.Background())
	if err != nil {
		return nil, err
	}
	out := make([]model.InvoiceWriteOff, len(rows))
	for i, r := range rows {
		out[i] = toModelInvoiceWriteOff(r)
	}
	return out, nil
}

func (u *invoiceWriteOffUsecase) GetByID(id uint) (model.InvoiceWriteOff, error) {
	row, err := u.q.GetInvoiceWriteOff(context.Background(), int64(id))
	if err != nil {
		return model.InvoiceWriteOff{}, err
	}
	return toModelInvoiceWriteOff(row), nil
}

func (u *invoiceWriteOffUsecase) GetPaginated(page, limit int, data dto.InvoiceWriteOffSearchParameters) ([]dto.InvoiceWriteOffPaginated, error) {
	ctx := context.Background()
	rows, err := u.q.ListInvoiceWriteOffsPaginated(ctx, db.ListInvoiceWriteOffsPaginatedParams{
		ProjectID: pgInt8(data.ProjectID),
		Column2:   data.WriteOffType,
		Limit:     int32(limit),
		Offset:    int32((page - 1) * limit),
	})
	if err != nil {
		return nil, err
	}

	out := make([]dto.InvoiceWriteOffPaginated, len(rows))
	for i, r := range rows {
		out[i] = dto.InvoiceWriteOffPaginated{
			ID:                 uint(r.ID),
			WriteOffType:       r.WriteOffType,
			WriteOffLocationID: uint(r.WriteOffLocationID),
			ReleasedWorkerID:   uintFromPgInt8(r.ReleasedWorkerID),
			ReleasedWorkerName: r.ReleasedWorkerName,
			DeliveryCode:       r.DeliveryCode,
			DateOfInvoice:      timeFromPgTimestamptz(r.DateOfInvoice),
			Confirmation:       r.Confirmation,
			DateOfConfirmation: timeFromPgTimestamptz(r.DateOfConfirmation),
		}
	}

	for index, invoiceWriteOff := range out {
		switch invoiceWriteOff.WriteOffType {
		case "writeoff-warehouse", "loss-warehouse":
			// no extra location name lookup
		case "loss-team":
			team, err := u.q.ListTeamNumberAndLeadersByID(ctx, db.ListTeamNumberAndLeadersByIDParams{
				ProjectID: pgInt8(data.ProjectID),
				ID:        int64(invoiceWriteOff.WriteOffLocationID),
			})
			if err != nil {
				return nil, err
			}
			if len(team) > 0 {
				out[index].WriteOffLocationName = team[0].TeamNumber + " (" + team[0].TeamLeaderName + ")"
			}
		case "writeoff-object", "loss-object":
			object, err := u.q.GetObject(ctx, int64(invoiceWriteOff.WriteOffLocationID))
			if err != nil {
				return nil, err
			}
			out[index].WriteOffLocationName = object.Name.String
		default:
			return nil, fmt.Errorf("Обноружен неправильный тип списание %v", invoiceWriteOff.WriteOffType)
		}
	}

	return out, nil
}

func (u *invoiceWriteOffUsecase) Create(data dto.InvoiceWriteOff) (model.InvoiceWriteOff, error) {
	ctx := context.Background()

	count, err := u.q.GetInvoiceCount(ctx, db.GetInvoiceCountParams{
		InvoiceType: pgText("writeoff"),
		ProjectID:   pgInt8(data.Details.ProjectID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		count = 0
	} else if err != nil {
		return model.InvoiceWriteOff{}, err
	}

	data.Details.DeliveryCode = utils.UniqueCodeGeneration("С", count+1, data.Details.ProjectID)

	writeOffLocation, err := writeOffLocationOf(data.Details.WriteOffType)
	if err != nil {
		return model.InvoiceWriteOff{}, err
	}

	invoiceMaterialForCreate, err := u.buildWriteOffItems(ctx, data, writeOffLocation)
	if err != nil {
		return model.InvoiceWriteOff{}, err
	}

	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return model.InvoiceWriteOff{}, err
	}
	defer tx.Rollback(ctx)
	qtx := u.q.WithTx(tx)

	invoiceRow, err := qtx.CreateInvoiceWriteOff(ctx, db.CreateInvoiceWriteOffParams{
		ProjectID:          pgInt8(data.Details.ProjectID),
		ReleasedWorkerID:   pgInt8(data.Details.ReleasedWorkerID),
		WriteOffType:       pgText(data.Details.WriteOffType),
		WriteOffLocationID: pgInt8(data.Details.WriteOffLocationID),
		DeliveryCode:       pgText(data.Details.DeliveryCode),
		DateOfInvoice:      pgTimestamptz(data.Details.DateOfInvoice),
		Confirmation:       pgBool(data.Details.Confirmation),
		DateOfConfirmation: pgTimestamptz(data.Details.DateOfConfirmation),
		Notes:              pgText(data.Details.Notes),
	})
	if err != nil {
		return model.InvoiceWriteOff{}, err
	}

	if err := writeInvoiceMaterialsBatch(ctx, qtx, uint(invoiceRow.ID), invoiceMaterialForCreate); err != nil {
		return model.InvoiceWriteOff{}, err
	}

	if err := qtx.IncrementInvoiceCount(ctx, db.IncrementInvoiceCountParams{
		InvoiceType: pgText("writeoff"),
		ProjectID:   invoiceRow.ProjectID,
	}); err != nil {
		return model.InvoiceWriteOff{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return model.InvoiceWriteOff{}, err
	}

	return toModelInvoiceWriteOff(invoiceRow), nil
}

func (u *invoiceWriteOffUsecase) buildWriteOffItems(ctx context.Context, data dto.InvoiceWriteOff, writeOffLocation string) ([]model.InvoiceMaterials, error) {
	out := []model.InvoiceMaterials{}
	for _, item := range data.Items {
		if len(item.SerialNumbers) != 0 {
			// Serial-number-tracked writeoffs aren't supported yet. The
			// pre-phase-7 GORM code silently dropped these items, which
			// could lose data; surface explicitly so the caller knows.
			return nil, fmt.Errorf("Списание материалов с серийными номерами пока не поддерживается")
		}

		rows, err := u.q.ListMaterialAmountSortedByCostM19InLocation(ctx, db.ListMaterialAmountSortedByCostM19InLocationParams{
			ProjectID:    pgInt8(data.Details.ProjectID),
			LocationType: pgText(writeOffLocation),
			LocationID:   pgInt8(data.Details.WriteOffLocationID),
			ID:           int64(item.MaterialID),
		})
		if err != nil {
			return nil, err
		}

		index := 0
		amountLeft := item.Amount
		for amountLeft > 0 {
			materialAmount := float64FromPgNumeric(rows[index].MaterialAmount)
			imc := model.InvoiceMaterials{
				ProjectID:      data.Details.ProjectID,
				MaterialCostID: uint(rows[index].MaterialCostID),
				InvoiceType:    "writeoff",
				Notes:          item.Notes,
			}
			if materialAmount <= amountLeft {
				imc.Amount = materialAmount
				amountLeft -= materialAmount
			} else {
				imc.Amount = amountLeft
				amountLeft = 0
			}
			out = append(out, imc)
			index++
		}
	}
	return out, nil
}

func (u *invoiceWriteOffUsecase) Update(data dto.InvoiceWriteOff) (model.InvoiceWriteOff, error) {
	ctx := context.Background()
	writeOffLocation, err := writeOffLocationOf(data.Details.WriteOffType)
	if err != nil {
		return model.InvoiceWriteOff{}, err
	}

	invoiceMaterialForCreate, err := u.buildWriteOffItems(ctx, data, writeOffLocation)
	if err != nil {
		return model.InvoiceWriteOff{}, err
	}

	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return model.InvoiceWriteOff{}, err
	}
	defer tx.Rollback(ctx)
	qtx := u.q.WithTx(tx)

	invoiceRow, err := qtx.UpdateInvoiceWriteOff(ctx, db.UpdateInvoiceWriteOffParams{
		ID:                 int64(data.Details.ID),
		ProjectID:          pgInt8(data.Details.ProjectID),
		ReleasedWorkerID:   pgInt8(data.Details.ReleasedWorkerID),
		WriteOffType:       pgText(data.Details.WriteOffType),
		WriteOffLocationID: pgInt8(data.Details.WriteOffLocationID),
		DeliveryCode:       pgText(data.Details.DeliveryCode),
		DateOfInvoice:      pgTimestamptz(data.Details.DateOfInvoice),
		Confirmation:       pgBool(data.Details.Confirmation),
		DateOfConfirmation: pgTimestamptz(data.Details.DateOfConfirmation),
		Notes:              pgText(data.Details.Notes),
	})
	if err != nil {
		return model.InvoiceWriteOff{}, err
	}

	if err := qtx.DeleteInvoiceMaterialsByInvoice(ctx, db.DeleteInvoiceMaterialsByInvoiceParams{
		InvoiceType: pgText("writeoff"),
		InvoiceID:   pgInt8(uint(invoiceRow.ID)),
	}); err != nil {
		return model.InvoiceWriteOff{}, err
	}

	if err := writeInvoiceMaterialsBatch(ctx, qtx, uint(invoiceRow.ID), invoiceMaterialForCreate); err != nil {
		return model.InvoiceWriteOff{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return model.InvoiceWriteOff{}, err
	}

	return toModelInvoiceWriteOff(invoiceRow), nil
}

func (u *invoiceWriteOffUsecase) Delete(id uint) error {
	ctx := context.Background()
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	qtx := u.q.WithTx(tx)

	if err := qtx.DeleteInvoiceMaterialsByInvoice(ctx, db.DeleteInvoiceMaterialsByInvoiceParams{
		InvoiceType: pgText("writeoff"),
		InvoiceID:   pgInt8(id),
	}); err != nil {
		return err
	}
	if err := qtx.DeleteInvoiceWriteOff(ctx, int64(id)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (u *invoiceWriteOffUsecase) Count(filter dto.InvoiceWriteOffSearchParameters) (int64, error) {
	return u.q.CountInvoiceWriteOffsFiltered(context.Background(), db.CountInvoiceWriteOffsFilteredParams{
		ProjectID: pgInt8(filter.ProjectID),
		Column2:   filter.WriteOffType,
	})
}

func (u *invoiceWriteOffUsecase) GetInvoiceMaterialsWithoutSerialNumbers(id uint) ([]dto.InvoiceMaterialsWithoutSerialNumberView, error) {
	rows, err := u.q.ListInvoiceMaterialsWithoutSerialNumbers(context.Background(), db.ListInvoiceMaterialsWithoutSerialNumbersParams{
		InvoiceType: pgText("writeoff"),
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

func (u *invoiceWriteOffUsecase) GetMaterialsForEdit(id uint, locationType string, locationID uint) ([]dto.InvoiceWriteOffMaterialsForEdit, error) {
	rows, err := u.q.ListInvoiceWriteOffMaterialsForEdit(context.Background(), db.ListInvoiceWriteOffMaterialsForEditParams{
		InvoiceID:    pgInt8(id),
		LocationType: pgText(locationType),
		LocationID:   pgInt8(locationID),
	})
	if err != nil {
		return []dto.InvoiceWriteOffMaterialsForEdit{}, nil
	}

	data := make([]dto.InvoiceWriteOffMaterialsForEdit, len(rows))
	for i, r := range rows {
		costM19, _ := decimalFromPgNumeric(r.MaterialCost).Float64()
		data[i] = dto.InvoiceWriteOffMaterialsForEdit{
			MaterialID:      uint(r.MaterialID),
			MaterialName:    r.MaterialName,
			Unit:            r.Unit,
			Amount:          float64FromPgNumeric(r.Amount),
			MaterialCostID:  uint(r.MaterialCostID),
			MaterialCost:    costM19,
			Notes:           r.Notes,
			HasSerialNumber: r.HasSerialNumber,
			LocationAmount:  float64FromPgNumeric(r.LocationAmount),
		}
	}

	var result []dto.InvoiceWriteOffMaterialsForEdit
	for index, entry := range data {
		if index == 0 {
			result = append(result, entry)
			continue
		}
		lastItemIndex := len(result) - 1
		if result[lastItemIndex].MaterialID == entry.MaterialID {
			result[lastItemIndex].Amount += entry.Amount
			result[lastItemIndex].LocationAmount += entry.LocationAmount
		} else {
			result = append(result, entry)
		}
	}
	return result, nil
}

func (u *invoiceWriteOffUsecase) Confirmation(id, projectID uint) error {
	ctx := context.Background()
	invoice, err := u.q.GetInvoiceWriteOff(ctx, int64(id))
	if err != nil {
		return err
	}

	invoiceMaterials, err := u.q.ListInvoiceMaterialsByInvoice(ctx, db.ListInvoiceMaterialsByInvoiceParams{
		InvoiceID:   pgInt8(id),
		InvoiceType: pgText("writeoff"),
		ProjectID:   pgInt8(projectID),
	})
	if err != nil {
		return err
	}

	var sourceLocation string
	var sourceLocationID uint
	switch invoice.WriteOffType.String {
	case "writeoff-warehouse", "loss-warehouse":
		sourceLocation = "warehouse"
		sourceLocationID = 0
	case "loss-team":
		sourceLocation = "team"
		sourceLocationID = uintFromPgInt8(invoice.WriteOffLocationID)
	case "loss-object", "writeoff-object":
		sourceLocation = "object"
		sourceLocationID = uintFromPgInt8(invoice.WriteOffLocationID)
	default:
		return fmt.Errorf("Неизвестный вид списания: %v", invoice.WriteOffType.String)
	}

	materialsInTheLocation, err := u.q.ListMaterialLocationsForInvoiceConfirmation(ctx, db.ListMaterialLocationsForInvoiceConfirmationParams{
		LocationType: pgText(sourceLocation),
		LocationID:   pgInt8(sourceLocationID),
		InvoiceType:  pgText("writeoff"),
		InvoiceID:    pgInt8(id),
	})
	if err != nil {
		return err
	}

	materialsInWriteOff, err := u.q.ListMaterialLocationsByLocationType(ctx, pgText(invoice.WriteOffType.String))
	if err != nil {
		return err
	}

	type loc struct {
		row    db.MaterialLocation
		amount float64
	}
	type writeOffLoc struct {
		row    *db.MaterialLocation
		params *db.CreateMaterialLocationParams
		amount float64
	}

	srcSlice := make([]loc, len(materialsInTheLocation))
	for i, ml := range materialsInTheLocation {
		srcSlice[i] = loc{row: ml, amount: float64FromPgNumeric(ml.Amount)}
	}

	dstSlice := make([]*writeOffLoc, 0, len(materialsInWriteOff))
	for i := range materialsInWriteOff {
		ml := materialsInWriteOff[i]
		dstSlice = append(dstSlice, &writeOffLoc{row: &ml, amount: float64FromPgNumeric(ml.Amount)})
	}

	for _, im := range invoiceMaterials {
		srcIndex := -1
		for index, s := range srcSlice {
			if uintFromPgInt8(s.row.MaterialCostID) == uintFromPgInt8(im.MaterialCostID) {
				srcIndex = index
				break
			}
		}
		if srcIndex == -1 {
			return fmt.Errorf("Ошибка, несанкционированный материал")
		}
		imAmount := float64FromPgNumeric(im.Amount)
		srcSlice[srcIndex].amount -= imAmount

		dstIndex := -1
		for index, d := range dstSlice {
			if d.row != nil && uintFromPgInt8(d.row.MaterialCostID) == uintFromPgInt8(im.MaterialCostID) {
				dstIndex = index
				break
			}
			if d.params != nil && uintFromPgInt8(d.params.MaterialCostID) == uintFromPgInt8(im.MaterialCostID) {
				dstIndex = index
				break
			}
		}
		if dstIndex != -1 {
			dstSlice[dstIndex].amount += imAmount
		} else {
			dstSlice = append(dstSlice, &writeOffLoc{
				params: &db.CreateMaterialLocationParams{
					MaterialCostID: im.MaterialCostID,
					ProjectID:      invoice.ProjectID,
					LocationID:     pgInt8(0),
					LocationType:   invoice.WriteOffType,
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

	if err := qtx.UpdateInvoiceWriteOffConfirmation(ctx, db.UpdateInvoiceWriteOffConfirmationParams{
		ID:                 int64(id),
		Confirmation:       pgBool(true),
		DateOfConfirmation: pgTimestamptz(time.Now()),
	}); err != nil {
		return err
	}

	for _, s := range srcSlice {
		if err := qtx.UpsertMaterialLocationByID(ctx, db.UpsertMaterialLocationByIDParams{
			Amount: pgNumericFromFloat64(s.amount),
			ID:     s.row.ID,
		}); err != nil {
			return err
		}
	}

	for _, d := range dstSlice {
		if d.row != nil {
			if err := qtx.UpsertMaterialLocationByID(ctx, db.UpsertMaterialLocationByIDParams{
				Amount: pgNumericFromFloat64(d.amount),
				ID:     d.row.ID,
			}); err != nil {
				return err
			}
		} else {
			params := *d.params
			params.Amount = pgNumericFromFloat64(d.amount)
			if _, err := qtx.CreateMaterialLocation(ctx, params); err != nil {
				return err
			}
		}
	}

	return tx.Commit(ctx)
}

func (u *invoiceWriteOffUsecase) Report(parameters dto.InvoiceWriteOffReportParameters) (string, error) {
	ctx := context.Background()
	invoices, err := u.q.ListInvoiceWriteOffReportData(ctx, db.ListInvoiceWriteOffReportDataParams{
		ProjectID:    pgInt8(parameters.ProjectID),
		WriteOffType: pgText(parameters.WriteOffType),
		Column3:      pgTimestamptz(parameters.DateFrom),
		Column4:      pgTimestamptz(parameters.DateTo),
	})
	if err != nil {
		return "", err
	}

	templateFilePath := filepath.Join("./internal/templates/", "Invoice Writeoff Report.xlsx")
	f, err := excelize.OpenFile(templateFilePath)
	if err != nil {
		return "", err
	}

	sheetName := "Sheet1"
	f.SetCellStr(sheetName, "J1", "ID материала")

	rowCount := 2
	for _, invoice := range invoices {
		invoiceMaterials, err := u.q.ListInvoiceMaterialsDataForReport(ctx, db.ListInvoiceMaterialsDataForReportParams{
			InvoiceType: pgText("writeoff"),
			InvoiceID:   pgInt8(uint(invoice.ID)),
		})
		if err != nil {
			return "", err
		}

		for _, im := range invoiceMaterials {
			f.SetCellStr(sheetName, "A"+fmt.Sprint(rowCount), invoice.DeliveryCode)
			f.SetCellStr(sheetName, "B"+fmt.Sprint(rowCount), invoice.ReleasedWorkerName)

			switch parameters.WriteOffType {
			case "writeoff-warehouse", "loss-warehouse":
				f.SetCellStr(sheetName, "C"+fmt.Sprint(rowCount), "Склад")
			case "loss-team":
				team, err := u.q.GetTeam(ctx, int64(parameters.WriteOffLocationID))
				if err != nil {
					return "", err
				}
				f.SetCellStr(sheetName, "C"+fmt.Sprint(rowCount), team.Number.String)
			case "loss-object", "writeoff-object":
				object, err := u.q.GetObject(ctx, int64(parameters.WriteOffLocationID))
				if err != nil {
					return "", err
				}
				f.SetCellStr(sheetName, "C"+fmt.Sprint(rowCount), object.Name.String)
			}

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
		"Отсчет накладной списание - %s.xlsx",
		currentTime.Format("02-01-2006"),
	)
	tempFilePath := filepath.Join("./storage/import_excel/temp/", fileName)
	f.SaveAs(tempFilePath)
	if err := f.Close(); err != nil {
		fmt.Println(err)
	}
	return fileName, nil
}

func (u *invoiceWriteOffUsecase) GetMaterialsInLocation(projectID, locationID uint, locationType string) ([]dto.InvoiceReturnMaterialForSelect, error) {
	ctx := context.Background()
	materials, err := u.q.ListUniqueMaterialsFromLocation(ctx, db.ListUniqueMaterialsFromLocationParams{
		ProjectID:    pgInt8(projectID),
		LocationID:   pgInt8(locationID),
		LocationType: pgText(locationType),
	})
	if err != nil {
		return nil, err
	}

	out := []dto.InvoiceReturnMaterialForSelect{}
	for _, m := range materials {
		amount, err := u.q.GetTotalAmountInLocation(ctx, db.GetTotalAmountInLocationParams{
			ProjectID:    pgInt8(projectID),
			ID:           m.ID,
			LocationType: pgText(locationType),
			LocationID:   pgInt8(locationID),
		})
		if errors.Is(err, pgx.ErrNoRows) {
			amount = pgNumericFromFloat64(0)
		} else if err != nil {
			return nil, err
		}
		out = append(out, dto.InvoiceReturnMaterialForSelect{
			MaterialID:      uint(m.ID),
			MaterialName:    m.Name.String,
			MaterialUnit:    m.Unit.String,
			Amount:          float64FromPgNumeric(amount),
			HasSerialNumber: m.HasSerialNumber.Bool,
		})
	}
	return out, nil
}

func toModelInvoiceWriteOff(r db.InvoiceWriteOff) model.InvoiceWriteOff {
	return model.InvoiceWriteOff{
		ID:                 uint(r.ID),
		ProjectID:          uintFromPgInt8(r.ProjectID),
		ReleasedWorkerID:   uintFromPgInt8(r.ReleasedWorkerID),
		WriteOffType:       r.WriteOffType.String,
		WriteOffLocationID: uintFromPgInt8(r.WriteOffLocationID),
		DeliveryCode:       r.DeliveryCode.String,
		DateOfInvoice:      timeFromPgTimestamptz(r.DateOfInvoice),
		Confirmation:       r.Confirmation.Bool,
		DateOfConfirmation: timeFromPgTimestamptz(r.DateOfConfirmation),
		Notes:              r.Notes.String,
	}
}
