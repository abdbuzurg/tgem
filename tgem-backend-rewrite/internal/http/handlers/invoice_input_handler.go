package handlers

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/http/response"
	"backend-v2/internal/usecase"
	"backend-v2/model"
	"backend-v2/pkg/tempfiles"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type invoiceInputHandler struct {
	invoiceInputUsecase usecase.IInvoiceInputUsecase
	userActionUsecase   usecase.IUserActionUsecase
}

func NewInvoiceInputHandler(
	invoiceInputUsecase usecase.IInvoiceInputUsecase,
	userActionUsecase usecase.IUserActionUsecase,
) IInvoiceInputHandler {
	return &invoiceInputHandler{
		invoiceInputUsecase: invoiceInputUsecase,
		userActionUsecase:   userActionUsecase,
	}
}

type IInvoiceInputHandler interface {
	GetAll(c *gin.Context)
	GetPaginated(c *gin.Context)
	GetInvoiceMaterialsWithoutSerialNumbers(c *gin.Context)
	GetInvoiceMaterialsWithSerialNumbers(c *gin.Context)
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	Confirmation(c *gin.Context)
	GetDocument(c *gin.Context)
	UniqueCode(c *gin.Context)
	UniqueReleased(c *gin.Context)
	UniqueWarehouseManager(c *gin.Context)
	Report(c *gin.Context)
	NewMaterial(c *gin.Context)
	NewMaterialCost(c *gin.Context)
	GetMaterialsForEdit(c *gin.Context)
	Import(c *gin.Context)
	GetParametersForSearch(c *gin.Context)
}

func (handler *invoiceInputHandler) GetAll(c *gin.Context) {

	data, err := handler.invoiceInputUsecase.GetAll()
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get Invoice Input data: %v", err))
		return
	}

	response.ResponseSuccess(c, data)

}

func (handler *invoiceInputHandler) GetPaginated(c *gin.Context) {

	projectID := c.GetUint("projectID")
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return
	}

	limitStr := c.DefaultQuery("limit", "25")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return
	}

  deliveryCode := c.DefaultQuery("deliveryCode", "")

	warehouseManagerWorkerIDStr := c.DefaultQuery("warehouseManagerWorkerID", "")
	warehouseManagerWorkerID, err := strconv.Atoi(warehouseManagerWorkerIDStr)
	if err != nil || warehouseManagerWorkerID < 0 {
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return
	}

	releasedWorkerIDStr := c.DefaultQuery("releasedWorkerID", "")
	releasedWorkerID, err := strconv.Atoi(releasedWorkerIDStr)
	if err != nil || releasedWorkerID < 0 {
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return
	}

  dateLayout := "Mon Jan 02 2006 03:04:05"

  dateFromStr := c.DefaultQuery("dateFrom", "")
  var dateFrom time.Time
  if dateFromStr != "" {
    dateFrom, err = time.Parse(dateLayout, dateFromStr)
    if err != nil {
      response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
      return
    }
  }

  dateToStr := c.DefaultQuery("dateTo", "")
  var dateTo time.Time
  if dateToStr != "" {
    dateTo, err = time.Parse(dateLayout, dateToStr)
    if err != nil {
      response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
    }
  }

	materialIDsStr := c.DefaultQuery("materials", "")
	materialIDs := []uint{}
	if materialIDsStr != "" {
		chunks := strings.Split(materialIDsStr, ",")
		for _, chunk := range chunks {
			id, err := strconv.Atoi(chunk)
			if err != nil || id <= 0 {
				response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
				return
			}
      
      materialIDs =append(materialIDs, uint(id))
		}
	}

	filter := dto.InvoiceInputSearchParameters{
		ProjectID: projectID,
    DeliveryCode: deliveryCode,
    WarehouseManagerWorkerID: uint(warehouseManagerWorkerID),
    ReleasedWorkerID: uint(releasedWorkerID),
    DateFrom: dateFrom,
    DateTo: dateTo,
    Materials: materialIDs,
	}

	data, err := handler.invoiceInputUsecase.GetPaginated(page, limit, filter)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	dataCount, err := handler.invoiceInputUsecase.Count(filter)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponsePaginatedData(c, data, dataCount)

}

