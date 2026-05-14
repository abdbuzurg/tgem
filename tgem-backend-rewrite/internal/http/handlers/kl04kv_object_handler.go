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

type kl04kvObjectHandler struct {
	kl04kvObjectUsecase usecase.IKL04KVObjectUsecase
}

func NewKl04KVObjectHandler(kl04kvObjectUsecase usecase.IKL04KVObjectUsecase) IKL04KVObjectHandler {
	return &kl04kvObjectHandler{
		kl04kvObjectUsecase: kl04kvObjectUsecase,
	}
}

type IKL04KVObjectHandler interface {
	GetPaginated(c *gin.Context)
	Create(c *gin.Context)
	Delete(c *gin.Context)
	Update(c *gin.Context)
	GetTemplateFile(c *gin.Context)
	Import(c *gin.Context)
	Export(c *gin.Context)
	GetObjectNamesForSearch(c *gin.Context)
}

func (handler *kl04kvObjectHandler) GetPaginated(c *gin.Context) {
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

	filter := dto.KL04KVObjectSearchParameters{
		ProjectID:          c.GetUint("projectID"),
		TeamID:             uint(teamID),
		SupervisorWorkerID: uint(supervisorWorkerID),
		TPObjectID:         uint(tpObjectID),
		ObjectName:         c.DefaultQuery("objectName", ""),
	}

	data, err := handler.kl04kvObjectUsecase.GetPaginated(page, limit, filter)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	dataCount, err := handler.kl04kvObjectUsecase.Count(filter)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponsePaginatedData(c, data, dataCount)
}

func (handler *kl04kvObjectHandler) Create(c *gin.Context) {
	var data dto.KL04KVObjectCreate
	if err := c.ShouldBindJSON(&data); err != nil {
		response.ResponseError(c, fmt.Sprintf("Неправильно тело запроса: %v", err))
		return
	}

	projectID := c.GetUint("projectID")
	data.BaseInfo.ProjectID = projectID

	_, err := handler.kl04kvObjectUsecase.Create(data)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, true)
}

func (handler *kl04kvObjectHandler) Delete(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Неправильный параметр в запросе: %v", err))
		return
	}

	projectID := c.GetUint("projectID")

	err = handler.kl04kvObjectUsecase.Delete(projectID, uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, true)
}

func (handler *kl04kvObjectHandler) Update(c *gin.Context) {
	var data dto.KL04KVObjectCreate
	if err := c.ShouldBindJSON(&data); err != nil {
		response.ResponseError(c, fmt.Sprintf("Неправильно тело запроса: %v", err))
		return
	}

	projectID := c.GetUint("projectID")
	data.BaseInfo.ProjectID = projectID

	_, err := handler.kl04kvObjectUsecase.Update(data)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, true)
}

func (handler *kl04kvObjectHandler) GetTemplateFile(c *gin.Context) {
	templateFilePath := filepath.Join("./internal/templates/Шаблон для импорта КЛ 04 КВ.xlsx")

	tmpFilePath, err := handler.kl04kvObjectUsecase.TemplateFile(templateFilePath, c.GetUint("projectID"))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	tempfiles.Track(c, tmpFilePath)
	c.FileAttachment(tmpFilePath, "Шаблон для импорта КЛ 04 КВ.xlsx")
}

func (handler *kl04kvObjectHandler) Import(c *gin.Context) {
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
	err = handler.kl04kvObjectUsecase.Import(projectID, importFilePath)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, true)
}

func (handler *kl04kvObjectHandler) Export(c *gin.Context) {
	projectID := c.GetUint("projectID")

	exportFileName, err := handler.kl04kvObjectUsecase.Export(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	exportFilePath := filepath.Join("./storage/import_excel/temp/", exportFileName)
	tempfiles.Track(c, exportFilePath)
	c.FileAttachment(exportFilePath, exportFileName)
}

func (handler *kl04kvObjectHandler) GetObjectNamesForSearch(c *gin.Context) {
	data, err := handler.kl04kvObjectUsecase.GetObjectNamesForSearch(c.GetUint("projectID"))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}
