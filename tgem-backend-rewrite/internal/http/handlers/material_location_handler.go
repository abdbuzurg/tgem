package handlers

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/http/response"
	"backend-v2/internal/usecase"
	"backend-v2/model"
	"backend-v2/pkg/tempfiles"
	"fmt"
	"net/url"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
)

type materialLocationHandler struct {
	materialLocationUsecase usecase.IMaterialLocationUsecase
}

func NewMaterialLocationHandler(materialLocationUsecase usecase.IMaterialLocationUsecase) IMaterialLocationHandler {
	return &materialLocationHandler{
		materialLocationUsecase: materialLocationUsecase,
	}
}

type IMaterialLocationHandler interface {
	GetAll(c *gin.Context)
	GetPaginated(c *gin.Context)
	GetByID(c *gin.Context)
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	GetMaterialInLocation(c *gin.Context)
	GetMaterialCostsInLocation(c *gin.Context)
	GetMaterialAmountBasedOnCost(c *gin.Context)
	UniqueObjects(c *gin.Context)
	UniqueTeams(c *gin.Context)
	ReportBalance(c *gin.Context)
	ReportBalanceWriteOff(c *gin.Context)
	ReportBalanceOutOfProject(c *gin.Context)
	Live(c *gin.Context)
}

func (handler *materialLocationHandler) GetAll(c *gin.Context) {
	data, err := handler.materialLocationUsecase.GetAll()
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get MaterialLocation data: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *materialLocationHandler) GetPaginated(c *gin.Context) {
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

	materialCostIDStr := c.DefaultQuery("materialCostID", "")
	materialCostID := 0
	if materialCostIDStr != "" {
		materialCostID, err = strconv.Atoi(materialCostIDStr)
		if err != nil {
			response.ResponseError(c, fmt.Sprintf("Cannot decode materialCostID parameter: %v", err))
			return
		}
	}

	locationIDStr := c.DefaultQuery("locationID", "")
	locationID := 0
	if locationIDStr != "" {
		locationID, err = strconv.Atoi(materialCostIDStr)
		if err != nil {
			response.ResponseError(c, fmt.Sprintf("Cannot decode locationID parameter: %v", err))
			return
		}
	}

	locationType := c.DefaultQuery("locationType", "")
	locationType, err = url.QueryUnescape(locationType)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Wrong query parameter provided for locationType: %v", err))
		return
	}

	amountStr := c.DefaultQuery("amount", "")
	var amount float64
	if amountStr != "" {
		amount, err = strconv.ParseFloat(amountStr, 64)
		if err != nil {
			response.ResponseError(c, fmt.Sprintf("Cannot decode amount parameter: %v", err))
			return
		}
	} else {
		amount = 0
	}

	filter := model.MaterialLocation{
		MaterialCostID: uint(materialCostID),
		LocationID:     uint(locationID),
		LocationType:   locationType,
		Amount:         amount,
	}

	data, err := handler.materialLocationUsecase.GetPaginated(page, limit, filter)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the paginated data of MaterialLocation: %v", err))
		return
	}

	dataCount, err := handler.materialLocationUsecase.Count()
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the total amount of MaterialLocation: %v", err))
		return
	}

	response.ResponsePaginatedData(c, data, dataCount)
}

