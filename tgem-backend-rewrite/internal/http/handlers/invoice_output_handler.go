package handlers

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/http/response"
	"backend-v2/internal/usecase"
	"backend-v2/model"
	"backend-v2/pkg/tempfiles"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type invoiceOutputHandler struct {
	invoiceOutputUsecase usecase.IInvoiceOutputUsecase
}

func NewInvoiceOutputHandler(invoiceOutputUsecase usecase.IInvoiceOutputUsecase) IInvoiceOutputHandler {
	return &invoiceOutputHandler{
		invoiceOutputUsecase: invoiceOutputUsecase,
	}
}

type IInvoiceOutputHandler interface {
	GetAll(c *gin.Context)
	GetPaginated(c *gin.Context)
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	GetDocument(c *gin.Context)
	GetInvoiceMaterialsWithoutSerialNumbers(c *gin.Context)
	GetInvoiceMaterialsWithSerialNumbers(c *gin.Context)
	Confirmation(c *gin.Context)
	UniqueCode(c *gin.Context)
	UniqueWarehouseManager(c *gin.Context)
	UniqueRecieved(c *gin.Context)
	UniqueDistrict(c *gin.Context)
	UniqueTeam(c *gin.Context)
	Report(c *gin.Context)
	GetTotalAmountInWarehouse(c *gin.Context)
	GetCodesByMaterialID(c *gin.Context)
	GetAvailableMaterialsInWarehouse(c *gin.Context)
	GetMaterialsForEdit(c *gin.Context)
	Import(c *gin.Context)
}

func (handler *invoiceOutputHandler) GetAll(c *gin.Context) {
	data, err := handler.invoiceOutputUsecase.GetAll()
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get Invoice Input data: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *invoiceOutputHandler) GetPaginated(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		response.ResponseError(c, fmt.Sprintf("Wrong query parameter provided for page: %v", err))
		return
	}

	limitStr := c.DefaultQuery("limit", "25")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		response.ResponseError(c, fmt.Sprintf("Wrong query parameter provided for limit: %v", err))
		return
	}

	projectID := c.GetUint("projectID")
	filter := model.InvoiceOutput{
		ProjectID: projectID,
	}

	data, err := handler.invoiceOutputUsecase.GetPaginated(page, limit, filter)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the paginated data of Invoice: %v", err))
		return
	}

	dataCount, err := handler.invoiceOutputUsecase.Count(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the total amount of Invoice: %v", err))
		return
	}

	response.ResponsePaginatedData(c, data, dataCount)
}

func (handler *invoiceOutputHandler) GetInvoiceMaterialsWithoutSerialNumbers(c *gin.Context) {

	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return
	}

	projectID := c.GetUint("projectID")

	data, err := handler.invoiceOutputUsecase.GetInvoiceMaterialsWithoutSerialNumbers(uint(id), projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *invoiceOutputHandler) GetInvoiceMaterialsWithSerialNumbers(c *gin.Context) {

	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return
	}

	projectID := c.GetUint("projectID")

	data, err := handler.invoiceOutputUsecase.GetInvoiceMaterialsWithSerialNumbers(uint(id), projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *invoiceOutputHandler) Create(c *gin.Context) {
	var createData dto.InvoiceOutput
	if err := c.ShouldBindJSON(&createData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	workerID := c.GetUint("workerID")
	createData.Details.ReleasedWorkerID = workerID

	projectID := c.GetUint("projectID")
	createData.Details.ProjectID = projectID

	data, err := handler.invoiceOutputUsecase.Create(createData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could perform the creation of Invoice: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *invoiceOutputHandler) Delete(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	projectID := c.GetUint("projectID")
	existing, err := handler.invoiceOutputUsecase.GetByID(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Не удалось найти накладную: %v", err))
		return
	}
	if existing.ProjectID != projectID {
		response.ResponseError(c, "Доступ запрещен: накладная принадлежит другому проекту")
		return
	}

	err = handler.invoiceOutputUsecase.Delete(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the deletion of Invoice: %v", err))
		return
	}

	response.ResponseSuccess(c, "deleted")
}

func (handler *invoiceOutputHandler) Confirmation(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	projectID := c.GetUint("projectID")
	invoiceOutput, err := handler.invoiceOutputUsecase.GetByID(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Cannot find invoice Output by id %v: %v", id, err))
		return
	}
	if invoiceOutput.ProjectID != projectID {
		response.ResponseError(c, "Доступ запрещен: накладная принадлежит другому проекту")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("cannot form file: %v", err))
		return
	}

	fileNameAndExtension := strings.Split(file.Filename, ".")
	fileExtension := fileNameAndExtension[len(fileNameAndExtension)-1]
	if fileExtension != "pdf" {
		response.ResponseError(c, fmt.Sprintf("Файл должен быть формата PDF"))
		return
	}
	file.Filename = invoiceOutput.DeliveryCode + "." + fileExtension
	filePath := filepath.Join("./storage/import_excel/output/", file.Filename)

	err = c.SaveUploadedFile(file, filePath)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Ошибка сохранения файла: %v", err))
		return
	}

	excelFilePath := filepath.Join("./storage/import_excel/output/", invoiceOutput.DeliveryCode+".xlsx")
	os.Remove(excelFilePath)

	err = handler.invoiceOutputUsecase.Confirmation(uint(id))
	if err != nil {
    response.ResponseError(c, fmt.Sprintf("Ошибка подтверждения: %v", err))
		os.Remove(filePath)
		return
	}

	response.ResponseSuccess(c, true)
}

