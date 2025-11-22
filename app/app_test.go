package app

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"
)

func TestRegister(t *testing.T) {
	t.Run("panics on nil component", func(t *testing.T) {
		defer expectPanic(t, "given component is nil")
		a := New()
		a.Register(nil)
	})
	t.Run("component start returns error", func(t *testing.T) {
		const want = "error from component"
		defer expectPanic(t, want)
		a := New()
		a.Register(&mockComp{
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
		a := New()
		a.Register(&mockComp{
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
			a.Stop()
		}()
		a.Start()

		if !startCalled {
			t.Errorf("expected to have the start function called but it wasn't")
		}
		if !stopCalled {
			t.Errorf("expected to have the stop function called but it wasn't")
		}
		ctxErr := a.Context().Err()
		if ctxErr == nil {
			t.Fatalf("expected context to contain an error. got nothing")
		}
		want := "context canceled"
		if got := ctxErr.Error(); got != want {
			t.Errorf("failed with a different context error.\nexpected: \n\t%s\ngot:\n\t%s", want, got)
		}
		ctxCause := context.Cause(a.ctx)
		if ctxCause == nil {
			t.Fatalf("expected context to contain an reason. got nothing")
		}
		want = "app stopped"
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
		a := New()
		a.Register(&mockComp{
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
			a.Stop()
		}()
		a.Start()

		if !startCalled {
			t.Errorf("expected to have the start function called but it wasn't")
		}
		if !stopCalled {
			t.Errorf("expected to have the stop function called but it wasn't")
		}
	})
	t.Run("when component.Stop takes too much time, app.Stop returns before component.Stop", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			var (
				startCalled   bool
				compStoppedAt atomic.Pointer[time.Time]
				appStoppedAt  atomic.Pointer[time.Time]
			)
			a := New()
			a.Register(&mockComp{
				startF: func() error { startCalled = true; return nil },
				stopF: func() error {
					<-time.After(5 * time.Second) // longer than the forcefullyTimeout
					now := time.Now()
					compStoppedAt.Store(&now)
					return nil
				},
			})
			go func() {
				<-time.After(time.Second)
				a.Stop()
				now := time.Now()
				appStoppedAt.Store(&now)
			}()
			synctest.Wait()
			a.Start()
			// NOTE: Do not sleep here. Start() is meant to block until everything is cleaned up
			if !startCalled {
				t.Errorf("expected to have the start function called but it wasn't")
			}

			select {
			case <-a.Context().Done():
			case <-time.After(5 * time.Second):
				t.Fatalf("expected the app to fail and close the channel")
			}
			compStoppedAtTime := compStoppedAt.Load()
			appStoppedAtTime := appStoppedAt.Load()
			if compStoppedAtTime.Compare(*appStoppedAtTime) <= 0 {
				t.Fatalf("expected the component to finish after the app because of the timeout")
			}
		})
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