func (handler *materialLocationHandler) GetByID(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	data, err := handler.materialLocationUsecase.GetByID(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the data with ID(%d): %v", id, err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *materialLocationHandler) Create(c *gin.Context) {
	var createData model.MaterialLocation
	if err := c.ShouldBindJSON(&createData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	data, err := handler.materialLocationUsecase.Create(createData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could perform the creation of MaterialLocation: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *materialLocationHandler) Update(c *gin.Context) {
	var updateData model.MaterialLocation
	if err := c.ShouldBindJSON(&updateData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	data, err := handler.materialLocationUsecase.Update(updateData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the updation of MaterialLocation: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *materialLocationHandler) Delete(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	err = handler.materialLocationUsecase.Delete(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the deletion of MaterialLocation: %v", err))
		return
	}

	response.ResponseSuccess(c, "deleted")
}

func (handler *materialLocationHandler) GetMaterialInLocation(c *gin.Context) {

	locationType := c.Param("locationType")

	locationIDRaw := c.Param("locationID")
	locationID, err := strconv.ParseUint(locationIDRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid Query parameters: %v", err))
		return
	}

	data, err := handler.materialLocationUsecase.GetMaterialsInLocation(locationType, uint(locationID), c.GetUint("projectID"))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server Error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *materialLocationHandler) UniqueTeams(c *gin.Context) {

	data, err := handler.materialLocationUsecase.UniqueTeams(c.GetUint("projectID"))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server Error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)

}

func (handler *materialLocationHandler) UniqueObjects(c *gin.Context) {
	data, err := handler.materialLocationUsecase.UniqueObjects(c.GetUint("projectID"))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server Error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)

}

func (handler *materialLocationHandler) ReportBalance(c *gin.Context) {

	var data dto.ReportBalanceFilterRequest
	if err := c.ShouldBindJSON(&data); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid body request: %v", err))
		return
	}

	projectID := c.GetUint("projectID")

	fileName, err := handler.materialLocationUsecase.BalanceReport(projectID, data)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
		return
	}

	filePath := filepath.Join("./storage/import_excel/temp/", fileName)

	tempfiles.Track(c, filePath)
	c.FileAttachment(filePath, fileName)
}

func (handler *materialLocationHandler) Live(c *gin.Context) {
	locationType := c.DefaultQuery("locationType", "")

	locationIDStr := c.DefaultQuery("locationID", "0")
	locationID, err := strconv.Atoi(locationIDStr)
	if err != nil || locationID < 0 {
		response.ResponseError(c, fmt.Sprintf("Wrong query parameter provided for limit: %v", err))
		return
	}

	materialIDStr := c.DefaultQuery("materialID", "0")
	materialID, err := strconv.Atoi(materialIDStr)
	if err != nil || materialID < 0 {
		response.ResponseError(c, fmt.Sprintf("Wrong query parameter provided for limit: %v", err))
		return
	}

	searchParameters := dto.MaterialLocationLiveSearchParameters{
		ProjectID:    c.GetUint("projectID"),
		LocationType: locationType,
		LocationID:   uint(locationID),
		MaterialID:   uint(materialID),
	}

	data, err := handler.materialLocationUsecase.Live(searchParameters)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get MaterialLocation data: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *materialLocationHandler) ReportBalanceWriteOff(c *gin.Context) {
	var filterData dto.ReportWriteOffBalanceFilter
	if err := c.ShouldBindJSON(&filterData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	fileName, err := handler.materialLocationUsecase.BalanceReportWriteOff(c.GetUint("projectID"), filterData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
		return
	}

	filePath := filepath.Join("./storage/import_excel/temp/", fileName)

	tempfiles.Track(c, filePath)
	c.FileAttachment(filePath, fileName)
}

func (handler *materialLocationHandler) ReportBalanceOutOfProject(c *gin.Context) {
	fileName, err := handler.materialLocationUsecase.BalanceReportOutOfProject(c.GetUint("projectID"))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
		return
	}

	filePath := filepath.Join("./storage/import_excel/temp/", fileName)

	tempfiles.Track(c, filePath)
	c.FileAttachment(filePath, fileName)
}

func (handler *materialLocationHandler) GetMaterialCostsInLocation(c *gin.Context) {
	locationType := c.Param("locationType")

	locationIDStr := c.Param("locationID")
	locationID, err := strconv.Atoi(locationIDStr)
	if err != nil || locationID < 0 {
		response.ResponseError(c, fmt.Sprintf("Wrong query parameter provided for limit: %v", err))
		return
	}

	materialIDStr := c.Param("materialID")
	materialID, err := strconv.Atoi(materialIDStr)
	if err != nil || materialID < 0 {
		response.ResponseError(c, fmt.Sprintf("Wrong query parameter provided for limit: %v", err))
		return
	}

	data, err := handler.materialLocationUsecase.GetMaterialCostsInLocation(c.GetUint("projectID"), uint(materialID), uint(locationID), locationType)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *materialLocationHandler) GetMaterialAmountBasedOnCost(c *gin.Context) {
	locationType := c.Param("locationType")

	locationIDStr := c.Param("locationID")
	locationID, err := strconv.Atoi(locationIDStr)
	if err != nil || locationID < 0 {
		response.ResponseError(c, fmt.Sprintf("Wrong query parameter provided for limit: %v", err))
		return
	}

	materialCostIDStr := c.Param("materialCostID")
	materialCostID, err := strconv.Atoi(materialCostIDStr)
	if err != nil || materialCostID < 0 {
		response.ResponseError(c, fmt.Sprintf("Wrong query parameter provided for limit: %v", err))
		return
	}

	data, err := handler.materialLocationUsecase.GetMaterialAmountBasedOnCost(c.GetUint("projectID"), uint(materialCostID), uint(locationID), locationType)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}
