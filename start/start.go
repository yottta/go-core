package start

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

type App struct {
	c []Component

	ctx    context.Context
	cancel context.CancelCauseFunc

	internalCtx context.Context
}

type Component interface {
	fmt.Stringer
	Start() error
	Stop() error
}

func New() *App {
	ctx, cancel := context.WithCancelCause(context.Background())
	return &App{
		ctx:    ctx,
		cancel: cancel,
	}
}

func (a *App) Register(c Component) {
	if c == nil {
		panic("given component is nil")
	}
	err := c.Start()
	if err != nil {
		panic(err)
	}
	// TODO add logging
	a.c = append(a.c, c)
}

func (a *App) Start() context.Context {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	go func() {
		defer func() { // TODO call components.stop
		}()
		select {
		case <-sigc:
			a.cancel(fmt.Errorf("app closing triggered"))
		case <-a.ctx.Done():
			// TODO logging
		}

	}()

	return a.ctx
}

func (a *App) Stop() {

	a.cancel(fmt.Errorf("app stopped"))
}
