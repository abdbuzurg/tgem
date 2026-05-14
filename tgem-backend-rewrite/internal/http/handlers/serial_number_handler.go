package handlers

import (
	"backend-v2/internal/usecase"
	"backend-v2/model"
	"backend-v2/internal/http/response"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

type serialNumberHandler struct {
	serialNumberUsecase usecase.ISerialNumberUsecase
}

func NewSerialNumberHandler(serialNumberUsecase usecase.ISerialNumberUsecase) ISerialNumberHandler {
	return &serialNumberHandler{
		serialNumberUsecase: serialNumberUsecase,
	}
}

type ISerialNumberHandler interface {
	GetAll(c *gin.Context)
	GetCodes(c *gin.Context)
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
}

func (handler *serialNumberHandler) GetAll(c *gin.Context) {
	data, err := handler.serialNumberUsecase.GetAll()
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server Error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *serialNumberHandler) GetCodes(c *gin.Context) {
	materialIDRaw := c.Param("materialID")
	materialID, err := strconv.ParseUint(materialIDRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect request parameter: %v", err))
		return
	}

	codes, err := handler.serialNumberUsecase.GetCodesByMaterialID(uint(materialID))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
		return
	}

	response.ResponseSuccess(c, codes)
}

func (handler *serialNumberHandler) Create(c *gin.Context) {
	var data model.SerialNumber
	if err := c.ShouldBindJSON(&data); err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect request body: %v", err))
		return
	}

	data, err := handler.serialNumberUsecase.Create(data)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server Error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *serialNumberHandler) Update(c *gin.Context) {
	var data model.SerialNumber
	if err := c.ShouldBindJSON(&data); err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect request body: %v", err))
		return
	}

	data, err := handler.serialNumberUsecase.Update(data)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server Errror: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *serialNumberHandler) Delete(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect request parameter: %v", err))
		return
	}

	err = handler.serialNumberUsecase.Delete(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server error: %v", err))
		return
	}

	response.ResponseSuccess(c, true)
}
