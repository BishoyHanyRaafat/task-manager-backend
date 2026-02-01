package validation

import (
	"fmt"
	"task_manager/internal/response"

	"github.com/go-playground/validator/v10"
)

var Validate = validator.New()

func Format(err error) (message string, details map[string]any, debug string) {
	if err == nil {
		return "", nil, ""
	}
	verrs, ok := err.(validator.ValidationErrors)
	if !ok {
		return "validation error", map[string]any{"error": err.Error()}, err.Error()
	}

	fieldErrors := map[string]any{}
	for _, fe := range verrs {
		// Field names are struct fields; good enough for debug. You can map to json tags later if desired.
		// in case of min or max tag, you can get the param using fe.Param()
		fieldErrors[fe.Field()] = formatTag(fe.Field(), fe.Tag(), fe.Param())
	}
	return "validation error", fieldErrors, "Invalid fields check the details for more information"
}

func formatTag(field string, tag string, param string) string {
	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return "You must enter a valid email address"
	case "min":
		min := 0
		fmt.Sscanf(param, "%d", &min)
		return fmt.Sprintf("%s must be at least %d characters long", field, min)
	case "max":
		max := 0
		fmt.Sscanf(param, "%d", &max)
		return fmt.Sprintf("%s must be at most %d characters long", field, max)
	default:
		return tag
	}
}

func RegisterValidations() {
	response.RegisterErrorCodeValidation(Validate)
}
