package app

import (
	"context"
	"fmt"
	"syscall"
	"testing"
	"time"
)

func TestRegister(t *testing.T) {
	t.Run("panics on nil component", func(t *testing.T) {
		defer expectPanic(t, "given component is nil")
		app, _ := New()
		app.Register(nil)
	})
	t.Run("component start returns error", func(t *testing.T) {
		const want = "error from component"
		defer expectPanic(t, want)
		app, _ := New()
		app.Register(&mockComp{
			startF: func() error {
				return fmt.Errorf(want)
			},
			stopF: nil,
		})
	})
}

func TestStartStop(t *testing.T) {
	t.Run("start and stop with the given methods", func(t *testing.T) {

		var (
			startCalled, stopCalled bool
		)
		app, ctx := New()
		app.Register(&mockComp{
			startF: func() error {
				startCalled = true
				return nil
			},
			stopF: func() error {
				stopCalled = true
				return nil
			},
		})
		go func() {
			<-time.After(time.Second)
			app.Stop()
		}()
		app.Start()

		if !startCalled {
			t.Errorf("expected to have the start function called but it wasn't")
		}
		if !stopCalled {
			t.Errorf("expected to have the stop function called but it wasn't")
		}
		ctxErr := ctx.Err()
		if ctxErr == nil {
			t.Fatalf("expected context to contain an error. got nothing")
		}
		want := "context canceled"
		if got := ctxErr.Error(); got != want {
			t.Errorf("failed with a different context error.\nexpected: \n\t%s\ngot:\n\t%s", want, got)
		}
		ctxCause := context.Cause(ctx)
		if ctxCause == nil {
			t.Fatalf("expected context to contain an reason. got nothing")
		}
		want = "app stopped"
		if got := ctxCause.Error(); got != want {
			t.Fatalf("failed with a different context cause.\nexpected: \n\t%s\ngot:\n\t%s", want, got)
		}
	})
	t.Run("shutting down on system call", func(t *testing.T) {
		var (
			startCalled, stopCalled bool
		)
		app, ctx := New()
		app.Register(&mockComp{
			startF: func() error {
				startCalled = true
				return nil
			},
			stopF: func() error {
				stopCalled = true
				return nil
			},
		})
		go func() {
			<-time.After(time.Second)
			app.shutdownCh <- syscall.SIGINT
		}()
		app.Start()
		// NOTE: Do not sleep here. app.Start() is meant to block until everything is cleaned up
		if !startCalled {
			t.Errorf("expected to have the start function called but it wasn't")
		}
		if !stopCalled {
			t.Errorf("expected to have the stop function called but it wasn't")
		}
		ctxErr := ctx.Err()
		if ctxErr == nil {
			t.Fatalf("expected context to contain an error. got nothing")
		}
		want := "context canceled"
		if got := ctxErr.Error(); got != want {
			t.Errorf("failed with a different context error.\nexpected: \n\t%s\ngot:\n\t%s", want, got)
		}
		ctxCause := context.Cause(ctx)
		if ctxCause == nil {
			t.Fatalf("expected context to contain an reason. got nothing")
		}
		want = "app closing triggered by system call"
		if got := ctxCause.Error(); got != want {
			t.Fatalf("failed with a different context cause.\nexpected: \n\t%s\ngot:\n\t%s", want, got)
		}
	})
}

func TestComponentErrors(t *testing.T) {
	t.Run("does not crash when stop returns error", func(t *testing.T) {
		var (
			startCalled, stopCalled bool
		)
		app, _ := New()
		app.Register(&mockComp{
			startF: func() error {
				startCalled = true
				return nil
			},
			stopF: func() error {
				stopCalled = true
				return fmt.Errorf("failed to stop")
			},
		})
		go func() {
			<-time.After(time.Second)
			app.Stop()
		}()
		app.Start()

		if !startCalled {
			t.Errorf("expected to have the start function called but it wasn't")
		}
		if !stopCalled {
			t.Errorf("expected to have the stop function called but it wasn't")
		}
	})
	t.Run("when component.Stop takes too much time, app.Stop returns before component.Stop", func(t *testing.T) {
		var (
			startCalled   bool
			compStoppedAt time.Time
			appStoppedAt  time.Time
		)
		app, ctx := New()
		app.forcefullyTimeout = 200 * time.Millisecond
		app.Register(&mockComp{
			startF: func() error { startCalled = true; return nil },
			stopF: func() error {
				<-time.After(500 * time.Millisecond)
				compStoppedAt = time.Now()
				return nil
			},
		})
		go func() {
			<-time.After(time.Second)
			app.Stop()
			appStoppedAt = time.Now()
		}()
		app.Start()
		// NOTE: Do not sleep here. app.Start() is meant to block until everything is cleaned up
		if !startCalled {
			t.Errorf("expected to have the start function called but it wasn't")
		}

		select {
		case <-ctx.Done():
		case <-time.After(5 * time.Second):
			t.Fatalf("expected the app to fail and close the channel")
		}
		if compStoppedAt.Compare(appStoppedAt) <= 0 {
			t.Fatalf("expected the component to finish after the app because of the timeout")
		}
	})
}

func expectPanic(t *testing.T, want string) {
	r := recover()
	if r == nil {
		t.Fatalf("test didn't fail as it was expected to")
	}
	got := fmt.Sprintf("%s", r)
	if got != want {
		t.Fatalf("failed with a different error.\nexpected: \n\t%s\ngot:\n\t%s", want, got)
	}
}

type mockComp struct {
	startF, stopF func() error
}

func (m mockComp) String() string {
	return "mockComp"
}

func (m mockComp) Start() error {
	return m.startF()
}

func (m mockComp) Stop() error {
	return m.stopF()
}
