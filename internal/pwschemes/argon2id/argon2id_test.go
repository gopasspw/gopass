package argon2id

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestArgon2ID(t *testing.T) {
	t.Parallel()

	pw := "foobar"
	hash, err := Generate(pw, 0)
	require.NoError(t, err)

	t.Logf("PW: %s - Hash: %s", pw, hash)
	ok, err := Validate(pw, hash)
	require.NoError(t, err)
	assert.True(t, ok)
}
