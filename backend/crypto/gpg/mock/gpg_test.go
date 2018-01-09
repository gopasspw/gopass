package mock

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/blang/semver"
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
	kl, err := m.ListPrivateKeyIDs(ctx)
	assert.NoError(t, err)
	assert.NotEmpty(t, kl)
	assert.Equal(t, "0xDEADBEEF", kl[0])

	kl, err = m.ListPublicKeyIDs(ctx)
	assert.NoError(t, err)
	assert.NotEmpty(t, kl, "ListPublicKeyIDs")

	rcs, err := m.RecipientIDs(ctx, []byte{})
	assert.NoError(t, err)
	assert.NotEmpty(t, rcs, "RecipientIDs")

	buf, err := m.Encrypt(ctx, []byte("foobar"), []string{"0xDEADBEEF"})
	assert.NoError(t, err)

	content, err := m.Decrypt(ctx, buf)
	assert.NoError(t, err)
	assert.Equal(t, string(content), "foobar")

	assert.Equal(t, "gpg", m.Binary())

	assert.Error(t, m.CreatePrivateKey(ctx))
	assert.Error(t, m.CreatePrivateKeyBatch(ctx, "", "", ""))

	kl, err = m.FindPublicKeys(ctx)
	assert.NoError(t, err)
	assert.Empty(t, kl, "FindPublicKeys()")

	kl, err = m.FindPublicKeys(ctx, "0xDEADBEEF")
	assert.NoError(t, err)
	assert.NotEmpty(t, kl, "FindPublicKeys(0xDEADBEEF)")

	_, err = m.FindPrivateKeys(ctx)
	assert.NoError(t, err)

	buf, err = m.ExportPublicKey(ctx, "")
	assert.NoError(t, err)
	assert.NoError(t, m.ImportPublicKey(ctx, buf))
	assert.Equal(t, semver.Version{}, m.Version(ctx))

	assert.Equal(t, "", m.EmailFromKey(ctx, ""))
	assert.Equal(t, "", m.NameFromKey(ctx, ""))
	assert.Equal(t, "", m.FormatKey(ctx, ""))
	assert.Nil(t, m.Initialized(ctx))
	assert.Equal(t, "gpgmock", m.Name())
	assert.Equal(t, "gpg", m.Ext())
	assert.Equal(t, ".gpg-id", m.IDFile())
	names, err := m.ReadNamesFromKey(ctx, nil)
	assert.NoError(t, err)
	assert.Equal(t, []string{"unsupported"}, names)
}

func TestSignVerify(t *testing.T) {
	td, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(td)
	}()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)

	m := New()

	in := filepath.Join(td, "in")
	assert.NoError(t, ioutil.WriteFile(in, []byte("in"), 0644))
	sigf := filepath.Join(td, "sigf")

	assert.NoError(t, m.Sign(ctx, in, sigf))
	assert.NoError(t, m.Verify(ctx, sigf, in))
}
