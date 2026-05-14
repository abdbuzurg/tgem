package middleware

import (
	"backend-v2/pkg/jwt"
	"backend-v2/internal/http/response"
	"strings"

	"github.com/gin-gonic/gin"
)

func Authentication() gin.HandlerFunc {
	return func(c *gin.Context) {

		authHeader := c.GetHeader("Authorization")
		if len(authHeader) == 0 {
      response.ResponseError(c, "Ошибка идентификации: вы не являетесь пользователем")
			c.Abort()
			return
		}

		fields := strings.Fields(authHeader)
		if len(fields) < 2 {
      response.ResponseError(c, "Ошибка идентификации: неправильные учетные данные")
			c.Abort()
			return
		}

		authType := strings.ToLower(fields[0])
		if authType != "bearer" {
      response.ResponseError(c, "Ошибка идентификации: неправильная аутентификация")
			c.Abort()
			return
		}

		accessToken := fields[1]
		payload, err := jwt.VerifyToken(accessToken)
		if err != nil {
      response.ResponseError(c, "Ошибка идентификации: ключ аутентификации недействителен или срок действия ключа истек. Выполните вход в систему заново.")
			c.Abort()
			return
		}

    c.Set("userID", payload.UserID)
		c.Set("projectID", payload.ProjectID)
		c.Set("workerID", payload.WorkerID)
		c.Set("roleID", payload.RoleID)

		c.Next()
	}
}
