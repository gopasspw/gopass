package identityfile

import (
	"errors"
	"strings"
	"testing"

	"filippo.io/age"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSkipsCommentsAndReportsSourceLine(t *testing.T) {
	wantErr := errors.New("invalid identity")
	parser := func(line string) (age.Identity, error) {
		assert.Equal(t, "invalid", line)

		return nil, wantErr
	}

	_, err := Parse(strings.NewReader("# comment\n\ninvalid\n"), parser)
	require.ErrorIs(t, err, wantErr)
	assert.EqualError(t, err, "error at line 3: invalid identity")
}

func TestParseRejectsEmptyInput(t *testing.T) {
	_, err := Parse(strings.NewReader("# comment\n\n"), func(string) (age.Identity, error) {
		t.Fatal("parser must not be called")

		return nil, errors.New("unreachable parser call")
	})
	require.EqualError(t, err, "no secret keys found")
}

func TestParseReturnsAllIdentities(t *testing.T) {
	first, err := age.GenerateX25519Identity()
	require.NoError(t, err)
	second, err := age.GenerateX25519Identity()
	require.NoError(t, err)

	input := first.String() + "\n" + second.String() + "\n"
	ids, err := Parse(strings.NewReader(input), func(line string) (age.Identity, error) {
		return age.ParseX25519Identity(line)
	})
	require.NoError(t, err)
	require.Equal(t, []age.Identity{first, second}, ids)
}
