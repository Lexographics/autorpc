package main

import (
	"context"
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

func CustomErrorFunc(ctx context.Context, params string) (any, error) {
	return nil, &CustomError{
		code:    -32000, // Custom error code
		message: "This is a custom error",
		data:    map[string]string{"param": params},
	}
}

func Join(ctx context.Context, params []string) (string, error) {
	return strings.Join(params, ""), nil
}

type ConcatParams struct {
	A string `json:"a" validate:"required"`
	B string `json:"b" validate:"required"`
}

func Concat(ctx context.Context, params ConcatParams) (string, error) {
	return params.A + params.B, nil
}

func LogMiddleware(text string) autorpc.Middleware {
	return func(ctx context.Context, req autorpc.RPCRequest, next autorpc.HandlerFunc) (autorpc.RPCResponse, error) {
		fmt.Printf("text: %s\n", text)
		return next(ctx, req)
	}
}

func main() {
	server := autorpc.NewServer()
	server.Use(LogMiddleware("global"))

	stringGroup := server.Group("string.")
	autorpc.RegisterMethod(stringGroup, "concat", Concat)
	autorpc.RegisterMethod(stringGroup, "join", Join)

	autorpc.RegisterMethod(server, "debug.custom-error", CustomErrorFunc)

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
