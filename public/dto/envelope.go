package dto

import (
	"log"
	"net/http"
	"strings"
	"task_manager/public/trace"

	"github.com/gin-gonic/gin"
)

// Envelope is the unified API response format.
//
// success: true/false
// data:
//   - on error: { "code": "...", "message": "...", "trace_id": "...", "details": {...} }
//   - on success: any object payload (tokens, user info, etc.)
type Envelope[T any] struct {
	StatusCode int    `json:"-"`
	Success    bool   `json:"success"`
	TraceID    string `json:"trace_id,omitempty"`
	Debug      string `json:"debug,omitempty"`
	Data       T      `json:"data"`
}

type (
	EnvelopeAny   = Envelope[any]
	ErrorEnvelope Envelope[ErrorData]
)

func OK[T any](c *gin.Context, status int, data T) {
	if status == 0 {
		status = http.StatusOK
	}
	tid := trace.Get(c)
	c.JSON(status, Envelope[T]{
		StatusCode: status,
		Success:    true,
		TraceID:    tid,
		Data:       data,
	})
}

func Fail(c *gin.Context, status int, code ErrorCode, message string, debug string, details map[string]any) {
	if status == 0 {
		status = http.StatusInternalServerError
	}
	tid := trace.Get(c)

	// Log debug details (never return them to clients).
	if strings.TrimSpace(debug) != "" {
		log.Printf("trace_id=%s event=error code=%s message=%q debug=%q", tid, code, message, debug)
	} else {
		log.Printf("trace_id=%s event=error code=%s message=%q", tid, code, message)
	}

	// Dev-only field: never include debug details in release mode.
	clientDebug := ""
	if gin.Mode() != gin.ReleaseMode {
		clientDebug = debug
	}

	payload := ErrorData{
		Code:    code,
		Message: message,
		Details: details,
	}
	c.AbortWithStatusJSON(status, ErrorEnvelope{
		StatusCode: status,
		Success:    false,
		TraceID:    tid,
		Debug:      clientDebug,
		Data:       payload,
	})
}
