package age

import "context"

type contextKey int

const (
	ctxKeyOnlyNative contextKey = iota
	ctxKeyAgentLauncher
)

// WithOnlyNative will return a context with the flag for only native set.
func WithOnlyNative(ctx context.Context, at bool) context.Context {
	return context.WithValue(ctx, ctxKeyOnlyNative, at)
}

// IsOnlyNative will return the value of the only native flag or the default
// (false).
func IsOnlyNative(ctx context.Context) bool {
	bv, ok := ctx.Value(ctxKeyOnlyNative).(bool)
	if !ok {
		return false
	}

	return bv
}

// WithAgentLauncher registers a launcher that starts the age agent process.
// Only callers that own the `age agent start` subcommand (the standalone
// gopass CLI) should register one; library embedders leave it unset so
// tryStartAgent skips autostart instead of fork-bombing. Mirrors the
// leaf.WithFsckFunc callback-in-context pattern.
func WithAgentLauncher(ctx context.Context, l func(context.Context) error) context.Context {
	return context.WithValue(ctx, ctxKeyAgentLauncher, l)
}

// GetAgentLauncher returns the registered agent launcher, or nil if none was
// set.
func GetAgentLauncher(ctx context.Context) func(context.Context) error {
	l, _ := ctx.Value(ctxKeyAgentLauncher).(func(context.Context) error)

	return l
}
