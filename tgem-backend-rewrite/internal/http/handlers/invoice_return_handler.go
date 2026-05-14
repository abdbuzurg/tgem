package handlers

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/http/response"
	"backend-v2/internal/usecase"
	"backend-v2/pkg/tempfiles"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type invoiceReturnHandler struct {
	invoiceReturnUsecase usecase.IInvoiceReturnUsecase
}

func NewInvoiceReturnHandler(invoiceReturnUsecase usecase.IInvoiceReturnUsecase) IInvoiceReturnHandler {
	return &invoiceReturnHandler{
		invoiceReturnUsecase: invoiceReturnUsecase,
	}
}

type IInvoiceReturnHandler interface {
	GetAll(c *gin.Context)
	GetPaginated(c *gin.Context)
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	GetDocument(c *gin.Context)
	Confirmation(c *gin.Context)
	UniqueCode(c *gin.Context)
	UniqueTeam(c *gin.Context)
	UniqueObject(c *gin.Context)
	Report(c *gin.Context)
	GetUniqueMaterialCostsFromLocation(c *gin.Context)
	GetMaterialsInLocation(c *gin.Context)
	GetMaterialAmountInLocation(c *gin.Context)
	GetSerialNumberCodesInLocation(c *gin.Context)
	GetInvoiceMaterialsWithSerialNumbers(c *gin.Context)
	GetInvoiceMaterialsWithoutSerialNumbers(c *gin.Context)
	GetMaterialsForEdit(c *gin.Context)
	GetMaterialAmountByMaterialID(c *gin.Context)
}

func (handler *invoiceReturnHandler) GetAll(c *gin.Context) {
	data, err := handler.invoiceReturnUsecase.GetAll()
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get Invoice Input data: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *invoiceReturnHandler) GetPaginated(c *gin.Context) {
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

	returnType := c.DefaultQuery("type", "")
	if returnType == "" {
		response.ResponseError(c, fmt.Sprintf("Wrong query parameter provided for limit: %v", err))
		return
	}

	projectID := c.GetUint("projectID")

	var data interface{}
	if returnType == "team" {
		data, err = handler.invoiceReturnUsecase.GetPaginatedTeam(page, limit, projectID)
		if err != nil {
			response.ResponseError(c, fmt.Sprintf("Could not get the paginated data of Invoice: %v", err))
			return
		}
	}

	if returnType == "object" {
		data, err = handler.invoiceReturnUsecase.GetPaginatedObject(page, limit, projectID)
		if err != nil {
			response.ResponseError(c, fmt.Sprintf("Could not get the paginated data of Invoice: %v", err))
			return
		}
	}

	dataCount, err := handler.invoiceReturnUsecase.CountBasedOnType(projectID, returnType)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the total amount of Invoice: %v", err))
		return
	}

	response.ResponsePaginatedData(c, data, dataCount)
}

