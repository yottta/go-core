package app

import (
	"context"
	"fmt"
	"log/slog"
	"syscall"
	"time"

	"github.com/yottta/go-core/shutdown"
)

// Component sets the contract for any construct that wants to be controller by the startup and the shutdown of the
// whole application.
// By using [Register], the construct will be initialized by calling [Start] on it and if the error occurs it breaks
// the startup considering it a bad configuration.
// The [Stop] can be used to cleanup and connections or opened resources before shutting down the whole application.
type Component interface {
	fmt.Stringer
	Start() error
	Stop() error
}

type App struct {
	components []Component

	ctx       context.Context
	cancel    context.CancelCauseFunc
	closingCh chan struct{}

	forcefullyTimeout time.Duration
}

func New() *App {
	ctx, cancel := context.WithCancelCause(context.Background())
	return &App{
		ctx:               ctx,
		cancel:            cancel,
		closingCh:         make(chan struct{}, 1),
		forcefullyTimeout: 3 * time.Second,
	}
}

// Register initialises a [Component] calling its [Component.Start].
// If the initialisation of the [Component] returns an error, any other [Component] previously
// registered, will be cleaned up (ie: call [Component.Stop]) and will panic to stop the startup.
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
		Debug("component registered successfully")
	a.components = append(a.components, c)
}

// Start is a blocking call that keeps the main goroutine from returning, allowing the other
// previously registered components to run properly.
// This method returns in only 2 cases: a system signal is received or the [Stop] is called specifically from another
// goroutine.
// The system signals that this listens for are: syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT.
func (a *App) Start() {
	ctx, cancel := shutdown.Context(a.ctx, syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer cancel()

	defer func() {
		a.cleanup()
		close(a.closingCh)
	}()
	slog.Info("started...")
	select {
	case <-ctx.Done():
		slog.Debug("app closing triggered")
	}
}

// Stop cancels the application [context.Context] and waits for the whole application to cleanup
func (a *App) Stop() {
	a.cancel(fmt.Errorf("app stopped"))

	select {
	case <-a.closingCh:
		slog.Debug("app stopped successfully")
	case <-time.After(a.forcefullyTimeout):
		slog.With("timeout", a.forcefullyTimeout).Warn("app stopped forcefully after timeout")
	}
}

// Context returns the context that is used to start the app.
// This is cancellable context whose [context.Done()] can be used
// to listen on the shutdown signals.
func (a *App) Context() context.Context {
	return context.WithValue(a.ctx, "", "")
}

// cleanup stops and successfully registered [Component].
func (a *App) cleanup() {
	for _, c := range a.components {
		if err := c.Stop(); err != nil {
			slog.
				With("error", err).
				With("component", c.String()).
				Warn("stop error encountered during closing component")
		}
	}
	a.components = nil
}

// exit is just a utility function that combines [cleanup] with a panic.
func (a *App) exit(err error) {
	a.cleanup()
	panic(err)
}
