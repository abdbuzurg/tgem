package apperr

import (
	"backend-v2/internal/http/response"
	"errors"
	"log"

	"github.com/gin-gonic/gin"
)

// WriteError emits the standard envelope for an error. HTTP status is always
// 200 — that is the locked-in contract from Phase 1. The response is routed
// through the existing internal/http/response helpers so the on-wire shape never
// diverges from the rest of the codebase during the migration.
//
// CodeInternal and unwrapped non-apperr errors get their cause logged via
// log.Printf so the user-facing message can stay generic without losing
// debug context. Structured logging is a later phase.
func WriteError(c *gin.Context, err error) {
	var ae *Error
	if errors.As(err, &ae) {
		switch ae.Code {
		case CodePermissionDenied:
			response.ResponsePermissionDenied(c)
		case CodeInternal:
			log.Printf("apperr internal: %s %s: %v", c.Request.Method, c.Request.URL.Path, ae.Cause)
			response.ResponseError(c, ae.Message)
		default:
			response.ResponseError(c, ae.Message)
		}
		return
	}

	log.Printf("apperr: non-apperr error from %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
	response.ResponseError(c, "Внутренняя ошибка сервера")
}
