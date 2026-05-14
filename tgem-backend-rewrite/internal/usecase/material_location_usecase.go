package usecase

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"backend-v2/internal/db"
	"backend-v2/internal/dto"
	"backend-v2/internal/utils"
	"backend-v2/model"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xuri/excelize/v2"
)

type materialLocationUsecase struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

func NewMaterialLocationUsecase(pool *pgxpool.Pool) IMaterialLocationUsecase {
	return &materialLocationUsecase{
		pool: pool,
		q:    db.New(pool),
	}
}

type IMaterialLocationUsecase interface {
	GetAll() ([]model.MaterialLocation, error)
	GetPaginated(page, limit int, data model.MaterialLocation) ([]model.MaterialLocation, error)
	GetByID(id uint) (model.MaterialLocation, error)
	Create(data model.MaterialLocation) (model.MaterialLocation, error)
	Update(data model.MaterialLocation) (model.MaterialLocation, error)
	Delete(id uint) error
	Count() (int64, error)
	GetMaterialsInLocation(locationType string, locationID uint, projectID uint) ([]model.Material, error)
	UniqueObjects(projectID uint) ([]dto.ObjectDataForSelect, error)
	UniqueTeams(projectID uint) ([]dto.TeamDataForSelect, error)
	BalanceReport(projectID uint, data dto.ReportBalanceFilterRequest) (string, error)
	BalanceReportWriteOff(projectID uint, data dto.ReportWriteOffBalanceFilter) (string, error)
	BalanceReportOutOfProject(projectID uint) (string, error)
	Live(searchParameters dto.MaterialLocationLiveSearchParameters) ([]dto.MaterialLocationLiveView, error)
	GetMaterialCostsInLocation(projectID, materialID, locationID uint, locationType string) ([]model.MaterialCost, error)
	GetMaterialAmountBasedOnCost(projectID, materialCost, locationID uint, locationType string) (float64, error)
}

func (u *materialLocationUsecase) GetAll() ([]model.MaterialLocation, error) {
	rows, err := u.q.ListMaterialLocations(context.Background())
	if err != nil {
		return nil, err
	}
	out := make([]model.MaterialLocation, len(rows))
	for i, r := range rows {
		out[i] = toModelMaterialLocation(r)
	}
	return out, nil
}

func (u *materialLocationUsecase) GetPaginated(page, limit int, data model.MaterialLocation) ([]model.MaterialLocation, error) {
	ctx := context.Background()
	if !utils.IsEmptyFields(data) {
		rows, err := u.q.ListMaterialLocationsPaginatedFiltered(ctx, db.ListMaterialLocationsPaginatedFilteredParams{
			Column1: int64(data.MaterialCostID),
			Column2: int64(data.LocationID),
			Column3: data.LocationType,
			Column4: pgNumericFromFloat64(data.Amount),
			Limit:   int32(limit),
			Offset:  int32((page - 1) * limit),
		})
		if err != nil {
			return nil, err
		}
		out := make([]model.MaterialLocation, len(rows))
		for i, r := range rows {
			out[i] = toModelMaterialLocation(r)
		}
		return out, nil
	}

	rows, err := u.q.ListMaterialLocationsPaginated(ctx, db.ListMaterialLocationsPaginatedParams{
		Limit:  int32(limit),
		Offset: int32((page - 1) * limit),
	})
	if err != nil {
		return nil, err
	}
	out := make([]model.MaterialLocation, len(rows))
	for i, r := range rows {
		out[i] = toModelMaterialLocation(r)
	}
	return out, nil
}

func (u *materialLocationUsecase) GetByID(id uint) (model.MaterialLocation, error) {
	row, err := u.q.GetMaterialLocation(context.Background(), int64(id))
	if err != nil {
		return model.MaterialLocation{}, err
	}
	return toModelMaterialLocation(row), nil
}

