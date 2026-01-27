package chix

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/yottta/go-core/shutdown"
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
		router: r,
	}
}

// Server wrapper for [chi.Router]
type Server struct {
	router chi.Router

	config Config

	ctx     context.Context
	closeFn func()

	started  bool
	startedM sync.Mutex
}

// Start is starting the listening for connections.
// The received [ctx] is used to close the server on cancellation.
//
// This method uses the [Config.Host] and [Config.Port] to start the listener. If
// these are not configured, the [net] package will allocate an available one.
//
// The call on this function is blocking.
func (r *Server) Start(ctx context.Context) error {
	var srv http.Server
	var cancel context.CancelFunc
	var l net.Listener
	var err error
	configure := func() { // anonymous function for locking
		r.startedM.Lock()
		defer r.startedM.Unlock()
		// No need to defer this cancel since this will be called in [Server.Close] or the cancel
		// will be canceled when a sys signal will be issued.
		ctx, cancel = shutdown.Context(ctx)
		r.closeFn = cancel

		addr := fmt.Sprintf("%s:%d", r.config.Host, r.config.Port)
		l, err = net.Listen("tcp", addr)
		if err != nil {
			return
		}

		r.started = true
		srv = http.Server{
			Handler: r.router,
		}
	}
	configure()
	if err != nil {
		return err
	}

	go func() {
		select {
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
	r.startedM.Lock()
	defer r.startedM.Unlock()
	if !r.started {
		return
	}
	slog.Info("http server closing triggered")
	r.closeFn()
}

// Router returns the inner router to allow configuration of routes.
// Calling this method after [Server.Start] has been called, will panic.
func (r *Server) Router() chi.Router {
	r.startedM.Lock()
	defer r.startedM.Unlock()
	if r.started {
		panic("server already started, cannot configure the router anymore")
	}
	return r.router
}