func (handler *invoiceInputHandler) GetInvoiceMaterialsWithoutSerialNumbers(c *gin.Context) {

	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return
	}

	data, err := handler.invoiceInputUsecase.GetInvoiceMaterialsWithoutSerialNumbers(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *invoiceInputHandler) GetInvoiceMaterialsWithSerialNumbers(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return
	}

	data, err := handler.invoiceInputUsecase.GetInvoiceMaterialsWithSerialNumbers(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *invoiceInputHandler) Create(c *gin.Context) {

	workerID := c.GetUint("workerID")
	projectID := c.GetUint("projectID")

	var createData dto.InvoiceInput
	if err := c.ShouldBindJSON(&createData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return
	}

	createData.Details.ProjectID = projectID
	createData.Details.ReleasedWorkerID = workerID

	data, err := handler.invoiceInputUsecase.Create(createData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *invoiceInputHandler) Update(c *gin.Context) {
	var updateData dto.InvoiceInput
	if err := c.ShouldBindJSON(&updateData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	workerID := c.GetUint("workerID")
	projectID := c.GetUint("projectID")

	updateData.Details.ProjectID = projectID
	updateData.Details.ReleasedWorkerID = workerID

	data, err := handler.invoiceInputUsecase.Update(updateData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the updation of Invoice: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *invoiceInputHandler) Delete(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return
	}

	err = handler.invoiceInputUsecase.Delete(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, "deleted")
}

func (handler *invoiceInputHandler) Confirmation(c *gin.Context) {

	projectID := c.GetUint("projectID")

	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return
	}

	invoiceInput, err := handler.invoiceInputUsecase.GetByID(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	fileNameAndExtension := strings.Split(file.Filename, ".")
	fileExtension := fileNameAndExtension[len(fileNameAndExtension)-1]
	if fileExtension != "pdf" {
		response.ResponseError(c, fmt.Sprintf("Файл должен быть формата PDF"))
		return
	}
	file.Filename = invoiceInput.DeliveryCode + "." + fileExtension
	filePath := filepath.Join("./storage/import_excel/input/", file.Filename)

	err = c.SaveUploadedFile(file, filePath)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	err = handler.invoiceInputUsecase.Confirmation(uint(id), projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, true)
}

func (handler *invoiceInputHandler) GetDocument(c *gin.Context) {
	fileName := c.Param("deliveryCode") + ".pdf"
	filePath := filepath.Join("./storage/import_excel/input/", fileName)
	if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
		response.ResponseError(c, fmt.Sprint("Внутренняя ошибка сервера: Файл не существует"))
		return
	}
	c.FileAttachment(filePath, fileName)
}

func (handler *invoiceInputHandler) UniqueCode(c *gin.Context) {

	projectID := c.GetUint("projectID")

	data, err := handler.invoiceInputUsecase.UniqueCode(projectID)
	if err != nil {

		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return

	}

	response.ResponseSuccess(c, data)
}

func (handler *invoiceInputHandler) UniqueWarehouseManager(c *gin.Context) {

	projectID := c.GetUint("projectID")

	data, err := handler.invoiceInputUsecase.UniqueWarehouseManager(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *invoiceInputHandler) UniqueReleased(c *gin.Context) {

	projectID := c.GetUint("projectID")

	data, err := handler.invoiceInputUsecase.UniqueReleased(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *invoiceInputHandler) Report(c *gin.Context) {

	projectID := c.GetUint("projectID")

	var filter dto.InvoiceInputReportFilterRequest
	if err := c.ShouldBindJSON(&filter); err != nil {
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return
	}

	filter.ProjectID = projectID
	filename, err := handler.invoiceInputUsecase.Report(filter)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	filePath := filepath.Join("./storage/import_excel/temp/", filename)
	tempfiles.Track(c, filePath)
	c.FileAttachment(filePath, filename)
	// response.ResponseSuccess(c, true)
}

func (handler *invoiceInputHandler) NewMaterial(c *gin.Context) {
	var data dto.NewMaterialDataFromInvoiceInput
	if err := c.ShouldBindJSON(&data); err != nil {
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса %v", err))
		return
	}

	projectID := c.GetUint("projectID")
	data.ProjectID = projectID

	err := handler.invoiceInputUsecase.NewMaterialAndItsCost(data)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, true)
}

func (handler *invoiceInputHandler) NewMaterialCost(c *gin.Context) {
	var data model.MaterialCost
	if err := c.ShouldBindJSON(&data); err != nil {
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса %v", err))
		return
	}

	err := handler.invoiceInputUsecase.NewMaterialCost(data)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, true)
}

func (handler *invoiceInputHandler) GetMaterialsForEdit(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)

	result, err := handler.invoiceInputUsecase.GetMaterialsForEdit(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, result)
}

func (handler *invoiceInputHandler) Import(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Файл не может быть сформирован, проверьте файл: %v", err))
		return
	}

	date := time.Now()
	importFileName := date.Format("2006-01-02 15-04-05") + file.Filename
	importFilePath := filepath.Join("./storage/import_excel/temp/", importFileName)
	tempfiles.Track(c, importFilePath)
	err = c.SaveUploadedFile(file, importFilePath)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Файл не может быть сохранен на сервере: %v", err))
		return
	}

	projectID := c.GetUint("projectID")
	workerID := c.GetUint("workerID")
	err = handler.invoiceInputUsecase.Import(importFilePath, projectID, workerID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, true)

}

func (handler *invoiceInputHandler) GetParametersForSearch(c *gin.Context) {
	data, err := handler.invoiceInputUsecase.GetParametersForSearch(c.GetUint("projectID"))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}
