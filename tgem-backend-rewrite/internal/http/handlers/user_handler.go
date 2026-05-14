package handlers

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/http/response"
	"backend-v2/internal/usecase"
	"backend-v2/model"
	"backend-v2/pkg/jwt"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type userHandler struct {
	userUsecase       usecase.IUserUsecase
	userActionUsecase usecase.IUserActionUsecase
}

func NewUserHandler(userUsecase usecase.IUserUsecase, userActionUsecase usecase.IUserActionUsecase) IUserHandler {
	return &userHandler{
		userUsecase:       userUsecase,
		userActionUsecase: userActionUsecase,
	}
}

type IUserHandler interface {
	GetAll(c *gin.Context)
	GetPaginated(c *gin.Context)
	GetByID(c *gin.Context)
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	Login(c *gin.Context)
	IsAuthenticated(c *gin.Context)
}

func (handler *userHandler) GetAll(c *gin.Context) {
	data, err := handler.userUsecase.GetAll()
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get User data: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *userHandler) GetPaginated(c *gin.Context) {
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

	workerIDStr := c.DefaultQuery("workerID", "")
	workerID := 0
	if workerIDStr != "" {
		workerID, err = strconv.Atoi(workerIDStr)
		if err != nil {
			response.ResponseError(c, fmt.Sprintf("Cannot decode workerID parameter: %v", err))
			return
		}
	}

	username := c.DefaultQuery("username", "")
	username, err = url.QueryUnescape(username)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Wrong query parameter provided for username: %v", err))
		return
	}

	filter := model.User{
		WorkerID: uint(workerID),
		Username: username,
	}

	data, err := handler.userUsecase.GetPaginated(page, limit, filter)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the paginated data of User: %v", err))
		return
	}

	dataCount, err := handler.userUsecase.Count()
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the total amount of User: %v", err))
		return
	}

	response.ResponsePaginatedData(c, data, dataCount)
}

func (handler *userHandler) GetByID(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	data, err := handler.userUsecase.GetByID(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the data with ID(%d): %v", id, err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (handler *userHandler) Create(c *gin.Context) {
	var createData dto.NewUserData
	if err := c.ShouldBindJSON(&createData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	err := handler.userUsecase.Create(createData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could perform the creation of User: %v", err))
		return
	}

	response.ResponseSuccess(c, true)
}

func (handler *userHandler) Update(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	var updateData dto.NewUserData
	if err := c.ShouldBindJSON(&updateData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}
	updateData.UserData.ID = uint(id)

	if err := handler.userUsecase.Update(updateData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the updation of User: %v", err))
		return
	}

	response.ResponseSuccess(c, true)
}

func (handler *userHandler) Delete(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	err = handler.userUsecase.Delete(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the deletion of User: %v", err))
		return
	}

	response.ResponseSuccess(c, "deleted")
}

func (handler *userHandler) Login(c *gin.Context) {
	var data dto.LoginData
	if err := c.ShouldBindJSON(&data); err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect data recieved by server: %v", err))
		return
	}

	result, err := handler.userUsecase.Login(data)

	// Record the login attempt before returning. We only ever persist the
	// username (failed lookups still get logged for security review). The
	// password is never touched. On success, the user id is recovered by
	// re-verifying the issued JWT — keeps the usecase signature unchanged.
	audit := model.UserAction{
		ActionURL:    "/api/user/login",
		ActionType:   "login",
		HTTPMethod:   "POST",
		RequestIP:    c.ClientIP(),
		ProjectID:    data.ProjectID,
		DateOfAction: time.Now(),
		ActionStatus: err == nil,
	}
	if err != nil {
		audit.ActionStatusMessage = fmt.Sprintf("username=%s; %v", data.Username, err)
	} else {
		audit.ActionStatusMessage = fmt.Sprintf("username=%s", data.Username)
		if payload, jerr := jwt.VerifyToken(result.Token); jerr == nil {
			audit.UserID = payload.UserID
		}
	}
	handler.userActionUsecase.Create(audit)

	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Ошибка при входе: %v", err))
		return
	}

	response.ResponseSuccess(c, result)
}

func (handler *userHandler) IsAuthenticated(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if len(authHeader) == 0 {
		response.ResponseError(c, "not authenticated based on first-level check")
		return
	}

	fields := strings.Fields(authHeader)
	if len(fields) < 2 {
		response.ResponseError(c, "not authenticated based on second-level check")
		return
	}

	authType := strings.ToLower(fields[0])
	if authType != "bearer" {
		response.ResponseError(c, "not authenticated based on third-level check")
		return
	}

	accessToken := fields[1]
	_, err := jwt.VerifyToken(accessToken)
	if err != nil {
		response.ResponseError(c, "not authenticated based on forth-level check")
		return
	}

	response.ResponseSuccess(c, "authenticated")
}
