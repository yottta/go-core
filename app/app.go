package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	components []Component

	ctx        context.Context
	cancel     context.CancelCauseFunc
	shutdownCh chan os.Signal
	closingCh  chan struct{}

	forcefullyTimeout time.Duration
)

func init() {
	reset()
}

func reset() {
	ctx, cancel = context.WithCancelCause(context.Background())
	shutdownCh = make(chan os.Signal, 1)
	closingCh = make(chan struct{}, 1)
	forcefullyTimeout = 3 * time.Second
}

type Component interface {
	fmt.Stringer
	Start() error
	Stop() error
}

func cleanup() {
	for _, c := range components {
		if err := c.Stop(); err != nil {
			slog.
				With("error", err).
				With("component", c.String()).
				WarnContext(ctx, "stop error encountered during closing component")
		}
	}
	components = nil
}

func exit(err error) {
	cleanup()
	panic(err)
}

func Register(c Component) {
	if c == nil {
		exit(fmt.Errorf("given component is nil"))
		return
	}
	err := c.Start()
	if err != nil {
		exit(err)
	}
	slog.
		With("component", c.String()).
		DebugContext(ctx, "component registered successfully")
	components = append(components, c)
}

func Start() {
	signal.Notify(shutdownCh,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	defer func() {
		cleanup()
		close(closingCh)
	}()
	slog.InfoContext(ctx, "started...")
	select {
	case <-shutdownCh:
		cancel(fmt.Errorf("app closing triggered by system call"))
	case <-ctx.Done():
		slog.DebugContext(ctx, "app closing triggered from inside the app")
	}
}

func Stop() {
	cancel(fmt.Errorf("app stopped"))

	select {
	case <-closingCh:
		slog.DebugContext(ctx, "app stopped successfully")
	case <-time.After(forcefullyTimeout):
		slog.With("timeout", forcefullyTimeout).WarnContext(ctx, "app stopped forcefully after timeout")
	}
}
