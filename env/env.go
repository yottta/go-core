package env

import (
	"log/slog"
	"os"
	"strconv"
)

func Expand(v string) string {
	return os.ExpandEnv(v)
}

func StringWithDefault(k string, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}

func String(k string) string {
	return StringWithDefault(k, "")
}

func BoolWithDefault(k string, def bool) bool {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	val, err := strconv.ParseBool(v)
	if err != nil {
		slog.With("key", k).Warn("env var not a bool")
		return def
	}
	return val
}

func Bool(k string) bool {
	return BoolWithDefault(k, false)
}

func IntWithDefault(k string, def int) int {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	val, err := strconv.Atoi(v)
	if err != nil {
		slog.With("key", k).Warn("env var not an int")
		return def
	}
	return val
}

func Int(k string) int {
	return IntWithDefault(k, 0)
}
