package handlers

import (
	"backend-v2/internal/usecase"
	"backend-v2/model"
	"backend-v2/internal/http/response"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

type roleHandler struct {
	roleUsecase usecase.IRoleUsecase
}

func NewRoleHandler(roleUsecase usecase.IRoleUsecase) IRoleHandler {
	return &roleHandler{
		roleUsecase: roleUsecase,
	}
}

type IRoleHandler interface {
	GetAll(c *gin.Context)
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
}

func (handler *roleHandler) GetAll(c *gin.Context) {
	data, err := handler.roleUsecase.GetAll()
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server Error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *roleHandler) Create(c *gin.Context) {
	var requestData model.Role
	if err := c.ShouldBindJSON(&requestData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid body request: %v", err))
		return
	}

	data, err := handler.roleUsecase.Create(requestData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server Error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *roleHandler) Update(c *gin.Context) {
	var requestData model.Role
	if err := c.ShouldBindJSON(&requestData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid body request: %v", err))
		return
	}

	data, err := handler.roleUsecase.Update(requestData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server Error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *roleHandler) Delete(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	err = handler.roleUsecase.Delete(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server Error: %v", err))
		return
	}

	response.ResponseSuccess(c, true)
}
