package handlers

import (
	"backend-v2/internal/usecase"
	"backend-v2/model"
	"backend-v2/internal/http/response"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

type permissionHandler struct {
	permissionUsecase usecase.IPermissionUsecase
}

func NewPermissionHandler(permissionUsecase usecase.IPermissionUsecase) IPermissionHandler {
	return &permissionHandler{
		permissionUsecase: permissionUsecase,
	}
}

type IPermissionHandler interface {
	GetAll(c *gin.Context)
	GetByRoleID(c *gin.Context)
	GetByRoleName(c *gin.Context)
	GetByResourceURL(c *gin.Context)
	GetCurrentUserEffectivePermissions(c *gin.Context)
	Create(c *gin.Context)
	CreateBatch(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
}

func (handler *permissionHandler) GetAll(c *gin.Context) {
	data, err := handler.permissionUsecase.GetAll()
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *permissionHandler) GetByRoleID(c *gin.Context) {
	roleIDRaw := c.Param("roleID")
	roleID, err := strconv.ParseUint(roleIDRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	data, err := handler.permissionUsecase.GetByRoleID(uint(roleID))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *permissionHandler) Create(c *gin.Context) {
	var requestData model.Permission
	if err := c.ShouldBindJSON(&requestData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid request body: %v", err))
		return
	}

	data, err := handler.permissionUsecase.Create(requestData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *permissionHandler) CreateBatch(c *gin.Context) {
	var requestData []model.Permission
	if err := c.ShouldBindJSON(&requestData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid request body: %v", err))
		return
	}

	err := handler.permissionUsecase.CreateBatch(requestData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
		return
	}

	response.ResponseSuccess(c, true)
}

func (handler *permissionHandler) Update(c *gin.Context) {
	var requestData model.Permission
	if err := c.ShouldBindJSON(&requestData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid body request: %v", err))
		return
	}

	data, err := handler.permissionUsecase.Update(requestData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server Error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *permissionHandler) Delete(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	err = handler.permissionUsecase.Delete(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server Error: %v", err))
		return
	}

	response.ResponseSuccess(c, true)
}

func (handler *permissionHandler) GetByRoleName(c *gin.Context) {

	roleName := c.Param("roleName")

	permissions, err := handler.permissionUsecase.GetByRoleName(roleName)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server Error: %v", err))
		return
	}

	response.ResponseSuccess(c, permissions)

}

func (handler *permissionHandler) GetByResourceURL(c *gin.Context) {

	roleID := c.GetUint("roleID")

	resourceURL := c.Param("resourceURL")

	err := handler.permissionUsecase.GetByResourceURL("/"+resourceURL, roleID)
	if err != nil {
		response.ResponseSuccess(c, false)
		return
	}

	response.ResponseSuccess(c, true)
}

// GetCurrentUserEffectivePermissions returns the v2 flat permission list for
// the authenticated user. The frontend declarative gate
// (Require action="..." resource="...") consumes this.
func (handler *permissionHandler) GetCurrentUserEffectivePermissions(c *gin.Context) {
	userID := c.GetUint("userID")

	permissions, err := handler.permissionUsecase.GetEffectivePermissionsForUser(userID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server Error: %v", err))
		return
	}

	response.ResponseSuccess(c, permissions)
}
