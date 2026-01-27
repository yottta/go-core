package httpx

import "net/http"

type Middlewares []func(handler http.Handler) http.Handler

func (m Middlewares) ApplyOn(handler http.HandlerFunc) http.Handler {
	h := http.Handler(handler)
	for i := len(m) - 1; i >= 0; i-- {
		h = m[i](h)
	}
	return h
}
