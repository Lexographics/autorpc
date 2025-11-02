package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/Lexographics/autorpc"
	"github.com/go-playground/validator/v10"
)

type CustomError struct {
	code    int
	message string
	data    interface{}
}

func (e *CustomError) Error() string {
	return e.message
}

func (e *CustomError) Code() int {
	return e.code
}

func (e *CustomError) Message() string {
	return e.message
}

func (e *CustomError) Data() interface{} {
	return e.data
}

func CustomErrorFunc(params string) (any, error) {
	return nil, &CustomError{
		code:    -32000, // Custom error code
		message: "This is a custom error",
		data:    map[string]string{"param": params},
	}
}

func Join(params []string) (string, error) {
	return strings.Join(params, ""), nil
}

type ConcatTwoParams struct {
	A string `json:"a" validate:"required"`
	B string `json:"b" validate:"required"`
}

func Concat(params ConcatTwoParams) (string, error) {
	return params.A + params.B, nil
}

func main() {
	server := autorpc.NewServer()

	autorpc.RegisterMethod(server, "debug.custom-error", CustomErrorFunc)
	autorpc.RegisterMethod(server, "string.join", Join)
	autorpc.RegisterMethod(server, "string.concat", Concat)

	server.SetValidateErrorHandler(func(errs *validator.ValidationErrors) *autorpc.RPCError {
		details := make([]string, 0, len(*errs))
		for _, err := range *errs {
			details = append(details, fmt.Sprintf("%s: %s", err.Field(), err.Tag()))
		}
		return &autorpc.RPCError{
			Code:    autorpc.CodeInvalidParams,
			Message: "Validation failed",
			Data:    details,
		}
	})

	http.Handle("/rpc", autorpc.HTTPHandler(server))
	http.Handle("/spec.json", autorpc.SpecJSONHandler(server))

	log.Println("Server started on port 8080")
	http.ListenAndServe(":8080", nil)
}
