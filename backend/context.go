package backend

import "context"

type contextKey int

const (
	ctxKeyCryptoBackend contextKey = iota
	ctxKeySyncBackend
	ctxKeyStoreBackend
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

// SyncBackendName returns the name of the given backend
func SyncBackendName(sb SyncBackend) string {
	return syncNameFromBackend(sb)
}

// WithSyncBackendString returns a context with the given sync backend set
func WithSyncBackendString(ctx context.Context, sb string) context.Context {
	if be := syncBackendFromName(sb); be >= 0 {
		return WithSyncBackend(ctx, be)
	}
	return WithSyncBackend(ctx, GitMock)
}

// WithSyncBackend returns a context with the given sync backend set
func WithSyncBackend(ctx context.Context, sb SyncBackend) context.Context {
	return context.WithValue(ctx, ctxKeySyncBackend, sb)
}

// HasSyncBackend returns true if a value for sync backend has been set in the context
func HasSyncBackend(ctx context.Context) bool {
	_, ok := ctx.Value(ctxKeySyncBackend).(SyncBackend)
	return ok
}

// GetSyncBackend returns the sync backend or the default (Git Mock)
func GetSyncBackend(ctx context.Context) SyncBackend {
	be, ok := ctx.Value(ctxKeySyncBackend).(SyncBackend)
	if !ok {
		return GitMock
	}
	return be
}

// WithStoreBackendString returns a context with the given store backend set
func WithStoreBackendString(ctx context.Context, sb string) context.Context {
	return WithStoreBackend(ctx, storeBackendFromName(sb))
}

// WithStoreBackend returns a context with the given store backend set
func WithStoreBackend(ctx context.Context, sb StoreBackend) context.Context {
	return context.WithValue(ctx, ctxKeyStoreBackend, sb)
}

// GetStoreBackend returns the store backend or the default (FS)
func GetStoreBackend(ctx context.Context) StoreBackend {
	be, ok := ctx.Value(ctxKeyStoreBackend).(StoreBackend)
	if !ok {
		return FS
	}
	return be
}

// HasStoreBackend returns true if a value for store backend was set
func HasStoreBackend(ctx context.Context) bool {
	_, ok := ctx.Value(ctxKeyStoreBackend).(StoreBackend)
	return ok
}

// StoreBackendName returns the name of the given backend
func StoreBackendName(sb StoreBackend) string {
	return storeNameFromBackend(sb)
}
