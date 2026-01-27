package chix

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

func TestWithPreMiddleware(t *testing.T) {
	c := configWithDefaults(t)
	c.NewServer(WithPreMiddleware(func(handler http.Handler) http.Handler {
		return middleware.Recoverer(handler)
	}))
	want := 4
	if got := len(c.middlewares); got != want {
		t.Fatalf("expected the config to have %d middlewares but got %d", want, got)
	}
}

func TestWithPostMiddleware(t *testing.T) {
	c := configWithDefaults(t)
	c.NewServer(WithPostMiddleware(func(handler http.Handler) http.Handler {
		return middleware.Recoverer(handler)
	}))
	want := 4
	if got := len(c.middlewares); got != want {
		t.Fatalf("expected the config to have %d middlewares but got %d", want, got)
	}
}

func TestWithMiddlewares(t *testing.T) {
	c := configWithDefaults(t)
	c.NewServer(WithMiddlewares(func(handler http.Handler) http.Handler {
		return middleware.Recoverer(handler)
	}))
	want := 1
	if got := len(c.middlewares); got != want {
		t.Fatalf("expected the config to have %d middlewares but got %d", want, got)
	}
}

func TestFullMiddlewares(t *testing.T) {
	newMiddleware := func(position int) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			fn := func(w http.ResponseWriter, r *http.Request) {
				ctx := context.WithValue(r.Context(), position, time.Now())
				<-time.After(10 * time.Nanosecond)
				next.ServeHTTP(w, r.WithContext(ctx))
			}
			return http.HandlerFunc(fn)
		}
	}
	c := &Config{}
	s := c.NewServer(
		// Overwrite the default middlewares
		WithMiddlewares(newMiddleware(3), newMiddleware(4)),
		WithPreMiddleware(newMiddleware(2)),
		WithPreMiddleware(newMiddleware(1)),
		WithPostMiddleware(newMiddleware(5)),
		WithPostMiddleware(newMiddleware(6)),
	)

	if got, want := len(c.middlewares), 6; got != want {
		t.Fatalf("expected the config to have %d middlewares but got %d", want, got)
	}
	handle := s.Router().Middlewares().HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		data := map[int]time.Time{}
		for i := 1; i <= 6; i++ {
			v := request.Context().Value(i)
			tm, ok := v.(time.Time)
			if !ok {
				t.Fatalf("middleware %d saved wrong value: %#v", i, v)
			}
			data[i] = tm
		}
		for i := 1; i < 6; i++ {
			t1 := data[i]
			t2 := data[i+1]
			if !t1.Before(t2) {
				t.Fatalf("wrong order of execution of the middlewares. %d executed after %d and expected to be before", i, i+1)
			}
		}
	})
	handle.ServeHTTP(&httptest.ResponseRecorder{}, &http.Request{})
}

func configWithDefaults(t *testing.T) *Config {
	c := &Config{}
	c.setDefaults()
	expectedNoOfDefault := 3
	if got := len(c.middlewares); got != expectedNoOfDefault {
		t.Fatalf("expected the config to have %d middlewares but got %d", expectedNoOfDefault, got)
	}
	return c
}
