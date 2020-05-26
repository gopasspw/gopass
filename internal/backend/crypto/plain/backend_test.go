package plain

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/gopasspw/gopass/pkg/ctxutil"

	"github.com/blang/semver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlain(t *testing.T) {
	td, err := ioutil.TempDir("", "gopass-")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(td)
	}()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)

	m := New()
	kl, err := m.ListIdentities(ctx)
	assert.NoError(t, err)
	assert.NotEmpty(t, kl)
	assert.Equal(t, "0xDEADBEEF", kl[0])

	kl, err = m.ListRecipients(ctx)
	assert.NoError(t, err)
	assert.NotEmpty(t, kl, "ListRecipients")

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
	assert.Equal(t, "", m.Fingerprint(ctx, ""))
	assert.Nil(t, m.Initialized(ctx))
	assert.Equal(t, "plain", m.Name())
	assert.Equal(t, "txt", m.Ext())
	assert.Equal(t, ".plain-id", m.IDFile())
	names, err := m.ReadNamesFromKey(ctx, nil)
	assert.NoError(t, err)
	assert.Equal(t, []string{"unsupported"}, names)
}

func TestLoader(t *testing.T) {
	l := &loader{}
	b, err := l.New(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, name, l.String())
	assert.Equal(t, "plain", b.Name())
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

	assert.Error(t, m.Sign(ctx, "/tmp", sigf))
	assert.Error(t, m.Verify(ctx, sigf, "/tmp"))
	assert.Error(t, m.Verify(ctx, "/tmp", in))

	assert.NoError(t, ioutil.WriteFile(sigf, []byte("in"), 0644))
	assert.Error(t, m.Verify(ctx, sigf, in))
}
