package httpx

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
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

		srv.Router.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
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

		srv.Router.Get("/test", func(w http.ResponseWriter, r *http.Request) {
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

		if srv.closeCh == nil {
			t.Fatal("server did not start properly")
		}

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
			Port: 2345,
		}
		srv1 := cfg.NewServer()
		srv2 := cfg.NewServer()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		errCh1 := make(chan error, 1)
		errCh2 := make(chan error, 1)
		go func() {
			errCh1 <- srv1.Start(ctx)
		}()
		go func() {
			errCh2 <- srv2.Start(ctx)
		}()
		<-time.After(100 * time.Millisecond)
		cancel()
		<-time.After(100 * time.Millisecond)
		close(errCh1)
		close(errCh2)
		e, ok := <-errCh1
		e2, ok2 := <-errCh2
		t.Logf("server1 status: %s and %t", e, ok)
		t.Logf("server2 status: %s and %t", e2, ok2)
		if e == nil && e2 == nil {
			t.Fatalf("no server started or none failed")
		}
		if e != nil && e2 != nil {
			t.Fatalf("both servers failed which is not expected")
		}
		expected := "address already in use"
		if e != nil && !strings.Contains(e.Error(), expected) {
			t.Errorf("expected error to contain %q but got %q", expected, e.Error())
		}
		if e2 != nil && !strings.Contains(e2.Error(), expected) {
			t.Errorf("expected error to contain %q but got %q", expected, e2.Error())
		}
	})
}
