package logging

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"testing"
)

func TestSetup(t *testing.T) {
	t.Run("level tests", func(t *testing.T) {
		defer func() { _ = os.Unsetenv("LOG_LEVEL") }()
		// returns bool for wantDebug, wantInfo, wantWarn, wantError
		genWantErrors := func(lvl string) (bool, bool, bool, bool) {
			switch lvl {
			case "debug":
				return true, true, true, true
			case "info":
				return false, true, true, true
			case "warn":
				return false, false, true, true
			case "error":
				return false, false, false, true
			}
			return false, false, false, false
		}
		for _, lvl := range []string{"debug", "info", "warn", "error"} {
			t.Run(lvl, func(t *testing.T) {
				if err := os.Setenv("LOG_LEVEL", lvl); err != nil {
					t.Fatalf("could not set the LOG_FORMAT env var")
				}
				var b bytes.Buffer
				setupWithWriter(&b)
				writeAllLevelLogs()
				wantD, wantI, wantW, wantE := genWantErrors(lvl)
				assertLogs(t, b.String(), wantD, wantI, wantW, wantE)
			})
		}
	})
	t.Run("format tests", func(t *testing.T) {
		defer func() { _ = os.Unsetenv("LOG_FORMAT") }()
		t.Run("text", func(t *testing.T) {
			if err := os.Setenv("LOG_FORMAT", "text"); err != nil {
				t.Fatalf("could not set the LOG_FORMAT env var")
			}
			var b bytes.Buffer
			setupWithWriter(&b)
			writeAllLevelLogs()
			t.Logf("content: %s", b.String())
			if content := b.String(); strings.Contains(content, "{") {
				t.Errorf("generated logs seems to contain json content but it shouldn't. content: %s", content)
			}
		})
		t.Run("json", func(t *testing.T) {
			if err := os.Setenv("LOG_FORMAT", "json"); err != nil {
				t.Fatalf("could not set the LOG_FORMAT env var")
			}
			var b bytes.Buffer
			setupWithWriter(&b)
			writeAllLevelLogs()
			t.Logf("content: %s", b.String())
			if content := b.String(); !strings.Contains(content, "{") {
				t.Errorf("generated logs seems to contain json content but it shouldn't. content: %s", content)
			}
		})
	})

	t.Run("sources tests", func(t *testing.T) {
		defer func() { _ = os.Unsetenv("LOG_SOURCE") }()
		t.Run("w/o source", func(t *testing.T) {
			if err := os.Setenv("LOG_SOURCE", "false"); err != nil {
				t.Fatalf("could not set the LOG_SOURCE env var")
			}
			var b bytes.Buffer
			setupWithWriter(&b)
			writeAllLevelLogs()
			t.Logf("content: %s", b.String())
			if content := b.String(); strings.Contains(content, "source=") {
				t.Errorf("generated logs seems to contain json content but it shouldn't. content: %s", content)
			}
		})
		t.Run("with source", func(t *testing.T) {
			if err := os.Setenv("LOG_SOURCE", "true"); err != nil {
				t.Fatalf("could not set the LOG_SOURCE env var")
			}
			var b bytes.Buffer
			setupWithWriter(&b)
			writeAllLevelLogs()
			t.Logf("content: %s", b.String())
			if content := b.String(); !strings.Contains(content, "source=") {
				t.Errorf("generated logs seems to contain json content but it shouldn't. content: %s", content)
			}
		})
	})
}

func writeAllLevelLogs() {
	slog.Debug("debug log here")
	slog.Info("info log here")
	slog.Warn("warn log here")
	slog.Error("error log here")
}

func assertLogs(t *testing.T, output string, debug, info, warn, error bool) {
	checkLogLevel := func(t *testing.T, expectLvl bool, wantLvl string) {
		contains := strings.Contains(output, fmt.Sprintf("%s log here", wantLvl))
		if expectLvl && !contains {
			t.Errorf("expected to contain %s but it didn't", wantLvl)
		}
		if !expectLvl && contains {
			t.Errorf("expected to not contain %s but it does", wantLvl)
		}
	}
	checkLogLevel(t, debug, "debug")
	checkLogLevel(t, info, "info")
	checkLogLevel(t, warn, "warn")
	checkLogLevel(t, error, "error")
}
