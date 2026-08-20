package age

import (
	"bytes"
	"crypto/ed25519"
	"fmt"
	"io"
	"sort"
	"testing"

	"filippo.io/age"
	"filippo.io/age/agessh"
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

func TestEncryptUsesOnlyConfiguredRecipients(t *testing.T) {
	ctx := t.Context()
	a := newTestAge(t)

	storeID, err := age.GenerateX25519Identity()
	require.NoError(t, err)

	localOnlyID, err := age.GenerateX25519Identity()
	require.NoError(t, err)
	require.NoError(t, a.addIdentity(ctx, localOnlyID))

	plaintext := []byte("top secret")
	ciphertext, err := a.Encrypt(ctx, plaintext, []string{storeID.Recipient().String()})
	require.NoError(t, err)

	r, err := age.Decrypt(bytes.NewReader(ciphertext), storeID)
	require.NoError(t, err)

	decrypted, err := io.ReadAll(r)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)

	_, err = age.Decrypt(bytes.NewReader(ciphertext), localOnlyID)
	require.Error(t, err)
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
