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

type invoiceOutputOutOfProjectHandler struct {
	invoiceOutputOutOfProjectUsecase usecase.IInvoiceOutputOutOfProjectUsecase
}

func NewInvoiceOutputOutOfProjectHandler(
	invoiceOutputOutOfProjectUsecase usecase.IInvoiceOutputOutOfProjectUsecase,
) IInvoiceOutputOutOfProjectHandler {
	return &invoiceOutputOutOfProjectHandler{
		invoiceOutputOutOfProjectUsecase: invoiceOutputOutOfProjectUsecase,
	}
}

type IInvoiceOutputOutOfProjectHandler interface {
	GetPaginated(c *gin.Context)
	Create(c *gin.Context)
	GetInvoiceMaterialsWithSerialNumbers(c *gin.Context)
	GetInvoiceMaterialsWithoutSerialNumbers(c *gin.Context)
	Confirmation(c *gin.Context)
	Update(c *gin.Context)
	GetMaterialsForEdit(c *gin.Context)
	UniqueNameOfProjects(c *gin.Context)
	Report(c *gin.Context)
	GetDocument(c *gin.Context)
  Delete(c *gin.Context)
}

func (handler *invoiceOutputOutOfProjectHandler) GetPaginated(c *gin.Context) {
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

	nameOfProject := c.DefaultQuery("nameOfProject", "")

	releasedWorkerIDStr := c.DefaultQuery("releasedWorkerID", "0")
	releasedWorkerID, err := strconv.Atoi(releasedWorkerIDStr)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Wrong query parameter provided for releasedWorkerID: %v", err))
		return
	}

	filter := dto.InvoiceOutputOutOfProjectSearchParameters{
		ProjectID:        c.GetUint("projectID"),
		NameOfProject:    nameOfProject,
		ReleasedWorkerID: uint(releasedWorkerID),
	}

	data, err := handler.invoiceOutputOutOfProjectUsecase.GetPaginated(page, limit, filter)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the paginated data of Invoice: %v", err))
		return
	}

	dataCount, err := handler.invoiceOutputOutOfProjectUsecase.Count(filter)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the total amount of Invoice: %v", err))
		return
	}

	response.ResponsePaginatedData(c, data, dataCount)
}

func (handler *invoiceOutputOutOfProjectHandler) Create(c *gin.Context) {
	createData := dto.InvoiceOutputOutOfProject{}
	if err := c.ShouldBindJSON(&createData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	workerID := c.GetUint("workerID")
	createData.Details.ReleasedWorkerID = workerID

	projectID := c.GetUint("projectID")
	createData.Details.ProjectID = projectID

	data, err := handler.invoiceOutputOutOfProjectUsecase.Create(createData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could perform the creation of Invoice: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *invoiceOutputOutOfProjectHandler) Delete(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	projectID := c.GetUint("projectID")
	existing, err := handler.invoiceOutputOutOfProjectUsecase.GetByID(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Не удалось найти накладную: %v", err))
		return
	}
	if existing.ProjectID != projectID {
		response.ResponseError(c, "Доступ запрещен: накладная принадлежит другому проекту")
		return
	}

	err = handler.invoiceOutputOutOfProjectUsecase.Delete(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the deletion of Invoice: %v", err))
		return
	}

	response.ResponseSuccess(c, "deleted")
}

func (handler *invoiceOutputOutOfProjectHandler) GetInvoiceMaterialsWithoutSerialNumbers(c *gin.Context) {

	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return
	}

	projectID := c.GetUint("projectID")

	data, err := handler.invoiceOutputOutOfProjectUsecase.GetInvoiceMaterialsWithoutSerialNumbers(uint(id), projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *invoiceOutputOutOfProjectHandler) GetInvoiceMaterialsWithSerialNumbers(c *gin.Context) {

	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return
	}

	projectID := c.GetUint("projectID")

	data, err := handler.invoiceOutputOutOfProjectUsecase.GetInvoiceMaterialsWithSerialNumbers(uint(id), projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *invoiceOutputOutOfProjectHandler) Update(c *gin.Context) {
	updateData := dto.InvoiceOutputOutOfProject{}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	workerID := c.GetUint("workerID")
	updateData.Details.ReleasedWorkerID = workerID

	projectID := c.GetUint("projectID")
	updateData.Details.ProjectID = projectID

	existing, err := handler.invoiceOutputOutOfProjectUsecase.GetByID(updateData.Details.ID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Не удалось найти накладную: %v", err))
		return
	}
	if existing.ProjectID != projectID {
		response.ResponseError(c, "Доступ запрещен: накладная принадлежит другому проекту")
		return
	}

	data, err := handler.invoiceOutputOutOfProjectUsecase.Update(updateData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could perform the creation of Invoice: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *invoiceOutputOutOfProjectHandler) Confirmation(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	projectID := c.GetUint("projectID")
	invoiceOutputOutOfProject, err := handler.invoiceOutputOutOfProjectUsecase.GetByID(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Cannot find invoice Output by id %v: %v", id, err))
		return
	}
	if invoiceOutputOutOfProject.ProjectID != projectID {
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
	file.Filename = invoiceOutputOutOfProject.DeliveryCode + "." + fileExtension
	filePath := filepath.Join("./storage/import_excel/output/", file.Filename)

	err = c.SaveUploadedFile(file, filePath)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("cannot save file: %v", err))
		return
	}

	excelFilePath := filepath.Join("./storage/import_excel/output/", invoiceOutputOutOfProject.DeliveryCode+".xlsx")
	os.Remove(excelFilePath)

	err = handler.invoiceOutputOutOfProjectUsecase.Confirmation(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("cannot confirm invoice input with id %v: %v", id, err))
		return
	}

	response.ResponseSuccess(c, true)
}

func (handler *invoiceOutputOutOfProjectHandler) GetMaterialsForEdit(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)

	projectID := c.GetUint("projectID")

	result, err := handler.invoiceOutputOutOfProjectUsecase.GetMaterialsForEdit(uint(id), projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, result)
}

func (handler *invoiceOutputOutOfProjectHandler) UniqueNameOfProjects(c *gin.Context) {
	result, err := handler.invoiceOutputOutOfProjectUsecase.GetUniqueNameOfProjects(c.GetUint("projectID"))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, result)
}

func (handler *invoiceOutputOutOfProjectHandler) Report(c *gin.Context) {
	var filter dto.InvoiceOutputOutOfProjectReportFilter
	if err := c.ShouldBindJSON(&filter); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	filter.ProjectID = c.GetUint("projectID")
	filename, err := handler.invoiceOutputOutOfProjectUsecase.Report(filter)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	filePath := filepath.Join("./storage/import_excel/temp/", filename)
	tempfiles.Track(c, filePath)
	c.FileAttachment(filePath, filename)

}

func (handler *invoiceOutputOutOfProjectHandler) GetDocument(c *gin.Context) {
	deliveryCode := c.Param("deliveryCode")
	projectID := c.GetUint("projectID")

  extension, err := handler.invoiceOutputOutOfProjectUsecase.GetDocument(deliveryCode, projectID)
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
