package mock

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/stretchr/testify/assert"
)

func TestMock(t *testing.T) {
	td, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(td)
	}()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)

	m := New()
	kl, err := m.ListPrivateKeys(ctx)
	assert.NoError(t, err)
	assert.NotEmpty(t, kl)
	assert.Equal(t, "0xDEADBEEF", kl[0].ID())

	kl, err = m.ListPublicKeys(ctx)
	assert.NoError(t, err)
	assert.Empty(t, kl)

	rcs, err := m.GetRecipients(ctx, "")
	assert.NoError(t, err)
	assert.Empty(t, rcs)

	fn := filepath.Join(td, "sec.gpg")
	assert.NoError(t, m.Encrypt(ctx, fn, []byte("foobar"), []string{"0xDEADBEEF"}))
	assert.FileExists(t, fn)

	content, err := m.Decrypt(ctx, fn)
	assert.NoError(t, err)
	assert.Equal(t, string(content), "foobar")

	assert.Equal(t, "gpg", m.Binary())

	sigfn := fn + ".sig"
	assert.NoError(t, m.Sign(ctx, fn, sigfn))
	assert.NoError(t, m.Verify(ctx, sigfn, fn))

	assert.Error(t, m.CreatePrivateKey(ctx))
	assert.Error(t, m.CreatePrivateKeyBatch(ctx, "", "", ""))
}
