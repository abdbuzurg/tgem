package handlers

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/http/response"
	"backend-v2/internal/usecase"
	"backend-v2/pkg/tempfiles"
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
)

type invoiceCorrectionHandler struct {
	invoiceCorrectionUsecase usecase.IInvoiceCorrectionUsecase
}

func NewInvoiceCorrectionHandler(
	invoiceCorrectionUsecase usecase.IInvoiceCorrectionUsecase,
) IInvoiceCorrectionHandler {
	return &invoiceCorrectionHandler{
		invoiceCorrectionUsecase: invoiceCorrectionUsecase,
	}
}

type IInvoiceCorrectionHandler interface {
	GetAll(c *gin.Context)
	GetTotalMaterialInTeamByTeamNumber(c *gin.Context)
	GetInvoiceMaterialsByInvoiceObjectID(c *gin.Context)
	GetSerialNumbersOfMaterial(c *gin.Context)
	Create(c *gin.Context)
	UniqueObject(c *gin.Context)
	UniqueTeam(c *gin.Context)
	Report(c *gin.Context)
	GetPaginated(c *gin.Context)
	GetOperationsByInvoiceObjectID(c *gin.Context)
	GetParametersForSearch(c *gin.Context)
}

func (handler *invoiceCorrectionHandler) GetPaginated(c *gin.Context) {
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

	teamIDStr := c.DefaultQuery("teamID", "0")
	teamID, err := strconv.Atoi(teamIDStr)
	if err != nil || teamID < 0 {
		response.ResponseError(c, fmt.Sprintf("Wrong query paramter provided for teamID: %v", err))
		return
	}

	objectIDStr := c.DefaultQuery("objectID", "0")
	objectID, err := strconv.Atoi(objectIDStr)
	if err != nil || objectID < 0 {
		response.ResponseError(c, fmt.Sprintf("Wrong query paramter provided for objectID: %v", err))
	}

	filter := dto.InvoiceCorrectionPaginatedParamters{
		ProjectID: c.GetUint("projectID"),
		TeamID:    uint(teamID),
		ObjectID:  uint(objectID),
	}

	data, err := handler.invoiceCorrectionUsecase.GetPaginated(page, limit, filter)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	dataCount, err := handler.invoiceCorrectionUsecase.Count(filter)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponsePaginatedData(c, data, dataCount)
}

func (handler *invoiceCorrectionHandler) GetAll(c *gin.Context) {

	projectID := c.GetUint("projectID")

	data, err := handler.invoiceCorrectionUsecase.GetAll(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *invoiceCorrectionHandler) GetTotalMaterialInTeamByTeamNumber(c *gin.Context) {

	materialIDRaw := c.Param("materialID")
	materialID, err := strconv.ParseUint(materialIDRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	teamNumber := c.Param("teamNumber")
	projectID := c.GetUint("projectID")

	data, err := handler.invoiceCorrectionUsecase.GetTotalAmounInLocationByTeamName(projectID, uint(materialID), teamNumber)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *invoiceCorrectionHandler) GetInvoiceMaterialsByInvoiceObjectID(c *gin.Context) {

	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	data, err := handler.invoiceCorrectionUsecase.GetInvoiceMaterialsByInvoiceObjectID(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)

}

func (handler *invoiceCorrectionHandler) GetSerialNumbersOfMaterial(c *gin.Context) {
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

	data, err := handler.invoiceCorrectionUsecase.GetSerialNumberOfMaterialInTeam(projectID, uint(materialID), uint(teamID))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server Error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)

}

func (handler *invoiceCorrectionHandler) Create(c *gin.Context) {
	var createData dto.InvoiceCorrectionCreate
	if err := c.ShouldBindJSON(&createData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return
	}

	createData.Details.OperatorWorkerID = c.GetUint("workerID")

	data, err := handler.invoiceCorrectionUsecase.Create(createData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *invoiceCorrectionHandler) UniqueObject(c *gin.Context) {
	projectID := c.GetUint("projectID")
	data, err := handler.invoiceCorrectionUsecase.UniqueObject(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server Error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *invoiceCorrectionHandler) UniqueTeam(c *gin.Context) {
	projectID := c.GetUint("projectID")
	data, err := handler.invoiceCorrectionUsecase.UniqueTeam(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server Error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *invoiceCorrectionHandler) Report(c *gin.Context) {
	var filter dto.InvoiceCorrectionReportFilter
	if err := c.ShouldBindJSON(&filter); err != nil {
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return
	}

	filter.ProjectID = c.GetUint("projectID")

	reportFileName, err := handler.invoiceCorrectionUsecase.Report(filter)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server Error: %v", err))
		return
	}

	reportFilePath := filepath.Join("./storage/import_excel/temp/", reportFileName)
	tempfiles.Track(c, reportFilePath)
	c.FileAttachment(reportFilePath, reportFileName)
}

func (handler *invoiceCorrectionHandler) GetOperationsByInvoiceObjectID(c *gin.Context) {

	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	data, err := handler.invoiceCorrectionUsecase.GetOperationsByInvoiceObjectID(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *invoiceCorrectionHandler) GetParametersForSearch(c *gin.Context) {
	data, err := handler.invoiceCorrectionUsecase.GetParametersForSearch(c.GetUint("projectID"))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}
