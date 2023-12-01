package plain

import (
	"testing"

	"github.com/blang/semver/v4"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlain(t *testing.T) {
	t.Parallel()

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithAlwaysYes(ctx, true)

	m := New()
	kl, err := m.ListIdentities(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, kl)
	assert.Equal(t, "0xDEADBEEF", kl[0])

	kl, err = m.ListRecipients(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, kl, "ListRecipients")

	rcs, err := m.RecipientIDs(ctx, []byte{})
	require.NoError(t, err)
	assert.NotEmpty(t, rcs, "RecipientIDs")

	buf, err := m.Encrypt(ctx, []byte("foobar"), []string{"0xDEADBEEF"})
	require.NoError(t, err)

	content, err := m.Decrypt(ctx, buf)
	require.NoError(t, err)
	assert.Equal(t, "foobar", string(content))

	assert.Equal(t, "gpg", m.Binary())

	require.Error(t, m.GenerateIdentity(ctx, "", "", ""))

	kl, err = m.FindRecipients(ctx)
	require.NoError(t, err)
	assert.Empty(t, kl, "FindRecipients()")

	kl, err = m.FindRecipients(ctx, "0xDEADBEEF")
	require.NoError(t, err)
	assert.NotEmpty(t, kl, "FindRecipients(0xDEADBEEF)")

	_, err = m.FindIdentities(ctx)
	require.NoError(t, err)

	require.NoError(t, m.ImportPublicKey(ctx, buf))
	assert.Equal(t, semver.Version{}, m.Version(ctx))

	assert.Equal(t, "", m.FormatKey(ctx, "", ""))
	assert.Equal(t, "", m.Fingerprint(ctx, ""))
	require.NoError(t, m.Initialized(ctx))
	assert.Equal(t, "plain", m.Name())
	assert.Equal(t, "txt", m.Ext())
	assert.Equal(t, ".plain-id", m.IDFile())
	names, err := m.ReadNamesFromKey(ctx, nil)
	require.NoError(t, err)
	assert.Equal(t, []string{"unsupported"}, names)
}

func TestLoader(t *testing.T) {
	t.Parallel()

	l := &loader{}
	b, err := l.New(config.NewContextInMemory())
	require.NoError(t, err)
	assert.Equal(t, name, l.String())
	assert.Equal(t, "plain", b.Name())
}
