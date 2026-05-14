package handlers

import (
	"backend-v2/internal/usecase"
	"backend-v2/internal/http/response"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

type statisticsHandler struct {
	statUsecase usecase.IStatisticsUsecase
}

type IStatisticsHandler interface {
	InvoiceCountStat(c *gin.Context)
	InvoiceInputCreatorStat(c *gin.Context)
	InvoiceOutputCreatorStat(c *gin.Context)
	MaterialInInvoice(c *gin.Context)
	MaterialInLocations(c *gin.Context)
}

func NewStatisticsHandler(statUsecase usecase.IStatisticsUsecase) IStatisticsHandler {
	return &statisticsHandler{
		statUsecase: statUsecase,
	}
}

func (handler *statisticsHandler) InvoiceCountStat(c *gin.Context) {
	data, err := handler.statUsecase.InvoiceCountStat(c.GetUint("projectID"))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Система не смогла собрать данные: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *statisticsHandler) InvoiceInputCreatorStat(c *gin.Context) {
	data, err := handler.statUsecase.InvoiceInputCreatorStat(c.GetUint("projectID"))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Система не смогла собрать данные: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *statisticsHandler) InvoiceOutputCreatorStat(c *gin.Context) {
	data, err := handler.statUsecase.InvoiceOutputCreatorStat(c.GetUint("projectID"))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Система не смогла собрать данные: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *statisticsHandler) MaterialInInvoice(c *gin.Context) {
	materialIDRaw := c.Param("materialID")
	materialID, err := strconv.ParseUint(materialIDRaw, 10, 64)

	data, err := handler.statUsecase.CountMaterialInInvoices(uint(materialID))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Система не смогла собрать данные: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *statisticsHandler) MaterialInLocations(c *gin.Context) {
	materialIDRaw := c.Param("materialID")
	materialID, err := strconv.ParseUint(materialIDRaw, 10, 64)

	data, err := handler.statUsecase.LocationMaterial(uint(materialID))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Система не смогла собрать данные: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}
