package handlers

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/http/response"
	"backend-v2/internal/usecase"
	"backend-v2/model"
	"backend-v2/pkg/tempfiles"
	"fmt"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type workerHandler struct {
	workerUsecase usecase.IWorkerUsecase
}

func NewWorkerHandler(workerUsecase usecase.IWorkerUsecase) IWorkerHandler {
	return &workerHandler{
		workerUsecase: workerUsecase,
	}
}

type IWorkerHandler interface {
	GetAll(c *gin.Context)
	GetPaginated(c *gin.Context)
	GetByID(c *gin.Context)
	GetByJobTitleInProject(c *gin.Context)
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	GetTemplateFile(c *gin.Context)
	Import(c *gin.Context)
	GetWorkerInformationForSearch(c *gin.Context)
	Export(c *gin.Context)
}

func (handler *workerHandler) GetAll(c *gin.Context) {
	data, err := handler.workerUsecase.GetAll(c.GetUint("projectID"))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get Worker data: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *workerHandler) GetPaginated(c *gin.Context) {
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

	filter := dto.WorkerSearchParameters{
		ProjectID:         c.GetUint("projectID"),
		Name:              c.DefaultQuery("name", ""),
		JobTitleInProject: c.DefaultQuery("jobTitleInProject", ""),
		JobTitleInCompany: c.DefaultQuery("jobTitleInCompany", ""),
		MobileNumber:      c.DefaultQuery("mobileNumber", ""),
		CompanyWorkerID:   c.DefaultQuery("companyWorkerID", ""),
	}

	data, err := handler.workerUsecase.GetPaginated(page, limit, filter)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the paginated data of Worker: %v", err))
		return
	}

	dataCount, err := handler.workerUsecase.Count(filter)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the total amount of Worker: %v", err))
		return
	}

	response.ResponsePaginatedData(c, data, dataCount)
}

func (handler *workerHandler) GetByID(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	data, err := handler.workerUsecase.GetByID(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the data with ID(%d): %v", id, err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *workerHandler) GetByJobTitleInProject(c *gin.Context) {
	jobTitleInProject := c.Param("jobTitleInProject")

	data, err := handler.workerUsecase.GetByJobTitleInProject(jobTitleInProject, c.GetUint("projectID"))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Cannot get workers by the job title: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *workerHandler) Create(c *gin.Context) {
	var createData model.Worker
	if err := c.ShouldBindJSON(&createData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	createData.ProjectID = c.GetUint("projectID")

	data, err := handler.workerUsecase.Create(createData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could perform the creation of Worker: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *workerHandler) Update(c *gin.Context) {
	var updateData model.Worker
	if err := c.ShouldBindJSON(&updateData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	updateData.ProjectID = c.GetUint("projectID")

	data, err := handler.workerUsecase.Update(updateData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the updation of Worker: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *workerHandler) Delete(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	err = handler.workerUsecase.Delete(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the deletion of Worker: %v", err))
		return
	}

	response.ResponseSuccess(c, "deleted")
}

func (handler *workerHandler) GetTemplateFile(c *gin.Context) {
	filepath := "./internal/templates/Шаблон для импорта Рабочего Персонала.xlsx"
	c.FileAttachment(filepath, "Шаблон для импорта Рабочего Персонала.xlsx")
}

func (handler *workerHandler) Import(c *gin.Context) {
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

	err = handler.workerUsecase.Import(filePath, projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, true)
}

func (handler *workerHandler) GetWorkerInformationForSearch(c *gin.Context) {
	data, err := handler.workerUsecase.GetWorkerInformationForSearch(c.GetUint("projectID"))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the updation of Worker: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *workerHandler) Export(c *gin.Context) {
	projectID := c.GetUint("projectID")

	exportFileName, err := handler.workerUsecase.Export(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	exportFilePath := filepath.Join("./storage/import_excel/temp/", exportFileName)
	tempfiles.Track(c, exportFilePath)
	c.FileAttachment(exportFilePath, exportFileName)
}
