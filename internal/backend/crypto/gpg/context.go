package gpg

import "context"

type contextKey int

const (
	ctxKeyAlwaysTrust contextKey = iota
	ctxKeyUseCache
)

// WithAlwaysTrust will return a context with the flag for always trust set.
func WithAlwaysTrust(ctx context.Context, at bool) context.Context {
	return context.WithValue(ctx, ctxKeyAlwaysTrust, at)
}

// IsAlwaysTrust will return the value of the always trust flag or the default
// (false).
func IsAlwaysTrust(ctx context.Context) bool {
	bv, ok := ctx.Value(ctxKeyAlwaysTrust).(bool)
	if !ok {
		return false
	}
	return bv
}

// WithUseCache returns a context with the value of NoCache set.
func WithUseCache(ctx context.Context, nc bool) context.Context {
	return context.WithValue(ctx, ctxKeyUseCache, nc)
}

// UseCache returns true if this request should ignore the cache.
func UseCache(ctx context.Context) bool {
	nc, ok := ctx.Value(ctxKeyUseCache).(bool)
	if !ok {
		return false
	}
	return nc
}
