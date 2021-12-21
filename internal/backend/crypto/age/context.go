package age

import "context"

type contextKey int

const (
	ctxKeyOnlyNative contextKey = iota
)

// WithOnlyNative will return a context with the flag for only native set
func WithOnlyNative(ctx context.Context, at bool) context.Context {
	return context.WithValue(ctx, ctxKeyOnlyNative, at)
}

// IsOnlyNative will return the value of the only native flag or the default
// (false)
func IsOnlyNative(ctx context.Context) bool {
	bv, ok := ctx.Value(ctxKeyOnlyNative).(bool)
	if !ok {
		return false
	}
	return bv
}
