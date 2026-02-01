package dto

import "github.com/go-playground/validator/v10"

// ErrorCode is a stable machine-readable error identifier.
// Keep these values backward-compatible once clients rely on them.
//
//swagger:enum ErrorCode
type ErrorCode string

const (
	CodeValidationError ErrorCode = "VALIDATION_ERROR"
	CodeInvalidRequest  ErrorCode = "INVALID_REQUEST"

	CodeUnauthorized ErrorCode = "UNAUTHORIZED"
	CodeForbidden    ErrorCode = "FORBIDDEN"

	CodeNotFound ErrorCode = "PAGE_NOT_FOUND"
	CodeConflict ErrorCode = "CONFLICT"

	CodeDatabaseError ErrorCode = "DATABASE_ERROR"
	CodeInternalError ErrorCode = "INTERNAL_ERROR"

	CodeInvalidToken ErrorCode = "INVALID_TOKEN"
	CodeMissingToken ErrorCode = "MISSING_TOKEN"

	CodeInvalidEmail ErrorCode = "INVALID_EMAIL"
)

type ErrorData struct {
	Code    ErrorCode      `json:"code" validate:"error_code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

func (c ErrorCode) IsValid() bool {
	switch c {
	case
		CodeValidationError,
		CodeInvalidRequest,
		CodeUnauthorized,
		CodeForbidden,
		CodeNotFound,
		CodeConflict,
		CodeDatabaseError,
		CodeInternalError,
		CodeInvalidToken,
		CodeMissingToken,
		CodeInvalidEmail:
		return true
	default:
		return false
	}
}

func RegisterErrorCodeValidation(v *validator.Validate) {
	v.RegisterValidation("error_code", func(fl validator.FieldLevel) bool {
		code, ok := fl.Field().Interface().(ErrorCode)
		return ok && code.IsValid()
	})
}
