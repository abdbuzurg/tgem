package handlers

import (
	"backend-v2/internal/apperr"
	"backend-v2/internal/dto"
	"backend-v2/internal/usecase"
	"backend-v2/model"
	"backend-v2/internal/http/response"
	"fmt"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"
)

type objectHandler struct {
	objectUsecase usecase.IObjectUsecase
}

func NewObjectHandler(objectUsecase usecase.IObjectUsecase) IObjectHandler {
	return &objectHandler{
		objectUsecase: objectUsecase,
	}
}

type IObjectHandler interface {
	GetAll(c *gin.Context)
	GetPaginated(c *gin.Context)
	GetByID(c *gin.Context)
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	GetTeamsByObject(c *gin.Context)
}

func (handler *objectHandler) GetAll(c *gin.Context) {
	projectID := c.GetUint("projectID")

	data, err := handler.objectUsecase.GetAll(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get Object data: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *objectHandler) GetPaginated(c *gin.Context) {
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

	objectDetailedIDStr := c.DefaultQuery("objectDetailedID", "")
	objectDetailedID := 0
	if objectDetailedIDStr != "" {
		objectDetailedID, err = strconv.Atoi(objectDetailedIDStr)
		if err != nil {
			response.ResponseError(c, fmt.Sprintf("Cannot decode objectDetailedID parameter: %v", err))
			return
		}
	}

	objectType := c.DefaultQuery("objectType", "")
	objectType, err = url.QueryUnescape(objectType)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Wrong query parameter provided for objectType: %v", err))
		return
	}

	name := c.DefaultQuery("name", "")
	name, err = url.QueryUnescape(name)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Wrong query parameter provided for objectType: %v", err))
		return
	}

	status := c.DefaultQuery("status", "")
	status, err = url.QueryUnescape(status)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Wrong query parameter provided for status: %v", err))
		return
	}

	filter := model.Object{
		ObjectDetailedID: uint(objectDetailedID),
		Type:             objectType,
		Name:             name,
		Status:           status,
	}

	data, err := handler.objectUsecase.GetPaginated(page, limit, filter)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the paginated data of Object: %v", err))
		return
	}

	dataCount, err := handler.objectUsecase.Count()
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the total amount of Object: %v", err))
		return
	}

	response.ResponsePaginatedData(c, data, dataCount)
}

func (handler *objectHandler) GetByID(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		apperr.WriteError(c, apperr.InvalidInput("Некорректный идентификатор", err))
		return
	}

	data, err := handler.objectUsecase.GetByID(uint(id))
	if err != nil {
		apperr.WriteError(c, err)
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *objectHandler) Create(c *gin.Context) {
	var createData dto.ObjectCreate
	if err := c.ShouldBindJSON(&createData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	data, err := handler.objectUsecase.Create(createData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could perform the creation of Object: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *objectHandler) Update(c *gin.Context) {
	var updateData model.Object
	if err := c.ShouldBindJSON(&updateData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	data, err := handler.objectUsecase.Update(updateData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the updation of Object: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *objectHandler) Delete(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	err = handler.objectUsecase.Delete(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the deletion of Object: %v", err))
		return
	}

	response.ResponseSuccess(c, "deleted")
}

func (handler *objectHandler) GetTeamsByObject(c *gin.Context) {
	objectIDRaw := c.Param("objectID")
	objectID, err := strconv.ParseUint(objectIDRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	data, err := handler.objectUsecase.GetTeamsByObjectID(uint(objectID))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the deletion of Object: %v", err))
		return
	}

  response.ResponseSuccess(c, data)
}
