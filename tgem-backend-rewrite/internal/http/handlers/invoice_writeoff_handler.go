package handlers

import (
	// "backend-v2/internal/dto"
	"backend-v2/internal/dto"
	"backend-v2/internal/http/response"
	"backend-v2/internal/usecase"
	"backend-v2/pkg/tempfiles"
	"fmt"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type invoiceWriteOffHandler struct {
	invoiceWriteOffUsecase usecase.IInvoiceWriteOffUsecase
}

func NewInvoiceWriteOffHandler(invoiceWriteOffUsecase usecase.IInvoiceWriteOffUsecase) IInvoiceWriteOffHandler {
	return &invoiceWriteOffHandler{
		invoiceWriteOffUsecase: invoiceWriteOffUsecase,
	}
}

type IInvoiceWriteOffHandler interface {
	GetPaginated(c *gin.Context)
	GetInvoiceMaterialsWithoutSerialNumber(c *gin.Context)
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	GetMaterialsForEdit(c *gin.Context)
	GetRawDocument(c *gin.Context)
	Confirmation(c *gin.Context)
	GetDocument(c *gin.Context)
	Report(c *gin.Context)
	GetMaterialsInLocation(c *gin.Context)
}

func (handler *invoiceWriteOffHandler) GetPaginated(c *gin.Context) {
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

	writeOffType := c.DefaultQuery("writeOffType", "")
	writeOffType, err = url.QueryUnescape(writeOffType)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Wrong query parameter provided for returnerType: %v", err))
		return
	}

	projectID := c.GetUint("projectID")
	filter := dto.InvoiceWriteOffSearchParameters{
		ProjectID:    projectID,
		WriteOffType: writeOffType,
	}

	data, err := handler.invoiceWriteOffUsecase.GetPaginated(page, limit, filter)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the paginated data of Invoice: %v", err))
		return
	}

	dataCount, err := handler.invoiceWriteOffUsecase.Count(filter)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the total amount of Invoice: %v", err))
		return
	}

	response.ResponsePaginatedData(c, data, dataCount)
}

func (handler *invoiceWriteOffHandler) Create(c *gin.Context) {
	var createData dto.InvoiceWriteOff
	if err := c.ShouldBindJSON(&createData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	createData.Details.ProjectID = c.GetUint("projectID")
	createData.Details.ReleasedWorkerID = c.GetUint("workerID")

	data, err := handler.invoiceWriteOffUsecase.Create(createData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could perform the creation of Invoice: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *invoiceWriteOffHandler) Update(c *gin.Context) {
	var updateData dto.InvoiceWriteOff
	if err := c.ShouldBindJSON(&updateData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	updateData.Details.ProjectID = c.GetUint("projectID")
	updateData.Details.ReleasedWorkerID = c.GetUint("workerID")

	data, err := handler.invoiceWriteOffUsecase.Update(updateData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the updation of Invoice: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *invoiceWriteOffHandler) Delete(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	err = handler.invoiceWriteOffUsecase.Delete(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the deletion of Invoice: %v", err))
		return
	}

	response.ResponseSuccess(c, "deleted")
}

func (handler *invoiceWriteOffHandler) GetRawDocument(c *gin.Context) {
	deliveryCode := c.Param("deliveryCode")
	c.FileAttachment("./storage/import_excel/writeoff/"+deliveryCode+".xlsx", deliveryCode+".xlsx")
}

func (handler *invoiceWriteOffHandler) GetInvoiceMaterialsWithoutSerialNumber(c *gin.Context) {

	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return
	}

	data, err := handler.invoiceWriteOffUsecase.GetInvoiceMaterialsWithoutSerialNumbers(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, data)

}

func (handler *invoiceWriteOffHandler) GetMaterialsForEdit(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)

	locationIDRaw := c.Param("locationID")
	locationID, err := strconv.Atoi(locationIDRaw)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid parameters in request: %v", err))
		return
	}

	locationType := c.Param("locationType")

	result, err := handler.invoiceWriteOffUsecase.GetMaterialsForEdit(uint(id), locationType, uint(locationID))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, result)
}

func (handler *invoiceWriteOffHandler) Confirmation(c *gin.Context) {

	projectID := c.GetUint("projectID")

	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return
	}

	invoiceWriteOff, err := handler.invoiceWriteOffUsecase.GetByID(uint(id))
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
	fileExtension := fileNameAndExtension[len(fileNameAndExtension) - 1]
	if fileExtension != "pdf" {
		response.ResponseError(c, fmt.Sprintf("Файл должен быть формата PDF"))
		return
	}
	file.Filename = invoiceWriteOff.DeliveryCode + "." + fileExtension
	filePath := filepath.Join("./storage/import_excel/writeoff/", file.Filename)

	err = c.SaveUploadedFile(file, filePath)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	err = handler.invoiceWriteOffUsecase.Confirmation(uint(id), projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, true)
}

func (handler *invoiceWriteOffHandler) GetDocument(c *gin.Context) {

	deliveryCode := c.Param("deliveryCode")

	filePath := filepath.Join("./storage/import_excel/writeoff/", deliveryCode)
	fileGlob, err := filepath.Glob(filePath + ".*")
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	filePath = fileGlob[0]
	pathSeparated := strings.Split(filePath, ".")
	deliveryCodeExtension := pathSeparated[len(pathSeparated)-1]

	c.FileAttachment(filePath, deliveryCode+"."+deliveryCodeExtension)
}

func (handler *invoiceWriteOffHandler) Report(c *gin.Context) {
	var reportParameters dto.InvoiceWriteOffReportParameters
	if err := c.ShouldBindJSON(&reportParameters); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	reportParameters.ProjectID = c.GetUint("projectID")

	filename, err := handler.invoiceWriteOffUsecase.Report(reportParameters)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	filePath := filepath.Join("./storage/import_excel/temp/", filename)
	tempfiles.Track(c, filePath)
	c.FileAttachment(filePath, filename)
}

func (handler *invoiceWriteOffHandler) GetMaterialsInLocation(c *gin.Context) {

	projectID := c.GetUint("projectID")

	locationIDRaw := c.Param("locationID")
	locationID, err := strconv.Atoi(locationIDRaw)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid parameters in request: %v", err))
		return
	}

	locationType := c.Param("locationType")

	data, err := handler.invoiceWriteOffUsecase.GetMaterialsInLocation(projectID, uint(locationID), locationType)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}
