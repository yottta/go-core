package chix

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v3"
)

// Config can be embedded in your configs and map flags and env vars directly to the
// [Config.Host] and [Config.Port] attributes.
//
// With the [Config.NewServer] a new [*Server] will be returned to handle an http
// handler.
type Config struct {
	Host string
	Port int

	middlewares []func(http.Handler) http.Handler
}

// setDefaults configures defaults on the config.
// At the moment, it's used to set some default middlewares.
func (c *Config) setDefaults() {
	// The middlewares here are executed in the same order as are defined here:
	// request -> middleware0 -> ... -> middlewareN -> handler
	c.middlewares = []func(http.Handler) http.Handler{
		middleware.RequestID,
		middleware.RealIP,
		httplog.RequestLogger(slog.Default(), &httplog.Options{}), // Using slog.Default() because this is configured at the app level. Check main.go
	}
}

type Opt func(*Config)

// WithPreMiddleware inserts a middleware before the the default chain configured by [Config#setDefaults].
// This is recommended only for specific cases, like recovery middlewares.
func WithPreMiddleware(m func(http.Handler) http.Handler) Opt {
	return func(config *Config) {
		config.middlewares = append([]func(http.Handler) http.Handler{m}, config.middlewares...)
	}
}

// WithPostMiddleware adds a middleware after the the default chain configured by [Config#setDefaults].
// This is the recommended way to configure middlewares, leaving untouched the default chain of
// middlewares.
func WithPostMiddleware(m func(http.Handler) http.Handler) Opt {
	return func(config *Config) {
		config.middlewares = append(config.middlewares, m)
	}
}

// WithMiddlewares overwrites all the middlewares, also the default ones.
func WithMiddlewares(m ...func(http.Handler) http.Handler) Opt {
	return func(config *Config) {
		config.middlewares = m
	}
}
