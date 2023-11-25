package config

import "context"

// Bool returns a bool value from the config in the context.
func Bool(ctx context.Context, key string) bool {
	cfg, mp := FromContext(ctx)

	return cfg.GetBoolM(mp, key)
}

// String returns a string value from the config in the context.
func String(ctx context.Context, key string) string {
	cfg, mp := FromContext(ctx)

	return cfg.GetM(mp, key)
}

// Int returns an integer value from the config in the context.
func Int(ctx context.Context, key string) int {
	cfg, mp := FromContext(ctx)

	return cfg.GetIntM(mp, key)
}
