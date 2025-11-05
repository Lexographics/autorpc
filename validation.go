package autorpc

import (
	"reflect"

	"github.com/go-playground/validator/v10"
)

type ValidateErrorHandler func(*validator.ValidationErrors) *RPCError

// Returns CodeInvalidParams with validation error details
func defaultValidateErrorHandler(errs *validator.ValidationErrors) *RPCError {
	errorDetails := make([]map[string]any, 0, len(*errs))
	for _, err := range *errs {
		details := map[string]any{}
		if err.Field() != "" {
			details["field"] = err.Namespace()
		}
		if err.Tag() != "" {
			details["tag"] = err.Tag()
		}
		if err.Value() != nil {
			details["value"] = err.Value()
		}
		errorDetails = append(errorDetails, details)
	}

	return &RPCError{
		Code:    CodeInvalidParams,
		Message: "Invalid params",
		Data:    errorDetails,
	}
}

func validateParams(params interface{}, handler ValidateErrorHandler) *RPCError {
	if reflect.TypeOf(params).Kind() != reflect.Struct {
		return nil
	}

	validate := validator.New()
	if err := validate.Struct(params); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			return handler(&validationErrors)
		}
		return &RPCError{
			Code:    CodeInternalError,
			Message: "Validation error: " + err.Error(),
		}
	}
	return nil
}