func (handler *invoiceOutputHandler) GetDocument(c *gin.Context) {
	deliveryCode := c.Param("deliveryCode")
	projectID := c.GetUint("projectID")
	extension, err := handler.invoiceOutputUsecase.GetDocument(deliveryCode, projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
		return
	}
	filePath, _ := resolveInvoiceDoc(deliveryCode, extension, "./storage/import_excel/output", "output")
	if filePath == "" {
		response.ResponseError(c, "Внутренняя ошибка сервера: Файл не существует")
		return
	}
	c.FileAttachment(filePath, deliveryCode+extension)
}

func (handler *invoiceOutputHandler) UniqueCode(c *gin.Context) {
	projectID := c.GetUint("projectID")
	data, err := handler.invoiceOutputUsecase.UniqueCode(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *invoiceOutputHandler) UniqueWarehouseManager(c *gin.Context) {
	projectID := c.GetUint("projectID")
	data, err := handler.invoiceOutputUsecase.UniqueWarehouseManager(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *invoiceOutputHandler) UniqueRecieved(c *gin.Context) {
	projectID := c.GetUint("projectID")
	data, err := handler.invoiceOutputUsecase.UniqueRecieved(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *invoiceOutputHandler) UniqueDistrict(c *gin.Context) {
	projectID := c.GetUint("projectID")
	data, err := handler.invoiceOutputUsecase.UniqueDistrict(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *invoiceOutputHandler) UniqueTeam(c *gin.Context) {
	projectID := c.GetUint("projectID")
	data, err := handler.invoiceOutputUsecase.UniqueTeam(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *invoiceOutputHandler) Report(c *gin.Context) {
	var filter dto.InvoiceOutputReportFilterRequest
	if err := c.ShouldBindJSON(&filter); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	filter.ProjectID = c.GetUint("projectID")
	filename, err := handler.invoiceOutputUsecase.Report(filter)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
		return
	}

	filePath := filepath.Join("./storage/import_excel/temp/", filename)
	tempfiles.Track(c, filePath)
	c.FileAttachment(filePath, filename)
}

func (handler *invoiceOutputHandler) GetTotalAmountInWarehouse(c *gin.Context) {
	materialIDRaw := c.Param("materialID")
	materialID, err := strconv.ParseUint(materialIDRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	if materialID == 0 {
		response.ResponseSuccess(c, 0)
		return
	}

	projectID := c.GetUint("projectID")

	totalAmount, err := handler.invoiceOutputUsecase.GetTotalMaterialAmount(projectID, uint(materialID))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
		return
	}

	response.ResponseSuccess(c, totalAmount)
}

func (handler *invoiceOutputHandler) GetCodesByMaterialID(c *gin.Context) {

	materialIDRaw := c.Param("materialID")
	materialID, err := strconv.ParseUint(materialIDRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	projectID := c.GetUint("projectID")

	data, err := handler.invoiceOutputUsecase.GetSerialNumbersByMaterial(projectID, uint(materialID))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)

}

func (handler *invoiceOutputHandler) GetAvailableMaterialsInWarehouse(c *gin.Context) {
	projectID := c.GetUint("projectID")

	data, err := handler.invoiceOutputUsecase.GetAvailableMaterialsInWarehouse(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)

}

func (handler *invoiceOutputHandler) GetMaterialsForEdit(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)

	projectID := c.GetUint("projectID")

	result, err := handler.invoiceOutputUsecase.GetMaterialsForEdit(uint(id), projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, result)
}

func (handler *invoiceOutputHandler) Update(c *gin.Context) {
	var updateData dto.InvoiceOutput
	if err := c.ShouldBindJSON(&updateData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	workerID := c.GetUint("workerID")
	updateData.Details.ReleasedWorkerID = workerID

	projectID := c.GetUint("projectID")
	updateData.Details.ProjectID = projectID

	existing, err := handler.invoiceOutputUsecase.GetByID(updateData.Details.ID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Не удалось найти накладную: %v", err))
		return
	}
	if existing.ProjectID != projectID {
		response.ResponseError(c, "Доступ запрещен: накладная принадлежит другому проекту")
		return
	}

	data, err := handler.invoiceOutputUsecase.Update(updateData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could perform the creation of Invoice: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *invoiceOutputHandler) Import(c *gin.Context) {
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
	err = handler.invoiceOutputUsecase.Import(importFilePath, projectID, workerID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, true)

}
