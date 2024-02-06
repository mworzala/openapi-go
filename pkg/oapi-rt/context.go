package oapi_rt

import (
	"context"
	"net/http"
)

// I am not in love with this because it introduces a dependency on the generator in the generated code
// (for fetching the accept header), but I don't have a better idea.

type contextKey string

const (
	requestKey contextKey = "request"
	acceptKey  contextKey = "accept"
)

func NewContext(ctx context.Context, r *http.Request) context.Context {
	return context.WithValue(
		context.WithValue(ctx, acceptKey, r.Header.Get("Accept")),
		requestKey, r,
	)
}

// RequestFromContext returns the request from the context.
// Bad bad very bad do not use unless necessary
func RequestFromContext(ctx context.Context) *http.Request {
	return ctx.Value(requestKey).(*http.Request)
}
