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

func TestCryptoBackendXC(t *testing.T) {
	ctx := context.Background()

	assert.Equal(t, "xc", backend.CryptoBackendName(backend.XC))
	assert.Equal(t, backend.XC, backend.GetCryptoBackend(backend.WithCryptoBackendString(ctx, "xc")))
	assert.Equal(t, backend.XC, backend.GetCryptoBackend(backend.WithCryptoBackend(ctx, backend.XC)))
	assert.Equal(t, true, backend.HasCryptoBackend(backend.WithCryptoBackend(ctx, backend.XC)))
}

func TestCompositeXC(t *testing.T) {
	ctx := context.Background()
	ctx = backend.WithCryptoBackend(ctx, backend.XC)
	ctx = backend.WithRCSBackend(ctx, backend.GoGit)
	ctx = backend.WithStorageBackend(ctx, backend.FS)

	assert.Equal(t, backend.XC, backend.GetCryptoBackend(ctx))
	assert.Equal(t, backend.GoGit, backend.GetRCSBackend(ctx))
	assert.Equal(t, backend.FS, backend.GetStorageBackend(ctx))
}
