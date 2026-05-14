package middleware

import (
	"log"

	"backend-v2/internal/auth"

	"github.com/gin-gonic/gin"
)

// EnforcePermissions controls whether RequirePermission denies requests on
// failure. Phase 2 ships with this set to false (log-only mode) so that
// missing grants surface in logs without breaking users. Phase 4 flips it.
//
// Settable from main() before the router is built; not safe to flip at runtime.
var EnforcePermissions = false

// RequirePermission gates a handler on (action, resource). The Authentication
// middleware must run first to populate userID and projectID on the context.
//
// In log-only mode, denials are logged but the request proceeds. In enforce
// mode, denials abort with the standard {success:false, permission:false}
// envelope.
func RequirePermission(resolver auth.Resolver, action auth.Action, resource auth.ResourceType) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		projectID := c.GetUint("projectID")

		allowed, err := resolver.Allowed(c.Request.Context(), userID, projectID, resource, action)
		if err != nil {
			log.Printf("auth: resolver error for user=%d project=%d %s/%s: %v", userID, projectID, action, resource, err)
			if EnforcePermissions {
				abortDenied(c)
				return
			}
			c.Next()
			return
		}

		if !allowed {
			log.Printf("auth: would-deny user=%d project=%d %s %s %s", userID, projectID, c.Request.Method, action, resource)
			if EnforcePermissions {
				abortDenied(c)
				return
			}
		}

		c.Next()
	}
}

func abortDenied(c *gin.Context) {
	c.JSON(200, gin.H{
		"data":       nil,
		"error":      "Доступ запрещен",
		"success":    false,
		"permission": false,
	})
	c.Abort()
}

// GroupGate returns a closure that produces RequirePermission middleware for
// a fixed resource. Lets a route group use a one-line per-handler gate:
//
//	gate := middleware.GroupGate(resolver, auth.ResInvoiceOutput)
//	g.GET("/",          gate(auth.ActionView),    handler.GetAll)
//	g.POST("/",         gate(auth.ActionCreate),  handler.Create)
//	g.POST("/confirm/:id", gate(auth.ActionConfirm), handler.Confirmation)
func GroupGate(resolver auth.Resolver, resource auth.ResourceType) func(auth.Action) gin.HandlerFunc {
	return func(action auth.Action) gin.HandlerFunc {
		return RequirePermission(resolver, action, resource)
	}
}
