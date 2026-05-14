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

type teamHandler struct {
	teamUsecase usecase.ITeamUsecase
}

func NewTeamHandler(teamUsecase usecase.ITeamUsecase) ITeamHandler {
	return &teamHandler{
		teamUsecase: teamUsecase,
	}
}

type ITeamHandler interface {
	GetAll(c *gin.Context)
	GetPaginated(c *gin.Context)
	GetByID(c *gin.Context)
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	GetTemplateFile(c *gin.Context)
	Import(c *gin.Context)
	GetAllForSelect(c *gin.Context)
	GetAllUniqueTeamNumbers(c *gin.Context)
	GetAllUniqueMobileNumber(c *gin.Context)
	GetAllUniqueCompanies(c *gin.Context)
	Export(c *gin.Context)
}

func (handler *teamHandler) GetAll(c *gin.Context) {
	projectID := c.GetUint("projectID")

	data, err := handler.teamUsecase.GetAll(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get Team data: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *teamHandler) GetPaginated(c *gin.Context) {
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

	teamLeaderIDStr := c.DefaultQuery("leaderID", "")
	teamLeaderID, err := strconv.Atoi(teamLeaderIDStr)
	if err != nil || teamLeaderID < 0 {
		response.ResponseError(c, fmt.Sprintf("Wrong query parameter provided for limit: %v", err))
		return
	}

	searchParameters := dto.TeamSearchParameters{
		ProjectID:    c.GetUint("projectID"),
		Number:       c.DefaultQuery("number", ""),
		MobileNumber: c.DefaultQuery("mobileNumber", ""),
		Company:      c.DefaultQuery("company", ""),
		TeamLeaderID: uint(teamLeaderID),
	}

	data, err := handler.teamUsecase.GetPaginated(page, limit, searchParameters)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the paginated data of Team: %v", err))
		return
	}

	dataCount, err := handler.teamUsecase.Count(searchParameters)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the total amount of Team: %v", err))
		return
	}

	response.ResponsePaginatedData(c, data, dataCount)
}

func (handler *teamHandler) GetByID(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	data, err := handler.teamUsecase.GetByID(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the data with ID(%d): %v", id, err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *teamHandler) Create(c *gin.Context) {

	var createData dto.TeamMutation
	if err := c.ShouldBindJSON(&createData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	projectID := c.GetUint("projectID")
	createData.ProjectID = projectID

	exist, err := handler.teamUsecase.DoesTeamNumberAlreadyExistForCreate(createData.Number, createData.ProjectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform team number check-up: %v", err))
		return
	}

	if exist {
		response.ResponseError(c, fmt.Sprintf("Бригада с таким номером уже существует"))
		return
	}

	data, err := handler.teamUsecase.Create(createData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could perform the creation of Team: %v", err))
		return
	}

	response.ResponseSuccess(c, data)

}

func (handler *teamHandler) Update(c *gin.Context) {
	var updateData dto.TeamMutation
	if err := c.ShouldBindJSON(&updateData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	projectID := c.GetUint("projectID")
	updateData.ProjectID = projectID

	exist, err := handler.teamUsecase.DoesTeamNumberAlreadyExistForUpdate(updateData.Number, updateData.ID, projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform team number check-up: %v", err))
		return
	}

	if exist {
		response.ResponseError(c, fmt.Sprintf("Бригада с таким номером уже существует"))
		return
	}

	data, err := handler.teamUsecase.Update(updateData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the updation of Team: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *teamHandler) Delete(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	err = handler.teamUsecase.Delete(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the deletion of Team: %v", err))
		return
	}

	response.ResponseSuccess(c, "deleted")
}

func (handler *teamHandler) GetTemplateFile(c *gin.Context) {
	filepath := "./internal/templates/Шаблон для импорта Бригады.xlsx"
	projectID := c.GetUint("projectID")
	tmpFilePath, err := handler.teamUsecase.TemplateFile(projectID, filepath)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	tempfiles.Track(c, tmpFilePath)
	c.FileAttachment(tmpFilePath, "Шаблон для импорта Бригады.xlsx")
}

func (handler *teamHandler) Import(c *gin.Context) {
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

	err = handler.teamUsecase.Import(projectID, filePath)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, true)
}

func (handler *teamHandler) GetAllForSelect(c *gin.Context) {
	projectID := c.GetUint("projectID")

	data, err := handler.teamUsecase.GetAllForSelect(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get Team data: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *teamHandler) GetAllUniqueTeamNumbers(c *gin.Context) {
	data, err := handler.teamUsecase.GetAllUniqueTeamNumbers(c.GetUint("projectID"))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера при получении номера бригад: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *teamHandler) GetAllUniqueMobileNumber(c *gin.Context) {
	data, err := handler.teamUsecase.GetAllUniqueMobileNumber(c.GetUint("projectID"))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера при получении телефона бригад: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *teamHandler) GetAllUniqueCompanies(c *gin.Context) {
	data, err := handler.teamUsecase.GetAllUniqueCompanies(c.GetUint("projectID"))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера при получении компании бригад: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *teamHandler) Export(c *gin.Context) {
	projectID := c.GetUint("projectID")

	exportFileName, err := handler.teamUsecase.Export(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	exportFilePath := filepath.Join("./storage/import_excel/temp/", exportFileName)
	tempfiles.Track(c, exportFilePath)
	c.FileAttachment(exportFilePath, exportFileName)
}
