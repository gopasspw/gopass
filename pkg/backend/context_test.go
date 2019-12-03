package backend

import (
	"context"
	"testing"

	"github.com/gopasspw/gopass/pkg/backend"
	_ "github.com/gopasspw/gopass/pkg/backend/crypto"
	_ "github.com/gopasspw/gopass/pkg/backend/rcs"
	_ "github.com/gopasspw/gopass/pkg/backend/storage"
	"github.com/stretchr/testify/assert"
)

func TestCryptoBackend(t *testing.T) {
	ctx := context.Background()

	assert.Equal(t, backend.GPGCLI, backend.GetCryptoBackend(ctx))
	assert.Equal(t, backend.GPGCLI, backend.GetCryptoBackend(backend.WithCryptoBackendString(ctx, "gpgcli")))
	assert.Equal(t, backend.GPGCLI, backend.GetCryptoBackend(backend.WithCryptoBackend(ctx, backend.GPGCLI)))
	assert.Equal(t, true, backend.HasCryptoBackend(backend.WithCryptoBackend(ctx, backend.GPGCLI)))
}

func TestRCSBackend(t *testing.T) {
	ctx := context.Background()

	assert.Equal(t, "gitcli", backend.RCSBackendName(backend.GitCLI))
	assert.Equal(t, backend.Noop, backend.GetRCSBackend(ctx))
	assert.Equal(t, backend.GitCLI, backend.GetRCSBackend(backend.WithRCSBackendString(ctx, "gitcli")))
	assert.Equal(t, backend.GitCLI, backend.GetRCSBackend(backend.WithRCSBackend(ctx, backend.GitCLI)))
	assert.Equal(t, backend.Noop, backend.GetRCSBackend(backend.WithRCSBackendString(ctx, "foobar")))
	assert.Equal(t, true, backend.HasRCSBackend(backend.WithRCSBackend(ctx, backend.GitCLI)))
}

func TestStorageBackend(t *testing.T) {
	ctx := context.Background()

	assert.Equal(t, "fs", backend.StorageBackendName(backend.FS))
	assert.Equal(t, backend.FS, backend.GetStorageBackend(ctx))
	assert.Equal(t, backend.FS, backend.GetStorageBackend(backend.WithStorageBackendString(ctx, "fs")))
	assert.Equal(t, backend.FS, backend.GetStorageBackend(backend.WithStorageBackend(ctx, backend.FS)))
	assert.Equal(t, true, backend.HasStorageBackend(backend.WithStorageBackend(ctx, backend.FS)))
}

func TestComposite(t *testing.T) {
	ctx := context.Background()
	ctx = backend.WithCryptoBackend(ctx, backend.XC)
	ctx = backend.WithRCSBackend(ctx, backend.GoGit)
	ctx = backend.WithStorageBackend(ctx, backend.FS)

	assert.Equal(t, backend.XC, backend.GetCryptoBackend(ctx))
	assert.Equal(t, backend.GoGit, backend.GetRCSBackend(ctx))
	assert.Equal(t, backend.FS, backend.GetStorageBackend(ctx))
}
