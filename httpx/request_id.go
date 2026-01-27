package httpx

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type ctxKeyRequestId int32

// Key to use when setting the request ID.
const (
	ctxKeyRequestID ctxKeyRequestId = 1
)

// RequestIDHeader is the name of the HTTP Header which contains the request id.
// Exported so that it can be changed by developers
const defaultRequestIDHeader = "X-Request-Id"

// RequestIDMiddleware is a middleware that generates an UUID and injects that into
// the context to be used down the line.
// This uses the default "X-Request-Id" header to propagate that from the caller downwards.
func RequestIDMiddleware(next http.Handler) http.Handler {
	return RequestIDMiddlewareFromHeader(next, defaultRequestIDHeader)
}

// RequestIDMiddlewareFromHeader is a middleware that generates an UUID and injects that into
// the context to be used down the line.
// This receives a string that will be used to read from the request header and propagate its value as request id.
func RequestIDMiddlewareFromHeader(next http.Handler, fromHeader string) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		requestID := r.Header.Get(fromHeader)
		if requestID == "" {
			requestID = uuid.NewString()
		}
		ctx = context.WithValue(ctx, ctxKeyRequestID, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}

// GetReqID returns a request ID from the given context if one is present.
// Returns the empty string if a request ID cannot be found.
func GetReqID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if reqID, ok := ctx.Value(ctxKeyRequestID).(string); ok {
		return reqID
	}
	return ""
}
