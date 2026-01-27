package chix

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestServerStartStop(t *testing.T) {
	t.Run("starts and stops gracefully", func(t *testing.T) {
		cfg := &Config{
			Host: "localhost",
			Port: 0,
		}
		srv := cfg.NewServer()

		srv.Router().Get("/ping", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("pong"))
		})

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		errCh := make(chan error, 1)
		go func() {
			errCh <- srv.Start(ctx)
		}()

		<-time.After(100 * time.Millisecond)

		cancel()

		select {
		case err := <-errCh:
			if err != nil {
				t.Errorf("expected no error on graceful shutdown, got: %v", err)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("server did not shut down in time")
		}
	})

	t.Run("stops via Close method", func(t *testing.T) {
		cfg := &Config{
			Host: "localhost",
			Port: 0,
		}
		srv := cfg.NewServer()

		ctx := context.Background()
		errCh := make(chan error, 1)
		go func() {
			errCh <- srv.Start(ctx)
		}()

		<-time.After(100 * time.Millisecond)

		srv.Close()

		select {
		case err := <-errCh:
			if err != nil {
				t.Errorf("expected no error on Close, got: %v", err)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("server did not shut down in time")
		}
	})

	t.Run("Close before Start does nothing", func(t *testing.T) {
		cfg := &Config{
			Host: "localhost",
			Port: 0,
		}
		srv := cfg.NewServer()

		srv.Close()
	})

	t.Run("handles requests correctly", func(t *testing.T) {
		cfg := &Config{
			Host: "localhost",
			Port: 1234,
		}
		srv := cfg.NewServer()

		srv.Router().Get("/test", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("test response"))
		})

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		errCh := make(chan error, 1)
		go func() {
			errCh <- srv.Start(ctx)
		}()

		<-time.After(100 * time.Millisecond)

		resp, err := http.Get(fmt.Sprintf("http://localhost:%d/test", cfg.Port))
		if err != nil {
			t.Fatal("server failed to answer to requests")
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal("failed to read the response from the request on the server")
		}
		if string(body) != "test response" {
			t.Errorf("expected 'test response', got '%s'", string(body))
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		cancel()

		select {
		case <-errCh:
		case <-time.After(2 * time.Second):
			t.Fatal("server did not shut down in time")
		}
	})

	t.Run("fails when port is already in use", func(t *testing.T) {
		cfg := &Config{
			Host: "localhost",
			Port: 2344,
		}
		srv1 := cfg.NewServer()
		srv2 := cfg.NewServer()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var srv1Err, srv2Err error
		var wg sync.WaitGroup
		wg.Go(func() {
			srv1Err = srv1.Start(ctx)
		})
		wg.Go(func() {
			srv2Err = srv2.Start(ctx)
		})
		<-time.After(200 * time.Millisecond)
		cancel()
		wg.Wait()
		<-time.After(100 * time.Millisecond)
		if srv1Err == nil && srv2Err == nil {
			t.Fatalf("no server started or none failed")
		}
		if srv1Err != nil && srv2Err != nil {
			t.Fatalf("both servers failed which is not expected")
		}
		expected := "address already in use"
		if srv1Err != nil && !strings.Contains(srv1Err.Error(), expected) {
			t.Errorf("expected error to contain %q but got %q", expected, srv1Err.Error())
		}
		if srv2Err != nil && !strings.Contains(srv2Err.Error(), expected) {
			t.Errorf("expected error to contain %q but got %q", expected, srv2Err.Error())
		}
	})
	t.Run("calling Router() after Start() panics", func(t *testing.T) {
		cfg := &Config{
			Host: "localhost",
			Port: 0,
		}
		srv := cfg.NewServer()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		errCh := make(chan error, 1)
		go func() {
			errCh <- srv.Start(ctx)
		}()

		<-time.After(100 * time.Millisecond)

		defer func() {
			const expectedPanicContent = "server already started, cannot configure the router anymore"
			var panicErrContent any
			if r := recover(); r != nil {
				panicErrContent = r
			}

			if panicErrContent != expectedPanicContent {
				t.Errorf("invalid panic content. expected %q but got %q", expectedPanicContent, panicErrContent)
			}
			cancel()
			select {
			case err := <-errCh:
				if err != nil {
					t.Errorf("expected no error on graceful shutdown, got: %v", err)
				}
			case <-time.After(2 * time.Second):
				t.Fatal("server did not shut down in time")
			}
		}()
		srv.Router().Get("/ping", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("pong"))
		})
	})
}
