package handlers

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/http/response"
	"backend-v2/internal/usecase"
	"backend-v2/pkg/tempfiles"
	"fmt"
	"net/url"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type operationHandler struct {
	operationUsecase usecase.IOperationUsecase
}

func NewOperationHandler(operationUsecase usecase.IOperationUsecase) IOperationHandler {
	return &operationHandler{
		operationUsecase: operationUsecase,
	}
}

type IOperationHandler interface {
	GetAll(c *gin.Context)
	GetPaginated(c *gin.Context)
	GetByID(c *gin.Context)
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	Import(c *gin.Context)
	GetTemplateFile(c *gin.Context)
}

func (handler *operationHandler) GetAll(c *gin.Context) {
	data, err := handler.operationUsecase.GetAll(c.GetUint("projectID"))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get Operation data: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *operationHandler) GetPaginated(c *gin.Context) {
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

	name := c.DefaultQuery("name", "")
	name, err = url.QueryUnescape(name)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Wrong query parameter provided for name: %v", err))
		return
	}

	code := c.DefaultQuery("code", "")
	code, err = url.QueryUnescape(code)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Wrong query parameter provided for code: %v", err))
		return
	}

	materialIDStr := c.DefaultQuery("materialID", "0")
	materialID, err := strconv.Atoi(materialIDStr)
	if err != nil || materialID < 0 {
		response.ResponseError(c, fmt.Sprintf("Wrong query parameter provided for materialID: %v", err))
		return
	}

	filter := dto.OperationSearchParameters{
		Name:       name,
		Code:       code,
		ProjectID:  c.GetUint("projectID"),
		MaterialID: uint(materialID),
	}

	data, err := handler.operationUsecase.GetPaginated(page, limit, filter)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the paginated data of Operation: %v", err))
		return
	}

	dataCount, err := handler.operationUsecase.Count(filter)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the total amount of Operation: %v", err))
		return
	}

	response.ResponsePaginatedData(c, data, dataCount)
}

func (handler *operationHandler) GetByID(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	data, err := handler.operationUsecase.GetByID(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the data with ID(%d): %v", id, err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *operationHandler) Create(c *gin.Context) {
	var createData dto.Operation
	if err := c.ShouldBindJSON(&createData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	createData.ProjectID = c.GetUint("projectID")

	operation, err := handler.operationUsecase.GetByName(createData.Name, createData.ProjectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Ошибка проверки имени услуги: %v", err))
		return
	}

	if operation.Name == createData.Name {
		response.ResponseError(c, fmt.Sprint("Услуга с таким именем уже существует"))
		return
	}

	data, err := handler.operationUsecase.Create(createData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could perform the creation of Operation: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *operationHandler) Update(c *gin.Context) {
	var updateData dto.Operation
	if err := c.ShouldBindJSON(&updateData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	updateData.ProjectID = c.GetUint("projectID")

	operation, err := handler.operationUsecase.GetByName(updateData.Name, updateData.ProjectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Ошибка проверки имени услуги: %v", err))
		return
	}

	if operation.Name == updateData.Name && operation.ID != updateData.ID {
		response.ResponseError(c, fmt.Sprint("Услуга с таким именем уже существует"))
		return
	}

	data, err := handler.operationUsecase.Update(updateData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the updation of Operation: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *operationHandler) Delete(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	err = handler.operationUsecase.Delete(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the deletion of Operation: %v", err))
		return
	}

	response.ResponseSuccess(c, "deleted")
}

func (handler *operationHandler) Import(c *gin.Context) {
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
	err = handler.operationUsecase.Import(projectID, importFilePath)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, true)
}

func (handler *operationHandler) GetTemplateFile(c *gin.Context) {
	templateFilePath := filepath.Join("./internal/templates/Шаблон для импорта Услуг.xlsx")

	tmpFilePath, err := handler.operationUsecase.TemplateFile(templateFilePath, c.GetUint("projectID"))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	tempfiles.Track(c, tmpFilePath)
	c.FileAttachment(tmpFilePath, "Шаблон для импорта КЛ 04 КВ.xlsx")
}
