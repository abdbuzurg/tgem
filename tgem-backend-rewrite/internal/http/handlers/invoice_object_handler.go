package handlers

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/usecase"
	"backend-v2/internal/http/response"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

type invoiceObjectHandler struct {
	invoiceObjectUsecase usecase.IInvoiceObjectUsecase
}

func NewInvoiceObjectHandler(
	invoiceObjectUsecase usecase.IInvoiceObjectUsecase,
) IInvoiceObjectHandler {
	return &invoiceObjectHandler{
		invoiceObjectUsecase: invoiceObjectUsecase,
	}
}

type IInvoiceObjectHandler interface {
	GetInvoiceObjectDescriptiveDataByID(c *gin.Context)
	GetTeamsMaterials(c *gin.Context)
	GetSerialNumbersOfMaterial(c *gin.Context)
	Create(c *gin.Context)
	GetMaterialAmountInTeam(c *gin.Context)
	GetPaginated(c *gin.Context)
	GetTeamsFromObjectID(c *gin.Context)
	GetOperationsBasedOnMaterialsInTeamID(c *gin.Context)
}

func (handler *invoiceObjectHandler) GetInvoiceObjectDescriptiveDataByID(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	data, err := handler.invoiceObjectUsecase.GetInvoiceObjectDescriptiveDataByID(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server Error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)

}

func (handler *invoiceObjectHandler) GetPaginated(c *gin.Context) {
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

	projectID := c.GetUint("projectID")

	data, err := handler.invoiceObjectUsecase.GetPaginated(limit, page, projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server Error: %v", err))
		return
	}

	dataCount, err := handler.invoiceObjectUsecase.Count(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server Error: %v", err))
		return
	}

	response.ResponsePaginatedData(c, data, dataCount)
}

func (handler *invoiceObjectHandler) Delete(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	err = handler.invoiceObjectUsecase.Delete(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server Error: %v", err))
		return
	}

	response.ResponseSuccess(c, true)
}

func (handler *invoiceObjectHandler) GetTeamsMaterials(c *gin.Context) {
	teamIDRaw := c.Param("teamID")
	teamID, err := strconv.ParseUint(teamIDRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	projectID := c.GetUint("projectID")

	data, err := handler.invoiceObjectUsecase.GetTeamsMaterials(projectID, uint(teamID))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server Error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)

}

func (handler *invoiceObjectHandler) GetSerialNumbersOfMaterial(c *gin.Context) {
	materialIDRaw := c.Param("materialID")
	materialID, err := strconv.ParseUint(materialIDRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	teamIDRaw := c.Param("teamID")
	teamID, err := strconv.ParseUint(teamIDRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	projectID := c.GetUint("projectID")

	data, err := handler.invoiceObjectUsecase.GetSerialNumberOfMaterial(projectID, uint(materialID), uint(teamID))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server Error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)

}

func (handler *invoiceObjectHandler) Create(c *gin.Context) {

	var data dto.InvoiceObjectCreate
	if err := c.ShouldBindJSON(&data); err != nil {
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return
	}

	workerID := c.GetUint("workerID")
	data.Details.SupervisorWorkerID = workerID

	projectID := c.GetUint("projectID")
	data.Details.ProjectID = projectID

	_, err := handler.invoiceObjectUsecase.Create(data)
	if err != nil {

		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return

	}

	response.ResponseSuccess(c, true)
}

func (handler *invoiceObjectHandler) GetMaterialAmountInTeam(c *gin.Context) {

	projectID := c.GetUint("projectID")

	materialIDRaw := c.Param("materialID")
	materialID, err := strconv.ParseUint(materialIDRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	teamIDRaw := c.Param("teamID")
	teamID, err := strconv.ParseUint(teamIDRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	data, err := handler.invoiceObjectUsecase.GetAvailableMaterialAmount(projectID, uint(materialID), uint(teamID))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *invoiceObjectHandler) GetTeamsFromObjectID(c *gin.Context) {

	objectIDRaw := c.Param("objectID")
	objectID, err := strconv.ParseUint(objectIDRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	data, err := handler.invoiceObjectUsecase.GetTeamsFromObjectID(uint(objectID))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *invoiceObjectHandler) GetOperationsBasedOnMaterialsInTeamID(c *gin.Context) {

	teamIDRaw := c.Param("teamID")
	teamID, err := strconv.ParseUint(teamIDRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	data, err := handler.invoiceObjectUsecase.GetOperationsBasedOnMaterialsInTeamID(c.GetUint("projectID"), uint(teamID))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}
