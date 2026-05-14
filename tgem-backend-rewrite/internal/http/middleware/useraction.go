// Package middleware: RecordUserAction is a post-handler middleware that
// emits one row per non-GET request into the user_actions audit log so the
// admin section can show "what each user did" across the app.
//
// Behavior:
//   - Skips GET / OPTIONS / HEAD requests outright.
//   - Skips the /api/user/login and /api/user/is-authenticated paths because
//     those run before Authentication() and have their own logging in the
//     login handler (where the username is available without leaking the
//     password).
//   - Reads the response envelope after the handler runs to capture the
//     success / error message fields, which the frontend conventionally
//     uses to signal failure (HTTP is always 200).
//   - Pulls userID/projectID from the gin context (set by Authentication).
//     If userID == 0 — i.e. the route is unauthenticated and slipped past
//     the skip list — the record is dropped so the audit log doesn't
//     accumulate anonymous noise.
//   - Handlers may enrich the record by setting any of these on the gin
//     context before returning:
//       c.Set("actionEntityID", uint(...))   // target entity id (Create/Update/Delete)
//       c.Set("actionType",     "import")    // override the default verb
//       c.Set("actionMessage",  "...")       // free-form note
//   - Audit insert errors are logged to stdout, never propagated.
package middleware

import (
	"bytes"
	"encoding/json"
	"log"
	"time"

	"backend-v2/internal/usecase"
	"backend-v2/model"

	"github.com/gin-gonic/gin"
)

const (
	maxMessageLen = 500
	maxBodyParse  = 8 * 1024 // cap envelope read size for JSON unmarshal
)

// skipPaths are exact URL paths the middleware never logs. Login is logged
// from inside the login handler where the username is known.
var skipPaths = map[string]struct{}{
	"/api/user/login":            {},
	"/api/user/is-authenticated": {},
}

type bodyCapture struct {
	gin.ResponseWriter
	buf *bytes.Buffer
}

func (w *bodyCapture) Write(b []byte) (int, error) {
	if w.buf.Len() < maxBodyParse {
		// Capture only the leading portion of the body — enough to parse
		// the envelope header without copying gigantic responses.
		remaining := maxBodyParse - w.buf.Len()
		if len(b) <= remaining {
			w.buf.Write(b)
		} else {
			w.buf.Write(b[:remaining])
		}
	}
	return w.ResponseWriter.Write(b)
}

func (w *bodyCapture) WriteString(s string) (int, error) {
	if w.buf.Len() < maxBodyParse {
		remaining := maxBodyParse - w.buf.Len()
		if len(s) <= remaining {
			w.buf.WriteString(s)
		} else {
			w.buf.WriteString(s[:remaining])
		}
	}
	return w.ResponseWriter.WriteString(s)
}

func RecordUserAction(uc usecase.IUserActionUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		if method == "GET" || method == "OPTIONS" || method == "HEAD" {
			c.Next()
			return
		}
		if _, skip := skipPaths[c.Request.URL.Path]; skip {
			c.Next()
			return
		}

		buf := &bytes.Buffer{}
		bc := &bodyCapture{ResponseWriter: c.Writer, buf: buf}
		c.Writer = bc

		c.Next()

		userID := c.GetUint("userID")
		if userID == 0 {
			return
		}
		projectID := c.GetUint("projectID")

		actionType := c.GetString("actionType")
		if actionType == "" {
			actionType = defaultActionType(method)
		}

		actionEntityID := c.GetUint("actionEntityID")
		actionMessage := c.GetString("actionMessage")

		envelope := parseEnvelope(buf.Bytes())
		// If the handler didn't set a message, surface the envelope's error
		// on failure (and leave the message empty on success — the timeline
		// is more readable that way).
		if actionMessage == "" && !envelope.Success && envelope.Error != "" {
			actionMessage = truncate(envelope.Error, maxMessageLen)
		}

		uc.Create(model.UserAction{
			ActionURL:           c.Request.URL.Path,
			ActionType:          actionType,
			ActionID:            actionEntityID,
			ActionStatus:        envelope.Success,
			ActionStatusMessage: actionMessage,
			HTTPMethod:          method,
			RequestIP:           c.ClientIP(),
			UserID:              userID,
			ProjectID:           projectID,
			DateOfAction:        time.Now(),
		})
	}
}

func defaultActionType(method string) string {
	switch method {
	case "POST":
		return "create"
	case "PUT", "PATCH":
		return "edit"
	case "DELETE":
		return "delete"
	default:
		return method
	}
}

type envelope struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

func parseEnvelope(b []byte) envelope {
	var e envelope
	if len(b) == 0 {
		return e
	}
	if err := json.Unmarshal(b, &e); err != nil {
		// The body wasn't a JSON envelope (e.g. a file download stream
		// from an export endpoint that sneaks past the GET filter via a
		// POST). Treat as success — handler wrote bytes and didn't 500.
		log.Printf("useraction: envelope parse failed: %v", err)
		return envelope{Success: true}
	}
	return e
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
