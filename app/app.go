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

type App struct {
	c []Component

	ctx        context.Context
	cancel     context.CancelCauseFunc
	shutdownCh chan os.Signal
	closingCh  chan struct{}

	internalCtx context.Context

	forcefullyTimeout time.Duration
}

type Component interface {
	fmt.Stringer
	Start() error
	Stop() error
}

// New returns also the context to make the user aware of its existence and encourage on using it for better
// app state management.
func New() (*App, context.Context) {
	ctx, cancel := context.WithCancelCause(context.Background())
	return &App{
		ctx:    ctx,
		cancel: cancel,

		shutdownCh:        make(chan os.Signal, 1),
		closingCh:         make(chan struct{}, 1),
		forcefullyTimeout: 3 * time.Second,
	}, ctx
}

func (a *App) cleanup() {
	for _, c := range a.c {
		if err := c.Stop(); err != nil {
			slog.
				With("error", err).
				With("component", c.String()).
				WarnContext(a.ctx, "stop error encountered during closing component")
		}
	}
	a.c = nil
}

func (a *App) exit(err error) {
	a.cleanup()
	panic(err)
}

func (a *App) Register(c Component) {
	if c == nil {
		a.exit(fmt.Errorf("given component is nil"))
		return
	}
	err := c.Start()
	if err != nil {
		a.exit(err)
	}
	slog.
		With("component", c.String()).
		DebugContext(a.ctx, "component registered successfully")
	a.c = append(a.c, c)
}

func (a *App) Start() {
	signal.Notify(a.shutdownCh,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	defer func() {
		a.cleanup()
		close(a.closingCh)
	}()
	slog.InfoContext(a.ctx, "started...")
	select {
	case <-a.shutdownCh:
		a.cancel(fmt.Errorf("app closing triggered by system call"))
	case <-a.ctx.Done():
		slog.DebugContext(a.ctx, "app closing triggered from inside the app")
	}
}

func (a *App) Stop() {
	a.cancel(fmt.Errorf("app stopped"))

	select {
	case <-a.closingCh:
		slog.DebugContext(a.ctx, "app stopped successfully")
	case <-time.After(a.forcefullyTimeout):
		slog.With("timeout", a.forcefullyTimeout).WarnContext(a.ctx, "app stopped forcefully after timeout")
	}
}
