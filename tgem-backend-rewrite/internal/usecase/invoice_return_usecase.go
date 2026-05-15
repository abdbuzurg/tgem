package usecase

import (
	"context"
	"errors"
	"fmt"
	"os"
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

type invoiceReturnUsecase struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

func NewInvoiceReturnUsecase(pool *pgxpool.Pool) IInvoiceReturnUsecase {
	return &invoiceReturnUsecase{
		pool: pool,
		q:    db.New(pool),
	}
}

type IInvoiceReturnUsecase interface {
	GetAll() ([]model.InvoiceReturn, error)
	GetByID(id uint) (model.InvoiceReturn, error)
	GetPaginatedTeam(page, limit int, projectID uint) ([]dto.InvoiceReturnTeamPaginatedQueryData, error)
	GetPaginatedObject(page, limit int, projectID uint) ([]dto.InvoiceReturnObjectPaginated, error)
	GetDocument(deliveryCode string, projectID uint) (string, error)
	Create(data dto.InvoiceReturn) (model.InvoiceReturn, error)
	Update(data dto.InvoiceReturn) (model.InvoiceReturn, error)
	Delete(id uint) error
	CountBasedOnType(projectID uint, invoiceType string) (int64, error)
	Confirmation(id uint) error
	UniqueCode(projectID uint) ([]string, error)
	UniqueTeam(projectID uint) ([]string, error)
	UniqueObject(projectID uint) ([]string, error)
	Report(filter dto.InvoiceReturnReportFilterRequest, projectID uint) (string, error)
	GetMaterialsInLocation(projectID, locationID uint, locationType string) ([]dto.InvoiceReturnMaterialForSelect, error)
	GetMaterialCostInLocation(projectID, locationID, materialID uint, locationType string) ([]model.MaterialCost, error)
	GetMaterialAmountInLocation(projectID, locationID, materialCostID uint, locationType string) (float64, error)
	GetSerialNumberCodesInLocation(projectID, materialID uint, locationType string, locationID uint) ([]string, error)
	GetInvoiceMaterialsWithoutSerialNumbers(id, projectID uint) ([]dto.InvoiceMaterialsWithoutSerialNumberView, error)
	GetInvoiceMaterialsWithSerialNumbers(id, projectID uint) ([]dto.InvoiceMaterialsWithSerialNumberView, error)
	GetMaterialsForEdit(id uint, locationType string, locationID, projectID uint) ([]dto.InvoiceReturnMaterialForEdit, error)
	GetMaterialAmountByMaterialID(projectID, materialID, locationID uint, locationType string) (float64, error)
}

func (u *invoiceReturnUsecase) GetAll() ([]model.InvoiceReturn, error) {
	rows, err := u.q.ListInvoiceReturns(context.Background())
	if err != nil {
		return nil, err
	}
	out := make([]model.InvoiceReturn, len(rows))
	for i, r := range rows {
		out[i] = toModelInvoiceReturn(r)
	}
	return out, nil
}

func (u *invoiceReturnUsecase) GetByID(id uint) (model.InvoiceReturn, error) {
	row, err := u.q.GetInvoiceReturn(context.Background(), int64(id))
	if err != nil {
		return model.InvoiceReturn{}, err
	}
	return toModelInvoiceReturn(row), nil
}

func (u *invoiceReturnUsecase) GetPaginatedTeam(page, limit int, projectID uint) ([]dto.InvoiceReturnTeamPaginatedQueryData, error) {
	rows, err := u.q.ListInvoiceReturnsPaginatedTeam(context.Background(), db.ListInvoiceReturnsPaginatedTeamParams{
		ProjectID: pgInt8(projectID),
		Limit:     int32(limit),
		Offset:    int32((page - 1) * limit),
	})
	if err != nil {
		return nil, err
	}
	out := make([]dto.InvoiceReturnTeamPaginatedQueryData, len(rows))
	for i, r := range rows {
		out[i] = dto.InvoiceReturnTeamPaginatedQueryData{
			ID:             uint(r.ID),
			DeliveryCode:   r.DeliveryCode,
			DistrictName:   r.DistrictName,
			AcceptorName:   r.AcceptorName,
			TeamNumber:     r.TeamNumber,
			TeamLeaderName: r.TeamLeaderName,
			DateOfInvoice:  timeFromPgTimestamptz(r.DateOfInvoice).String(),
			Confirmation:   r.Confirmation,
		}
	}
	return out, nil
}

func (u *invoiceReturnUsecase) GetPaginatedObject(page, limit int, projectID uint) ([]dto.InvoiceReturnObjectPaginated, error) {
	rows, err := u.q.ListInvoiceReturnsPaginatedObject(context.Background(), db.ListInvoiceReturnsPaginatedObjectParams{
		ProjectID: pgInt8(projectID),
		Limit:     int32(limit),
		Offset:    int32((page - 1) * limit),
	})
	if err != nil {
		return nil, err
	}

	queryData := make([]dto.InvoiceReturnObjectPaginatedQueryData, len(rows))
	for i, r := range rows {
		queryData[i] = dto.InvoiceReturnObjectPaginatedQueryData{
			ID:                   uint(r.ID),
			DeliveryCode:         r.DeliveryCode,
			AcceptorName:         r.AcceptorName,
			DistrictName:         r.DistrictName,
			ObjectName:           r.ObjectName,
			ObjectSupervisorName: r.ObjectSupervisorName,
			ObjectType:           r.ObjectType,
			TeamNumber:           r.TeamNumber,
			TeamLeaderName:       r.TeamLeaderName,
			DateOfInvoice:        timeFromPgTimestamptz(r.DateOfInvoice).String(),
			Confirmation:         r.Confirmation,
		}
	}

	result := []dto.InvoiceReturnObjectPaginated{}
	currentInvoice := dto.InvoiceReturnObjectPaginated{}
	for index, entry := range queryData {
		if index == 0 {
			currentInvoice = dto.InvoiceReturnObjectPaginated{
				ID:                    entry.ID,
				DeliveryCode:          entry.DeliveryCode,
				DateOfInvoice:         entry.DateOfInvoice,
				ObjectName:            entry.ObjectName,
				ObjectType:            entry.ObjectType,
				AcceptorName:          entry.AcceptorName,
				DistrictName:          entry.DistrictName,
				TeamNumber:            entry.TeamNumber,
				TeamLeaderName:        entry.TeamLeaderName,
				ObjectSupervisorNames: []string{},
				Confirmation:          entry.Confirmation,
			}
		}

		if currentInvoice.ID == entry.ID {
			currentInvoice.ObjectSupervisorNames = append(currentInvoice.ObjectSupervisorNames, entry.ObjectSupervisorName)
		} else {
			result = append(result, currentInvoice)
			currentInvoice = dto.InvoiceReturnObjectPaginated{
				ID:                    entry.ID,
				DeliveryCode:          entry.DeliveryCode,
				DateOfInvoice:         entry.DateOfInvoice,
				AcceptorName:          entry.AcceptorName,
				DistrictName:          entry.DistrictName,
				ObjectName:            entry.ObjectName,
				ObjectSupervisorNames: []string{entry.ObjectSupervisorName},
				ObjectType:            entry.ObjectType,
				TeamNumber:            entry.TeamNumber,
				TeamLeaderName:        entry.TeamLeaderName,
				Confirmation:          entry.Confirmation,
			}
		}
	}

	if len(queryData) != 0 {
		result = append(result, currentInvoice)
	}
	return result, nil
}

func (u *invoiceReturnUsecase) Create(data dto.InvoiceReturn) (model.InvoiceReturn, error) {
	ctx := context.Background()

	count, err := u.q.GetInvoiceCount(ctx, db.GetInvoiceCountParams{
		InvoiceType: pgText("return"),
		ProjectID:   pgInt8(data.Details.ProjectID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		count = 0
	} else if err != nil {
		return model.InvoiceReturn{}, err
	}

	data.Details.DeliveryCode = utils.UniqueCodeGeneration("В", count+1, data.Details.ProjectID)

	invoiceMaterialsForCreate, err := u.buildInvoiceReturnItems(ctx, data)
	if err != nil {
		return model.InvoiceReturn{}, err
	}

	if err := u.GenerateExcel(data); err != nil {
		return model.InvoiceReturn{}, err
	}

	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return model.InvoiceReturn{}, err
	}
	defer tx.Rollback(ctx)
	qtx := u.q.WithTx(tx)

	invoiceRow, err := qtx.CreateInvoiceReturn(ctx, db.CreateInvoiceReturnParams{
		ProjectID:          pgInt8(data.Details.ProjectID),
		DistrictID:         pgInt8(data.Details.DistrictID),
		ReturnerType:       pgText(data.Details.ReturnerType),
		ReturnerID:         pgInt8(data.Details.ReturnerID),
		AcceptorType:       pgText(data.Details.AcceptorType),
		AcceptorID:         pgInt8(data.Details.AcceptorID),
		AcceptedByWorkerID: pgInt8(data.Details.AcceptedByWorkerID),
		DateOfInvoice:      pgTimestamptz(data.Details.DateOfInvoice),
		Notes:              pgText(data.Details.Notes),
		DeliveryCode:       pgText(data.Details.DeliveryCode),
		Confirmation:       pgBool(data.Details.Confirmation),
	})
	if err != nil {
		return model.InvoiceReturn{}, err
	}

	if err := writeInvoiceMaterialsBatch(ctx, qtx, uint(invoiceRow.ID), invoiceMaterialsForCreate); err != nil {
		return model.InvoiceReturn{}, err
	}

	if err := qtx.IncrementInvoiceCount(ctx, db.IncrementInvoiceCountParams{
		InvoiceType: pgText("return"),
		ProjectID:   invoiceRow.ProjectID,
	}); err != nil {
		return model.InvoiceReturn{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return model.InvoiceReturn{}, err
	}
	return toModelInvoiceReturn(invoiceRow), nil
}

// buildInvoiceReturnItems mirrors the GORM-era logic that splits an
// invoice's items into per-cost invoice_materials rows using the cost-ASC
// material_locations lookup (returner gives back the lowest-cost material
// first). Serial-number returns are explicitly stubbed in the GORM version
// (returns an "Operation under testing" error); preserved here.
func (u *invoiceReturnUsecase) buildInvoiceReturnItems(ctx context.Context, data dto.InvoiceReturn) ([]model.InvoiceMaterials, error) {
	out := []model.InvoiceMaterials{}
	for _, item := range data.Items {
		if len(item.SerialNumbers) != 0 {
			return nil, fmt.Errorf("Операция возврат через серийный номер тестируется")
		}

		rows, err := u.q.ListMaterialAmountReverseSortedByCostM19InLocation(ctx, db.ListMaterialAmountReverseSortedByCostM19InLocationParams{
			ProjectID:    pgInt8(data.Details.ProjectID),
			LocationType: pgText(data.Details.ReturnerType),
			LocationID:   pgInt8(data.Details.ReturnerID),
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
				MaterialCostID: uint(rows[index].MaterialCostID),
				ProjectID:      data.Details.ProjectID,
				InvoiceType:    "return",
				IsDefected:     item.IsDefected,
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

func (u *invoiceReturnUsecase) Update(data dto.InvoiceReturn) (model.InvoiceReturn, error) {
	ctx := context.Background()
	invoiceMaterialsForCreate, err := u.buildInvoiceReturnItems(ctx, data)
	if err != nil {
		return model.InvoiceReturn{}, err
	}

	excelFilePath := filepath.Join("./storage/import_excel/return/", data.Details.DeliveryCode+".xlsx")
	if err := os.Remove(excelFilePath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return model.InvoiceReturn{}, err
	}

	if err := u.GenerateExcel(data); err != nil {
		return model.InvoiceReturn{}, err
	}

	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return model.InvoiceReturn{}, err
	}
	defer tx.Rollback(ctx)
	qtx := u.q.WithTx(tx)

	invoiceRow, err := qtx.UpdateInvoiceReturn(ctx, db.UpdateInvoiceReturnParams{
		ID:                 int64(data.Details.ID),
		ProjectID:          pgInt8(data.Details.ProjectID),
		DistrictID:         pgInt8(data.Details.DistrictID),
		ReturnerType:       pgText(data.Details.ReturnerType),
		ReturnerID:         pgInt8(data.Details.ReturnerID),
		AcceptorType:       pgText(data.Details.AcceptorType),
		AcceptorID:         pgInt8(data.Details.AcceptorID),
		AcceptedByWorkerID: pgInt8(data.Details.AcceptedByWorkerID),
		DateOfInvoice:      pgTimestamptz(data.Details.DateOfInvoice),
		Notes:              pgText(data.Details.Notes),
		DeliveryCode:       pgText(data.Details.DeliveryCode),
		Confirmation:       pgBool(data.Details.Confirmation),
	})
	if err != nil {
		return model.InvoiceReturn{}, err
	}

	if err := qtx.DeleteInvoiceMaterialsByInvoice(ctx, db.DeleteInvoiceMaterialsByInvoiceParams{
		InvoiceType: pgText("return"),
		InvoiceID:   pgInt8(uint(invoiceRow.ID)),
	}); err != nil {
		return model.InvoiceReturn{}, err
	}

	if err := qtx.DeleteSerialNumberMovementsByInvoice(ctx, db.DeleteSerialNumberMovementsByInvoiceParams{
		InvoiceType: pgText("return"),
		InvoiceID:   pgInt8(uint(invoiceRow.ID)),
	}); err != nil {
		return model.InvoiceReturn{}, err
	}

	if err := writeInvoiceMaterialsBatch(ctx, qtx, uint(invoiceRow.ID), invoiceMaterialsForCreate); err != nil {
		return model.InvoiceReturn{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return model.InvoiceReturn{}, err
	}
	return toModelInvoiceReturn(invoiceRow), nil
}

func (u *invoiceReturnUsecase) Delete(id uint) error {
	ctx := context.Background()
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	qtx := u.q.WithTx(tx)

	if err := qtx.DeleteInvoiceReturn(ctx, int64(id)); err != nil {
		return err
	}
	if err := qtx.DeleteInvoiceMaterialsByInvoice(ctx, db.DeleteInvoiceMaterialsByInvoiceParams{
		InvoiceType: pgText("return"),
		InvoiceID:   pgInt8(id),
	}); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (u *invoiceReturnUsecase) CountBasedOnType(projectID uint, invoiceType string) (int64, error) {
	return u.q.CountInvoiceReturnsBasedOnType(context.Background(), db.CountInvoiceReturnsBasedOnTypeParams{
		ProjectID:    pgInt8(projectID),
		ReturnerType: pgText(invoiceType),
	})
}

func (u *invoiceReturnUsecase) Confirmation(id uint) error {
	ctx := context.Background()
	invoice, err := u.q.GetInvoiceReturn(ctx, int64(id))
	if err != nil {
		return err
	}

	invoiceMaterials, err := u.q.ListInvoiceMaterialsByInvoice(ctx, db.ListInvoiceMaterialsByInvoiceParams{
		InvoiceID:   pgInt8(id),
		InvoiceType: pgText("return"),
		ProjectID:   invoice.ProjectID,
	})
	if err != nil {
		return err
	}

	materialsInReturnerLocation, err := u.q.ListMaterialLocationsForInvoiceConfirmation(ctx, db.ListMaterialLocationsForInvoiceConfirmationParams{
		LocationType: invoice.ReturnerType,
		LocationID:   invoice.ReturnerID,
		InvoiceType:  pgText("return"),
		InvoiceID:    pgInt8(id),
	})
	if err != nil {
		return err
	}

	materialsInAcceptorLocation, err := u.q.ListMaterialLocationsForInvoiceConfirmation(ctx, db.ListMaterialLocationsForInvoiceConfirmationParams{
		LocationType: invoice.AcceptorType,
		LocationID:   invoice.AcceptorID,
		InvoiceType:  pgText("return"),
		InvoiceID:    pgInt8(id),
	})
	if err != nil {
		return err
	}

	type returnerLoc struct {
		row    db.MaterialLocation
		amount float64
	}
	type acceptorLoc struct {
		row    *db.MaterialLocation
		params *db.CreateMaterialLocationParams
		amount float64
	}

	returnerSlice := make([]returnerLoc, len(materialsInReturnerLocation))
	for i, ml := range materialsInReturnerLocation {
		returnerSlice[i] = returnerLoc{row: ml, amount: float64FromPgNumeric(ml.Amount)}
	}

	acceptorSlice := make([]*acceptorLoc, 0, len(materialsInAcceptorLocation))
	for i := range materialsInAcceptorLocation {
		ml := materialsInAcceptorLocation[i]
		acceptorSlice = append(acceptorSlice, &acceptorLoc{row: &ml, amount: float64FromPgNumeric(ml.Amount)})
	}

	// We need to apply, after writing the upserts, two kinds of defect ops:
	// (a) update an existing material_defect row's amount + material_location_id
	//     when both the acceptor location AND the material_defect already exist;
	// (b) create a new material_defect row pointing at a freshly-created
	//     material_location row when the acceptor location didn't exist.
	type existingDefect struct {
		defectID            int64
		newAmount           float64
		newMaterialLocation int64
	}
	type newDefect struct {
		acceptorParamsIndex int
		amount              float64
	}
	existingDefects := []existingDefect{}
	newDefects := []newDefect{}

	for _, im := range invoiceMaterials {
		// returner side
		returnerIndex := -1
		for index, r := range returnerSlice {
			if uintFromPgInt8(r.row.MaterialCostID) == uintFromPgInt8(im.MaterialCostID) {
				returnerIndex = index
				break
			}
		}
		if returnerIndex == -1 {
			return fmt.Errorf("Возвращаемый материал не найден на источнике")
		}
		imAmount := float64FromPgNumeric(im.Amount)
		returnerSlice[returnerIndex].amount -= imAmount

		// acceptor side
		acceptorIndex := -1
		for index, a := range acceptorSlice {
			var ml *db.MaterialLocation
			if a.row != nil {
				ml = a.row
			}
			if ml != nil && uintFromPgInt8(ml.MaterialCostID) == uintFromPgInt8(im.MaterialCostID) {
				acceptorIndex = index
				break
			}
			if a.params != nil && uintFromPgInt8(a.params.MaterialCostID) == uintFromPgInt8(im.MaterialCostID) {
				acceptorIndex = index
				break
			}
		}

		newAcceptorRow := false
		if acceptorIndex == -1 {
			acceptorSlice = append(acceptorSlice, &acceptorLoc{
				params: &db.CreateMaterialLocationParams{
					ProjectID:      invoice.ProjectID,
					MaterialCostID: im.MaterialCostID,
					LocationType:   invoice.AcceptorType,
					LocationID:     invoice.AcceptorID,
					Amount:         im.Amount,
				},
				amount: imAmount,
			})
			acceptorIndex = len(acceptorSlice) - 1
			newAcceptorRow = true
		} else {
			acceptorSlice[acceptorIndex].amount += imAmount
		}

		if !im.IsDefected.Bool {
			continue
		}

		// material_defect bookkeeping
		// GORM looked up the existing defect by the *returner-side* material_location_id
		// (which it had just decremented). We do the same — preserved verbatim.
		returnerLocID := returnerSlice[returnerIndex].row.ID
		existingDefect_, err := u.q.GetMaterialDefectByMaterialLocationID(ctx, pgInt8(uint(returnerLocID)))
		hasExistingDefect := true
		if errors.Is(err, pgx.ErrNoRows) {
			hasExistingDefect = false
		} else if err != nil {
			return err
		}

		if newAcceptorRow {
			newDefects = append(newDefects, newDefect{
				acceptorParamsIndex: acceptorIndex,
				amount:              imAmount,
			})
			_ = existingDefect_
			_ = hasExistingDefect
			continue
		}

		// Acceptor row exists; bind the defect to its id.
		// GORM-era logic added imAmount to the existing-defect amount; preserved.
		newAmount := imAmount
		if hasExistingDefect {
			newAmount += float64FromPgNumeric(existingDefect_.Amount)
		}
		existingDefects = append(existingDefects, existingDefect{
			defectID:            existingDefect_.ID,
			newAmount:           newAmount,
			newMaterialLocation: acceptorSlice[acceptorIndex].row.ID,
		})
	}

	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	qtx := u.q.WithTx(tx)

	if err := qtx.ConfirmInvoiceReturn(ctx, int64(id)); err != nil {
		return err
	}

	for _, r := range returnerSlice {
		if err := qtx.UpsertMaterialLocationByID(ctx, db.UpsertMaterialLocationByIDParams{
			Amount: pgNumericFromFloat64(r.amount),
			ID:     r.row.ID,
		}); err != nil {
			return err
		}
	}

	createdAcceptorIDs := make([]int64, len(acceptorSlice))
	for i, a := range acceptorSlice {
		if a.row != nil {
			if err := qtx.UpsertMaterialLocationByID(ctx, db.UpsertMaterialLocationByIDParams{
				Amount: pgNumericFromFloat64(a.amount),
				ID:     a.row.ID,
			}); err != nil {
				return err
			}
			createdAcceptorIDs[i] = a.row.ID
		} else {
			params := *a.params
			params.Amount = pgNumericFromFloat64(a.amount)
			row, err := qtx.CreateMaterialLocation(ctx, params)
			if err != nil {
				return err
			}
			createdAcceptorIDs[i] = row.ID
		}
	}

	for _, d := range existingDefects {
		if err := qtx.UpsertMaterialDefectByID(ctx, db.UpsertMaterialDefectByIDParams{
			Amount:             pgNumericFromFloat64(d.newAmount),
			MaterialLocationID: d.newMaterialLocation,
			ID:                 d.defectID,
		}); err != nil {
			return err
		}
	}

	for _, d := range newDefects {
		if _, err := qtx.CreateMaterialDefect(ctx, db.CreateMaterialDefectParams{
			Amount:             pgNumericFromFloat64(d.amount),
			MaterialLocationID: pgInt8(uint(createdAcceptorIDs[d.acceptorParamsIndex])),
		}); err != nil {
			return err
		}
	}

	if err := qtx.ConfirmSerialNumberMovementsByInvoice(ctx, db.ConfirmSerialNumberMovementsByInvoiceParams{
		InvoiceID:   pgInt8(id),
		InvoiceType: pgText("return"),
	}); err != nil {
		return err
	}

	if err := qtx.ConfirmSerialNumberLocationsByReturnInvoice(ctx, db.ConfirmSerialNumberLocationsByReturnInvoiceParams{
		AcceptorID: int64(uintFromPgInt8(invoice.AcceptorID)),
		InvoiceID:  int64(id),
	}); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (u *invoiceReturnUsecase) UniqueCode(projectID uint) ([]string, error) {
	return u.q.ListInvoiceReturnUniqueCodes(context.Background(), pgInt8(projectID))
}

func (u *invoiceReturnUsecase) UniqueTeam(projectID uint) ([]string, error) {
	ctx := context.Background()
	teamIDs, err := u.q.ListInvoiceReturnUniqueTeams(ctx, pgInt8(projectID))
	if err != nil {
		return nil, err
	}
	out := []string{}
	for _, id := range teamIDs {
		team, err := u.q.GetTeam(ctx, id)
		if err != nil {
			return nil, err
		}
		out = append(out, team.Number.String)
	}
	return out, nil
}

func (u *invoiceReturnUsecase) UniqueObject(projectID uint) ([]string, error) {
	ctx := context.Background()
	objectIDs, err := u.q.ListInvoiceReturnUniqueObjects(ctx, pgInt8(projectID))
	if err != nil {
		return nil, err
	}
	out := []string{}
	for _, id := range objectIDs {
		object, err := u.q.GetObject(ctx, id)
		if err != nil {
			return nil, err
		}
		out = append(out, object.Name.String)
	}
	return out, nil
}

func (u *invoiceReturnUsecase) Report(filter dto.InvoiceReturnReportFilterRequest, projectID uint) (string, error) {
	ctx := context.Background()
	newFilter := dto.InvoiceReturnReportFilter{
		Code:     filter.Code,
		DateFrom: filter.DateFrom,
		DateTo:   filter.DateTo,
	}

	if filter.ReturnerType == "team" {
		newFilter.ReturnerType = "team"
		if filter.Returner != "" {
			team, err := u.q.GetTeamByNumber(ctx, pgText(filter.Returner))
			if err != nil {
				return "", err
			}
			newFilter.ReturnerID = uint(team.ID)
		}
	}
	if filter.ReturnerType == "object" {
		newFilter.ReturnerType = "object"
		if filter.Returner != "" {
			object, err := u.q.GetObjectByName(ctx, pgText(filter.Returner))
			if err != nil {
				return "", err
			}
			newFilter.ReturnerID = uint(object.ID)
		}
	}
	if filter.ReturnerType == "all" {
		newFilter.ReturnerType = ""
		newFilter.ReturnerID = 0
	}

	invoiceRows, err := u.q.ListInvoiceReturnReportData(ctx, db.ListInvoiceReturnReportDataParams{
		ProjectID: pgInt8(projectID),
		Column2:   newFilter.Code,
		Column3:   newFilter.ReturnerType,
		Column4:   int64(newFilter.ReturnerID),
		Column5:   pgTimestamptz(newFilter.DateFrom),
		Column6:   pgTimestamptz(newFilter.DateTo),
	})
	if err != nil {
		return "", err
	}

	templateFilePath := filepath.Join("./internal/templates/", "Invoice Return Report.xlsx")
	f, err := excelize.OpenFile(templateFilePath)
	defer f.Close()
	if err != nil {
		return "", err
	}

	sheetName := "Sheet1"
	f.SetCellStr(sheetName, "K1", "ID материала")
	f.SetCellStr(sheetName, "L1", "Код материала")

	rowCount := 2

	type InvoiceReturnReportData struct {
		DeliveryCode      string
		InvoiceReturnType string
		Returner          string
		DateOfInvoice     time.Time
		MaterialID        uint
		MaterialName      string
		MaterialCode      string
		MaterialUnit      string
		Amount            float64
		Price             float64
		IsDefected        string
		Notes             string
	}

	reportData := []InvoiceReturnReportData{}
	for _, invoice := range invoiceRows {
		invoiceMaterials, err := u.q.ListInvoiceMaterialsByInvoice(ctx, db.ListInvoiceMaterialsByInvoiceParams{
			InvoiceID:   pgInt8(uint(invoice.ID)),
			InvoiceType: pgText("return"),
			ProjectID:   pgInt8(projectID),
		})
		if err != nil {
			return "", err
		}

		for _, im := range invoiceMaterials {
			oneEntry := InvoiceReturnReportData{
				DeliveryCode:  invoice.DeliveryCode.String,
				DateOfInvoice: timeFromPgTimestamptz(invoice.DateOfInvoice),
				Amount:        float64FromPgNumeric(im.Amount),
			}

			if invoice.ReturnerType.String == "team" {
				team, err := u.q.GetTeam(ctx, int64(uintFromPgInt8(invoice.ReturnerID)))
				if err != nil {
					return "", err
				}
				oneEntry.InvoiceReturnType = "Бригада"
				oneEntry.Returner = team.Number.String
			}

			if invoice.ReturnerType.String == "object" {
				object, err := u.q.GetObject(ctx, int64(uintFromPgInt8(invoice.ReturnerID)))
				if err != nil {
					return "", err
				}
				oneEntry.InvoiceReturnType = "Объект"
				oneEntry.Returner = object.Name.String
			}

			materialCost, err := u.q.GetMaterialCost(ctx, int64(uintFromPgInt8(im.MaterialCostID)))
			if err != nil {
				return "", nil
			}

			material, err := u.q.GetMaterial(ctx, int64(uintFromPgInt8(materialCost.MaterialID)))
			if err != nil {
				return "", nil
			}

			oneEntry.MaterialID = uint(material.ID)
			oneEntry.MaterialName = material.Name.String
			oneEntry.MaterialCode = material.Code.String
			oneEntry.MaterialUnit = material.Unit.String
			oneEntry.Price, _ = decimalFromPgNumeric(materialCost.CostM19).Float64()
			if im.IsDefected.Bool {
				oneEntry.IsDefected = "Да"
			} else {
				oneEntry.IsDefected = "Нет"
			}
			oneEntry.Notes = im.Notes.String

			reportData = append(reportData, oneEntry)
		}
	}

	for _, oneEntry := range reportData {
		f.SetCellValue(sheetName, "A"+fmt.Sprint(rowCount), oneEntry.DeliveryCode)
		f.SetCellValue(sheetName, "B"+fmt.Sprint(rowCount), oneEntry.InvoiceReturnType)
		f.SetCellValue(sheetName, "C"+fmt.Sprint(rowCount), oneEntry.Returner)

		dateOfInvoice := oneEntry.DateOfInvoice.String()
		if len(dateOfInvoice) > 10 {
			dateOfInvoice = dateOfInvoice[:len(dateOfInvoice)-10]
		}
		f.SetCellValue(sheetName, "D"+fmt.Sprint(rowCount), dateOfInvoice)

		f.SetCellValue(sheetName, "E"+fmt.Sprint(rowCount), oneEntry.MaterialName)
		f.SetCellValue(sheetName, "F"+fmt.Sprint(rowCount), oneEntry.MaterialUnit)
		f.SetCellValue(sheetName, "G"+fmt.Sprint(rowCount), oneEntry.Amount)
		f.SetCellValue(sheetName, "H"+fmt.Sprint(rowCount), oneEntry.Price)
		f.SetCellValue(sheetName, "I"+fmt.Sprint(rowCount), oneEntry.IsDefected)
		f.SetCellValue(sheetName, "J"+fmt.Sprint(rowCount), oneEntry.Notes)
		f.SetCellInt(sheetName, "K"+fmt.Sprint(rowCount), int(oneEntry.MaterialID))
		f.SetCellStr(sheetName, "L"+fmt.Sprint(rowCount), oneEntry.MaterialCode)
		rowCount++
	}

	currentTime := time.Now()
	fileName := fmt.Sprintf(
		"Отсчет накладной возврат - %s.xlsx",
		currentTime.Format("02-01-2006"),
	)
	tempFilePath := filepath.Join("./storage/import_excel/temp/", fileName)
	f.SaveAs(tempFilePath)
	return fileName, nil
}

func (u *invoiceReturnUsecase) GetMaterialsInLocation(projectID, locationID uint, locationType string) ([]dto.InvoiceReturnMaterialForSelect, error) {
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
		amount, err := u.q.GetUniqueMaterialTotalAmount(ctx, db.GetUniqueMaterialTotalAmountParams{
			ProjectID:      pgInt8(projectID),
			LocationType:   pgText(locationType),
			LocationID:     pgInt8(locationID),
			MaterialCostID: pgInt8(uint(m.ID)),
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

func (u *invoiceReturnUsecase) GetMaterialCostInLocation(projectID, locationID, materialID uint, locationType string) ([]model.MaterialCost, error) {
	rows, err := u.q.ListUniqueMaterialCostsFromLocation(context.Background(), db.ListUniqueMaterialCostsFromLocationParams{
		ProjectID:    pgInt8(projectID),
		LocationType: pgText(locationType),
		LocationID:   pgInt8(locationID),
		ID:           int64(materialID),
	})
	if err != nil {
		return nil, err
	}
	out := make([]model.MaterialCost, len(rows))
	for i, r := range rows {
		out[i] = toModelMaterialCost(r)
	}
	return out, nil
}

func (u *invoiceReturnUsecase) GetMaterialAmountInLocation(projectID, locationID, materialCostID uint, locationType string) (float64, error) {
	amount, err := u.q.GetUniqueMaterialTotalAmount(context.Background(), db.GetUniqueMaterialTotalAmountParams{
		ProjectID:      pgInt8(projectID),
		LocationType:   pgText(locationType),
		LocationID:     pgInt8(locationID),
		MaterialCostID: pgInt8(materialCostID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return float64FromPgNumeric(amount), nil
}

func (u *invoiceReturnUsecase) GetSerialNumberCodesInLocation(projectID, materialID uint, locationType string, locationID uint) ([]string, error) {
	return u.q.GetSerialNumberCodesByMaterialIDAndLocation(context.Background(), db.GetSerialNumberCodesByMaterialIDAndLocationParams{
		ProjectID:    pgInt8(projectID),
		ID:           int64(materialID),
		LocationType: pgText(locationType),
		LocationID:   pgInt8(locationID),
	})
}

func (u *invoiceReturnUsecase) GetInvoiceMaterialsWithoutSerialNumbers(id, projectID uint) ([]dto.InvoiceMaterialsWithoutSerialNumberView, error) {
	rows, err := u.q.ListInvoiceMaterialsWithoutSerialNumbers(context.Background(), db.ListInvoiceMaterialsWithoutSerialNumbersParams{
		InvoiceType: pgText("return"),
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

func (u *invoiceReturnUsecase) GetInvoiceMaterialsWithSerialNumbers(id, projectID uint) ([]dto.InvoiceMaterialsWithSerialNumberView, error) {
	rows, err := u.q.ListInvoiceMaterialsWithSerialNumbers(context.Background(), db.ListInvoiceMaterialsWithSerialNumbersParams{
		InvoiceType: pgText("return"),
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
				IsDefected:    materialInfo.IsDefected,
				SerialNumbers: []string{},
				Amount:        materialInfo.Amount,
				CostM19:       materialInfo.CostM19,
				Notes:         materialInfo.Notes,
			}
		}

		if current.MaterialName == materialInfo.MaterialName &&
			current.CostM19.Equal(materialInfo.CostM19) &&
			current.IsDefected == materialInfo.IsDefected {
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
				IsDefected:    materialInfo.IsDefected,
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

func (u *invoiceReturnUsecase) GetMaterialsForEdit(id uint, locationType string, locationID, projectID uint) ([]dto.InvoiceReturnMaterialForEdit, error) {
	rows, err := u.q.ListInvoiceReturnMaterialsForEdit(context.Background(), db.ListInvoiceReturnMaterialsForEditParams{
		InvoiceID:    pgInt8(id),
		LocationType: pgText(locationType),
		LocationID:   pgInt8(locationID),
		ProjectID:    pgInt8(projectID),
	})
	if err != nil {
		return []dto.InvoiceReturnMaterialForEdit{}, nil
	}

	data := make([]dto.InvoiceReturnMaterialForEdit, len(rows))
	for i, r := range rows {
		data[i] = dto.InvoiceReturnMaterialForEdit{
			MaterialID:      uint(r.MaterialID),
			MaterialName:    r.MaterialName,
			Unit:            r.Unit,
			Amount:          float64FromPgNumeric(r.Amount),
			Notes:           r.Notes,
			HasSerialNumber: r.HasSerialNumber,
			IsDefective:     r.IsDefective,
			HolderAmount:    float64FromPgNumeric(r.HolderAmount),
		}
	}

	var result []dto.InvoiceReturnMaterialForEdit
	for index, entry := range data {
		if index == 0 {
			result = append(result, entry)
			continue
		}
		lastItemIndex := len(result) - 1
		if result[lastItemIndex].MaterialID == entry.MaterialID {
			result[lastItemIndex].Amount += entry.Amount
			result[lastItemIndex].HolderAmount += entry.HolderAmount
		} else {
			result = append(result, entry)
		}
	}
	return result, nil
}

func (u *invoiceReturnUsecase) GetMaterialAmountByMaterialID(projectID, materialID, locationID uint, locationType string) (float64, error) {
	amount, err := u.q.GetTotalAmountInLocation(context.Background(), db.GetTotalAmountInLocationParams{
		ProjectID:    pgInt8(projectID),
		ID:           int64(materialID),
		LocationType: pgText(locationType),
		LocationID:   pgInt8(locationID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return float64FromPgNumeric(amount), nil
}

func (u *invoiceReturnUsecase) GenerateExcel(data dto.InvoiceReturn) error {
	ctx := context.Background()
	templateFilePath := filepath.Join("./internal/templates/return.xlsx")
	f, err := excelize.OpenFile(templateFilePath)
	if err != nil {
		return err
	}

	sheetName := "Возврат"
	startingRow := 5

	f.InsertRows(sheetName, startingRow, len(data.Items))

	defaultStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Size: 8, VertAlign: "center", Family: "Times New Roman"},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		},
		Alignment: &excelize.Alignment{Horizontal: "center", WrapText: true, Vertical: "center"},
	})

	namingStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Size: 8, VertAlign: "center", Family: "Times New Roman"},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center", WrapText: true},
	})

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 10, VertAlign: "center", Bold: true, Family: "Times New Roman"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "top", WrapText: true},
	})

	project, err := u.q.GetProject(ctx, int64(data.Details.ProjectID))
	if err != nil {
		return err
	}
	district, err := u.q.GetDistrict(ctx, int64(data.Details.DistrictID))
	if err != nil {
		return err
	}

	f.SetCellStyle(sheetName, "C1", "C1", headerStyle)
	f.SetCellStr(sheetName, "C1", fmt.Sprintf(`НАКЛАДНАЯ № %s
от %s года
на возврат материала
      `, data.Details.DeliveryCode, utils.DateConverter(data.Details.DateOfInvoice)))

	f.MergeCell(sheetName, "G1", "I1")
	f.SetCellStyle(sheetName, "G1", "I1", headerStyle)
	f.SetCellStr(sheetName, "G1", fmt.Sprintf(`%s
в г. Душанбе
Регион: %s
      `, project.Name.String, district.Name.String))

	if data.Details.AcceptorType == "warehouse" {
		teamData, err := u.q.ListTeamNumberAndLeadersByID(ctx, db.ListTeamNumberAndLeadersByIDParams{
			ProjectID: pgInt8(data.Details.ProjectID),
			ID:        int64(data.Details.ReturnerID),
		})
		if err != nil {
			return err
		}

		f.SetCellStr(sheetName, "B2", "")
		f.SetCellStr(sheetName, "B3", "")
		if len(teamData) > 0 {
			f.SetCellStr(sheetName, "C"+fmt.Sprint(6+len(data.Items)), teamData[0].TeamLeaderName)
			f.SetCellStr(sheetName, "C"+fmt.Sprint(8+len(data.Items)), teamData[0].TeamLeaderName)
		}

		acceptor, err := u.q.GetWorker(ctx, int64(data.Details.AcceptedByWorkerID))
		if err != nil {
			return err
		}
		f.SetCellValue(sheetName, "C"+fmt.Sprint(10+len(data.Items)), acceptor.Name.String)
	}

	if data.Details.AcceptorType == "team" {
		object, err := u.q.GetObject(ctx, int64(data.Details.ReturnerID))
		if err != nil {
			return err
		}

		f.SetCellStr(sheetName, "D2", utils.ObjectTypeConverter(object.Type.String))
		f.SetCellStr(sheetName, "C3", object.Name.String)

		teamData, err := u.q.ListTeamNumberAndLeadersByID(ctx, db.ListTeamNumberAndLeadersByIDParams{
			ProjectID: pgInt8(data.Details.ProjectID),
			ID:        int64(data.Details.AcceptorID),
		})
		if err != nil {
			return err
		}
		if len(teamData) > 0 {
			f.SetCellStr(sheetName, "C"+fmt.Sprint(6+len(data.Items)), teamData[0].TeamLeaderName)
			f.SetCellStr(sheetName, "C"+fmt.Sprint(10+len(data.Items)), teamData[0].TeamLeaderName)
		}

		supervisorNames, err := u.q.ListSupervisorNamesByObjectID(ctx, int64(data.Details.ReturnerID))
		if err != nil {
			return err
		}
		if len(supervisorNames) > 0 {
			f.SetCellStr(sheetName, "C"+fmt.Sprint(8+len(data.Items)), supervisorNames[len(supervisorNames)-1])
		}
	}

	f.SetCellStyle(sheetName, "J4", "J4", headerStyle)
	f.SetCellStr(sheetName, "J4", "ID материала")
	for index, oneEntry := range data.Items {
		f.MergeCell(sheetName, "G"+fmt.Sprint(startingRow+index), "I"+fmt.Sprint(startingRow+index))

		f.SetCellStyle(sheetName, "A"+fmt.Sprint(startingRow+index), "J"+fmt.Sprint(startingRow+index), defaultStyle)
		f.SetCellStyle(sheetName, "B"+fmt.Sprint(startingRow+index), "B"+fmt.Sprint(startingRow+index), namingStyle)

		material, err := u.q.GetMaterial(ctx, int64(oneEntry.MaterialID))
		if err != nil {
			return err
		}

		f.SetCellValue(sheetName, "A"+fmt.Sprint(startingRow+index), index+1)
		f.SetCellValue(sheetName, "B"+fmt.Sprint(startingRow+index), material.Code.String)
		f.SetCellValue(sheetName, "C"+fmt.Sprint(startingRow+index), material.Name.String)
		f.SetCellValue(sheetName, "D"+fmt.Sprint(startingRow+index), material.Unit.String)
		f.SetCellValue(sheetName, "E"+fmt.Sprint(startingRow+index), oneEntry.Amount)
		materialDefect := "Нет"
		if oneEntry.IsDefected {
			materialDefect = "Да"
		}
		f.SetCellValue(sheetName, "F"+fmt.Sprint(startingRow+index), materialDefect)
		f.SetCellValue(sheetName, "G"+fmt.Sprint(startingRow+index), oneEntry.Notes)
		f.SetCellInt(sheetName, "J"+fmt.Sprint(startingRow+index), int(material.ID))
	}

	savePath := filepath.Join("./storage/import_excel/return/", data.Details.DeliveryCode+".xlsx")
	f.SaveAs(savePath)
	if err := f.Close(); err != nil {
		fmt.Println(err)
	}
	return nil
}

func (u *invoiceReturnUsecase) GetDocument(deliveryCode string, projectID uint) (string, error) {
	invoice, err := u.q.GetInvoiceReturnByDeliveryCode(context.Background(), db.GetInvoiceReturnByDeliveryCodeParams{
		DeliveryCode: pgText(deliveryCode),
		ProjectID:    pgInt8(projectID),
	})
	if err != nil {
		return "", err
	}
	if invoice.Confirmation.Bool {
		return ".pdf", nil
	}
	return ".xlsx", nil
}

func toModelInvoiceReturn(r db.InvoiceReturn) model.InvoiceReturn {
	return model.InvoiceReturn{
		ID:                 uint(r.ID),
		ProjectID:          uintFromPgInt8(r.ProjectID),
		DistrictID:         uintFromPgInt8(r.DistrictID),
		ReturnerType:       r.ReturnerType.String,
		ReturnerID:         uintFromPgInt8(r.ReturnerID),
		AcceptorType:       r.AcceptorType.String,
		AcceptorID:         uintFromPgInt8(r.AcceptorID),
		AcceptedByWorkerID: uintFromPgInt8(r.AcceptedByWorkerID),
		DateOfInvoice:      timeFromPgTimestamptz(r.DateOfInvoice),
		Notes:              r.Notes.String,
		DeliveryCode:       r.DeliveryCode.String,
		Confirmation:       r.Confirmation.Bool,
	}
}
