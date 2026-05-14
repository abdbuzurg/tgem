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

type materialCostHandler struct {
	materialCostUsecase usecase.IMaterialCostUsecase
}

func NewMaterialCostHandler(materialCostUsecase usecase.IMaterialCostUsecase) IMaterialCostHandler {
	return &materialCostHandler{
		materialCostUsecase: materialCostUsecase,
	}
}

type IMaterialCostHandler interface {
	GetAll(c *gin.Context)
	GetPaginated(c *gin.Context)
	GetByID(c *gin.Context)
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	GetAllMaterialCostByMaterialID(c *gin.Context)
	ImportTemplate(c *gin.Context)
	Import(c *gin.Context)
	Export(c *gin.Context)
}

func (handler *materialCostHandler) GetAll(c *gin.Context) {
	data, err := handler.materialCostUsecase.GetAll()
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get MaterialCost data: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *materialCostHandler) GetPaginated(c *gin.Context) {
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

	filter := dto.MaterialCostSearchFilter{
		ProjectID:    c.GetUint("projectID"),
		MaterialName: c.DefaultQuery("materialName", ""),
	}

	data, err := handler.materialCostUsecase.GetPaginated(page, limit, filter)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the paginated data of MaterialCost: %v", err))
		return
	}

	dataCount, err := handler.materialCostUsecase.Count(filter)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the total amount of MaterialCost: %v", err))
		return
	}

	response.ResponsePaginatedData(c, data, dataCount)
}

func (handler *materialCostHandler) GetByID(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	data, err := handler.materialCostUsecase.GetByID(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the data with ID(%d): %v", id, err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *materialCostHandler) GetAllMaterialCostByMaterialID(c *gin.Context) {
	materialIDRaw := c.Param("materialID")
	materialID, err := strconv.ParseUint(materialIDRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	data, err := handler.materialCostUsecase.GetByMaterialID(uint(materialID))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server Error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *materialCostHandler) Create(c *gin.Context) {
	var createData model.MaterialCost
	if err := c.ShouldBindJSON(&createData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	data, err := handler.materialCostUsecase.Create(createData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could perform the creation of MaterialCost: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *materialCostHandler) Update(c *gin.Context) {
	var updateData model.MaterialCost
	if err := c.ShouldBindJSON(&updateData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	data, err := handler.materialCostUsecase.Update(updateData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the updation of MaterialCost: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *materialCostHandler) Delete(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	err = handler.materialCostUsecase.Delete(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the deletion of MaterialCost: %v", err))
		return
	}

	response.ResponseSuccess(c, "deleted")
}

func (handler *materialCostHandler) ImportTemplate(c *gin.Context) {
	projectID := c.GetUint("projectID")
	tmpFilePath, err := handler.materialCostUsecase.ImportTemplateFile(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	importTemplateFileName := "Шаблон импорта ценников для материалов.xlsx"
	tempfiles.Track(c, tmpFilePath)
	c.FileAttachment(tmpFilePath, importTemplateFileName)
}

func (handler *materialCostHandler) Import(c *gin.Context) {
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
	err = handler.materialCostUsecase.Import(projectID, importFilePath)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, true)
}

func (handler *materialCostHandler) Export(c *gin.Context) {
	projectID := c.GetUint("projectID")

	exportFileName, err := handler.materialCostUsecase.Export(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	exportFilePath := filepath.Join("./storage/import_excel/temp/", exportFileName)
	tempfiles.Track(c, exportFilePath)
	c.FileAttachment(exportFilePath, exportFileName)
}
