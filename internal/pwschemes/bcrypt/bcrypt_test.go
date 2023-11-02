package bcrypt

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBcrypt(t *testing.T) {
	t.Parallel()

	pw := "foobar"

	hash, err := Generate(pw)
	require.NoError(t, err)

	t.Logf("PW: %s - Hash: %s", pw, hash)

	require.NoError(t, Validate(pw, hash))
}
