package backend_test

import (
	"context"
	"testing"

	"github.com/gopasspw/gopass/pkg/backend"
	"github.com/gopasspw/gopass/pkg/backend/crypto/plain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeCryptoLoader struct{}

func (l fakeCryptoLoader) New(context.Context) (backend.Crypto, error) {
	return plain.New(), nil
}

func (l fakeCryptoLoader) String() string {
	return "fakecryptoloader"
}

func TestCryptoLoader(t *testing.T) {
	ctx := context.Background()
	backend.RegisterCrypto(backend.Plain, "plain", fakeCryptoLoader{})
	c, err := backend.NewCrypto(ctx, backend.Plain)
	require.NoError(t, err)
	assert.Equal(t, c.Name(), "plain")
}
