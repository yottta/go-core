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

// Register initialises a [Component] calling its [Component.Start].
// If the initialisation of the [Component] returns an error, any other [Component] previously
// registered, will be cleaned up (ie: call [Component.Stop]) and will panic to stop the startup.
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

// Start is a blocking call that keeps the main goroutine from returning, allowing the other
// previously registered components to run properly.
// This method returns in only 2 cases: a system signal is received or the [Stop] is called specifically from another
// goroutine.
// The system signals that this listens for are: syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT.
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

// Stop cancels the application [context.Context] and waits for the whole application to cleanup
func Stop() {
	cancel(fmt.Errorf("app stopped"))

	select {
	case <-closingCh:
		slog.DebugContext(ctx, "app stopped successfully")
	case <-time.After(forcefullyTimeout):
		slog.With("timeout", forcefullyTimeout).WarnContext(ctx, "app stopped forcefully after timeout")
	}
}

// Context returns the context that is used to start the app.
// This is cancellable context whose [context.Done()] can be used
// to listen on the shutdown signals.
func Context() context.Context {
	return ctx
}

// reset cleans up all the registered components and recreates all the other structs
// to make this package usable again. This is mainly for testing
func reset() {
	cleanup()
	ctx, cancel = context.WithCancelCause(context.Background())
	shutdownCh = make(chan os.Signal, 1)
	closingCh = make(chan struct{}, 1)
	forcefullyTimeout = 3 * time.Second
}

// cleanup stops and successfully registered [Component].
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

// exit is just aa utility function that combines [cleanup] with a panic.
func exit(err error) {
	cleanup()
	panic(err)
}
