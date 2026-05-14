package handlers

import (
	"backend-v2/internal/usecase"
	"backend-v2/model"
	"backend-v2/internal/http/response"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

type projectHandler struct {
	projectUsecase usecase.IProjectUsecase
}

func NewProjectHandler(projectUsecase usecase.IProjectUsecase) IProjectHandler {
	return &projectHandler{
		projectUsecase: projectUsecase,
	}
}

type IProjectHandler interface {
	GetAll(c *gin.Context)
	GetPaginated(c *gin.Context)
	GetByID(c *gin.Context)
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	GetProjectName(c *gin.Context)
}

func (handler *projectHandler) GetAll(c *gin.Context) {
	data, err := handler.projectUsecase.GetAll()
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get Project data: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *projectHandler) GetPaginated(c *gin.Context) {
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

	data, err := handler.projectUsecase.GetPaginated(page, limit)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the paginated data of Project: %v", err))
		return
	}

	dataCount, err := handler.projectUsecase.Count()
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the total amount of Project: %v", err))
		return
	}

	response.ResponsePaginatedData(c, data, dataCount)
}

func (handler *projectHandler) GetByID(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	data, err := handler.projectUsecase.GetByID(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the data with ID(%d): %v", id, err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *projectHandler) Create(c *gin.Context) {
	var createData model.Project
	if err := c.ShouldBindJSON(&createData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	data, err := handler.projectUsecase.Create(createData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could perform the creation of Project: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *projectHandler) Update(c *gin.Context) {
	var updateData model.Project
	if err := c.ShouldBindJSON(&updateData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	data, err := handler.projectUsecase.Update(updateData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the updation of Project: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *projectHandler) Delete(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	err = handler.projectUsecase.Delete(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the deletion of Project: %v", err))
		return
	}

	response.ResponseSuccess(c, "deleted")
}

func (handler *projectHandler) GetProjectName(c *gin.Context) {
	projectName, err := handler.projectUsecase.GetProjectName(c.GetUint("projectID"))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the deletion of Project: %v", err))
		return
	}

  response.ResponseSuccess(c, projectName)
}
