package handlers

import (
	"backend-v2/internal/http/response"
	"backend-v2/internal/usecase"
	"backend-v2/model"
	"backend-v2/pkg/tempfiles"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type materialHandler struct {
	materialUsecase usecase.IMaterialUsecase
}

func NewMaterialHandler(materialUsecase usecase.IMaterialUsecase) IMaterialHandler {
	return &materialHandler{
		materialUsecase: materialUsecase,
	}
}

type IMaterialHandler interface {
	GetAll(c *gin.Context)
	GetPaginated(c *gin.Context)
	GetByID(c *gin.Context)
	GetTemplateFile(c *gin.Context)
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	Import(c *gin.Context)
	Export(c *gin.Context)
}

func (handler *materialHandler) GetAll(c *gin.Context) {
	projectID := c.GetUint("projectID")
	data, err := handler.materialUsecase.GetAll(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get Material data: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *materialHandler) GetPaginated(c *gin.Context) {
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

	category := c.DefaultQuery("category", "")
	category, err = url.QueryUnescape(category)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Wrong query parameter provided for category: %v", err))
		return
	}

	code := c.DefaultQuery("code", "")
	code, err = url.QueryUnescape(code)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Wrong query parameter provided for code: %v", err))
		return
	}

	name := c.DefaultQuery("name", "")
	name, err = url.QueryUnescape(name)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Wrong query parameter provided for name: %v", err))
		return
	}

	unit := c.DefaultQuery("unit", "")
	unit, err = url.QueryUnescape(unit)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Wrong query parameter provided for unit: %v", err))
		return
	}

	projectID := c.GetUint("projectID")

	filter := model.Material{
		Category:  category,
		Code:      code,
		Name:      name,
		Unit:      unit,
		ProjectID: projectID,
	}

	data, err := handler.materialUsecase.GetPaginated(page, limit, filter)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the paginated data of Material: %v", err))
		return
	}

	dataCount, err := handler.materialUsecase.Count(filter)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the total amount of Materials: %v", err))
		return
	}

	response.ResponsePaginatedData(c, data, dataCount)
}

func (handler *materialHandler) GetByID(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	data, err := handler.materialUsecase.GetByID(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the data with ID(%d): %v", id, err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *materialHandler) Create(c *gin.Context) {
	var createData model.Material
	if err := c.ShouldBindJSON(&createData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	projectID := c.GetUint("projectID")
	createData.ProjectID = projectID

	data, err := handler.materialUsecase.Create(createData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could perform the creation of Material: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *materialHandler) Update(c *gin.Context) {
	var updateData model.Material
	if err := c.ShouldBindJSON(&updateData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	projectID := c.GetUint("projectID")
	updateData.ProjectID = projectID

	data, err := handler.materialUsecase.Update(updateData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the updation of Material: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *materialHandler) Delete(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	err = handler.materialUsecase.Delete(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the deletion of Material: %v", err))
		return
	}

	response.ResponseSuccess(c, "deleted")
}

func (handler *materialHandler) GetTemplateFile(c *gin.Context) {
	filePath := filepath.Join("./internal/templates/", "Шаблон для импорта Материалов.xlsx")
	if _, err := os.Stat(filePath); err == nil {
		c.FileAttachment(filePath, "Шаблон для импорта Материалов.xlsx")
	} else if errors.Is(err, os.ErrNotExist) {
		response.ResponseError(c, fmt.Sprintf("Файл Шаблон импорта материала не существует или был удалён: %v", err))
	} else {
		response.ResponseError(c, fmt.Sprintf("Неизвестная ошибка при проверке существования файла: %v", err))
	}
}

func (handler *materialHandler) Import(c *gin.Context) {
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
	err = handler.materialUsecase.Import(projectID, importFilePath)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, true)
}

func (handler *materialHandler) Export(c *gin.Context) {
	projectID := c.GetUint("projectID")

	exportFileName, err := handler.materialUsecase.Export(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	exportFilePath := filepath.Join("./storage/import_excel/temp/", exportFileName)
	tempfiles.Track(c, exportFilePath)
	c.FileAttachment(exportFilePath, exportFileName)
}
