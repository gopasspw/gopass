package backend

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCryptoBackend(t *testing.T) {
	ctx := context.Background()

	assert.Equal(t, GPGCLI, GetCryptoBackend(ctx))
	assert.Equal(t, GPGCLI, GetCryptoBackend(WithCryptoBackendString(ctx, "gpgcli")))
	assert.Equal(t, GPGCLI, GetCryptoBackend(WithCryptoBackend(ctx, GPGCLI)))
	assert.Equal(t, true, HasCryptoBackend(WithCryptoBackend(ctx, GPGCLI)))
}

func TestRCSBackend(t *testing.T) {
	ctx := context.Background()

	assert.Equal(t, "gitcli", RCSBackendName(GitCLI))
	assert.Equal(t, Noop, GetRCSBackend(ctx))
	assert.Equal(t, GitCLI, GetRCSBackend(WithRCSBackendString(ctx, "gitcli")))
	assert.Equal(t, GitCLI, GetRCSBackend(WithRCSBackend(ctx, GitCLI)))
	assert.Equal(t, Noop, GetRCSBackend(WithRCSBackendString(ctx, "foobar")))
	assert.Equal(t, true, HasRCSBackend(WithRCSBackend(ctx, GitCLI)))
}

func TestStorageBackend(t *testing.T) {
	ctx := context.Background()

	assert.Equal(t, "fs", StorageBackendName(FS))
	assert.Equal(t, FS, GetStorageBackend(ctx))
	assert.Equal(t, FS, GetStorageBackend(WithStorageBackendString(ctx, "fs")))
	assert.Equal(t, FS, GetStorageBackend(WithStorageBackend(ctx, FS)))
	assert.Equal(t, true, HasStorageBackend(WithStorageBackend(ctx, FS)))
}

func TestComposite(t *testing.T) {
	ctx := context.Background()
	ctx = WithCryptoBackend(ctx, XC)
	ctx = WithRCSBackend(ctx, GoGit)
	ctx = WithStorageBackend(ctx, FS)

	assert.Equal(t, XC, GetCryptoBackend(ctx))
	assert.Equal(t, GoGit, GetRCSBackend(ctx))
	assert.Equal(t, FS, GetStorageBackend(ctx))
}