func (handler *invoiceReturnHandler) Create(c *gin.Context) {
	var createData dto.InvoiceReturn
	if err := c.ShouldBindJSON(&createData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	projectID := c.GetUint("projectID")
	createData.Details.ProjectID = projectID
	data, err := handler.invoiceReturnUsecase.Create(createData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could perform the creation of Invoice: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *invoiceReturnHandler) Update(c *gin.Context) {
	var updateData dto.InvoiceReturn
	if err := c.ShouldBindJSON(&updateData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	projectID := c.GetUint("projectID")
	updateData.Details.ProjectID = projectID

	existing, err := handler.invoiceReturnUsecase.GetByID(updateData.Details.ID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Не удалось найти накладную: %v", err))
		return
	}
	if existing.ProjectID != projectID {
		response.ResponseError(c, "Доступ запрещен: накладная принадлежит другому проекту")
		return
	}

	data, err := handler.invoiceReturnUsecase.Update(updateData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could perform the creation of Invoice: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *invoiceReturnHandler) Delete(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	projectID := c.GetUint("projectID")
	existing, err := handler.invoiceReturnUsecase.GetByID(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Не удалось найти накладную: %v", err))
		return
	}
	if existing.ProjectID != projectID {
		response.ResponseError(c, "Доступ запрещен: накладная принадлежит другому проекту")
		return
	}

	err = handler.invoiceReturnUsecase.Delete(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the deletion of Invoice: %v", err))
		return
	}

	response.ResponseSuccess(c, "deleted")
}

func (handler *invoiceReturnHandler) Confirmation(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	projectID := c.GetUint("projectID")
	invoiceReturn, err := handler.invoiceReturnUsecase.GetByID(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Cannot find invoice Return by id %v: %v", id, err))
		return
	}
	if invoiceReturn.ProjectID != projectID {
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
		response.ResponseError(c, "Файл должен быть формата PDF")
		return
	}
	file.Filename = invoiceReturn.DeliveryCode + "." + fileExtension
	filePath := filepath.Join("./storage/import_excel/return/", file.Filename)

	err = c.SaveUploadedFile(file, filePath)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("cannot save file: %v", err))
		return
	}

	excelFilePath := filepath.Join("./storage/import_excel/return/", invoiceReturn.DeliveryCode+".xlsx")
	os.Remove(excelFilePath)

	err = handler.invoiceReturnUsecase.Confirmation(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("cannot confirm invoice input with id %v: %v", id, err))
		return
	}

	response.ResponseSuccess(c, true)
}

func (handler *invoiceReturnHandler) GetDocument(c *gin.Context) {
	deliveryCode := c.Param("deliveryCode")
	projectID := c.GetUint("projectID")
	extension, err := handler.invoiceReturnUsecase.GetDocument(deliveryCode, projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
		return
	}
	filePath, _ := resolveInvoiceDoc(deliveryCode, extension, "./storage/import_excel/return", "return")
	if filePath == "" {
		response.ResponseError(c, "Внутренняя ошибка сервера: Файл не существует")
		return
	}
	c.FileAttachment(filePath, deliveryCode+extension)
}

func (handler *invoiceReturnHandler) UniqueCode(c *gin.Context) {
	projectID := c.GetUint("projectID")
	data, err := handler.invoiceReturnUsecase.UniqueCode(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server Error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *invoiceReturnHandler) UniqueTeam(c *gin.Context) {
	projectID := c.GetUint("projectID")
	data, err := handler.invoiceReturnUsecase.UniqueTeam(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server Error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *invoiceReturnHandler) UniqueObject(c *gin.Context) {
	projectID := c.GetUint("projectID")
	data, err := handler.invoiceReturnUsecase.UniqueObject(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server Error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *invoiceReturnHandler) Report(c *gin.Context) {
	var filter dto.InvoiceReturnReportFilterRequest
	if err := c.ShouldBindJSON(&filter); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid body request: %v", err))
		return
	}

	projectID := c.GetUint("projectID")
	filename, err := handler.invoiceReturnUsecase.Report(filter, projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
		return
	}

	filePath := filepath.Join("./storage/import_excel/temp/", filename)
	tempfiles.Track(c, filePath)
	c.FileAttachment(filePath, filename)
}

func (handler *invoiceReturnHandler) GetUniqueMaterialCostsFromLocation(c *gin.Context) {

	projectID := c.GetUint("projectID")

	locationIDRaw := c.Param("locationID")
	locationID, err := strconv.Atoi(locationIDRaw)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid parameters in request: %v", err))
		return
	}

	locationType := c.Param("locationType")

	materialIDRaw := c.Param("materialID")
	materialID, err := strconv.Atoi(materialIDRaw)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid parameters in request: %v", err))
		return
	}

	data, err := handler.invoiceReturnUsecase.GetMaterialCostInLocation(projectID, uint(locationID), uint(materialID), locationType)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *invoiceReturnHandler) GetMaterialsInLocation(c *gin.Context) {

	projectID := c.GetUint("projectID")

	locationIDRaw := c.Param("locationID")
	locationID, err := strconv.Atoi(locationIDRaw)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid parameters in request: %v", err))
		return
	}

	locationType := c.Param("locationType")

	data, err := handler.invoiceReturnUsecase.GetMaterialsInLocation(projectID, uint(locationID), locationType)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *invoiceReturnHandler) GetMaterialAmountInLocation(c *gin.Context) {

	projectID := c.GetUint("projectID")

	locationIDRaw := c.Param("locationID")
	locationID, err := strconv.Atoi(locationIDRaw)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid parameters in request: %v", err))
		return
	}

	locationType := c.Param("locationType")

	materialCostIDRaw := c.Param("materialCostID")
	materialCostID, err := strconv.Atoi(materialCostIDRaw)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid parameters in request: %v", err))
		return
	}

	data, err := handler.invoiceReturnUsecase.GetMaterialAmountInLocation(projectID, uint(locationID), uint(materialCostID), locationType)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)

}

func (handler *invoiceReturnHandler) GetSerialNumberCodesInLocation(c *gin.Context) {

	materialIDRaw := c.Param("materialID")
	materialID, err := strconv.Atoi(materialIDRaw)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid parameters in request: %v", err))
		return
	}

	locationType := c.Param("locationType")
	locationIDRaw := c.Param("locationID")
	locationID, err := strconv.Atoi(locationIDRaw)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid parameters in request: %v", err))
		return
	}
	projectID := c.GetUint("projectID")

	data, err := handler.invoiceReturnUsecase.GetSerialNumberCodesInLocation(projectID, uint(materialID), locationType, uint(locationID))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *invoiceReturnHandler) GetInvoiceMaterialsWithoutSerialNumbers(c *gin.Context) {

	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return
	}

	projectID := c.GetUint("projectID")

	data, err := handler.invoiceReturnUsecase.GetInvoiceMaterialsWithoutSerialNumbers(uint(id), projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *invoiceReturnHandler) GetInvoiceMaterialsWithSerialNumbers(c *gin.Context) {

	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return
	}

	projectID := c.GetUint("projectID")

	data, err := handler.invoiceReturnUsecase.GetInvoiceMaterialsWithSerialNumbers(uint(id), projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *invoiceReturnHandler) GetMaterialsForEdit(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)

	locationIDRaw := c.Param("locationID")
	locationID, err := strconv.Atoi(locationIDRaw)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid parameters in request: %v", err))
		return
	}

	locationType := c.Param("locationType")

	projectID := c.GetUint("projectID")

	result, err := handler.invoiceReturnUsecase.GetMaterialsForEdit(uint(id), locationType, uint(locationID), projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, result)
}

func (handler *invoiceReturnHandler) GetMaterialAmountByMaterialID(c *gin.Context) {

	locationType := c.Param("locationType")

	locationIDRaw := c.Param("locationID")
	locationID, err := strconv.ParseUint(locationIDRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return
	}

	materialIDRaw := c.Param("materialID")
	materialID, err := strconv.ParseUint(materialIDRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return
	}

	result, err := handler.invoiceReturnUsecase.GetMaterialAmountByMaterialID(c.GetUint("projectID"), uint(materialID), uint(locationID), locationType)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, result)
}
