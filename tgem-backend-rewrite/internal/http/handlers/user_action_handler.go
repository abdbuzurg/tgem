package handlers

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/http/response"
	"backend-v2/internal/usecase"
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type userActionHandler struct {
	userActionUsecase usecase.IUserActionUsecase
}

func NewUserActionHandler(userActionUsecase usecase.IUserActionUsecase) IUserActionHandler {
	return &userActionHandler{
		userActionUsecase: userActionUsecase,
	}
}

type IUserActionHandler interface {
	GetAllByUserID(c *gin.Context)
	GetPaginated(c *gin.Context)
	GetFilterUsers(c *gin.Context)
}

// GetFilterUsers returns the (id, username, workerName) list that backs the
// admin audit-log page's "filter by user" search box.
func (handler *userActionHandler) GetFilterUsers(c *gin.Context) {
	data, err := handler.userActionUsecase.ListFilterUserOptions()
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not load user options: %v", err))
		return
	}
	response.ResponseSuccess(c, data)
}

func (handler *userActionHandler) GetAllByUserID(c *gin.Context) {
	userIDRaw := c.Param("userID")
	userID, err := strconv.ParseUint(userIDRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	data, err := handler.userActionUsecase.GetAllByUserID(uint(userID))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server Error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

// GetPaginated reads optional filters from the query string and returns a
// page of user actions plus the total count for the same WHERE clause.
//
// Recognized query parameters (all optional):
//
//	page         int   default 1
//	limit        int   default 25
//	userID       uint  filter to a specific user
//	projectID    uint  filter to a specific project
//	actionType   text  e.g. create, edit, delete, login, import
//	httpMethod   text  POST / PATCH / DELETE / etc.
//	status       "true" | "false"  filter by success/failure
//	dateFrom     RFC3339 or YYYY-MM-DD  inclusive lower bound
//	dateTo       RFC3339 or YYYY-MM-DD  inclusive upper bound
func (handler *userActionHandler) GetPaginated(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page <= 0 {
		response.ResponseError(c, fmt.Sprintf("Wrong query parameter provided for page: %v", err))
		return
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "25"))
	if err != nil || limit <= 0 {
		response.ResponseError(c, fmt.Sprintf("Wrong query parameter provided for limit: %v", err))
		return
	}

	filter := dto.UserActionFilter{Page: page, Limit: limit}

	if v := c.Query("userID"); v != "" {
		n, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			response.ResponseError(c, fmt.Sprintf("Wrong query parameter for userID: %v", err))
			return
		}
		filter.UserID = uint(n)
	}
	if v := c.Query("projectID"); v != "" {
		n, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			response.ResponseError(c, fmt.Sprintf("Wrong query parameter for projectID: %v", err))
			return
		}
		filter.ProjectID = uint(n)
	}
	filter.ActionType = c.Query("actionType")
	filter.HTTPMethod = c.Query("httpMethod")
	if v := c.Query("status"); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			response.ResponseError(c, fmt.Sprintf("Wrong query parameter for status: %v", err))
			return
		}
		filter.HasStatus = true
		filter.Status = b
	}
	if v := c.Query("dateFrom"); v != "" {
		t, err := parseFlexibleDate(v, false)
		if err != nil {
			response.ResponseError(c, fmt.Sprintf("Wrong query parameter for dateFrom: %v", err))
			return
		}
		filter.DateFrom = t
	}
	if v := c.Query("dateTo"); v != "" {
		t, err := parseFlexibleDate(v, true)
		if err != nil {
			response.ResponseError(c, fmt.Sprintf("Wrong query parameter for dateTo: %v", err))
			return
		}
		filter.DateTo = t
	}

	data, total, err := handler.userActionUsecase.GetPaginated(filter)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the paginated data of UserAction: %v", err))
		return
	}

	response.ResponsePaginatedData(c, data, total)
}

// parseFlexibleDate accepts RFC3339 or a bare YYYY-MM-DD; with the bare form
// the upper bound is shifted to 23:59:59 so "dateTo=2026-05-10" includes the
// whole day.
func parseFlexibleDate(s string, endOfDay bool) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return time.Time{}, err
	}
	if endOfDay {
		t = t.Add(24*time.Hour - time.Second)
	}
	return t, nil
}
