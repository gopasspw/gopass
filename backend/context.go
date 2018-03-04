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
	switch cb {
	case GPGMock:
		return "gpgmock"
	case GPGCLI:
		return "gpgcli"
	case XC:
		return "xc"
	case OpenPGP:
		return "openpgp"
	default:
		return ""
	}
}

// WithCryptoBackendString returns a context with the given crypto backend set
func WithCryptoBackendString(ctx context.Context, be string) context.Context {
	switch be {
	case "gpg":
		fallthrough
	case "gpgcli":
		return WithCryptoBackend(ctx, GPGCLI)
	case "gpgmock":
		return WithCryptoBackend(ctx, GPGMock)
	case "xc":
		return WithCryptoBackend(ctx, XC)
	case "openpgp":
		return WithCryptoBackend(ctx, OpenPGP)
	default:
		return ctx
	}
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
	switch sb {
	case GitMock:
		return "gitmock"
	case GitCLI:
		return "gitcli"
	case GoGit:
		return "gogit"
	default:
		return ""
	}
}

// WithSyncBackendString returns a context with the given sync backend set
func WithSyncBackendString(ctx context.Context, sb string) context.Context {
	switch sb {
	case "git":
		fallthrough
	case "gitcli":
		return WithSyncBackend(ctx, GitCLI)
	case "gogit":
		return WithSyncBackend(ctx, GoGit)
	case "gitmock":
		return WithSyncBackend(ctx, GitMock)
	default:
		return WithSyncBackend(ctx, GitMock)
	}
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
	switch sb {
	case "kvmock":
		return WithStoreBackend(ctx, KVMock)
	case "fs":
		return WithStoreBackend(ctx, FS)
	default:
		return WithStoreBackend(ctx, FS)
	}
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

// StoreBackendName returns the name of the given backend
func StoreBackendName(sb StoreBackend) string {
	switch sb {
	case FS:
		return "fs"
	case KVMock:
		return "kvmock"
	default:
		return ""
	}
}
