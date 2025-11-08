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

type Config struct {
	Host string
	Port int
}

func (c *Config) NewServer() *Server {
	return &Server{
		config: *c,
		Router: chi.NewRouter(),
	}
}

type Server struct {
	Router chi.Router

	config Config

	closeFn func()
	closeCh chan struct{}
}

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

func (r *Server) Close() {
	if r.closeFn == nil {
		slog.Info("http server closing skipped since it was not started")
		return
	}
	slog.Info("http server closing triggered")
	r.closeFn()
}