func (u *materialLocationUsecase) Create(data model.MaterialLocation) (model.MaterialLocation, error) {
	row, err := u.q.CreateMaterialLocation(context.Background(), db.CreateMaterialLocationParams{
		ProjectID:      pgInt8(data.ProjectID),
		MaterialCostID: pgInt8(data.MaterialCostID),
		LocationID:     pgInt8(data.LocationID),
		LocationType:   pgText(data.LocationType),
		Amount:         pgNumericFromFloat64(data.Amount),
	})
	if err != nil {
		return model.MaterialLocation{}, err
	}
	return toModelMaterialLocation(row), nil
}

func (u *materialLocationUsecase) Update(data model.MaterialLocation) (model.MaterialLocation, error) {
	if err := u.q.UpdateMaterialLocation(context.Background(), db.UpdateMaterialLocationParams{
		ID:             int64(data.ID),
		ProjectID:      pgInt8(data.ProjectID),
		MaterialCostID: pgInt8(data.MaterialCostID),
		LocationID:     pgInt8(data.LocationID),
		LocationType:   pgText(data.LocationType),
		Amount:         pgNumericFromFloat64(data.Amount),
	}); err != nil {
		return model.MaterialLocation{}, err
	}
	return data, nil
}

func (u *materialLocationUsecase) Delete(id uint) error {
	return u.q.DeleteMaterialLocation(context.Background(), int64(id))
}

func (u *materialLocationUsecase) Count() (int64, error) {
	return u.q.CountMaterialLocations(context.Background())
}

