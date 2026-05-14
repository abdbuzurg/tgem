package jobs

import (
	"backend-v2/model"
	"fmt"
	"time"

	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type InvoiceMaterialDataForProjectProgressReportDaily struct {
	MaterialCostID uint
	InvoiceAmount  float64
	InvoiceType    string
}

type MaterialDataForProjectProgressReportDaily struct {
	MaterialCostID uint
	LocationType   string
	LocationAmount float64
}

type InvoiceOperationDataForProgressReportDaily struct {
	OperationID     uint
	AmountInInvoice float64
}

// Progress Report job
// The job will execute every day at 23:50:00
func ProgressReportDaily() {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
		viper.GetString("Database.Host"),
		viper.GetString("Database.Username"),
		viper.GetString("Database.Password"),
		viper.GetString("Database.DBName"),
		viper.GetInt("Database.Port"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	projectIDs := []uint{}
	err = db.Raw(`SELECT id FROM projects WHERE name <> 'Администрирование'`).Scan(&projectIDs).Error
	if err != nil {
		fmt.Printf("Ошибка при получении данных проекта: %v", err)
		return
	}

	//date and time of the daily report
	dushanbeLocation, err := time.LoadLocation("Asia/Dushanbe")
	if err != nil {
		fmt.Printf("Не удалось получить время по Душанбе: %v", err)
		return
	}
	dailyDate := time.Now().In(dushanbeLocation)

	// dataLimit is used in queries for LIMITing the data received from queries
	dataLimit := 1000

	for _, projectID := range projectIDs {
		go dailyMaterialProgressBasedOnProjectID(db, projectID, dataLimit, dailyDate)
    go dailyOperationProgressBasedOnProjectID(db, projectID, dataLimit, dailyDate)
	}
}

func dailyMaterialProgressBasedOnProjectID(db *gorm.DB, projectID uint, dataLimit int, date time.Time) {
	dailyProjectProgressReport := []model.ProjectProgressMaterials{}
	dailyProjectProgressReportIndex := -1

	var materialDataCountInProject int
	err := db.Raw(`
      SELECT COUNT(*)
      FROM material_locations
      INNER JOIN material_costs ON material_locations.material_cost_id = material_costs.id
      INNER JOIN materials ON material_costs.material_id = materials.id
      WHERE
        materials.project_id = ? AND	
        materials.show_planned_amount_in_report = true
    `, projectID).Scan(&materialDataCountInProject).Error
	if err != nil {
		fmt.Printf("Ошибка при подсчете данных материалов в проекте %d: %v", projectID, err)
		return
	}

	pageCount := 0
	for materialDataCountInProject > 0 {
		materialDataForProjectPaginated := []MaterialDataForProjectProgressReportDaily{}
		err := db.Raw(`
        SELECT 
          material_costs.id as material_cost_id,
          material_locations.amount as location_amount,
          material_locations.location_type as location_type
        FROM material_locations
        INNER JOIN material_costs ON material_locations.material_cost_id = material_costs.id
        INNER JOIN materials ON material_costs.material_id = materials.id
        WHERE
          materials.project_id = ? AND	
          materials.show_planned_amount_in_report = true
        ORDER BY materials.id, material_costs.id, material_locations.id
        LIMIT ? OFFSET ?
        `, projectID, dataLimit, pageCount*dataLimit).Scan(&materialDataForProjectPaginated).Error
		if err != nil {
			fmt.Printf("Ошибка при получении данных материала в проекте %d, LIMIT %d, OFFSET %d: %v", projectID, dataLimit, (pageCount+1)*dataLimit, err)
			return
		}

		for _, materialData := range materialDataForProjectPaginated {
			oneEntry := model.ProjectProgressMaterials{
				MaterialCostID: materialData.MaterialCostID,
				ProjectID:      projectID,
				Date:           date,
			}

			switch materialData.LocationType {
			case "warehouse":
				oneEntry.AmountInWarehouse = materialData.LocationAmount
				break
			case "team":
				oneEntry.AmountInTeams = materialData.LocationAmount
				break
			case "object":
				oneEntry.AmountInObjects = materialData.LocationAmount
				break
			case "writeoff-warehouse", "loss-warehouse", "loss-team", "writeoff-object", "loss-object":
				oneEntry.AmountWriteOff = materialData.LocationAmount
				break
			}

			if dailyProjectProgressReportIndex == -1 {
				dailyProjectProgressReport = append(dailyProjectProgressReport, oneEntry)
				dailyProjectProgressReportIndex++
				continue
			}

			if dailyProjectProgressReport[dailyProjectProgressReportIndex].MaterialCostID == oneEntry.MaterialCostID {
				dailyProjectProgressReport[dailyProjectProgressReportIndex].AmountInWarehouse += oneEntry.AmountInWarehouse
				dailyProjectProgressReport[dailyProjectProgressReportIndex].AmountInTeams += oneEntry.AmountInTeams
				dailyProjectProgressReport[dailyProjectProgressReportIndex].AmountInObjects += oneEntry.AmountInObjects
				dailyProjectProgressReport[dailyProjectProgressReportIndex].AmountWriteOff += oneEntry.AmountWriteOff
			} else {
				dailyProjectProgressReport = append(dailyProjectProgressReport, oneEntry)
				dailyProjectProgressReportIndex++
			}
		}

		pageCount++
		materialDataCountInProject -= dataLimit
	}

	invoiceMaterialDataForProjectCount := 0
	err = db.Raw(`
      SELECT COUNT(*)
      FROM materials
      INNER JOIN material_costs ON material_costs.material_id = materials.id
      RIGHT JOIN invoice_materials ON invoice_materials.material_cost_id = material_costs.id
      WHERE 
        materials.project_id = ? AND
        materials.show_planned_amount_in_report = true AND
        (invoice_materials.invoice_type = 'input' OR invoice_materials.invoice_type = 'object-correction')
      `, projectID).Scan(&invoiceMaterialDataForProjectCount).Error
	if err != nil {
		fmt.Printf("Ошибка при подсчете данных материалов в проекте %d: %v", projectID, err)
		return
	}

	pageCount = 0
	for invoiceMaterialDataForProjectCount > 0 {
		invoiceMaterialDataForProject := []InvoiceMaterialDataForProjectProgressReportDaily{}
		err := db.Raw(`
        SELECT 
          material_costs.id as material_cost_id,
          invoice_materials.amount as invoice_amount,
          invoice_materials.invoice_type as invoice_type
        FROM materials
        INNER JOIN material_costs ON material_costs.material_id = materials.id
        RIGHT JOIN invoice_materials ON invoice_materials.material_cost_id = material_costs.id
        WHERE 
          materials.project_id = ? AND
          materials.show_planned_amount_in_report = true AND
          (invoice_materials.invoice_type = 'input' OR invoice_materials.invoice_type = 'object-correction')
        ORDER BY materials.id, material_costs.id, invoice_materials.id
        LIMIT ? OFFSET ?
        `, projectID, dataLimit, pageCount*dataLimit).Scan(&invoiceMaterialDataForProject).Error
		if err != nil {
			fmt.Printf("Ошибка при получении данных материала в накладных ПРИХОД и РАСХОД в проекте %d, LIMIT %d, OFFSET %d: %v", projectID, dataLimit, (pageCount+1)*dataLimit, err)
			return
		}

		for _, invoiceMaterialData := range invoiceMaterialDataForProject {
			for index, projectProgress := range dailyProjectProgressReport {
				if projectProgress.MaterialCostID == invoiceMaterialData.MaterialCostID {
					switch invoiceMaterialData.InvoiceType {
					case "input":
						dailyProjectProgressReport[index].Received += invoiceMaterialData.InvoiceAmount
						break
					case "object-correction":
						dailyProjectProgressReport[index].Installed += invoiceMaterialData.InvoiceAmount
						break
					}
				}
			}
		}

		pageCount++
		invoiceMaterialDataForProjectCount -= dataLimit
	}

	if err := db.CreateInBatches(dailyProjectProgressReport, dataLimit).Error; err != nil {
		var projectName string
		db.Raw(`SELECT name FROM projects WHERE id = ?`, projectID).Scan(&projectName)
		fmt.Printf("Неудалось сохранить данные материалов для ежедневного прогресса в проекте %s", projectName)
	}
}

func dailyOperationProgressBasedOnProjectID(db *gorm.DB, projectID uint, dataLimit int, date time.Time) {
	dataForOperations := []model.ProjectProgressOperations{}
	dataForOperationsIndex := -1

	var invoiceOperationsDataForProjectCount int
	err := db.Raw(`
    SELECT COUNT(*)
    FROM invoice_objects
    INNER JOIN invoice_operations ON invoice_operations.invoice_id = invoice_objects.id
    INNER JOIN operations ON operations.id = invoice_operations.operation_id
    WHERE
      invoice_objects.confirmed_by_operator = true AND
      invoice_operations.invoice_type = 'object-correction' AND
      operations.project_id = ? AND
      operations.show_planned_amount_in_report = true
    `, projectID).Scan(&invoiceOperationsDataForProjectCount).Error
	if err != nil {
		fmt.Printf("Ошибка при подсчете услуг в проекте %d: %v", projectID, err)
		return
	}

	pageCount := 0
	for invoiceOperationsDataForProjectCount > 0 {
		invoiceOperationDataForProject := []InvoiceOperationDataForProgressReportDaily{}
		err := db.Raw(`
        SELECT 
          operations.id as operation_id,
          invoice_operations.amount as amount_in_invoice
        FROM invoice_objects
        INNER JOIN invoice_operations ON invoice_operations.invoice_id = invoice_objects.id
        INNER JOIN operations ON operations.id = invoice_operations.operation_id
        WHERE
          invoice_objects.confirmed_by_operator = true AND
          invoice_operations.invoice_type = 'object-correction' AND
          operations.project_id = ? AND
          operations.show_planned_amount_in_report = true
        ORDER BY operations.id, invoice_operations.id, invoice_objects.id
        LIMIT ? OFFSET ?
        `, projectID, dataLimit, pageCount*dataLimit).Scan(&invoiceOperationDataForProject).Error
		if err != nil {
			fmt.Printf("Ошибка при получении данных материала в накладных ПРИХОД и РАСХОД в проекте %d, LIMIT %d, OFFSET %d: %v", projectID, dataLimit, (pageCount+1)*dataLimit, err)
			return
		}

		for _, invoiceOperationData := range invoiceOperationDataForProject {
			oneEntry := model.ProjectProgressOperations{
				ProjectID:   projectID,
				OperationID: invoiceOperationData.OperationID,
				Installed:   invoiceOperationData.AmountInInvoice,
				Date:        date,
			}

			if dataForOperationsIndex == -1 {
				dataForOperations = append(dataForOperations, oneEntry)
				dataForOperationsIndex++
				continue
			}

			if dataForOperations[dataForOperationsIndex].OperationID == oneEntry.OperationID {
				dataForOperations[dataForOperationsIndex].Installed += oneEntry.Installed
			} else {
				dataForOperations = append(dataForOperations, oneEntry)
				dataForOperationsIndex++
			}
		}

		pageCount++
		invoiceOperationsDataForProjectCount -= dataLimit
	}

	if err := db.CreateInBatches(dataForOperations, dataLimit).Error; err != nil {
		var projectName string
		db.Raw(`SELECT name FROM projects WHERE id = ?`, projectID).Scan(&projectName)
		fmt.Printf("Неудалось сохранить данные услуг для ежедневного прогресса в проекте %s", projectName)
	}
}
