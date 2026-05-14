package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ResponseFormat struct {
	Data       interface{} `json:"data"`
	Error      string      `json:"error"`
	Success    bool        `json:"success"`
	Permission bool        `json:"permission"`
}

// Будет использвана эта фунцкия если запрос который пришел на сервер выполнен успешно
func ResponseSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, ResponseFormat{
		Data:       data,
		Success:    true,
		Permission: true,
	})
}

// Будет использована эта функция если запрос который пришел на сервер дал сбой
func ResponseError(c *gin.Context, errorMessage string) {
	c.JSON(http.StatusOK, ResponseFormat{
		Error:      errorMessage,
		Success:    false,
		Permission: true,
	})
}

// Будет использована эта фунцкия если пользователь не имеет доступа к определенному ресурсу
func ResponsePermissionDenied(c *gin.Context) {
	c.JSON(http.StatusOK, ResponseFormat{
		Success:    false,
		Permission: false,
	})
}

// Будет использвана эта фунцкия если данные котороые получает сервер не соответствует проприсанными правилами
func ResponseInvalidData(c *gin.Context, data map[string]interface{}) {
	c.JSON(http.StatusOK, ResponseFormat{
		Success:    false,
		Permission: true,
		Data:       data,
		Error:      "Data did not pass validation",
	})
}

func ResponsePaginatedData(c *gin.Context, data interface{}, count int64) {
	c.JSON(http.StatusOK, ResponseFormat{
		Success:    true,
		Permission: true,
		Data: gin.H{
			"data":  data,
			"count": count,
		},
	})
}
