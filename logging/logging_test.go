package logging

import (
	"bytes"
	"fmt"
	"log/slog"
	"strings"
	"testing"
)

func TestSetup(t *testing.T) {
	t.Run("level tests", func(t *testing.T) {
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
			return true, true, true, true // same as debug
		}
		for _, lvl := range []string{"definitely_not_acceptable_log_level", "debug", "info", "warn", "error"} {
			t.Run(lvl, func(t *testing.T) {
				t.Setenv("LOG_LEVEL", lvl)
				var b bytes.Buffer
				setupWithWriter(&b)
				writeAllLevelLogs()
				wantD, wantI, wantW, wantE := genWantErrors(lvl)
				assertLogs(t, b.String(), wantD, wantI, wantW, wantE)
			})
		}
	})
	t.Run("format tests", func(t *testing.T) {
		t.Run("text", func(t *testing.T) {
			t.Setenv("LOG_FORMAT", "text")
			var b bytes.Buffer
			setupWithWriter(&b)
			writeAllLevelLogs()
			t.Logf("content: %s", b.String())
			if content := b.String(); strings.Contains(content, "{") {
				t.Errorf("generated logs seems to contain json content but it shouldn't. content: %s", content)
			}
		})
		t.Run("json", func(t *testing.T) {
			t.Setenv("LOG_FORMAT", "json")
			var b bytes.Buffer
			setupWithWriter(&b)
			writeAllLevelLogs()
			t.Logf("content: %s", b.String())
			if content := b.String(); !strings.Contains(content, "{") {
				t.Errorf("generated logs seems to contain json content but it shouldn't. content: %s", content)
			}
		})
		t.Run("wrong format", func(t *testing.T) {
			t.Setenv("LOG_FORMAT", "wrong")
			var b bytes.Buffer
			setupWithWriter(&b)
			writeAllLevelLogs()
			t.Logf("content: %s", b.String())
			if content := b.String(); strings.Contains(content, "{") {
				t.Errorf("generated logs seems to contain json content but it shouldn't. content: %s", content)
			}
		})
	})

	t.Run("sources tests", func(t *testing.T) {
		t.Run("w/o source", func(t *testing.T) {
			t.Setenv("LOG_SOURCE", "false")
			var b bytes.Buffer
			setupWithWriter(&b)
			writeAllLevelLogs()
			t.Logf("content: %s", b.String())
			if content := b.String(); strings.Contains(content, "source=") {
				t.Errorf("generated logs seems to contain json content but it shouldn't. content: %s", content)
			}
		})
		t.Run("with source", func(t *testing.T) {
			t.Setenv("LOG_SOURCE", "true")
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
