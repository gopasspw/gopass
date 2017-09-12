package gpg

import "context"

type contextKey int

const (
	ctxKeyAlwaysTrust contextKey = iota
)

// WithAlwaysTrust will return a context with the flag for always trust set
func WithAlwaysTrust(ctx context.Context, at bool) context.Context {
	return context.WithValue(ctx, ctxKeyAlwaysTrust, at)
}

// IsAlwaysTrust will return the value of the always trust flag or the default
// (false)
func IsAlwaysTrust(ctx context.Context) bool {
	bv, ok := ctx.Value(ctxKeyAlwaysTrust).(bool)
	if !ok {
		return false
	}
	return bv
}
