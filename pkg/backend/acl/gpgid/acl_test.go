package gpgid

import (
	"context"
	"testing"

	gpgmock "github.com/justwatchcom/gopass/pkg/backend/crypto/plain"
	gitmock "github.com/justwatchcom/gopass/pkg/backend/rcs/noop"
	"github.com/justwatchcom/gopass/pkg/backend/storage/kv/inmem"
	"github.com/stretchr/testify/assert"
)

func TestInitLoadVerify(t *testing.T) {
	ctx := context.Background()
	crypto := gpgmock.New()
	rcs := gitmock.New()
	fs := inmem.New()

	assert.NoError(t, fs.Set(ctx, crypto.IDFile(), []byte("0xDEADBEEF")))

	a, err := Init(ctx, crypto, rcs, fs)
	assert.NoError(t, err)
	t.Logf("a.Recipients: %s", a.Recipients())

	assert.Equal(t, a.Recipients(), []string{"0xDEADBEEF"})

	b, err := Load(ctx, crypto, rcs, fs)
	assert.NoError(t, err, "Load()ing store")
	t.Logf("b.Recipients: %s", b.Recipients())

	assert.Equal(t, b.Recipients(), []string{"0xDEADBEEF"})
	assert.NoError(t, b.Save(ctx))
	assert.NoError(t, b.Add(ctx, "0xFEEDBEEF"))

	c, err := Load(ctx, crypto, rcs, fs)
	assert.NoError(t, err)
	t.Logf("c.Recipients: %s", c.Recipients())

	assert.Equal(t, c.Recipients(), []string{"0xDEADBEEF", "0xFEEDBEEF"})

	assert.NoError(t, b.Remove(ctx, "0xDEADBEEF"), "removing recipient")

	d, err := Load(ctx, crypto, rcs, fs)
	assert.NoError(t, err, "Load()ing store")
	t.Logf("d.Recipients: %s", d.Recipients())

	assert.Equal(t, d.Recipients(), []string{"0xFEEDBEEF"})
	assert.NoError(t, d.verify(ctx))
}

func TestInitVerify(t *testing.T) {
	ctx := context.Background()
	crypto := gpgmock.New()
	rcs := gitmock.New()
	fs := inmem.New()

	assert.NoError(t, fs.Set(ctx, crypto.IDFile(), []byte("0xDEADBEEF")))

	a, err := Init(ctx, crypto, rcs, fs)
	assert.NoError(t, err, "Init() store")
	t.Logf("a.Recipients: %s", a.Recipients())

	assert.Equal(t, a.Recipients(), []string{"0xDEADBEEF"})
	assert.NoError(t, a.verify(ctx), "initialized store is valid")

	t.Logf("Tokens: %s", a.tokens)
	a.tokens = nil
	assert.NoError(t, a.marshalTokenFile(ctx))

	assert.NoError(t, a.unmarshalTokenFile(ctx))
	t.Logf("Tokens: %s", a.tokens)

	assert.Error(t, a.verify(ctx), "should not verify invalid (empty) token file")
}
