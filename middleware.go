package autorpc

import (
	"context"
)

type Middleware func(ctx context.Context, req RPCRequest, next HandlerFunc) (RPCResponse, error)

type HandlerFunc func(ctx context.Context, req RPCRequest) (RPCResponse, error)

type MiddlewareChain struct {
	middlewares []Middleware
}

func NewMiddlewareChain(middlewares ...Middleware) *MiddlewareChain {
	return &MiddlewareChain{
		middlewares: middlewares,
	}
}

func (mc *MiddlewareChain) Add(middleware Middleware) {
	mc.middlewares = append(mc.middlewares, middleware)
}

func (mc *MiddlewareChain) Build(finalHandler HandlerFunc) HandlerFunc {
	handler := finalHandler

	for i := len(mc.middlewares) - 1; i >= 0; i-- {
		middleware := mc.middlewares[i]
		next := handler
		handler = func(ctx context.Context, req RPCRequest) (RPCResponse, error) {
			return middleware(ctx, req, next)
		}
	}

	return handler
}

func (mc *MiddlewareChain) Len() int {
	return len(mc.middlewares)
}
