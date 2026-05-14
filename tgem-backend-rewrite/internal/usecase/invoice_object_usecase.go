package usecase

import (
	"context"
	"errors"

	"backend-v2/internal/db"
	"backend-v2/internal/dto"
	"backend-v2/internal/utils"
	"backend-v2/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type invoiceObjectUsecase struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

func NewInvoiceObjectUsecase(pool *pgxpool.Pool) IInvoiceObjectUsecase {
	return &invoiceObjectUsecase{
		pool: pool,
		q:    db.New(pool),
	}
}

type IInvoiceObjectUsecase interface {
	GetPaginated(limit, page int, projectID uint) ([]dto.InvoiceObjectPaginated, error)
	Create(data dto.InvoiceObjectCreate) (model.InvoiceObject, error)
	Delete(id uint) error
	GetInvoiceObjectDescriptiveDataByID(id uint) (dto.InvoiceObjectWithMaterialsDescriptive, error)
	GetTeamsMaterials(projectID, teamID uint) ([]dto.InvoiceObjectTeamMaterials, error)
	GetSerialNumberOfMaterial(projectID, materialID uint, locationID uint) ([]string, error)
	GetAvailableMaterialAmount(projectID, materialID, teamID uint) (float64, error)
	Count(projectID uint) (int64, error)
	GetTeamsFromObjectID(objectID uint) ([]dto.TeamDataForSelect, error)
	GetOperationsBasedOnMaterialsInTeamID(projectID, teamID uint) ([]dto.InvoiceObjectOperationsBasedOnTeam, error)
}

func (u *invoiceObjectUsecase) GetInvoiceObjectDescriptiveDataByID(id uint) (dto.InvoiceObjectWithMaterialsDescriptive, error) {
	ctx := context.Background()
	row, err := u.q.GetInvoiceObjectDescriptiveDataByID(ctx, int64(id))
	if err != nil {
		return dto.InvoiceObjectWithMaterialsDescriptive{}, err
	}
	invoiceData := dto.InvoiceObjectPaginated{
		ID:                  uint(row.ID),
		DeliveryCode:        row.DeliveryCode,
		SupervisorName:      row.SupervisorName,
		DistrictID:          uintFromPgInt8(row.DistrictID),
		DistrictName:        row.DistrictName,
		ObjectName:          row.ObjectName,
		ObjectType:          row.ObjectType,
		TeamNumber:          row.TeamNumber,
		DateOfInvoice:       timeFromPgTimestamptz(row.DateOfInvoice),
		ConfirmedByOperator: row.ConfirmedByOperator,
	}

	withSNRows, err := u.q.ListInvoiceMaterialsWithSerialNumbers(ctx, db.ListInvoiceMaterialsWithSerialNumbersParams{
		InvoiceType: pgText("object"),
		InvoiceID:   pgInt8(id),
	})
	if err != nil {
		return dto.InvoiceObjectWithMaterialsDescriptive{}, err
	}

	withoutSNRows, err := u.q.ListInvoiceMaterialsWithoutSerialNumbers(ctx, db.ListInvoiceMaterialsWithoutSerialNumbersParams{
		InvoiceType: pgText("object"),
		InvoiceID:   pgInt8(id),
	})
	if err != nil {
		return dto.InvoiceObjectWithMaterialsDescriptive{}, err
	}

	withSNQuery := make([]dto.InvoiceMaterialsWithSerialNumberQuery, len(withSNRows))
	for i, r := range withSNRows {
		withSNQuery[i] = dto.InvoiceMaterialsWithSerialNumberQuery{
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

	invoiceMaterialsWithSerialNumber := []dto.InvoiceMaterialsWithSerialNumberView{}
	current := dto.InvoiceMaterialsWithSerialNumberView{}
	for index, materialInfo := range withSNQuery {
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
			invoiceMaterialsWithSerialNumber = append(invoiceMaterialsWithSerialNumber, current)
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
	if len(withSNQuery) != 0 {
		invoiceMaterialsWithSerialNumber = append(invoiceMaterialsWithSerialNumber, current)
	}

	withoutSN := make([]dto.InvoiceMaterialsWithoutSerialNumberView, len(withoutSNRows))
	for i, r := range withoutSNRows {
		withoutSN[i] = dto.InvoiceMaterialsWithoutSerialNumberView{
			ID:           uint(r.ID),
			MaterialName: r.MaterialName,
			MaterialUnit: r.MaterialUnit,
			IsDefected:   r.IsDefected,
			CostM19:      decimalFromPgNumeric(r.CostM19),
			Amount:       float64FromPgNumeric(r.Amount),
			Notes:        r.Notes,
		}
	}

	return dto.InvoiceObjectWithMaterialsDescriptive{
		InvoiceData:                  invoiceData,
		MaterialsWithSerialNumber:    invoiceMaterialsWithSerialNumber,
		MaterialsWithoutSerialNumber: withoutSN,
	}, nil
}

func (u *invoiceObjectUsecase) GetPaginated(limit, page int, projectID uint) ([]dto.InvoiceObjectPaginated, error) {
	rows, err := u.q.ListInvoiceObjectsPaginated(context.Background(), db.ListInvoiceObjectsPaginatedParams{
		ProjectID: pgInt8(projectID),
		Limit:     int32(limit),
		Offset:    int32((page - 1) * limit),
	})
	if err != nil {
		return nil, err
	}
	out := make([]dto.InvoiceObjectPaginated, len(rows))
	for i, r := range rows {
		out[i] = dto.InvoiceObjectPaginated{
			ID:                  uint(r.ID),
			DeliveryCode:        r.DeliveryCode,
			SupervisorName:      r.SupervisorName,
			DistrictID:          uintFromPgInt8(r.DistrictID),
			DistrictName:        r.DistrictName,
			ObjectName:          r.ObjectName,
			ObjectType:          r.ObjectType,
			TeamNumber:          r.TeamNumber,
			DateOfInvoice:       timeFromPgTimestamptz(r.DateOfInvoice),
			ConfirmedByOperator: r.ConfirmedByOperator,
		}
	}
	return out, nil
}

func (u *invoiceObjectUsecase) Create(data dto.InvoiceObjectCreate) (model.InvoiceObject, error) {
	ctx := context.Background()

	count, err := u.q.CountInvoiceObjectsByProject(ctx, pgInt8(data.Details.ProjectID))
	if err != nil {
		return model.InvoiceObject{}, err
	}

	code := utils.UniqueCodeGeneration("ПО", count+1, data.Details.ProjectID)
	data.Details.DeliveryCode = code

	invoiceMaterialForCreate := []model.InvoiceMaterials{}
	serialNumberMovements := []model.SerialNumberMovement{}
	for _, item := range data.Items {
		if len(item.SerialNumbers) == 0 {
			rows, err := u.q.ListMaterialAmountSortedByCostM19InLocation(ctx, db.ListMaterialAmountSortedByCostM19InLocationParams{
				ProjectID:    pgInt8(data.Details.ProjectID),
				LocationType: pgText("team"),
				LocationID:   pgInt8(data.Details.TeamID),
				ID:           int64(item.MaterialID),
			})
			if err != nil {
				return model.InvoiceObject{}, err
			}

			index := 0
			amountLeft := item.Amount
			for amountLeft > 0 {
				materialAmount := float64FromPgNumeric(rows[index].MaterialAmount)
				imc := model.InvoiceMaterials{
					ProjectID:      data.Details.ProjectID,
					MaterialCostID: uint(rows[index].MaterialCostID),
					InvoiceType:    "object",
					Notes:          item.Notes,
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
			continue
		}

		mcSnRows, err := u.q.ListMaterialCostIDAndSerialNumberIDByCodes(ctx, db.ListMaterialCostIDAndSerialNumberIDByCodesParams{
			ID:           int64(item.MaterialID),
			LocationType: pgText("team"),
			LocationID:   pgInt8(data.Details.TeamID),
			Codes:        item.SerialNumbers,
		})
		if err != nil {
			return model.InvoiceObject{}, err
		}

		var imc model.InvoiceMaterials
		for index, oneEntry := range mcSnRows {
			serialNumberMovements = append(serialNumberMovements, model.SerialNumberMovement{
				SerialNumberID: uint(oneEntry.SerialNumberID),
				ProjectID:      data.Details.ProjectID,
				InvoiceType:    "object",
			})

			if index == 0 {
				imc = model.InvoiceMaterials{
					ProjectID:      data.Details.ProjectID,
					MaterialCostID: uint(oneEntry.MaterialCostID),
					InvoiceType:    "object",
					Notes:          item.Notes,
				}
			}

			if uint(oneEntry.MaterialCostID) == imc.MaterialCostID {
				imc.Amount++
			} else {
				invoiceMaterialForCreate = append(invoiceMaterialForCreate, imc)
				imc = model.InvoiceMaterials{
					ProjectID:      data.Details.ProjectID,
					MaterialCostID: uint(oneEntry.MaterialCostID),
					InvoiceType:    "object",
					Notes:          item.Notes,
				}
			}
		}
		invoiceMaterialForCreate = append(invoiceMaterialForCreate, imc)
	}

	invoiceOperationsForCreate := []model.InvoiceOperations{}
	for _, op := range data.Operations {
		invoiceOperationsForCreate = append(invoiceOperationsForCreate, model.InvoiceOperations{
			ProjectID:   data.Details.ProjectID,
			OperationID: op.OperationID,
			InvoiceType: "object",
			Amount:      op.Amount,
			Notes:       op.Notes,
		})
	}

	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return model.InvoiceObject{}, err
	}
	defer tx.Rollback(ctx)
	qtx := u.q.WithTx(tx)

	invoiceRow, err := qtx.CreateInvoiceObject(ctx, db.CreateInvoiceObjectParams{
		DistrictID:          pgInt8(data.Details.DistrictID),
		DeliveryCode:        pgText(data.Details.DeliveryCode),
		ProjectID:           pgInt8(data.Details.ProjectID),
		SupervisorWorkerID:  pgInt8(data.Details.SupervisorWorkerID),
		ObjectID:            pgInt8(data.Details.ObjectID),
		TeamID:              pgInt8(data.Details.TeamID),
		DateOfInvoice:       pgTimestamptz(data.Details.DateOfInvoice),
		ConfirmedByOperator: pgBool(data.Details.ConfirmedByOperator),
		DateOfCorrection:    pgTimestamptz(data.Details.DateOfCorrection),
	})
	if err != nil {
		return model.InvoiceObject{}, err
	}

	if err := writeInvoiceMaterialsBatch(ctx, qtx, uint(invoiceRow.ID), invoiceMaterialForCreate); err != nil {
		return model.InvoiceObject{}, err
	}

	if len(invoiceOperationsForCreate) > 0 {
		opBatch := make([]db.CreateInvoiceOperationsBatchParams, len(invoiceOperationsForCreate))
		for i, op := range invoiceOperationsForCreate {
			opBatch[i] = db.CreateInvoiceOperationsBatchParams{
				ProjectID:   pgInt8(op.ProjectID),
				OperationID: pgInt8(op.OperationID),
				InvoiceID:   pgInt8(uint(invoiceRow.ID)),
				InvoiceType: pgText(op.InvoiceType),
				Amount:      pgNumericFromFloat64(op.Amount),
				Notes:       pgText(op.Notes),
			}
		}
		if _, err := qtx.CreateInvoiceOperationsBatch(ctx, opBatch); err != nil {
			return model.InvoiceObject{}, err
		}
	}

	if len(serialNumberMovements) > 0 {
		snBatch := make([]db.CreateSerialNumberMovementsBatchParams, len(serialNumberMovements))
		for i, m := range serialNumberMovements {
			snBatch[i] = db.CreateSerialNumberMovementsBatchParams{
				SerialNumberID: pgInt8(m.SerialNumberID),
				ProjectID:      pgInt8(m.ProjectID),
				InvoiceID:      pgInt8(uint(invoiceRow.ID)),
				InvoiceType:    pgText(m.InvoiceType),
				IsDefected:     pgBool(m.IsDefected),
				Confirmation:   pgBool(m.Confirmation),
			}
		}
		if _, err := qtx.CreateSerialNumberMovementsBatch(ctx, snBatch); err != nil {
			return model.InvoiceObject{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return model.InvoiceObject{}, err
	}

	return toModelInvoiceObject(invoiceRow), nil
}

func (u *invoiceObjectUsecase) Delete(id uint) error {
	return u.q.DeleteInvoiceObject(context.Background(), int64(id))
}

func (u *invoiceObjectUsecase) GetTeamsMaterials(projectID, teamID uint) ([]dto.InvoiceObjectTeamMaterials, error) {
	ctx := context.Background()
	materials, err := u.q.ListUniqueMaterialsFromLocation(ctx, db.ListUniqueMaterialsFromLocationParams{
		ProjectID:    pgInt8(projectID),
		LocationID:   pgInt8(teamID),
		LocationType: pgText("team"),
	})
	if err != nil {
		return nil, err
	}
	out := []dto.InvoiceObjectTeamMaterials{}
	for _, m := range materials {
		amount, err := u.q.GetTotalAmountInLocation(ctx, db.GetTotalAmountInLocationParams{
			ProjectID:    pgInt8(projectID),
			ID:           m.ID,
			LocationType: pgText("team"),
			LocationID:   pgInt8(teamID),
		})
		if errors.Is(err, pgx.ErrNoRows) {
			amount = pgNumericFromFloat64(0)
		} else if err != nil {
			return nil, err
		}
		out = append(out, dto.InvoiceObjectTeamMaterials{
			MaterialID:      uint(m.ID),
			MaterialName:    m.Name.String,
			MaterialUnit:    m.Unit.String,
			HasSerialNumber: m.HasSerialNumber.Bool,
			Amount:          float64FromPgNumeric(amount),
		})
	}
	return out, nil
}

func (u *invoiceObjectUsecase) GetSerialNumberOfMaterial(projectID, materialID uint, locationID uint) ([]string, error) {
	return u.q.GetSerialNumberCodesByMaterialIDAndLocation(context.Background(), db.GetSerialNumberCodesByMaterialIDAndLocationParams{
		ProjectID:    pgInt8(projectID),
		ID:           int64(materialID),
		LocationType: pgText("team"),
		LocationID:   pgInt8(locationID),
	})
}

func (u *invoiceObjectUsecase) GetAvailableMaterialAmount(projectID, materialID, teamID uint) (float64, error) {
	amount, err := u.q.GetTotalAmountInLocation(context.Background(), db.GetTotalAmountInLocationParams{
		ProjectID:    pgInt8(projectID),
		ID:           int64(materialID),
		LocationType: pgText("team"),
		LocationID:   pgInt8(teamID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return float64FromPgNumeric(amount), nil
}

func (u *invoiceObjectUsecase) Count(projectID uint) (int64, error) {
	return u.q.CountInvoiceObjectsByProject(context.Background(), pgInt8(projectID))
}

func (u *invoiceObjectUsecase) GetTeamsFromObjectID(objectID uint) ([]dto.TeamDataForSelect, error) {
	rows, err := u.q.ListTeamsForSelectByObjectID(context.Background(), pgInt8(objectID))
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

func (u *invoiceObjectUsecase) GetOperationsBasedOnMaterialsInTeamID(projectID, teamID uint) ([]dto.InvoiceObjectOperationsBasedOnTeam, error) {
	ctx := context.Background()
	result := []dto.InvoiceObjectOperationsBasedOnTeam{}

	operationsAvailable, err := u.q.ListInvoiceObjectOperationsBasedOnMaterialsInTeam(ctx, pgInt8(teamID))
	if err != nil {
		return result, err
	}

	operationsWithoutMaterials, err := u.q.ListOperationsWithoutMaterials(ctx, pgInt8(projectID))
	if err != nil {
		return result, err
	}

	for _, operation := range operationsAvailable {
		operationMaterial, err := u.q.GetOperationMaterialByOperationID(ctx, pgInt8(uint(operation.ID)))
		if err != nil {
			return []dto.InvoiceObjectOperationsBasedOnTeam{}, err
		}

		material, err := u.q.GetMaterial(ctx, int64(uintFromPgInt8(operationMaterial.MaterialID)))
		if err != nil {
			return []dto.InvoiceObjectOperationsBasedOnTeam{}, err
		}

		result = append(result, dto.InvoiceObjectOperationsBasedOnTeam{
			OperationID:   uint(operation.ID),
			OperationName: operation.Name,
			MaterialID:    uintFromPgInt8(operationMaterial.MaterialID),
			MaterialName:  material.Name.String,
		})
	}

	for _, operation := range operationsWithoutMaterials {
		result = append(result, dto.InvoiceObjectOperationsBasedOnTeam{
			OperationID:   uint(operation.ID),
			OperationName: operation.Name.String,
			MaterialID:    0,
			MaterialName:  "",
		})
	}

	return result, nil
}

func toModelInvoiceObject(r db.InvoiceObject) model.InvoiceObject {
	return model.InvoiceObject{
		ID:                  uint(r.ID),
		DistrictID:          uintFromPgInt8(r.DistrictID),
		DeliveryCode:        r.DeliveryCode.String,
		ProjectID:           uintFromPgInt8(r.ProjectID),
		SupervisorWorkerID:  uintFromPgInt8(r.SupervisorWorkerID),
		ObjectID:            uintFromPgInt8(r.ObjectID),
		TeamID:              uintFromPgInt8(r.TeamID),
		DateOfInvoice:       timeFromPgTimestamptz(r.DateOfInvoice),
		ConfirmedByOperator: r.ConfirmedByOperator.Bool,
		DateOfCorrection:    timeFromPgTimestamptz(r.DateOfCorrection),
	}
}
