package backend

import "context"

type contextKey int

const (
	ctxKeyCryptoBackend contextKey = iota
	ctxKeyRCSBackend
	ctxKeyStorageBackend
)

// CryptoBackendName returns the name of the given backend.
func CryptoBackendName(cb CryptoBackend) string {
	if name, err := CryptoRegistry.BackendName(cb); err == nil {
		return name
	}
	return ""
}

// WithCryptoBackendString returns a context with the given crypto backend set.
func WithCryptoBackendString(ctx context.Context, be string) context.Context {
	if cb, err := CryptoRegistry.Backend(be); err == nil {
		ctx = WithCryptoBackend(ctx, cb)
	}
	return ctx
}

// WithCryptoBackend returns a context with the given crypto backend set.
func WithCryptoBackend(ctx context.Context, be CryptoBackend) context.Context {
	return context.WithValue(ctx, ctxKeyCryptoBackend, be)
}

// HasCryptoBackend returns true if a value for crypto backend has been set in the context.
func HasCryptoBackend(ctx context.Context) bool {
	_, ok := ctx.Value(ctxKeyCryptoBackend).(CryptoBackend)
	return ok
}

// GetCryptoBackend returns the selected crypto backend or the default (GPGCLI).
func GetCryptoBackend(ctx context.Context) CryptoBackend {
	be, ok := ctx.Value(ctxKeyCryptoBackend).(CryptoBackend)
	if !ok {
		return GPGCLI
	}
	return be
}

// WithStorageBackendString returns a context with the given store backend set.
func WithStorageBackendString(ctx context.Context, sb string) context.Context {
	if be, err := StorageRegistry.Backend(sb); err == nil {
		return WithStorageBackend(ctx, be)
	}
	return WithStorageBackend(ctx, FS)
}

// WithStorageBackend returns a context with the given store backend set.
func WithStorageBackend(ctx context.Context, sb StorageBackend) context.Context {
	return context.WithValue(ctx, ctxKeyStorageBackend, sb)
}

// GetStorageBackend returns the store backend or the default (FS).
func GetStorageBackend(ctx context.Context) StorageBackend {
	be, ok := ctx.Value(ctxKeyStorageBackend).(StorageBackend)
	if !ok {
		return FS
	}
	return be
}

// HasStorageBackend returns true if a value for store backend was set.
func HasStorageBackend(ctx context.Context) bool {
	_, ok := ctx.Value(ctxKeyStorageBackend).(StorageBackend)
	return ok
}

// StorageBackendName returns the name of the given backend.
func StorageBackendName(sb StorageBackend) string {
	if name, err := StorageRegistry.BackendName(sb); err == nil {
		return name
	}
	return ""
}
