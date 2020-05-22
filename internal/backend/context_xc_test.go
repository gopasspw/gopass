// +build xc

package backend

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCryptoBackendXC(t *testing.T) {
	ctx := context.Background()

	assert.Equal(t, "xc", CryptoBackendName(XC))
	assert.Equal(t, XC, GetCryptoBackend(WithCryptoBackendString(ctx, "xc")))
	assert.Equal(t, XC, GetCryptoBackend(WithCryptoBackend(ctx, XC)))
	assert.Equal(t, true, HasCryptoBackend(WithCryptoBackend(ctx, XC)))
}

func TestCompositeXC(t *testing.T) {
	ctx := context.Background()
	ctx = WithCryptoBackend(ctx, XC)
	ctx = WithStorageBackend(ctx, FS)

	assert.Equal(t, XC, GetCryptoBackend(ctx))
	assert.Equal(t, FS, GetStorageBackend(ctx))
}
