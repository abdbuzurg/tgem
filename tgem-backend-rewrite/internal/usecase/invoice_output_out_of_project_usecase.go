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

type invoiceOutputOutOfProjectUsecase struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

func NewInvoiceOutputOutOfProjectUsecase(pool *pgxpool.Pool) IInvoiceOutputOutOfProjectUsecase {
	return &invoiceOutputOutOfProjectUsecase{
		pool: pool,
		q:    db.New(pool),
	}
}

type IInvoiceOutputOutOfProjectUsecase interface {
	GetPaginated(page, limit int, filter dto.InvoiceOutputOutOfProjectSearchParameters) ([]dto.InvoiceOutputOutOfProjectPaginated, error)
	GetByID(id uint) (model.InvoiceOutputOutOfProject, error)
	Count(data dto.InvoiceOutputOutOfProjectSearchParameters) (int64, error)
	Create(data dto.InvoiceOutputOutOfProject) (model.InvoiceOutputOutOfProject, error)
	Delete(id uint) error
	GetInvoiceMaterialsWithoutSerialNumbers(id uint) ([]dto.InvoiceMaterialsWithoutSerialNumberView, error)
	GetInvoiceMaterialsWithSerialNumbers(id uint) ([]dto.InvoiceMaterialsWithSerialNumberView, error)
	Update(data dto.InvoiceOutputOutOfProject) (model.InvoiceOutputOutOfProject, error)
	Confirmation(id uint) error
	GetMaterialsForEdit(id uint) ([]dto.InvoiceOutputMaterialsForEdit, error)
	GetUniqueNameOfProjects(projectID uint) ([]string, error)
	Report(filter dto.InvoiceOutputOutOfProjectReportFilter) (string, error)
	GetDocument(deliveryCode string) (string, error)
}

func (u *invoiceOutputOutOfProjectUsecase) GetPaginated(page, limit int, filter dto.InvoiceOutputOutOfProjectSearchParameters) ([]dto.InvoiceOutputOutOfProjectPaginated, error) {
	rows, err := u.q.ListInvoiceOutputOutOfProjectsPaginated(context.Background(), db.ListInvoiceOutputOutOfProjectsPaginatedParams{
		ProjectID: pgInt8(filter.ProjectID),
		Column2:   filter.NameOfProject,
		Column3:   int64(filter.ReleasedWorkerID),
		Limit:     int32(limit),
		Offset:    int32((page - 1) * limit),
	})
	if err != nil {
		return nil, err
	}
	out := make([]dto.InvoiceOutputOutOfProjectPaginated, len(rows))
	for i, r := range rows {
		out[i] = dto.InvoiceOutputOutOfProjectPaginated{
			ID:                 uint(r.ID),
			NameOfProject:      r.NameOfProject,
			DeliveryCode:       r.DeliveryCode,
			ReleasedWorkerName: r.ReleasedWorkerName,
			DateOfInvoice:      timeFromPgTimestamptz(r.DateOfInvoice),
			Confirmation:       r.Confirmation,
		}
	}
	return out, nil
}

func (u *invoiceOutputOutOfProjectUsecase) Count(filter dto.InvoiceOutputOutOfProjectSearchParameters) (int64, error) {
	return u.q.CountInvoiceOutputOutOfProjects(context.Background(), db.CountInvoiceOutputOutOfProjectsParams{
		ProjectID: pgInt8(filter.ProjectID),
		Column2:   filter.NameOfProject,
		Column3:   int64(filter.ReleasedWorkerID),
	})
}

