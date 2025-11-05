package autorpc

import (
	"context"
	"encoding/json"
	"reflect"
	"sync"
)

type methodHandler struct {
	fnValue     reflect.Value
	middlewares *MiddlewareChain
}

type Server struct {
	methods              sync.Map
	validateErrorHandler ValidateErrorHandler
	globalMiddlewares    *MiddlewareChain
}

func NewServer() *Server {
	return &Server{
		globalMiddlewares: NewMiddlewareChain(),
	}
}

func (s *Server) SetValidateErrorHandler(handler ValidateErrorHandler) {
	s.validateErrorHandler = handler
}

func (s *Server) Use(middlewares ...Middleware) {
	for _, mw := range middlewares {
		s.globalMiddlewares.Add(mw)
	}
}

// RegisterMethod registers a method with the given name and function.
// The first parameter can be either *Server or *Group.
// The function must have the signature: func(context.Context, ParamsType) (ResultType, error)
//
// Example with Server:
//
//	RegisterMethod(server, "add", func(ctx context.Context, params []int) (int, error) {
//	    if len(params) != 2 {
//	        return 0, errors.New("expected 2 numbers")
//	    }
//	    return params[0] + params[1], nil
//	})
//
// Example with Group:
//
//	mathGroup := server.Group("math.")
//	RegisterMethod(mathGroup, "add", AddFunc)
//
// With middleware:
//
//	RegisterMethod(server, "add", AddFunc, AuthMiddleware(), LoggingMiddleware())
func RegisterMethod[P, R any](
	r Registerer,
	name string,
	fn func(context.Context, P) (R, error),
	middlewares ...Middleware,
) {

	fnValue := reflect.ValueOf(fn)
	if fnValue.Kind() != reflect.Func {
		panic("RegisterMethod: fn must be a function")
	}

	fnType := fnValue.Type()
	if fnType.NumIn() != 2 || fnType.NumOut() != 2 {
		panic("RegisterMethod: function must have signature func(context.Context, ParamsType) (ResultType, error)")
	}

	contextType := fnType.In(0)
	if contextType.String() != "context.Context" {
		panic("RegisterMethod: first parameter must be context.Context")
	}

	errType := fnType.Out(1)
	if !errType.Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		panic("RegisterMethod: second return value must be error")
	}

	middlewareChain := NewMiddlewareChain(middlewares...)
	r.register(name, fn, middlewareChain)
}

func (s *Server) register(name string, fn interface{}, allMiddlewares *MiddlewareChain) {
	fnValue := reflect.ValueOf(fn)
	if fnValue.Kind() != reflect.Func {
		panic("register: fn must be a function")
	}

	fnType := fnValue.Type()
	if fnType.NumIn() != 2 || fnType.NumOut() != 2 {
		panic("register: function must have signature func(context.Context, ParamsType) (ResultType, error)")
	}

	contextType := fnType.In(0)
	if contextType.String() != "context.Context" {
		panic("register: first parameter must be context.Context")
	}

	errType := fnType.Out(1)
	if !errType.Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		panic("register: second return value must be error")
	}

	combinedMiddlewares := NewMiddlewareChain()

	if s.globalMiddlewares != nil && s.globalMiddlewares.Len() > 0 {
		for i := 0; i < s.globalMiddlewares.Len(); i++ {
			combinedMiddlewares.Add(s.globalMiddlewares.middlewares[i])
		}
	}

	if allMiddlewares != nil && allMiddlewares.Len() > 0 {
		for i := 0; i < allMiddlewares.Len(); i++ {
			combinedMiddlewares.Add(allMiddlewares.middlewares[i])
		}
	}

	handler := methodHandler{
		fnValue:     fnValue,
		middlewares: combinedMiddlewares,
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

	finalHandler := func(ctx context.Context, req RPCRequest) (RPCResponse, error) {
		paramType := fnType.In(1)
		paramPtr := reflect.New(paramType).Interface()

		if err := json.Unmarshal(req.Params, paramPtr); err != nil {
			return newErrorResponse(req.ID, CodeInvalidParams, "Failed to unmarshal params: "+err.Error()), err
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
			}, nil
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
				}, err
			}
		}

		return RPCResponse{
			JSONRPC: "2.0",
			Result:  resultValue,
			ID:      req.ID,
		}, nil
	}

	chainHandler := finalHandler

	if handler.middlewares != nil && handler.middlewares.Len() > 0 {
		chainHandler = handler.middlewares.Build(finalHandler)
	}

	resp, err := chainHandler(ctx, req)
	if err != nil {
		if resp.Error == nil {
			resp = newErrorResponse(req.ID, CodeInternalError, "Internal error")
		}
	}

	return resp
}
