package bcrypt

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBcrypt(t *testing.T) {
	pw := "foobar"

	hash, err := Generate(pw)
	require.NoError(t, err)

	t.Logf("PW: %s - Hash: %s", pw, hash)

	assert.NoError(t, Validate(pw, hash))
}
