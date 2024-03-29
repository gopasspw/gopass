package config

import (
	"context"
	"strconv"
)

// AsBool converts a string to a bool value.
func AsBool(s string) bool {
	return AsBoolWithDefault(s, false)
}

// AsBoolWithDefault converts a string to a bool value with a default value.
func AsBoolWithDefault(s string, def bool) bool {
	if s == "" {
		return def
	}

	switch s {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return def
	}
}

// AsInt converts a string to an integer value.
func AsInt(s string) int {
	return AsIntWithDefault(s, 0)
}

// AsIntWithDefault converts a string to an integer value with a default value.
func AsIntWithDefault(s string, def int) int {
	if s == "" {
		return def
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		return def
	}

	return i
}

// Bool returns a bool value from the config in the context.
func Bool(ctx context.Context, key string) bool {
	cfg, mp := FromContext(ctx)

	return AsBool(cfg.GetM(mp, key))
}

// String returns a string value from the config in the context.
func String(ctx context.Context, key string) string {
	cfg, mp := FromContext(ctx)

	return cfg.GetM(mp, key)
}

// Int returns an integer value from the config in the context.
func Int(ctx context.Context, key string) int {
	cfg, mp := FromContext(ctx)

	return AsInt(cfg.GetM(mp, key))
}
