package handlers

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/usecase"
	"backend-v2/internal/http/response"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

type auctionHandler struct {
	auctionUsecase usecase.IAuctionUsecase
}

func NewAuctionHandler(auctionUsecase usecase.IAuctionUsecase) IAuctionHandler {
	return &auctionHandler{
		auctionUsecase: auctionUsecase,
	}
}

type IAuctionHandler interface {
	GetAuctionDataForPublic(c *gin.Context)
	GetAuctionDataForPrivate(c *gin.Context)
	SaveParticipantChanges(c *gin.Context)
}

func (handler *auctionHandler) GetAuctionDataForPublic(c *gin.Context) {
	auctionIDRaw := c.Param("auctionID")
	auctionID, err := strconv.ParseUint(auctionIDRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Неправильный параметер запроса: %v", err))
		return
	}

	result, err := handler.auctionUsecase.GetAuctionDataForPublic(uint(auctionID))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, result)
}

func (handler *auctionHandler) GetAuctionDataForPrivate(c *gin.Context) {
	auctionIDRaw := c.Param("auctionID")
	auctionID, err := strconv.ParseUint(auctionIDRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Неправильный параметер запроса: %v", err))
		return
	}

	result, err := handler.auctionUsecase.GetAuctionDataForPrivate(uint(auctionID), c.GetUint("userID"))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, result)
}

func (handler *auctionHandler) SaveParticipantChanges(c *gin.Context) {
	var participantChanges []dto.ParticipantDataForSave
	if err := c.ShouldBindJSON(&participantChanges); err != nil {
		response.ResponseError(c, fmt.Sprintf("Request Error: %v", err))
		return
	}

	err := handler.auctionUsecase.SaveParticipantChanges(c.GetUint("userID"), participantChanges)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server Error: %v", err))
		return
	}

	response.ResponseSuccess(c, true)
}
