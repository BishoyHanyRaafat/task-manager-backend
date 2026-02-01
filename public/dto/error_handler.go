package dto

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (e *ErrorEnvelope) Send(c *gin.Context) {
	if e == nil {
		return
	}
	if e.StatusCode == 0 {
		e.StatusCode = http.StatusInternalServerError
	}
	if e.Data.Code == "" {
		e.Data.Code = CodeInternalError
	}
	Fail(c, e.StatusCode, e.Data.Code, e.Data.Message, e.Debug, e.Data.Details)
}

func BadRequest(code ErrorCode, msg string, details map[string]any) *ErrorEnvelope {
	return &ErrorEnvelope{StatusCode: http.StatusBadRequest, Data: ErrorData{Code: code, Message: msg, Details: details}}
}

func Unauthorized(code ErrorCode, msg string, details map[string]any) *ErrorEnvelope {
	return &ErrorEnvelope{StatusCode: http.StatusUnauthorized, Data: ErrorData{Code: code, Message: msg, Details: details}}
}

func Forbidden(code ErrorCode, msg string, details map[string]any) *ErrorEnvelope {
	return &ErrorEnvelope{StatusCode: http.StatusForbidden, Data: ErrorData{Code: code, Message: msg, Details: details}}
}

func Conflict(code ErrorCode, msg string, details map[string]any) *ErrorEnvelope {
	return &ErrorEnvelope{StatusCode: http.StatusConflict, Data: ErrorData{Code: code, Message: msg, Details: details}}
}

func Internal(code ErrorCode, msg string, debug string, details map[string]any) *ErrorEnvelope {
	return &ErrorEnvelope{
		StatusCode: http.StatusInternalServerError,
		Data:       ErrorData{Code: code, Message: msg, Details: details},
		Debug:      debug,
	}
}

func NotFound(code ErrorCode, msg string, debug string, details map[string]any) *ErrorEnvelope {
	return &ErrorEnvelope{
		StatusCode: http.StatusNotFound,
		Data:       ErrorData{Code: code, Message: msg, Details: details},
		Debug:      debug,
	}
}
