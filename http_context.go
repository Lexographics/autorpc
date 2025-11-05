package autorpc

import (
	"context"
	"net/http"
)

type httpRequestKey struct{}

func WithHTTPRequest(ctx context.Context, r *http.Request) context.Context {
	return context.WithValue(ctx, httpRequestKey{}, r)
}

func HTTPRequestFromContext(ctx context.Context) *http.Request {
	if r, ok := ctx.Value(httpRequestKey{}).(*http.Request); ok {
		return r
	}
	return nil
}
