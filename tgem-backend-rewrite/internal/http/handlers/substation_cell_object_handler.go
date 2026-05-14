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

type substationCellObjectHandler struct {
	substationCellObjectUsecase usecase.ISubstationCellObjectUsecase
}

func NewSubstationCellObjectHandler(
	substationCellObjectUsecase usecase.ISubstationCellObjectUsecase,
) ISubstationCellObjectHandler {
	return &substationCellObjectHandler{
		substationCellObjectUsecase: substationCellObjectUsecase,
	}
}

type ISubstationCellObjectHandler interface {
	GetPaginated(c *gin.Context)
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	GetTemplateFile(c *gin.Context)
	Import(c *gin.Context)
	Export(c *gin.Context)
	GetObjectNamesForSearch(c *gin.Context)
}

func (handler *substationCellObjectHandler) GetPaginated(c *gin.Context) {
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

  substationObjectIDStr := c.DefaultQuery("substationObjectID", "0")
	substationObjectID, err := strconv.Atoi(substationObjectIDStr)
	if err != nil || substationObjectID < 0 {
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса substationObjectID: %v", err))
		return
	}


	filter := dto.SubstationCellObjectSearchParameters{
		ProjectID:          c.GetUint("projectID"),
		TeamID:             uint(teamID),
		SupervisorWorkerID: uint(supervisorWorkerID),
		ObjectName:         c.DefaultQuery("objectName", ""),
    SubstationObjectID: uint(substationObjectID),
	}

  fmt.Println(filter)

	data, err := handler.substationCellObjectUsecase.GetPaginated(page, limit, filter)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	dataCount, err := handler.substationCellObjectUsecase.Count(filter)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponsePaginatedData(c, data, dataCount)

}

func (handler *substationCellObjectHandler) Create(c *gin.Context) {
	var createData dto.SubstationCellObjectCreate
	if err := c.ShouldBindJSON(&createData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return
	}

	projectID := c.GetUint("projectID")
	createData.BaseInfo.ProjectID = projectID

	data, err := handler.substationCellObjectUsecase.Create(createData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, data)

}

func (handler *substationCellObjectHandler) Update(c *gin.Context) {
	var updateData dto.SubstationCellObjectCreate
	if err := c.ShouldBindJSON(&updateData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return
	}

	projectID := c.GetUint("projectID")
	updateData.BaseInfo.ProjectID = projectID

	data, err := handler.substationCellObjectUsecase.Update(updateData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *substationCellObjectHandler) Delete(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Неправильный параметр в запросе: %v", err))
		return
	}

	projectID := c.GetUint("projectID")

	err = handler.substationCellObjectUsecase.Delete(uint(id), projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, true)
}

func (handler *substationCellObjectHandler) GetTemplateFile(c *gin.Context) {
	filepath := "./internal/templates/Шаблон для импорт Ячеек Подстанции.xlsx"

	tmpFilePath, err := handler.substationCellObjectUsecase.TemplateFile(filepath, c.GetUint("projectID"))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	tempfiles.Track(c, tmpFilePath)
	c.FileAttachment(tmpFilePath, "Шаблон для импорт Ячеек Подстанции.xlsx")
}

func (handler *substationCellObjectHandler) Import(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Файл не может быть сформирован, проверьте файл: %v", err))
		return
	}

	date := time.Now()
	filePath := "./storage/import_excel/temp/" + date.Format("2006-01-02 15-04-05") + file.Filename
	tempfiles.Track(c, filePath)
	err = c.SaveUploadedFile(file, filePath)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Файл не может быть сохранен на сервере: %v", err))
		return
	}

	projectID := c.GetUint("projectID")
	err = handler.substationCellObjectUsecase.Import(projectID, filePath)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, true)
}

func (handler *substationCellObjectHandler) Export(c *gin.Context) {
	projectID := c.GetUint("projectID")

	exportFileName, err := handler.substationCellObjectUsecase.Export(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	exportFilePath := filepath.Join("./storage/import_excel/temp/", exportFileName)
	tempfiles.Track(c, exportFilePath)
	c.FileAttachment(exportFilePath, exportFileName)
}

func (handler *substationCellObjectHandler) GetObjectNamesForSearch(c *gin.Context) {
	data, err := handler.substationCellObjectUsecase.GetObjectNamesForSearch(c.GetUint("projectID"))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}
