package agecrypto

import (
	"crypto/ed25519"
	"fmt"
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
