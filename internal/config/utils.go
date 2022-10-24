package config

import "context"

// Bool returns a bool value from the config in the context.
func Bool(ctx context.Context, key string) bool {
	return FromContext(ctx).GetBool(key)
}

// String returns a string value from the config in the context.
func String(ctx context.Context, key string) string {
	return FromContext(ctx).Get(key)
}

// Int returns an integer value from the config in the context.
func Int(ctx context.Context, key string) int {
	return FromContext(ctx).GetInt(key)
}