func (u *materialLocationUsecase) GetMaterialsInLocation(locationType string, locationID uint, projectID uint) ([]model.Material, error) {
	rows, err := u.q.ListUniqueMaterialsFromLocation(context.Background(), db.ListUniqueMaterialsFromLocationParams{
		ProjectID:    pgInt8(projectID),
		LocationID:   pgInt8(locationID),
		LocationType: pgText(locationType),
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

func (u *materialLocationUsecase) UniqueTeams(projectID uint) ([]dto.TeamDataForSelect, error) {
	rows, err := u.q.ListUniqueTeamsForSelectFromMaterialLocations(context.Background(), pgInt8(projectID))
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

func (u *materialLocationUsecase) UniqueObjects(projectID uint) ([]dto.ObjectDataForSelect, error) {
	rows, err := u.q.ListUniqueObjectsForSelectFromMaterialLocations(context.Background(), pgInt8(projectID))
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

func (u *materialLocationUsecase) BalanceReport(projectID uint, data dto.ReportBalanceFilterRequest) (string, error) {
	ctx := context.Background()

	filter := dto.ReportBalanceFilter{
		LocationType: data.Type,
	}

	templateFilePath := filepath.Join("./internal/templates/", "Отчет Остатка.xlsx")
	f, err := excelize.OpenFile(templateFilePath)
	if err != nil {
		return "", err
	}

	sheetName := "Отчет"
	f.SetCellValue(sheetName, "L1", "ID материала")
	rowCount := 2

	switch data.Type {
	case "team":
		f.SetCellValue(sheetName, "J1", "№ Бригады")
		f.SetCellValue(sheetName, "I1", "Бригадир")
		if data.TeamID != 0 {
			filter.LocationID = data.TeamID
			break
		}

		filter.LocationID = 0

	case "object":
		f.SetCellValue(sheetName, "I1", "Супервайзер")
		f.SetCellValue(sheetName, "J1", "Объект")
		f.SetCellValue(sheetName, "K1", "Тип Объекта")

		if data.ObjectID != 0 {
			filter.LocationID = data.ObjectID
			break
		}

		filter.LocationID = 0

	case "warehouse":
		filter.LocationID = 0

	default:
		return "", fmt.Errorf("incorrect type")
	}

	materialsData, err := u.q.ListBalanceReportData(ctx, db.ListBalanceReportDataParams{
		ProjectID:    pgInt8(projectID),
		LocationType: pgText(filter.LocationType),
		Column3:      int64(filter.LocationID),
	})
	if err != nil {
		return "", err
	}

	locationInformation := struct {
		LocationID        uint
		LocationName      string
		LocationOwnerName string
		LocationType      string
	}{}

	for _, entry := range materialsData {
		entryLocationID := uint(entry.LocationID)

		if entryLocationID != locationInformation.LocationID {

			locationInformation.LocationID = entryLocationID
			locationInformation.LocationOwnerName = ""

			if filter.LocationType == "team" {
				teamData, err := u.q.ListTeamNumberAndLeadersByID(ctx, db.ListTeamNumberAndLeadersByIDParams{
					ProjectID: pgInt8(projectID),
					ID:        int64(entryLocationID),
				})
				if err != nil {
					return "", fmt.Errorf("Ошибка базы: %v", err)
				}
				if len(teamData) > 0 {
					locationInformation.LocationName = teamData[0].TeamNumber
				}

				for index, td := range teamData {
					if index == len(teamData)-1 {
						locationInformation.LocationOwnerName += td.TeamLeaderName
						break
					}
					locationInformation.LocationOwnerName += td.TeamLeaderName + ", "
				}
			}

			if filter.LocationType == "object" {
				objectData, err := u.q.ListSupervisorAndObjectNamesByObjectID(ctx, db.ListSupervisorAndObjectNamesByObjectIDParams{
					ProjectID: pgInt8(projectID),
					ID:        int64(entryLocationID),
				})
				if err != nil {
					return "", fmt.Errorf("Ошибка базы: %v", err)
				}

				if len(objectData) > 0 {
					locationInformation.LocationName = objectData[0].ObjectName
					locationInformation.LocationType = utils.ObjectTypeConverter(objectData[0].ObjectType)
				}

				for index, od := range objectData {
					if index == len(objectData)-1 {
						locationInformation.LocationOwnerName += od.SupervisorName
						break
					}
					locationInformation.LocationOwnerName += od.SupervisorName + ", "
				}
			}
		}

		if data.Type == "object" {
			operationSheetName := "Услуги"
			_, err := f.NewSheet(operationSheetName)
			if err != nil {
				return "", fmt.Errorf("Ошибка создание нового листа: %v", err)
			}

			f.SetCellStr(operationSheetName, "A1", "Имя")
			f.SetCellStr(operationSheetName, "B1", "Код")
		}

		totalAmount := float64FromPgNumeric(entry.TotalAmount)
		defectAmount := float64FromPgNumeric(entry.DefectAmount)
		costM19, _ := decimalFromPgNumeric(entry.MaterialCostM19).Float64()
		totalCost, _ := decimalFromPgNumeric(entry.TotalCost).Float64()
		totalDefectCost, _ := decimalFromPgNumeric(entry.TotalDefectCost).Float64()

		f.SetCellStr(sheetName, "A"+fmt.Sprint(rowCount), entry.MaterialCode)
		f.SetCellStr(sheetName, "B"+fmt.Sprint(rowCount), entry.MaterialName)
		f.SetCellStr(sheetName, "C"+fmt.Sprint(rowCount), entry.MaterialUnit)
		f.SetCellFloat(sheetName, "D"+fmt.Sprint(rowCount), totalAmount, 2, 64)
		f.SetCellFloat(sheetName, "E"+fmt.Sprint(rowCount), defectAmount, 2, 64)
		f.SetCellFloat(sheetName, "F"+fmt.Sprint(rowCount), costM19, 2, 64)
		f.SetCellFloat(sheetName, "G"+fmt.Sprint(rowCount), totalCost, 2, 64)
		f.SetCellFloat(sheetName, "H"+fmt.Sprint(rowCount), totalDefectCost, 2, 64)

		f.SetCellStr(sheetName, "I"+fmt.Sprint(rowCount), locationInformation.LocationOwnerName)
		f.SetCellStr(sheetName, "J"+fmt.Sprint(rowCount), locationInformation.LocationName)
		f.SetCellStr(sheetName, "K"+fmt.Sprint(rowCount), locationInformation.LocationType)
		f.SetCellInt(sheetName, "L"+fmt.Sprint(rowCount), int(entry.MaterialID))

		rowCount++
	}

	currentTime := time.Now()
	var fileName string
	if filter.LocationID == 0 {
		fileName = fmt.Sprintf(
			"Report Balance %s %s.xlsx",
			strings.ToUpper(filter.LocationType),
			currentTime.Format("02-01-2006"),
		)
	} else {
		fileName = fmt.Sprintf(
			"Report Balance %s-%d %s.xlsx",
			strings.ToUpper(filter.LocationType),
			filter.LocationID,
			currentTime.Format("02-01-2006"),
		)
	}

	tempFilePath := filepath.Join("./storage/import_excel/temp/", fileName)
	f.SaveAs(tempFilePath)
	if err := f.Close(); err != nil {
		fmt.Println(err)
	}

	return fileName, nil
}

func (u *materialLocationUsecase) Live(data dto.MaterialLocationLiveSearchParameters) ([]dto.MaterialLocationLiveView, error) {
	ctx := context.Background()
	rows, err := u.q.ListMaterialLocationsLive(ctx, db.ListMaterialLocationsLiveParams{
		LocationType: pgText(data.LocationType),
		ProjectID:    pgInt8(data.ProjectID),
		Column3:      int64(data.LocationID),
		Column4:      int64(data.MaterialID),
	})
	if err != nil {
		return nil, err
	}

	result := make([]dto.MaterialLocationLiveView, len(rows))
	for i, r := range rows {
		result[i] = dto.MaterialLocationLiveView{
			MaterialID:      uint(r.MaterialID),
			MaterialName:    r.MaterialName,
			MaterialUnit:    r.MaterialUnit,
			MaterialCostID:  uint(r.MaterialCostID),
			MaterialCostM19: decimalFromPgNumeric(r.MaterialCostM19).String(),
			LocationType:    r.LocationType,
			LocationID:      uint(r.LocationID),
			Amount:          float64FromPgNumeric(r.Amount),
		}
	}

	if data.LocationType == "team" {
		for index, materialLocation := range result {
			team, err := u.q.GetTeam(ctx, int64(materialLocation.LocationID))
			if err != nil {
				return nil, err
			}
			result[index].LocationName = team.Number.String
		}
	}

	if data.LocationType == "object" {
		for index, materialLocation := range result {
			object, err := u.q.GetObject(ctx, int64(materialLocation.LocationID))
			if err != nil {
				return nil, err
			}
			result[index].LocationName = object.Name.String
		}
	}

	return result, nil
}

func (u *materialLocationUsecase) BalanceReportWriteOff(projectID uint, data dto.ReportWriteOffBalanceFilter) (string, error) {
	ctx := context.Background()
	templateFilePath := filepath.Join("./internal/templates/", "Отчет Остатка.xlsx")
	f, err := excelize.OpenFile(templateFilePath)
	if err != nil {
		return "", err
	}

	sheetName := "Отчет"
	f.SetCellValue(sheetName, "L1", "ID материала")
	rowCount := 2

	switch data.WriteOffType {
	case "loss-team":
		f.SetCellValue(sheetName, "I1", "№ Бригады")
		f.SetCellValue(sheetName, "J1", "Бригадир")
	case "loss-object":
		f.SetCellValue(sheetName, "I1", "Супервайзер")
		f.SetCellValue(sheetName, "J1", "Объект")
		f.SetCellValue(sheetName, "K1", "Тип Объекта")
	case "writoff-object":
		f.SetCellValue(sheetName, "I1", "Супервайзер")
		f.SetCellValue(sheetName, "J1", "Объект")
		f.SetCellValue(sheetName, "K1", "Тип Объекта")
	case "writeoff-warehouse":
	case "loss-warehouse":
	default:
		return "", fmt.Errorf("incorrect type")
	}

	materialsData, err := u.q.ListBalanceReportData(ctx, db.ListBalanceReportDataParams{
		ProjectID:    pgInt8(projectID),
		LocationType: pgText(data.WriteOffType),
		Column3:      int64(data.LocationID),
	})
	if err != nil {
		return "", err
	}

	locationInformation := struct {
		LocationID        uint
		LocationName      string
		LocationOwnerName string
		LocationType      string
	}{}

	for _, entry := range materialsData {
		entryLocationID := uint(entry.LocationID)

		if entryLocationID != locationInformation.LocationID {
			locationInformation.LocationID = entryLocationID
			locationInformation.LocationOwnerName = ""

			if data.WriteOffType == "loss-team" {
				teamData, err := u.q.ListTeamNumberAndLeadersByID(ctx, db.ListTeamNumberAndLeadersByIDParams{
					ProjectID: pgInt8(projectID),
					ID:        int64(entryLocationID),
				})
				if err != nil {
					return "", fmt.Errorf("Ошибка базы: %v", err)
				}
				if len(teamData) > 0 {
					locationInformation.LocationName = teamData[0].TeamNumber
				}

				for index, td := range teamData {
					if index == len(teamData)-1 {
						locationInformation.LocationOwnerName += td.TeamLeaderName
						break
					}
					locationInformation.LocationOwnerName += td.TeamLeaderName + ", "
				}
			}

			if data.WriteOffType == "loss-object" || data.WriteOffType == "writeoff-object" {
				objectData, err := u.q.ListSupervisorAndObjectNamesByObjectID(ctx, db.ListSupervisorAndObjectNamesByObjectIDParams{
					ProjectID: pgInt8(projectID),
					ID:        int64(entryLocationID),
				})
				if err != nil {
					return "", fmt.Errorf("Ошибка базы: %v", err)
				}
				if len(objectData) > 0 {
					locationInformation.LocationName = objectData[0].ObjectName
					locationInformation.LocationType = utils.ObjectTypeConverter(objectData[0].ObjectType)
				}

				for index, od := range objectData {
					if index == len(objectData)-1 {
						locationInformation.LocationOwnerName += od.SupervisorName
						break
					}
					locationInformation.LocationOwnerName += od.SupervisorName + ", "
				}
			}
		}

		totalAmount := float64FromPgNumeric(entry.TotalAmount)
		defectAmount := float64FromPgNumeric(entry.DefectAmount)
		costM19, _ := decimalFromPgNumeric(entry.MaterialCostM19).Float64()
		totalCost, _ := decimalFromPgNumeric(entry.TotalCost).Float64()
		totalDefectCost, _ := decimalFromPgNumeric(entry.TotalDefectCost).Float64()

		f.SetCellStr(sheetName, "A"+fmt.Sprint(rowCount), entry.MaterialCode)
		f.SetCellStr(sheetName, "B"+fmt.Sprint(rowCount), entry.MaterialName)
		f.SetCellStr(sheetName, "C"+fmt.Sprint(rowCount), entry.MaterialUnit)
		f.SetCellFloat(sheetName, "D"+fmt.Sprint(rowCount), totalAmount, 2, 64)
		f.SetCellFloat(sheetName, "E"+fmt.Sprint(rowCount), defectAmount, 2, 64)
		f.SetCellFloat(sheetName, "F"+fmt.Sprint(rowCount), costM19, 2, 64)
		f.SetCellFloat(sheetName, "G"+fmt.Sprint(rowCount), totalCost, 2, 64)
		f.SetCellFloat(sheetName, "H"+fmt.Sprint(rowCount), totalDefectCost, 2, 64)

		f.SetCellStr(sheetName, "I"+fmt.Sprint(rowCount), locationInformation.LocationOwnerName)
		f.SetCellStr(sheetName, "J"+fmt.Sprint(rowCount), locationInformation.LocationName)
		f.SetCellStr(sheetName, "K"+fmt.Sprint(rowCount), locationInformation.LocationType)
		f.SetCellInt(sheetName, "L"+fmt.Sprint(rowCount), int(entry.MaterialID))

		rowCount++
	}

	currentTime := time.Now()
	fileName := fmt.Sprintf(
		"Report Balance %s.xlsx",
		currentTime.Format("02-01-2006"),
	)

	tempFilePath := filepath.Join("./storage/import_excel/temp/", fileName)
	f.SaveAs(tempFilePath)
	if err := f.Close(); err != nil {
		fmt.Println(err)
	}

	return fileName, nil
}

func (u *materialLocationUsecase) BalanceReportOutOfProject(projectID uint) (string, error) {
	ctx := context.Background()
	templateFilePath := filepath.Join("./internal/templates/", "Отчет Остатка.xlsx")
	f, err := excelize.OpenFile(templateFilePath)
	if err != nil {
		return "", err
	}

	sheetName := "Отчет"
	f.SetCellValue(sheetName, "L1", "ID материала")
	rowCount := 2

	materialsData, err := u.q.ListBalanceReportData(ctx, db.ListBalanceReportDataParams{
		ProjectID:    pgInt8(projectID),
		LocationType: pgText("out-of-project"),
		Column3:      0,
	})
	if err != nil {
		return "", err
	}

	for _, entry := range materialsData {
		totalAmount := float64FromPgNumeric(entry.TotalAmount)
		defectAmount := float64FromPgNumeric(entry.DefectAmount)
		costM19, _ := decimalFromPgNumeric(entry.MaterialCostM19).Float64()
		totalCost, _ := decimalFromPgNumeric(entry.TotalCost).Float64()
		totalDefectCost, _ := decimalFromPgNumeric(entry.TotalDefectCost).Float64()

		f.SetCellStr(sheetName, "A"+fmt.Sprint(rowCount), entry.MaterialCode)
		f.SetCellStr(sheetName, "B"+fmt.Sprint(rowCount), entry.MaterialName)
		f.SetCellStr(sheetName, "C"+fmt.Sprint(rowCount), entry.MaterialUnit)
		f.SetCellFloat(sheetName, "D"+fmt.Sprint(rowCount), totalAmount, 2, 64)
		f.SetCellFloat(sheetName, "E"+fmt.Sprint(rowCount), defectAmount, 2, 64)
		f.SetCellFloat(sheetName, "F"+fmt.Sprint(rowCount), costM19, 2, 64)
		f.SetCellFloat(sheetName, "G"+fmt.Sprint(rowCount), totalCost, 2, 64)
		f.SetCellFloat(sheetName, "H"+fmt.Sprint(rowCount), totalDefectCost, 2, 64)
		f.SetCellInt(sheetName, "L"+fmt.Sprint(rowCount), int(entry.MaterialID))

		rowCount++
	}

	currentTime := time.Now()
	fileName := fmt.Sprintf(
		"Report Balance %s.xlsx",
		currentTime.Format("02-01-2006"),
	)

	tempFilePath := filepath.Join("./storage/import_excel/temp/", fileName)
	f.SaveAs(tempFilePath)
	if err := f.Close(); err != nil {
		fmt.Println(err)
	}

	return fileName, nil
}

func (u *materialLocationUsecase) GetMaterialCostsInLocation(projectID, materialID, locationID uint, locationType string) ([]model.MaterialCost, error) {
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

func (u *materialLocationUsecase) GetMaterialAmountBasedOnCost(projectID, materialCost, locationID uint, locationType string) (float64, error) {
	amount, err := u.q.GetUniqueMaterialTotalAmount(context.Background(), db.GetUniqueMaterialTotalAmountParams{
		ProjectID:      pgInt8(projectID),
		LocationType:   pgText(locationType),
		LocationID:     pgInt8(locationID),
		MaterialCostID: pgInt8(materialCost),
	})
	if err != nil {
		return 0, err
	}
	return float64FromPgNumeric(amount), nil
}

func toModelMaterialLocation(m db.MaterialLocation) model.MaterialLocation {
	return model.MaterialLocation{
		ID:             uint(m.ID),
		ProjectID:      uintFromPgInt8(m.ProjectID),
		MaterialCostID: uintFromPgInt8(m.MaterialCostID),
		LocationID:     uintFromPgInt8(m.LocationID),
		LocationType:   m.LocationType.String,
		Amount:         float64FromPgNumeric(m.Amount),
	}
}

