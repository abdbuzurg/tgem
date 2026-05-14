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

type invoiceCorrectionUsecase struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

func NewInvoiceCorrectionUsecase(pool *pgxpool.Pool) IInvoiceCorrectionUsecase {
	return &invoiceCorrectionUsecase{
		pool: pool,
		q:    db.New(pool),
	}
}

type IInvoiceCorrectionUsecase interface {
	GetPaginated(page, limit int, filter dto.InvoiceCorrectionPaginatedParamters) ([]dto.InvoiceCorrectionPaginated, error)
	GetAll(projectID uint) ([]dto.InvoiceCorrectionPaginated, error)
	GetTotalAmounInLocationByTeamName(projectID, materialID uint, teamNumber string) (float64, error)
	GetInvoiceMaterialsByInvoiceObjectID(id uint) ([]dto.InvoiceCorrectionMaterialsData, error)
	GetSerialNumberOfMaterialInTeam(projectID uint, materialID uint, teamID uint) ([]string, error)
	Create(data dto.InvoiceCorrectionCreate) (model.InvoiceObject, error)
	UniqueObject(projectID uint) ([]dto.ObjectDataForSelect, error)
	UniqueTeam(projectID uint) ([]dto.DataForSelect[uint], error)
	Report(filter dto.InvoiceCorrectionReportFilter) (string, error)
	Count(filter dto.InvoiceCorrectionPaginatedParamters) (int64, error)
	GetOperationsByInvoiceObjectID(id uint) ([]dto.InvoiceCorrectionOperationsData, error)
	GetParametersForSearch(projectID uint) (dto.InvoiceCorrectionSearchData, error)
}

func (u *invoiceCorrectionUsecase) GetPaginated(page, limit int, filter dto.InvoiceCorrectionPaginatedParamters) ([]dto.InvoiceCorrectionPaginated, error) {
	rows, err := u.q.ListInvoiceCorrectionsPaginated(context.Background(), db.ListInvoiceCorrectionsPaginatedParams{
		ProjectID: pgInt8(filter.ProjectID),
		Column2:   int64(filter.TeamID),
		Column3:   int64(filter.ObjectID),
		Limit:     int32(limit),
		Offset:    int32((page - 1) * limit),
	})
	if err != nil {
		return nil, err
	}
	out := make([]dto.InvoiceCorrectionPaginated, len(rows))
	for i, r := range rows {
		out[i] = dto.InvoiceCorrectionPaginated{
			ID:                  uint(r.ID),
			DeliveryCode:        r.DeliveryCode,
			SupervisorName:      r.SupervisorName,
			DistrictID:          uintFromPgInt8(r.DistrictID),
			DistrictName:        r.DistrictName,
			ObjectName:          r.ObjectName,
			ObjectType:          r.ObjectType,
			TeamLeaderName:      r.TeamLeaderName,
			TeamID:              uintFromPgInt8(r.TeamID),
			DateOfInvoice:       timeFromPgTimestamptz(r.DateOfInvoice),
			ConfirmedByOperator: r.ConfirmedByOperator,
		}
	}
	return out, nil
}

func (u *invoiceCorrectionUsecase) GetAll(projectID uint) ([]dto.InvoiceCorrectionPaginated, error) {
	rows, err := u.q.ListInvoiceObjectsForCorrection(context.Background(), pgInt8(projectID))
	if err != nil {
		return nil, err
	}
	out := make([]dto.InvoiceCorrectionPaginated, len(rows))
	for i, r := range rows {
		out[i] = dto.InvoiceCorrectionPaginated{
			ID:                  uint(r.ID),
			DeliveryCode:        r.DeliveryCode,
			SupervisorName:      r.SupervisorName,
			DistrictID:          uintFromPgInt8(r.DistrictID),
			DistrictName:        r.DistrictName,
			ObjectName:          r.ObjectName,
			TeamID:              uint(r.TeamID),
			TeamLeaderName:      "",
			DateOfInvoice:       timeFromPgTimestamptz(r.DateOfInvoice),
			ConfirmedByOperator: r.ConfirmedByOperator,
		}
	}
	return out, nil
}

