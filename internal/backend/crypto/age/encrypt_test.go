package age

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"sort"
	"testing"

	"filippo.io/age"
	"filippo.io/age/agessh"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

func TestDedupe(t *testing.T) {
	t.Parallel()

	i1, err := age.GenerateX25519Identity()
	require.NoError(t, err)

	i2, err := age.GenerateX25519Identity()
	require.NoError(t, err)

	i3pub, _, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)
	i3ssh, err := ssh.NewPublicKey(i3pub)
	require.NoError(t, err)
	i3, err := agessh.NewEd25519Recipient(i3ssh)
	require.NoError(t, err)

	in := []age.Recipient{i1.Recipient(), i2.Recipient(), i2.Recipient(), i3, i3}
	out := dedupe(in)
	want := []age.Recipient{i3, i3, i1.Recipient(), i2.Recipient()}

	sort.Sort(Recipients(out))
	sort.Sort(Recipients(want))
	assert.Equal(t, want, out)
}

type Recipients []age.Recipient

func (r Recipients) Len() int {
	return len(r)
}

func (r Recipients) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func (r Recipients) Less(i, j int) bool {
	return fmt.Sprintf("%s", r[i]) < fmt.Sprintf("%s", r[j])
}

func TestHasAnyRecipient(t *testing.T) {
	t.Parallel()

	i1, err := age.GenerateX25519Identity()
	require.NoError(t, err)
	i2, err := age.GenerateX25519Identity()
	require.NoError(t, err)

	// no local identities -> nothing to check, must not block.
	assert.True(t, hasAnyRecipient(nil, nil))

	// local identity is among the recipients.
	assert.True(t, hasAnyRecipient([]age.Recipient{i1.Recipient()}, []age.Recipient{i1.Recipient()}))

	// local identity is NOT among the recipients.
	assert.False(t, hasAnyRecipient([]age.Recipient{i2.Recipient()}, []age.Recipient{i1.Recipient()}))

	// no recipients at all, but local identities exist.
	assert.False(t, hasAnyRecipient(nil, []age.Recipient{i1.Recipient()}))
}

func TestEncryptNoLocalIdentity(t *testing.T) {
	ctx := context.Background()

	i1, err := age.GenerateX25519Identity()
	require.NoError(t, err)
	i2, err := age.GenerateX25519Identity()
	require.NoError(t, err)

	a := newTestAge(t)
	require.NoError(t, a.addIdentity(ctx, i1))

	// encrypting for a recipient we do not have an identity for must fail ...
	_, err = a.Encrypt(ctx, []byte("foobar"), []string{i2.Recipient().String()})
	assert.ErrorIs(t, err, ErrNoLocalIdentity) //nolint:testifylint

	// ... unless forced.
	ctx = ctxutil.WithForce(ctx, true)
	_, err = a.Encrypt(ctx, []byte("foobar"), []string{i2.Recipient().String()})
	require.NoError(t, err)

	// encrypting for our own recipient must succeed without force.
	ctx = context.Background()
	_, err = a.Encrypt(ctx, []byte("foobar"), []string{i1.Recipient().String()})
	require.NoError(t, err)
}
