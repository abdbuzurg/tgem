package handlers

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/http/response"
	"backend-v2/internal/usecase"
	"backend-v2/pkg/tempfiles"
	"fmt"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type mjdObjectHandler struct {
	mjdObjectUsecase usecase.IMJDObjectUsecase
}

func NewMJDObjectHandler(mjdObjectUsecase usecase.IMJDObjectUsecase) IMJDObjectHandler {
	return &mjdObjectHandler{
		mjdObjectUsecase: mjdObjectUsecase,
	}
}

type IMJDObjectHandler interface {
	GetPaginated(c *gin.Context)
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	GetTemplateFile(c *gin.Context)
	Import(c *gin.Context)
	Export(c *gin.Context)
	GetObjectNamesForSearch(c *gin.Context)
}

func (handler *mjdObjectHandler) GetPaginated(c *gin.Context) {

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

	teamIDStr := c.DefaultQuery("teamID", "0")
	teamID, err := strconv.Atoi(teamIDStr)
	if err != nil || teamID < 0 {
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса teamID: %v", err))
		return
	}

	supervisorWorkerIDStr := c.DefaultQuery("supervisorWorkerID", "0")
	supervisorWorkerID, err := strconv.Atoi(supervisorWorkerIDStr)
	if err != nil || supervisorWorkerID < 0 {
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса supervisorWorkerID: %v", err))
		return
	}

	tpObjectIDStr := c.DefaultQuery("tpObjectID", "0")
	tpObjectID, err := strconv.Atoi(tpObjectIDStr)
	if err != nil || tpObjectID < 0 {
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса tpObjectID: %v", err))
		return
	}

	filter := dto.MJDObjectSearchParameters{
		ProjectID:          c.GetUint("projectID"),
		TeamID:             uint(teamID),
		SupervisorWorkerID: uint(supervisorWorkerID),
		TPObjectID:         uint(tpObjectID),
		ObjectName:         c.DefaultQuery("objectName", ""),
	}

	data, err := handler.mjdObjectUsecase.GetPaginated(page, limit, filter)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	dataCount, err := handler.mjdObjectUsecase.Count(filter)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponsePaginatedData(c, data, dataCount)
}

func (handler *mjdObjectHandler) Create(c *gin.Context) {
	var createData dto.MJDObjectCreate
	if err := c.ShouldBindJSON(&createData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return
	}

	projectID := c.GetUint("projectID")
	createData.BaseInfo.ProjectID = projectID

	data, err := handler.mjdObjectUsecase.Create(createData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *mjdObjectHandler) Update(c *gin.Context) {
	var updateData dto.MJDObjectCreate
	if err := c.ShouldBindJSON(&updateData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return
	}

	projectID := c.GetUint("projectID")
	updateData.BaseInfo.ProjectID = projectID

	data, err := handler.mjdObjectUsecase.Update(updateData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *mjdObjectHandler) Delete(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Неправильный параметр в запросе: %v", err))
		return
	}

	projectID := c.GetUint("projectID")

	err = handler.mjdObjectUsecase.Delete(uint(id), projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, true)
}

func (handler *mjdObjectHandler) GetTemplateFile(c *gin.Context) {
	templateFilePath := filepath.Join("./internal/templates/", "Шаблон для импорта МЖД.xlsx")

	tmpFilePath, err := handler.mjdObjectUsecase.TemplateFile(templateFilePath, c.GetUint("projectID"))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	tempfiles.Track(c, tmpFilePath)
	c.FileAttachment(tmpFilePath, "Шаблон для импорта Подстанции.xlsx")
}

func (handler *mjdObjectHandler) Import(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Файл не может быть сформирован, проверьте файл: %v", err))
		return
	}

	date := time.Now()
	importFilePath := filepath.Join("./storage/import_excel/temp/", date.Format("2006-01-02 15-04-05")+file.Filename)
	tempfiles.Track(c, importFilePath)
	err = c.SaveUploadedFile(file, importFilePath)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Файл не может быть сохранен на сервере: %v", err))
		return
	}

	projectID := c.GetUint("projectID")
	err = handler.mjdObjectUsecase.Import(projectID, importFilePath)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, true)
}

func (handler *mjdObjectHandler) Export(c *gin.Context) {
	projectID := c.GetUint("projectID")

	exportFileName, err := handler.mjdObjectUsecase.Export(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	exportFilePath := filepath.Join("./storage/import_excel/temp/", exportFileName)
	tempfiles.Track(c, exportFilePath)
	c.FileAttachment(exportFilePath, exportFileName)
}

func (handler *mjdObjectHandler) GetObjectNamesForSearch(c *gin.Context) {
	data, err := handler.mjdObjectUsecase.GetObjectNamesForSearch(c.GetUint("projectID"))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}
