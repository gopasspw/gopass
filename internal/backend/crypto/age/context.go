package age

import "context"

type contextKey int

const (
	ctxKeyOnlyNative contextKey = iota
	ctxKeyUseKeychain
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

// WithUseKeychain returns a context with the value of use keychain
// set.
func WithUseKeychain(ctx context.Context, bv bool) context.Context {
	return context.WithValue(ctx, ctxKeyUseKeychain, bv)
}

// IsUseKeychain returns the value of use keychain.
func IsUseKeychain(ctx context.Context) bool {
	bv, ok := ctx.Value(ctxKeyUseKeychain).(bool)
	if !ok {
		return false
	}

	return bv
}

// HasUseKeychain returns true if a value for use keychain
// was set in the context.
func HasUseKeychain(ctx context.Context) bool {
	_, ok := ctx.Value(ctxKeyUseKeychain).(bool)

	return ok
}
