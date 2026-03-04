package backend

import (
	"testing"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCryptoBackend(t *testing.T) {
	t.Parallel()

	ctx := config.NewContextInMemory()

	assert.Equal(t, GPGCLI, GetCryptoBackend(ctx))
	ctx1, err := WithCryptoBackendString(ctx, "gpgcli")
	require.NoError(t, err)
	assert.Equal(t, GPGCLI, GetCryptoBackend(ctx1))
	assert.Equal(t, GPGCLI, GetCryptoBackend(WithCryptoBackend(ctx, GPGCLI)))
	assert.True(t, HasCryptoBackend(WithCryptoBackend(ctx, GPGCLI)))
}

func TestStorageBackend(t *testing.T) {
	t.Parallel()

	ctx := config.NewContextInMemory()

	assert.Equal(t, "fs", StorageBackendName(FS))
	assert.Equal(t, FS, GetStorageBackend(ctx))
	ctx1, err := WithStorageBackendString(ctx, "fs")
	require.NoError(t, err)
	assert.Equal(t, FS, GetStorageBackend(ctx1))
	assert.Equal(t, FS, GetStorageBackend(WithStorageBackend(ctx, FS)))
	assert.True(t, HasStorageBackend(WithStorageBackend(ctx, FS)))
}

func TestComposite(t *testing.T) {
	t.Parallel()

	ctx := config.NewContextInMemory()
	ctx = WithCryptoBackend(ctx, Age)
	ctx = WithStorageBackend(ctx, FS)

	assert.Equal(t, Age, GetCryptoBackend(ctx))
	assert.Equal(t, FS, GetStorageBackend(ctx))
}
