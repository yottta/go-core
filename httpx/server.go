package httpx

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"
)

// NewServer creates a new server from the given opts.
// This returns the struct that can be used to start and close a http server.
// For the options available, check [Opt].
func (c *Config) NewServer(opts ...Opt) *Server {
	r := chi.NewRouter()
	c.setDefaults()

	for _, opt := range opts {
		opt(c)
	}
	r.Use(
		c.middlewares...,
	)
	return &Server{
		config: *c,
		Router: r,
	}
}

// Server wrapper for [chi.Router]
type Server struct {
	Router chi.Router

	config Config

	closeFn func()
	closeCh chan struct{}
}

// Start is starting the listening for connections.
// The received [ctx] is used to close the server on cancellation.
//
// This method uses the [Config.Host] and [Config.Port] to start the listener. If
// these are not configured, the [net] package will allocate an available one.
//
// The call on this function is blocking.
func (r *Server) Start(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", r.config.Host, r.config.Port)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	srv := http.Server{
		Handler: r.Router,
	}

	r.closeFn = sync.OnceFunc(func() {
		defer close(r.closeCh)
		if err := srv.Close(); err != nil {
			slog.With("error", err).Info("http server closing on closeFn returned error")
		}
	})
	r.closeCh = make(chan struct{})
	go func() {
		select {
		case <-r.closeCh:
			return
		case <-ctx.Done():
			if err := srv.Close(); err != nil {
				slog.With("error", err).Info("http server closing on context.Done returned error")
			}
		}
	}()

	slog.With("addr", l.Addr().String()).Info("http server started")
	if err := srv.Serve(l); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.With("error", err).Warn("http server closed with error")
		return err
	}
	slog.Debug("http server closed gracefully")
	return nil
}

// Close is stopping the listening. If the server was not started, this
// method will do nothing.
func (r *Server) Close() {
	if r.closeFn == nil {
		slog.Debug("http server closing skipped since it was not started")
		return
	}
	slog.Info("http server closing triggered")
	r.closeFn()
}
