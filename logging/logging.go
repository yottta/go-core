package logging

import (
	"io"
	"log/slog"
	"os"

	"github.com/yottta/go-core/env"
)

// Config is a utility struct that can be injected in applications to map the args/env vars directly to it
// and call [Config.Setup] to configure the logging of the application.
type Config struct {
	// The stream where the logs will be written
	OutStream io.Writer
	// LogLevel: vals: debug, info, warn, error. This is controlling the logging level. Default: debug
	LogLevel string
	// Format: vals: text, json. This is controlling the format of the logs. Default: text
	LogFormat string
	// AddSource: true, false. This is controlling to include or not the sources of the logs. Default: false
	LogSource bool
}

func (c Config) Setup() {
	setupWithWriter(c.OutStream, c.LogLevel, c.LogFormat, c.LogSource)
}

// Setup is setting up slog with different options
// This is handling the following env vars:
// * LOG_LEVEL: vals: debug, info, warn, error. This is controlling the logging level. Default: debug
// * LOG_FORMAT: vals: text, json. This is controlling the format of the logs. Default: text
// * LOG_SOURCE: true, false. This is controlling to include or not the sources of the logs. Default: false
func Setup() {
	level := env.StringWithDefault("LOG_LEVEL", "debug")
	format := env.StringWithDefault("LOG_FORMAT", "text")
	addSource := env.BoolWithDefault("LOG_SOURCE", false)
	setupWithWriter(os.Stderr, level, format, addSource)
}

// Setup is setting up slog with different options
// This is handling the following env vars:
// * level: vals: debug, info, warn, error. This is controlling the logging level. Default: debug
// * format: vals: text, json. This is controlling the format of the logs. Default: text
// * addSource: true, false. This is controlling to include or not the sources of the logs. Default: false
func SetupWithArgs(level string, format string, addSource bool) {
	setupWithWriter(os.Stderr, level, format, addSource)
}

// setupWithWriter is mainly created for testing
func setupWithWriter(w io.Writer, level string, format string, addSource bool) {
	lvl := &slog.LevelVar{}
	err := lvl.UnmarshalText([]byte(level))
	if err != nil {
		lvl.Set(slog.LevelDebug)
	}

	opts := slog.HandlerOptions{
		AddSource: addSource,
		Level:     lvl,
	}
	var h slog.Handler
	switch format {
	case "text":
		h = slog.NewTextHandler(w, &opts)
	case "json":
		h = slog.NewJSONHandler(w, &opts)
	default:
		h = slog.NewTextHandler(w, &opts)
	}
	slog.SetDefault(slog.New(h))
}
