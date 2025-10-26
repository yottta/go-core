package app

import (
	"context"
	"fmt"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

func TestRegister(t *testing.T) {
	t.Run("panics on nil component", func(t *testing.T) {
		defer expectPanic(t, "given component is nil")
		Register(nil)
	})
	t.Run("component start returns error", func(t *testing.T) {
		const want = "error from component"
		defer expectPanic(t, want)
		Register(&mockComp{
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
		reset()
		Register(&mockComp{
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
			Stop()
		}()
		Start()

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
		reset()
		Register(&mockComp{
			startF: func() error {
				startCalled = true
				return nil
			},
			stopF: func() error {
				stopCalled = true
				return nil
			},
		})
		// setup a listening on the Context().Done() to be sure that it works as expected
		var contextClosed atomic.Bool
		contextCheckDone := make(chan struct{}, 1)
		go func() {
			<-Context().Done()
			contextClosed.Store(true)
			close(contextCheckDone)
		}()
		// After 1 second from now, simulate the system SIGINT signal
		go func() {
			<-time.After(time.Second)
			shutdownCh <- syscall.SIGINT
		}()
		Start()
		// NOTE: Do not sleep here. Start() is meant to block until everything is cleaned up
		if !startCalled {
			t.Errorf("expected to have the start function called but it wasn't")
		}
		if !stopCalled {
			t.Errorf("expected to have the stop function called but it wasn't")
		}
		ctxErr := Context().Err()
		if ctxErr == nil {
			t.Fatalf("expected context to contain an error. got nothing")
		}
		want := "context canceled"
		if got := ctxErr.Error(); got != want {
			t.Errorf("failed with a different context error.\nexpected: \n\t%s\ngot:\n\t%s", want, got)
		}
		ctxCause := context.Cause(Context())
		if ctxCause == nil {
			t.Fatalf("expected context to contain an reason. got nothing")
		}
		want = "app closing triggered by system call"
		if got := ctxCause.Error(); got != want {
			t.Fatalf("failed with a different context cause.\nexpected: \n\t%s\ngot:\n\t%s", want, got)
		}
		select {
		case _, ok := <-contextCheckDone:
			if ok {
				t.Errorf("received signal from the context closing check but the channel is still open")
				break
			}
			t.Log("context closed as expected", ok)
		case <-time.After(500 * time.Millisecond):
			t.Fatalf("context didn't close correctly")
		}
	})
}

func TestComponentErrors(t *testing.T) {
	t.Run("does not crash when stop returns error", func(t *testing.T) {
		var (
			startCalled, stopCalled bool
		)
		reset()
		Register(&mockComp{
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
			Stop()
		}()
		Start()

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
		reset()
		forcefullyTimeout = 200 * time.Millisecond
		Register(&mockComp{
			startF: func() error { startCalled = true; return nil },
			stopF: func() error {
				<-time.After(500 * time.Millisecond)
				compStoppedAt = time.Now()
				return nil
			},
		})
		go func() {
			<-time.After(time.Second)
			Stop()
			appStoppedAt = time.Now()
		}()
		Start()
		// NOTE: Do not sleep here. Start() is meant to block until everything is cleaned up
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