func (u *invoiceOutputOutOfProjectUsecase) Create(data dto.InvoiceOutputOutOfProject) (model.InvoiceOutputOutOfProject, error) {
	ctx := context.Background()

	count, err := u.q.GetInvoiceCount(ctx, db.GetInvoiceCountParams{
		InvoiceType: pgText("output-out-of-project"),
		ProjectID:   pgInt8(data.Details.ProjectID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		count = 0
	} else if err != nil {
		return model.InvoiceOutputOutOfProject{}, err
	}

	data.Details.DeliveryCode = utils.UniqueCodeGeneration("ОВ", count+1, data.Details.ProjectID)

	invoiceMaterialForCreate, err := u.buildOutOfProjectItems(ctx, data)
	if err != nil {
		return model.InvoiceOutputOutOfProject{}, err
	}

	if err := u.GenerateExcelFile(data.Details, invoiceMaterialForCreate); err != nil {
		return model.InvoiceOutputOutOfProject{}, err
	}

	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return model.InvoiceOutputOutOfProject{}, err
	}
	defer tx.Rollback(ctx)
	qtx := u.q.WithTx(tx)

	invoiceRow, err := qtx.CreateInvoiceOutputOutOfProject(ctx, db.CreateInvoiceOutputOutOfProjectParams{
		ProjectID:        pgInt8(data.Details.ProjectID),
		DeliveryCode:     pgText(data.Details.DeliveryCode),
		ReleasedWorkerID: pgInt8(data.Details.ReleasedWorkerID),
		NameOfProject:    pgText(data.Details.NameOfProject),
		DateOfInvoice:    pgTimestamptz(data.Details.DateOfInvoice),
		Notes:            pgText(data.Details.Notes),
		Confirmation:     pgBool(data.Details.Confirmation),
	})
	if err != nil {
		return model.InvoiceOutputOutOfProject{}, err
	}

	if err := writeInvoiceMaterialsBatch(ctx, qtx, uint(invoiceRow.ID), invoiceMaterialForCreate); err != nil {
		return model.InvoiceOutputOutOfProject{}, err
	}

	if err := qtx.IncrementInvoiceCount(ctx, db.IncrementInvoiceCountParams{
		InvoiceType: pgText("output-out-of-project"),
		ProjectID:   invoiceRow.ProjectID,
	}); err != nil {
		return model.InvoiceOutputOutOfProject{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return model.InvoiceOutputOutOfProject{}, err
	}
	return toModelInvoiceOutputOutOfProject(invoiceRow), nil
}

// buildOutOfProjectItems mirrors the GORM-era logic that splits an invoice's
// items into per-cost invoice_materials rows using the cost-DESC FIFO
// material_locations lookup. Serial numbers were stubbed out in the GORM
// version (commented blocks) so they're stubbed here too.
func (u *invoiceOutputOutOfProjectUsecase) buildOutOfProjectItems(ctx context.Context, data dto.InvoiceOutputOutOfProject) ([]model.InvoiceMaterials, error) {
	out := []model.InvoiceMaterials{}
	for _, item := range data.Items {
		if len(item.SerialNumbers) != 0 {
			// GORM version had an empty `if len != 0 {}` block — preserved.
			continue
		}

		rows, err := u.q.ListMaterialAmountSortedByCostM19InLocation(ctx, db.ListMaterialAmountSortedByCostM19InLocationParams{
			ProjectID:    pgInt8(data.Details.ProjectID),
			LocationType: pgText("warehouse"),
			LocationID:   pgInt8(0),
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
				InvoiceType:    "output-out-of-project",
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

// writeInvoiceMaterialsBatch is a small helper used by both the
// out-of-project Create/Update paths.
func writeInvoiceMaterialsBatch(ctx context.Context, qtx *db.Queries, invoiceID uint, materials []model.InvoiceMaterials) error {
	if len(materials) == 0 {
		return nil
	}
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
	return nil
}

func (u *invoiceOutputOutOfProjectUsecase) Delete(id uint) error {
	ctx := context.Background()
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	qtx := u.q.WithTx(tx)

	if err := qtx.DeleteInvoiceMaterialsByInvoice(ctx, db.DeleteInvoiceMaterialsByInvoiceParams{
		InvoiceType: pgText("output-out-of-project"),
		InvoiceID:   pgInt8(id),
	}); err != nil {
		return err
	}
	if err := qtx.DeleteInvoiceOutputOutOfProject(ctx, int64(id)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (u *invoiceOutputOutOfProjectUsecase) GetInvoiceMaterialsWithoutSerialNumbers(id uint) ([]dto.InvoiceMaterialsWithoutSerialNumberView, error) {
	rows, err := u.q.ListInvoiceMaterialsWithoutSerialNumbers(context.Background(), db.ListInvoiceMaterialsWithoutSerialNumbersParams{
		InvoiceType: pgText("output-out-of-project"),
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

func (u *invoiceOutputOutOfProjectUsecase) GetInvoiceMaterialsWithSerialNumbers(id uint) ([]dto.InvoiceMaterialsWithSerialNumberView, error) {
	rows, err := u.q.ListInvoiceMaterialsWithSerialNumbers(context.Background(), db.ListInvoiceMaterialsWithSerialNumbersParams{
		InvoiceType: pgText("output-out-of-project"),
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

func (u *invoiceOutputOutOfProjectUsecase) Update(data dto.InvoiceOutputOutOfProject) (model.InvoiceOutputOutOfProject, error) {
	ctx := context.Background()
	invoiceMaterialForCreate, err := u.buildOutOfProjectItems(ctx, data)
	if err != nil {
		return model.InvoiceOutputOutOfProject{}, err
	}

	excelFilePath := filepath.Join("./storage/import_excel/output/", data.Details.DeliveryCode+".xlsx")
	if err := os.Remove(excelFilePath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return model.InvoiceOutputOutOfProject{}, err
	}

	if err := u.GenerateExcelFile(data.Details, invoiceMaterialForCreate); err != nil {
		return model.InvoiceOutputOutOfProject{}, err
	}

	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return model.InvoiceOutputOutOfProject{}, err
	}
	defer tx.Rollback(ctx)
	qtx := u.q.WithTx(tx)

	invoiceRow, err := qtx.UpdateInvoiceOutputOutOfProject(ctx, db.UpdateInvoiceOutputOutOfProjectParams{
		ID:               int64(data.Details.ID),
		ProjectID:        pgInt8(data.Details.ProjectID),
		DeliveryCode:     pgText(data.Details.DeliveryCode),
		ReleasedWorkerID: pgInt8(data.Details.ReleasedWorkerID),
		NameOfProject:    pgText(data.Details.NameOfProject),
		DateOfInvoice:    pgTimestamptz(data.Details.DateOfInvoice),
		Notes:            pgText(data.Details.Notes),
		Confirmation:     pgBool(data.Details.Confirmation),
	})
	if err != nil {
		return model.InvoiceOutputOutOfProject{}, err
	}

	if err := qtx.DeleteInvoiceMaterialsByInvoice(ctx, db.DeleteInvoiceMaterialsByInvoiceParams{
		InvoiceType: pgText("output-out-of-project"),
		InvoiceID:   pgInt8(uint(invoiceRow.ID)),
	}); err != nil {
		return model.InvoiceOutputOutOfProject{}, err
	}

	if err := writeInvoiceMaterialsBatch(ctx, qtx, uint(invoiceRow.ID), invoiceMaterialForCreate); err != nil {
		return model.InvoiceOutputOutOfProject{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return model.InvoiceOutputOutOfProject{}, err
	}
	return toModelInvoiceOutputOutOfProject(invoiceRow), nil
}

func (u *invoiceOutputOutOfProjectUsecase) GetByID(id uint) (model.InvoiceOutputOutOfProject, error) {
	row, err := u.q.GetInvoiceOutputOutOfProject(context.Background(), int64(id))
	if err != nil {
		return model.InvoiceOutputOutOfProject{}, err
	}
	return toModelInvoiceOutputOutOfProject(row), nil
}

func (u *invoiceOutputOutOfProjectUsecase) Confirmation(id uint) error {
	ctx := context.Background()
	invoice, err := u.q.GetInvoiceOutputOutOfProject(ctx, int64(id))
	if err != nil {
		return err
	}

	invoiceMaterials, err := u.q.ListInvoiceMaterialsByInvoice(ctx, db.ListInvoiceMaterialsByInvoiceParams{
		InvoiceID:   pgInt8(id),
		InvoiceType: pgText("output-out-of-project"),
		ProjectID:   invoice.ProjectID,
	})
	if err != nil {
		return err
	}

	materialsInWarehouse, err := u.q.ListMaterialLocationsForInvoiceConfirmation(ctx, db.ListMaterialLocationsForInvoiceConfirmationParams{
		LocationType: pgText("warehouse"),
		LocationID:   pgInt8(0),
		InvoiceType:  pgText("output-out-of-project"),
		InvoiceID:    pgInt8(id),
	})
	if err != nil {
		return err
	}

	materialsOutOfProject, err := u.q.ListMaterialLocationsForInvoiceConfirmation(ctx, db.ListMaterialLocationsForInvoiceConfirmationParams{
		LocationType: pgText("out-of-project"),
		LocationID:   pgInt8(0),
		InvoiceType:  pgText("output-out-of-project"),
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

	type outProjLoc struct {
		row    *db.MaterialLocation
		params *db.CreateMaterialLocationParams
		amount float64
	}
	outSlice := make([]*outProjLoc, 0, len(materialsOutOfProject))
	for i := range materialsOutOfProject {
		ml := materialsOutOfProject[i]
		outSlice = append(outSlice, &outProjLoc{row: &ml, amount: float64FromPgNumeric(ml.Amount)})
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
			material, err := u.q.GetMaterialByMaterialCostID(ctx, int64(uintFromPgInt8(im.MaterialCostID)))
			if err != nil {
				return fmt.Errorf("Недостаточно материалов в складе и данные про материал не были получены: %v", err)
			}
			return fmt.Errorf("Недостаточно материала %v на складе: в накладной указано - %v а на складе - 0",
				material.Name.String, float64FromPgNumeric(im.Amount))
		}
		imAmount := float64FromPgNumeric(im.Amount)
		if warehouseSlice[whIndex].amount < imAmount {
			material, err := u.q.GetMaterialByMaterialCostID(ctx, int64(uintFromPgInt8(im.MaterialCostID)))
			if err != nil {
				return fmt.Errorf("Недостаточно материалов в складе и данные про материал не были получены: %v", err)
			}
			return fmt.Errorf("Недостаточно материала %v на складе: в накладной указано - %v а на складе - %v",
				material.Name.String, imAmount, warehouseSlice[whIndex].amount)
		}
		warehouseSlice[whIndex].amount -= imAmount

		outIndex := -1
		for index, o := range outSlice {
			if o.row != nil && uintFromPgInt8(o.row.MaterialCostID) == uintFromPgInt8(im.MaterialCostID) {
				outIndex = index
				break
			}
			if o.params != nil && uintFromPgInt8(o.params.MaterialCostID) == uintFromPgInt8(im.MaterialCostID) {
				outIndex = index
				break
			}
		}
		if outIndex != -1 {
			outSlice[outIndex].amount += imAmount
		} else {
			outSlice = append(outSlice, &outProjLoc{
				params: &db.CreateMaterialLocationParams{
					ProjectID:      invoice.ProjectID,
					MaterialCostID: im.MaterialCostID,
					LocationID:     pgInt8(0),
					LocationType:   pgText("out-of-project"),
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

	if err := qtx.ConfirmInvoiceOutputOutOfProject(ctx, int64(id)); err != nil {
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

	for _, o := range outSlice {
		if o.row != nil {
			if err := qtx.UpsertMaterialLocationByID(ctx, db.UpsertMaterialLocationByIDParams{
				Amount: pgNumericFromFloat64(o.amount),
				ID:     o.row.ID,
			}); err != nil {
				return err
			}
		} else {
			params := *o.params
			params.Amount = pgNumericFromFloat64(o.amount)
			if _, err := qtx.CreateMaterialLocation(ctx, params); err != nil {
				return err
			}
		}
	}

	return tx.Commit(ctx)
}

func (u *invoiceOutputOutOfProjectUsecase) GetMaterialsForEdit(id uint) ([]dto.InvoiceOutputMaterialsForEdit, error) {
	rows, err := u.q.ListInvoiceOutputOutOfProjectMaterialsForEdit(context.Background(), pgInt8(id))
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

func (u *invoiceOutputOutOfProjectUsecase) GetUniqueNameOfProjects(projectID uint) ([]string, error) {
	return u.q.ListInvoiceOutputOutOfProjectUniqueNameOfProjects(context.Background(), pgInt8(projectID))
}

func (u *invoiceOutputOutOfProjectUsecase) Report(filter dto.InvoiceOutputOutOfProjectReportFilter) (string, error) {
	ctx := context.Background()
	invoices, err := u.q.ListInvoiceOutputOutOfProjectReportData(ctx, db.ListInvoiceOutputOutOfProjectReportDataParams{
		ProjectID: pgInt8(filter.ProjectID),
		Column2:   pgTimestamptz(filter.DateFrom),
		Column3:   pgTimestamptz(filter.DateTo),
	})
	if err != nil {
		return "", err
	}

	templateFilePath := filepath.Join("./internal/templates/", "Invoice Output Out Of Project.xlsx")
	f, err := excelize.OpenFile(templateFilePath)
	if err != nil {
		return "", err
	}
	sheetName := "Sheet1"
	f.SetCellStr(sheetName, "J1", "ID материала")

	rowCount := 2
	for _, invoice := range invoices {
		invoiceMaterials, err := u.q.ListInvoiceMaterialsDataForReport(ctx, db.ListInvoiceMaterialsDataForReportParams{
			InvoiceType: pgText("output-out-of-project"),
			InvoiceID:   pgInt8(uint(invoice.ID)),
		})
		if err != nil {
			return "", err
		}

		for _, im := range invoiceMaterials {
			f.SetCellStr(sheetName, "A"+fmt.Sprint(rowCount), invoice.DeliveryCode)
			f.SetCellStr(sheetName, "B"+fmt.Sprint(rowCount), invoice.ReleasedWorkerName)
			f.SetCellStr(sheetName, "C"+fmt.Sprint(rowCount), invoice.NameOfProject)

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
		"Отсчет накладной отпуск вне проекта - %s.xlsx",
		currentTime.Format("02-01-2006"),
	)

	tempFilePath := filepath.Join("./storage/import_excel/temp/", fileName)
	f.SaveAs(tempFilePath)
	if err := f.Close(); err != nil {
		fmt.Println(err)
	}
	return fileName, nil
}

func (u *invoiceOutputOutOfProjectUsecase) GenerateExcelFile(details model.InvoiceOutputOutOfProject, items []model.InvoiceMaterials) error {
	ctx := context.Background()
	templateFilePath := filepath.Join("./internal/templates/output out of project.xlsx")
	f, err := excelize.OpenFile(templateFilePath)
	if err != nil {
		return err
	}

	sheetName := "Отпуск"
	startingRow := 5
	f.InsertRows(sheetName, startingRow, len(items))

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
		Font:      &excelize.Font{Size: 8, VertAlign: "center"},
		Alignment: &excelize.Alignment{Horizontal: "left", WrapText: true, Vertical: "center"},
	})

	f.SetCellValue(sheetName, "C1", fmt.Sprintf(`НАКЛАДНАЯ
№ %s
от %s года
на отпуск материала
`, details.DeliveryCode, utils.DateConverter(details.DateOfInvoice)))

	f.SetCellStyle(sheetName, "H4", "H4", defaultStyle)
	f.SetCellStr(sheetName, "H4", "ID материала")
	for index, oneEntry := range items {
		material, err := u.q.GetMaterialByMaterialCostID(ctx, int64(oneEntry.MaterialCostID))
		if err != nil {
			return err
		}

		materialCost, err := u.q.GetMaterialCost(ctx, int64(oneEntry.MaterialCostID))
		if err != nil {
			return err
		}
		f.SetCellStyle(sheetName, "A"+fmt.Sprint(startingRow+index), "H"+fmt.Sprint(startingRow+index), defaultStyle)
		f.SetCellStyle(sheetName, "B"+fmt.Sprint(startingRow+index), "B"+fmt.Sprint(startingRow+index), materialNamingStyle)

		f.SetCellInt(sheetName, "A"+fmt.Sprint(startingRow+index), index+1)
		f.SetCellStr(sheetName, "B"+fmt.Sprint(startingRow+index), material.Code.String)
		f.SetCellStr(sheetName, "C"+fmt.Sprint(startingRow+index), material.Name.String)
		f.SetCellStr(sheetName, "D"+fmt.Sprint(startingRow+index), material.Unit.String)
		f.SetCellFloat(sheetName, "E"+fmt.Sprint(startingRow+index), oneEntry.Amount, 3, 64)

		costM19, _ := decimalFromPgNumeric(materialCost.CostM19).Float64()
		f.SetCellFloat(sheetName, "F"+fmt.Sprint(startingRow+index), costM19, 3, 64)
		f.SetCellStr(sheetName, "G"+fmt.Sprint(startingRow+index), oneEntry.Notes)
		f.SetCellInt(sheetName, "H"+fmt.Sprint(startingRow+index), int(material.ID))
	}

	released, err := u.q.GetWorker(ctx, int64(details.ReleasedWorkerID))
	if err != nil {
		return err
	}
	f.SetCellStyle(sheetName, "C"+fmt.Sprint(6+len(items)), "C"+fmt.Sprint(6+len(items)), workerNamingStyle)
	f.SetCellStr(sheetName, "C"+fmt.Sprint(6+len(items)), released.Name.String)

	excelFilePath := filepath.Join("./storage/import_excel/output/", details.DeliveryCode+".xlsx")
	if err := f.SaveAs(excelFilePath); err != nil {
		return err
	}
	return nil
}

func (u *invoiceOutputOutOfProjectUsecase) GetDocument(deliveryCode string) (string, error) {
	invoice, err := u.q.GetInvoiceOutputOutOfProjectByDeliveryCode(context.Background(), pgText(deliveryCode))
	if err != nil {
		return "", err
	}
	if invoice.Confirmation.Bool {
		return ".pdf", nil
	}
	return ".xlsx", nil
}

func toModelInvoiceOutputOutOfProject(r db.InvoiceOutputOutOfProject) model.InvoiceOutputOutOfProject {
	return model.InvoiceOutputOutOfProject{
		ID:               uint(r.ID),
		ProjectID:        uintFromPgInt8(r.ProjectID),
		DeliveryCode:     r.DeliveryCode.String,
		ReleasedWorkerID: uintFromPgInt8(r.ReleasedWorkerID),
		NameOfProject:    r.NameOfProject.String,
		DateOfInvoice:    timeFromPgTimestamptz(r.DateOfInvoice),
		Notes:            r.Notes.String,
		Confirmation:     r.Confirmation.Bool,
	}
}
