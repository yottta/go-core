package httpx

import (
	"log/slog"
	"net/http"
	"time"
)

// SloggingMiddleware is a basic middleware that prints basic information into logs by using [slog].
func SloggingMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		reqAttrs := requestAttributes(r)
		start := time.Now()
		slog.
			With(reqAttrs...).
			With("at", start.Format(time.RFC3339Nano)).
			Debug("request received")
		rw := NewInterceptor(w)
		next.ServeHTTP(rw, r)
		end := time.Now()
		duration := end.Sub(start)
		slog.
			With(responseInfo(rw)...).
			With("at", end.Format(time.RFC3339Nano)).
			With("duration", duration).
			Debug("request finished")

	}
	return http.HandlerFunc(fn)
}

func requestAttributes(r *http.Request) []any {
	var attrs []any
	if ra := r.RemoteAddr; len(ra) > 0 {
		attrs = append(attrs, "remote.addr", ra)
	}
	attrs = append(attrs, "headers", r.Header)
	attrs = append(attrs, "url.full", r.RequestURI)
	return attrs
}

func responseInfo(w *ResponseWriterCoder) []any {
	var attrs []any
	attrs = append(attrs, "response.size", w.Size)
	attrs = append(attrs, "response.code", w.StatusCode)
	return attrs
}

type ResponseWriterCoder struct {
	base       http.ResponseWriter
	Size       int
	StatusCode int
}

var _ http.ResponseWriter = &ResponseWriterCoder{}

func NewInterceptor(w http.ResponseWriter) *ResponseWriterCoder {
	return &ResponseWriterCoder{
		base:       w,
		StatusCode: http.StatusOK,
	}
}

func (i *ResponseWriterCoder) Header() http.Header {
	return i.base.Header()
}

func (i *ResponseWriterCoder) Write(bb []byte) (int, error) {
	i.Size += len(bb)
	return i.base.Write(bb)
}

func (i *ResponseWriterCoder) WriteHeader(statusCode int) {
	i.StatusCode = statusCode
	i.base.WriteHeader(statusCode)
}
