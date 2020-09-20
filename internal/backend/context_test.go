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
	ctx = WithCryptoBackend(ctx, Age)
	ctx = WithStorageBackend(ctx, FS)

	assert.Equal(t, Age, GetCryptoBackend(ctx))
	assert.Equal(t, FS, GetStorageBackend(ctx))
}
