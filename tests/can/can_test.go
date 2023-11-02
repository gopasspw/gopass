package can

import (
	"testing"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPubring(t *testing.T) {
	t.Parallel()

	fh, err := can.Open("gnupg/pubring.gpg")
	require.NoError(t, err)
	defer fh.Close() //nolint:errcheck

	el, err := openpgp.ReadKeyRing(fh)
	require.NoError(t, err)

	require.Len(t, el, 1)
	assert.Equal(t, "BE73F104", el[0].PrimaryKey.KeyIdShortString())
}
