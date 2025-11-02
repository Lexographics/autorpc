package autorpc

import (
	"context"
	"encoding/json"
	"reflect"
	"sync"
)

type methodHandler struct {
	fnValue reflect.Value
}

type Server struct {
	methods              sync.Map
	validateErrorHandler ValidateErrorHandler
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) SetValidateErrorHandler(handler ValidateErrorHandler) {
	s.validateErrorHandler = handler
}

// RegisterMethod registers a method with the given name and function.
// The function must have the signature: func(context.Context, ParamsType) (ResultType, error)
//
// Example:
//
//	RegisterMethod(server, "add", func(ctx context.Context, params []int) (int, error) {
//	    if len(params) != 2 {
//	        return 0, errors.New("expected 2 numbers")
//	    }
//	    return params[0] + params[1], nil
//	})
func RegisterMethod[P, R any](s *Server, name string, fn func(context.Context, P) (R, error)) {
	handler := methodHandler{
		fnValue: reflect.ValueOf(fn),
	}
	s.methods.Store(name, handler)
}

func (s *Server) processRequest(ctx context.Context, req RPCRequest) (resp RPCResponse) {
	defer func() {
		if r := recover(); r != nil {
			resp = newErrorResponse(req.ID, CodeInternalError, "Internal error")
		}
	}()

	if req.JSONRPC != "2.0" {
		return newErrorResponse(req.ID, CodeInvalidRequest, "Invalid JSON-RPC version")
	}

	handlerValue, ok := s.methods.Load(req.Method)
	if !ok {
		return newErrorResponse(req.ID, CodeMethodNotFound, "Method not found")
	}

	handler, ok := handlerValue.(methodHandler)
	if !ok {
		return newErrorResponse(req.ID, CodeInternalError, "Internal error: invalid method handler")
	}

	fnType := handler.fnValue.Type()
	if fnType.NumIn() != 2 || fnType.NumOut() != 2 {
		return newErrorResponse(req.ID, CodeInternalError, "Internal error: invalid method signature")
	}

	contextType := fnType.In(0)
	if contextType.String() != "context.Context" {
		return newErrorResponse(req.ID, CodeInternalError, "Internal error: first parameter must be context.Context")
	}

	paramType := fnType.In(1)
	paramPtr := reflect.New(paramType).Interface()

	if err := json.Unmarshal(req.Params, paramPtr); err != nil {
		return newErrorResponse(req.ID, CodeInvalidParams, "Failed to unmarshal params: "+err.Error())
	}

	paramValue := reflect.ValueOf(paramPtr).Elem().Interface()

	handlerFunc := s.validateErrorHandler
	if handlerFunc == nil {
		handlerFunc = defaultValidateErrorHandler
	}

	if validationErr := validateParams(paramValue, handlerFunc); validationErr != nil {
		return RPCResponse{
			JSONRPC: "2.0",
			Error:   validationErr,
			ID:      req.ID,
		}
	}

	results := handler.fnValue.Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(paramValue)})
	resultValue := results[0].Interface()
	errValue := results[1].Interface()

	if errValue != nil {
		if err, ok := errValue.(error); ok && err != nil {
			rpcErr := errorToRPCError(err)
			return RPCResponse{
				JSONRPC: "2.0",
				Error:   rpcErr,
				ID:      req.ID,
			}
		}
	}

	return RPCResponse{
		JSONRPC: "2.0",
		Result:  resultValue,
		ID:      req.ID,
	}
}
