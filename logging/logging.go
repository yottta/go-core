package logging

import (
	"io"
	"log/slog"
	"os"

	"github.com/yottta/reno-core/env"
)

// Setup is setting up slog with different options
// This is handling the following env vars:
// * LOG_LEVEL: vals: debug, info, warn, error. This is controlling the logging level. Default: debug
// * LOG_FORMAT: vals: text, json. This is controlling the format of the logs. Default: text
// * LOG_SOURCE: true, false. This is controlling to include or not the sources of the logs. Default: false
func Setup() {
	setupWithWriter(os.Stderr)
}

// setupWithWriter is mainly created for testing
func setupWithWriter(w io.Writer) {
	level := env.StringWithDefault("LOG_LEVEL", "debug")
	format := env.StringWithDefault("LOG_FORMAT", "text")
	addSource := env.BoolWithDefault("LOG_SOURCE", false)

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
