package httpx

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
)

// Config can be embedded in your configs and map flags and env vars directly to the
// [Config.Host] and [Config.Port] attributes.
//
// With the [Config.NewServer] a new [*Server] will be returned to handle an http
// handler.
type Config struct {
	Host string
	Port int
}

// Start is starting the listening for connections.
// The received [ctx] is used to close the server on cancellation.
//
// This method uses the [Config.Host] and [Config.Port] to start the listener. If
// these are not configured, the [net] package will allocate an available one.
//
// The call on this function is blocking.
func (c *Config) Start(ctx context.Context, h http.Handler) error {
	var srv http.Server
	var cancel context.CancelFunc
	var l net.Listener
	var err error
	ctx, cancel = context.WithCancel(ctx)
	defer cancel()

	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)
	l, err = net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	srv = http.Server{
		Handler: h,
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
