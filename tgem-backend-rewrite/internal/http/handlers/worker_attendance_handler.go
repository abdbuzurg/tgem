package handlers

import (
	"backend-v2/internal/http/response"
	"backend-v2/internal/usecase"
	"backend-v2/pkg/tempfiles"
	"fmt"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

type workerAttendanceHandler struct {
	workerAttendanceUsecase usecase.IWorkerAttendanceUsecase
}

type IWorkerAttendanceHandler interface {
	Import(c *gin.Context)
	GetPaginated(c *gin.Context)
	Progress(c *gin.Context)
  Analysis(c *gin.Context)
}

func NewWorkerAttendanceHandler(workerAttendanceUsecase usecase.IWorkerAttendanceUsecase) IWorkerAttendanceHandler {
	return &workerAttendanceHandler{
		workerAttendanceUsecase: workerAttendanceUsecase,
	}
}

func (handler *workerAttendanceHandler) Import(c *gin.Context) {
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
	err = handler.workerAttendanceUsecase.Import(projectID, importFilePath)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, true)
}

func (handler *workerAttendanceHandler) GetPaginated(c *gin.Context) {
	projectID := c.GetUint("projectID")

	data, err := handler.workerAttendanceUsecase.GetPaginated(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Ошибка при обработке данных: %v", err))
		return
	}

	dataCount, err := handler.workerAttendanceUsecase.Count(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the total amount of Materials: %v", err))
		return
	}

	response.ResponsePaginatedData(c, data, dataCount)
}

func (handler *workerAttendanceHandler) Progress(c *gin.Context) {
	filepath := "./internal/templates/Прогресс.xlsx"
	c.FileAttachment(filepath, "Прогресс.xlsx")
}

func (handler *workerAttendanceHandler) Analysis(c *gin.Context) {
	filepath := "./internal/templates/Анализ.xlsx"
	c.FileAttachment(filepath, "Анализ.xlsx")
}
