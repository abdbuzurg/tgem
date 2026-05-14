package handlers

import (
	"backend-v2/internal/usecase"
	"backend-v2/internal/http/response"
	"fmt"

	"github.com/gin-gonic/gin"
)

type resourceHandler struct {
	resourceUsecase usecase.IResourceUsecase
}

func NewResourceHandler(
	resourceUsecase usecase.IResourceUsecase,
) IResourceHandler {
	return &resourceHandler{
		resourceUsecase: resourceUsecase,
	}
}

type IResourceHandler interface {
	GetAll(c *gin.Context)
}

func (handler *resourceHandler) GetAll(c *gin.Context) {
	data, err := handler.resourceUsecase.GetAll()
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутреняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}
