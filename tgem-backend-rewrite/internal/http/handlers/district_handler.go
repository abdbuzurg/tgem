package handlers

import (
	"backend-v2/internal/usecase"
	"backend-v2/model"
	"backend-v2/internal/http/response"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

type districtHandler struct {
	districtUsecase   usecase.IDistrictUsecase
	userActionUsecase usecase.IUserActionUsecase
}

func NewDistrictHandler(
	districtUsecase usecase.IDistrictUsecase,
	userActionUsecase usecase.IUserActionUsecase,
) IDistictHandler {
	return &districtHandler{
		districtUsecase:   districtUsecase,
		userActionUsecase: userActionUsecase,
	}
}

type IDistictHandler interface {
	GetAll(c *gin.Context)
	GetPaginated(c *gin.Context)
	GetByID(c *gin.Context)
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
}

func (handler *districtHandler) GetAll(c *gin.Context) {

	projectID := c.GetUint("projectID")

	data, err := handler.districtUsecase.GetAll(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return

	}

	response.ResponseSuccess(c, data)
}

func (handler *districtHandler) GetPaginated(c *gin.Context) {

	projectID := c.GetUint("projectID")

	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {

		response.ResponseError(c, fmt.Sprintf("Сервер получил неправильные данные: %v", err))
		return

	}

	limitStr := c.DefaultQuery("limit", "25")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {

		response.ResponseError(c, fmt.Sprintf("Сервер получил неправильные данные: %v", err))
		return

	}

	data, err := handler.districtUsecase.GetPaginated(page, limit, projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	dataCount, err := handler.districtUsecase.Count(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponsePaginatedData(c, data, dataCount)
}

func (handler *districtHandler) GetByID(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Сервер получил неправильные данные: %v", err))
		return
	}

	data, err := handler.districtUsecase.GetByID(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *districtHandler) Create(c *gin.Context) {

	projectID := c.GetUint("projectID")

	var createData model.District
	if err := c.ShouldBindJSON(&createData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return
	}

	createData.ProjectID = projectID

	data, err := handler.districtUsecase.Create(createData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутрення ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *districtHandler) Update(c *gin.Context) {

	projectID := c.GetUint("projectID")

	var updateData model.District
	if err := c.ShouldBindJSON(&updateData); err != nil {

		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return
	}

	updateData.ProjectID = projectID

	data, err := handler.districtUsecase.Update(updateData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *districtHandler) Delete(c *gin.Context) {

	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Неправильный параметер запроса: %v", err))
		return
	}

	err = handler.districtUsecase.Delete(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, "deleted")
}
