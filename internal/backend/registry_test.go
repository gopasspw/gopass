package backend_test

import (
	"context"
	"testing"

	"github.com/gopasspw/gopass/internal/backend"
	_ "github.com/gopasspw/gopass/internal/backend/crypto"
	"github.com/gopasspw/gopass/internal/backend/crypto/plain"
	_ "github.com/gopasspw/gopass/internal/backend/rcs"
	_ "github.com/gopasspw/gopass/internal/backend/storage"
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

func (l fakeCryptoLoader) Handles(_ backend.Storage) error {
	return nil
}

func (l fakeCryptoLoader) Priority() int {
	return 1
}

func TestCryptoLoader(t *testing.T) {
	ctx := context.Background()
	backend.RegisterCrypto(backend.Plain, "plain", fakeCryptoLoader{})
	c, err := backend.NewCrypto(ctx, backend.Plain)
	require.NoError(t, err)
	assert.Equal(t, c.Name(), "plain")
}
