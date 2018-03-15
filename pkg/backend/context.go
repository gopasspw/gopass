package backend

import "context"

type contextKey int

const (
	ctxKeyCryptoBackend contextKey = iota
	ctxKeyRCSBackend
	ctxKeyStorageBackend
)

// CryptoBackendName returns the name of the given backend
func CryptoBackendName(cb CryptoBackend) string {
	return cryptoNameFromBackend(cb)
}

// WithCryptoBackendString returns a context with the given crypto backend set
func WithCryptoBackendString(ctx context.Context, be string) context.Context {
	if cb := cryptoBackendFromName(be); cb >= 0 {
		ctx = WithCryptoBackend(ctx, cb)
	}
	return ctx
}

// WithCryptoBackend returns a context with the given crypto backend set
func WithCryptoBackend(ctx context.Context, be CryptoBackend) context.Context {
	return context.WithValue(ctx, ctxKeyCryptoBackend, be)
}

// HasCryptoBackend returns true if a value for crypto backend has been set in the context
func HasCryptoBackend(ctx context.Context) bool {
	_, ok := ctx.Value(ctxKeyCryptoBackend).(CryptoBackend)
	return ok
}

// GetCryptoBackend returns the selected crypto backend or the default (GPGCLI)
func GetCryptoBackend(ctx context.Context) CryptoBackend {
	be, ok := ctx.Value(ctxKeyCryptoBackend).(CryptoBackend)
	if !ok {
		return GPGCLI
	}
	return be
}

// RCSBackendName returns the name of the given backend
func RCSBackendName(sb RCSBackend) string {
	return rcsNameFromBackend(sb)
}

// WithRCSBackendString returns a context with the given sync backend set
func WithRCSBackendString(ctx context.Context, sb string) context.Context {
	if be := rcsBackendFromName(sb); be >= 0 {
		return WithRCSBackend(ctx, be)
	}
	return WithRCSBackend(ctx, Noop)
}

// WithRCSBackend returns a context with the given sync backend set
func WithRCSBackend(ctx context.Context, sb RCSBackend) context.Context {
	return context.WithValue(ctx, ctxKeyRCSBackend, sb)
}

// HasRCSBackend returns true if a value for sync backend has been set in the context
func HasRCSBackend(ctx context.Context) bool {
	_, ok := ctx.Value(ctxKeyRCSBackend).(RCSBackend)
	return ok
}

// GetRCSBackend returns the sync backend or the default (Git Mock)
func GetRCSBackend(ctx context.Context) RCSBackend {
	be, ok := ctx.Value(ctxKeyRCSBackend).(RCSBackend)
	if !ok {
		return Noop
	}
	return be
}

// WithStorageBackendString returns a context with the given store backend set
func WithStorageBackendString(ctx context.Context, sb string) context.Context {
	return WithStorageBackend(ctx, storageBackendFromName(sb))
}

// WithStorageBackend returns a context with the given store backend set
func WithStorageBackend(ctx context.Context, sb StorageBackend) context.Context {
	return context.WithValue(ctx, ctxKeyStorageBackend, sb)
}

// GetStorageBackend returns the store backend or the default (FS)
func GetStorageBackend(ctx context.Context) StorageBackend {
	be, ok := ctx.Value(ctxKeyStorageBackend).(StorageBackend)
	if !ok {
		return FS
	}
	return be
}

// HasStorageBackend returns true if a value for store backend was set
func HasStorageBackend(ctx context.Context) bool {
	_, ok := ctx.Value(ctxKeyStorageBackend).(StorageBackend)
	return ok
}

// StorageBackendName returns the name of the given backend
func StorageBackendName(sb StorageBackend) string {
	return storageNameFromBackend(sb)
}