func (u *invoiceCorrectionUsecase) GetTotalAmounInLocationByTeamName(projectID, materialID uint, teamNumber string) (float64, error) {
	amount, err := u.q.GetTotalAmountInTeamsByTeamNumber(context.Background(), db.GetTotalAmountInTeamsByTeamNumberParams{
		ProjectID: pgInt8(projectID),
		ID:        int64(materialID),
		Number:    pgText(teamNumber),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return float64FromPgNumeric(amount), nil
}

func (u *invoiceCorrectionUsecase) GetInvoiceMaterialsByInvoiceObjectID(id uint) ([]dto.InvoiceCorrectionMaterialsData, error) {
	rows, err := u.q.ListInvoiceCorrectionMaterialsByInvoiceObjectID(context.Background(), pgInt8(id))
	if err != nil {
		return nil, err
	}
	data := make([]dto.InvoiceCorrectionMaterialsData, len(rows))
	for i, r := range rows {
		data[i] = dto.InvoiceCorrectionMaterialsData{
			MaterialName:   r.MaterialName,
			MaterialID:     uint(r.MaterialID),
			MaterialAmount: float64FromPgNumeric(r.MaterialAmount),
			Notes:          r.Notes,
		}
	}

	result := []dto.InvoiceCorrectionMaterialsData{}
	resultIndex := 0
	for index, entry := range data {
		if index == 0 {
			result = append(result, entry)
			continue
		}

		if entry.MaterialID == result[resultIndex].MaterialID {
			result[resultIndex].MaterialAmount += entry.MaterialAmount
		} else {
			result = append(result, entry)
			resultIndex++
		}
	}

	return result, nil
}

func (u *invoiceCorrectionUsecase) GetSerialNumberOfMaterialInTeam(projectID uint, materialID uint, teamID uint) ([]string, error) {
	return u.q.ListInvoiceCorrectionSerialNumberOfMaterialInTeam(context.Background(), db.ListInvoiceCorrectionSerialNumberOfMaterialInTeamParams{
		ProjectID: pgInt8(projectID),
		ID:        int64(materialID),
		ID_2:      int64(teamID),
	})
}

func (u *invoiceCorrectionUsecase) Count(filter dto.InvoiceCorrectionPaginatedParamters) (int64, error) {
	return u.q.CountInvoiceCorrections(context.Background(), db.CountInvoiceCorrectionsParams{
		ProjectID: pgInt8(filter.ProjectID),
		Column2:   int64(filter.TeamID),
		Column3:   int64(filter.ObjectID),
	})
}

func (u *invoiceCorrectionUsecase) Create(data dto.InvoiceCorrectionCreate) (model.InvoiceObject, error) {
	ctx := context.Background()

	invoiceObjectRow, err := u.q.GetInvoiceObject(ctx, int64(data.Details.InvoiceObjectID))
	if err != nil {
		return model.InvoiceObject{}, err
	}
	invoiceObject := toModelInvoiceObject(invoiceObjectRow)
	invoiceObject.ConfirmedByOperator = true
	invoiceObject.DateOfCorrection = data.Details.DateOfCorrection

	invoiceMaterialForCreate := []model.InvoiceMaterials{}
	for _, im := range data.Items {
		rows, err := u.q.ListMaterialAmountSortedByCostM19InLocation(ctx, db.ListMaterialAmountSortedByCostM19InLocationParams{
			ProjectID:    pgInt8(invoiceObject.ProjectID),
			LocationType: pgText("team"),
			LocationID:   pgInt8(invoiceObject.TeamID),
			ID:           int64(im.MaterialID),
		})
		if err != nil {
			return model.InvoiceObject{}, err
		}

		index := 0
		amountLeft := im.MaterialAmount
		for amountLeft > 0 {
			if len(rows) == 0 || index >= len(rows) {
				return model.InvoiceObject{}, fmt.Errorf("Ошибка корректировки: количество материала внутри корректировки превышает количество материала у бригады")
			}
			materialAmount := float64FromPgNumeric(rows[index].MaterialAmount)
			imc := model.InvoiceMaterials{
				ProjectID:      invoiceObject.ProjectID,
				MaterialCostID: uint(rows[index].MaterialCostID),
				InvoiceID:      invoiceObject.ID,
				InvoiceType:    "object-correction",
				Notes:          im.Notes,
			}
			if materialAmount <= amountLeft {
				imc.Amount = materialAmount
				amountLeft -= materialAmount
			} else {
				imc.Amount = amountLeft
				amountLeft = 0
			}
			invoiceMaterialForCreate = append(invoiceMaterialForCreate, imc)
			index++
		}
	}

	invoiceOperationsForCreate := []model.InvoiceOperations{}
	for _, op := range data.Operations {
		invoiceOperationsForCreate = append(invoiceOperationsForCreate, model.InvoiceOperations{
			ProjectID:   invoiceObject.ProjectID,
			OperationID: op.OperationID,
			InvoiceID:   invoiceObject.ID,
			InvoiceType: "object-correction",
			Amount:      op.Amount,
		})
	}

	type ml struct {
		row     *db.MaterialLocation
		params  *db.CreateMaterialLocationParams
		amount  float64
	}
	teamLocs := []*ml{}
	objLocs := []*ml{}

	getOrCreateLoc := func(slice *[]*ml, mcID, locID uint, locType string) (*ml, error) {
		for _, l := range *slice {
			if l.row != nil && uintFromPgInt8(l.row.MaterialCostID) == mcID {
				return l, nil
			}
			if l.params != nil && uintFromPgInt8(l.params.MaterialCostID) == mcID {
				return l, nil
			}
		}
		row, err := u.q.GetMaterialLocationByCostAndLocation(ctx, db.GetMaterialLocationByCostAndLocationParams{
			ProjectID:      pgInt8(invoiceObject.ProjectID),
			MaterialCostID: pgInt8(mcID),
			LocationType:   pgText(locType),
			LocationID:     pgInt8(locID),
		})
		if errors.Is(err, pgx.ErrNoRows) {
			entry := &ml{
				params: &db.CreateMaterialLocationParams{
					ProjectID:      pgInt8(invoiceObject.ProjectID),
					MaterialCostID: pgInt8(mcID),
					LocationType:   pgText(locType),
					LocationID:     pgInt8(locID),
					Amount:         pgNumericFromFloat64(0),
				},
				amount: 0,
			}
			*slice = append(*slice, entry)
			return entry, nil
		}
		if err != nil {
			return nil, err
		}
		entry := &ml{row: &row, amount: float64FromPgNumeric(row.Amount)}
		*slice = append(*slice, entry)
		return entry, nil
	}

	for _, im := range invoiceMaterialForCreate {
		teamLoc, err := getOrCreateLoc(&teamLocs, im.MaterialCostID, invoiceObject.TeamID, "team")
		if err != nil {
			return model.InvoiceObject{}, err
		}
		objectLoc, err := getOrCreateLoc(&objLocs, im.MaterialCostID, invoiceObject.ObjectID, "object")
		if err != nil {
			return model.InvoiceObject{}, err
		}
		teamLoc.amount -= im.Amount
		objectLoc.amount += im.Amount
	}

	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return model.InvoiceObject{}, err
	}
	defer tx.Rollback(ctx)
	qtx := u.q.WithTx(tx)

	if err := qtx.UpdateInvoiceObjectConfirmation(ctx, db.UpdateInvoiceObjectConfirmationParams{
		ID:                  int64(invoiceObject.ID),
		ConfirmedByOperator: pgBool(invoiceObject.ConfirmedByOperator),
		DateOfCorrection:    pgTimestamptz(invoiceObject.DateOfCorrection),
	}); err != nil {
		return model.InvoiceObject{}, err
	}

	if err := qtx.CreateInvoiceObjectOperator(ctx, db.CreateInvoiceObjectOperatorParams{
		OperatorWorkerID: pgInt8(data.Details.OperatorWorkerID),
		InvoiceObjectID:  pgInt8(invoiceObject.ID),
	}); err != nil {
		return model.InvoiceObject{}, err
	}

	if err := writeInvoiceMaterialsBatch(ctx, qtx, invoiceObject.ID, invoiceMaterialForCreate); err != nil {
		return model.InvoiceObject{}, err
	}

	if len(invoiceOperationsForCreate) > 0 {
		batch := make([]db.CreateInvoiceOperationsBatchParams, len(invoiceOperationsForCreate))
		for i, op := range invoiceOperationsForCreate {
			batch[i] = db.CreateInvoiceOperationsBatchParams{
				ProjectID:   pgInt8(op.ProjectID),
				OperationID: pgInt8(op.OperationID),
				InvoiceID:   pgInt8(op.InvoiceID),
				InvoiceType: pgText(op.InvoiceType),
				Amount:      pgNumericFromFloat64(op.Amount),
				Notes:       pgText(op.Notes),
			}
		}
		if _, err := qtx.CreateInvoiceOperationsBatch(ctx, batch); err != nil {
			return model.InvoiceObject{}, err
		}
	}

	for _, l := range teamLocs {
		if l.row != nil {
			if err := qtx.UpsertMaterialLocationByID(ctx, db.UpsertMaterialLocationByIDParams{
				Amount: pgNumericFromFloat64(l.amount),
				ID:     l.row.ID,
			}); err != nil {
				return model.InvoiceObject{}, err
			}
		} else {
			params := *l.params
			params.Amount = pgNumericFromFloat64(l.amount)
			if _, err := qtx.CreateMaterialLocation(ctx, params); err != nil {
				return model.InvoiceObject{}, err
			}
		}
	}
	for _, l := range objLocs {
		if l.row != nil {
			if err := qtx.UpsertMaterialLocationByID(ctx, db.UpsertMaterialLocationByIDParams{
				Amount: pgNumericFromFloat64(l.amount),
				ID:     l.row.ID,
			}); err != nil {
				return model.InvoiceObject{}, err
			}
		} else {
			params := *l.params
			params.Amount = pgNumericFromFloat64(l.amount)
			if _, err := qtx.CreateMaterialLocation(ctx, params); err != nil {
				return model.InvoiceObject{}, err
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return model.InvoiceObject{}, err
	}

	return invoiceObject, nil
}

func (u *invoiceCorrectionUsecase) UniqueObject(projectID uint) ([]dto.ObjectDataForSelect, error) {
	rows, err := u.q.ListInvoiceCorrectionUniqueObjects(context.Background(), pgInt8(projectID))
	if err != nil {
		return nil, err
	}
	out := make([]dto.ObjectDataForSelect, len(rows))
	for i, r := range rows {
		out[i] = dto.ObjectDataForSelect{
			ID:         uint(r.ID),
			ObjectName: r.ObjectName,
			ObjectType: r.ObjectType,
		}
	}
	return out, nil
}

func (u *invoiceCorrectionUsecase) UniqueTeam(projectID uint) ([]dto.DataForSelect[uint], error) {
	rows, err := u.q.ListInvoiceCorrectionUniqueTeams(context.Background(), pgInt8(projectID))
	if err != nil {
		return nil, err
	}
	out := make([]dto.DataForSelect[uint], len(rows))
	for i, r := range rows {
		out[i] = dto.DataForSelect[uint]{Label: r.Label, Value: uint(r.Value)}
	}
	return out, nil
}

func (u *invoiceCorrectionUsecase) Report(filter dto.InvoiceCorrectionReportFilter) (string, error) {
	ctx := context.Background()
	invoices, err := u.q.ListInvoiceCorrectionReportData(ctx, db.ListInvoiceCorrectionReportDataParams{
		ProjectID: pgInt8(filter.ProjectID),
		Column2:   int64(filter.ObjectID),
		Column3:   int64(filter.TeamID),
		Column4:   pgTimestamptz(filter.DateFrom),
		Column5:   pgTimestamptz(filter.DateTo),
	})
	if err != nil {
		return "", err
	}

	templateFilePath := filepath.Join("./internal/templates/", "Object Spenditure Report.xlsx")
	f, err := excelize.OpenFile(templateFilePath)
	if err != nil {
		return "", err
	}
	sheetName := "Sheet1"
	f.SetCellStr(sheetName, "M1", "ID материала")

	rowCount := 2
	for _, invoice := range invoices {
		invoiceMaterials, err := u.q.ListInvoiceMaterialsDataForReport(ctx, db.ListInvoiceMaterialsDataForReportParams{
			InvoiceType: pgText("object-correction"),
			InvoiceID:   pgInt8(uint(invoice.ID)),
		})
		if err != nil {
			return "", err
		}

		for index, im := range invoiceMaterials {
			f.SetCellStr(sheetName, "A"+fmt.Sprint(rowCount+index), invoice.DeliveryCode)
			f.SetCellStr(sheetName, "B"+fmt.Sprint(rowCount+index), invoice.ObjectName)
			f.SetCellStr(sheetName, "C"+fmt.Sprint(rowCount+index), utils.ObjectTypeConverter(invoice.ObjectType))
			f.SetCellStr(sheetName, "D"+fmt.Sprint(rowCount+index), invoice.TeamNumber)
			f.SetCellStr(sheetName, "E"+fmt.Sprint(rowCount+index), invoice.TeamLeaderName)
			dateOfInvoice := timeFromPgTimestamptz(invoice.DateOfInvoice).String()
			if len(dateOfInvoice) > 10 {
				dateOfInvoice = dateOfInvoice[:len(dateOfInvoice)-10]
			}
			f.SetCellStr(sheetName, "F"+fmt.Sprint(rowCount+index), dateOfInvoice)
			f.SetCellStr(sheetName, "G"+fmt.Sprint(rowCount+index), invoice.OperatorName)
			dateOfCorrection := timeFromPgTimestamptz(invoice.DateOfCorrection).String()
			if len(dateOfCorrection) > 10 {
				dateOfCorrection = dateOfCorrection[:len(dateOfCorrection)-10]
			}
			f.SetCellStr(sheetName, "H"+fmt.Sprint(rowCount+index), dateOfCorrection)

			f.SetCellStr(sheetName, "I"+fmt.Sprint(rowCount+index), im.MaterialName)
			f.SetCellStr(sheetName, "J"+fmt.Sprint(rowCount+index), im.MaterialUnit)
			f.SetCellFloat(sheetName, "K"+fmt.Sprint(rowCount+index), float64FromPgNumeric(im.InvoiceMaterialAmount), 2, 64)
			f.SetCellStr(sheetName, "L"+fmt.Sprint(rowCount+index), im.InvoiceMaterialNotes)
			f.SetCellInt(sheetName, "M"+fmt.Sprint(rowCount+index), int(im.MaterialID))
		}

		rowCount += len(invoiceMaterials)
	}

	currentTime := time.Now()
	fileName := fmt.Sprintf(
		"Отсчет Расхода - %s.xlsx",
		currentTime.Format("02-01-2006"),
	)
	tempFilePath := filepath.Join("./storage/import_excel/temp/", fileName)
	f.SaveAs(tempFilePath)
	if err := f.Close(); err != nil {
		fmt.Println(err)
	}
	return fileName, nil
}

func (u *invoiceCorrectionUsecase) GetOperationsByInvoiceObjectID(id uint) ([]dto.InvoiceCorrectionOperationsData, error) {
	rows, err := u.q.ListInvoiceCorrectionOperationsByInvoiceObjectID(context.Background(), pgInt8(id))
	if err != nil {
		return nil, err
	}
	out := make([]dto.InvoiceCorrectionOperationsData, len(rows))
	for i, r := range rows {
		out[i] = dto.InvoiceCorrectionOperationsData{
			OperationID:   uint(r.OperationID),
			OperationName: r.OperationName,
			Amount:        float64FromPgNumeric(r.Amount),
			MaterialName:  r.MaterialName,
		}
	}
	return out, nil
}

func (u *invoiceCorrectionUsecase) GetParametersForSearch(projectID uint) (dto.InvoiceCorrectionSearchData, error) {
	ctx := context.Background()
	teamRows, err := u.q.ListInvoiceCorrectionTeamsForSearch(ctx, pgInt8(projectID))
	if err != nil {
		return dto.InvoiceCorrectionSearchData{}, err
	}
	teams := make([]dto.DataForSelect[uint], len(teamRows))
	for i, r := range teamRows {
		teams[i] = dto.DataForSelect[uint]{Label: r.Label, Value: uint(r.Value)}
	}

	objectRows, err := u.q.ListInvoiceCorrectionObjectsForSearch(ctx, pgInt8(projectID))
	if err != nil {
		return dto.InvoiceCorrectionSearchData{}, err
	}
	objects := make([]dto.DataForSelect[uint], len(objectRows))
	for i, r := range objectRows {
		objects[i] = dto.DataForSelect[uint]{Label: r.Label, Value: uint(r.Value)}
	}

	return dto.InvoiceCorrectionSearchData{
		Teams:   teams,
		Objects: objects,
	}, nil
}
