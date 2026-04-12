package agent

import (
	"strings"
	"testing"

	"filippo.io/age"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseIdentity(t *testing.T) {
	tests := []struct {
		name       string
		encoding   string
		shouldFail bool
		errMsg     string
	}{
		{
			name:       "native X25519 key",
			encoding:   "AGE-SECRET-KEY-1RLNPSS8EV69RL40DKHUFUPU9SNWHUYYJQQMF3ZXQ7S4F3PTPS8EQ2RWVNA",
			shouldFail: false,
		},
		{
			name:       "plugin identity (AGE-PLUGIN-*)",
			encoding:   "AGE-PLUGIN-YUBIKEY-1GKZKJQYZL98RLMC67F9PJ",
			shouldFail: false,
		},
		{
			name:       "plugin identity with recipient suffix (gopass custom format)",
			encoding:   "AGE-PLUGIN-YUBIKEY-1GKZKJQYZL98RLMC67F9PJ|age1yubikey1qt2r3tfk7wvlykudm7ew28dqqm3h8ln9zfsxsq4lcd2w8rh4n4hhz46ur24",
			shouldFail: false,
		},
		{
			name:       "X25519 key with recipient suffix (gopass custom format)",
			encoding:   "AGE-SECRET-KEY-1RLNPSS8EV69RL40DKHUFUPU9SNWHUYYJQQMF3ZXQ7S4F3PTPS8EQ2RWVNA|age1xmwwc06ly3ee5rytxm9mflaz2u56jjj36s0mypdrwsvlul66mv6s23e5ep",
			shouldFail: false,
		},
		{
			name:       "unknown identity type",
			encoding:   "AGE-NONSECRET-KEY-TEST",
			shouldFail: true,
			errMsg:     "unknown identity type",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := parseIdentity(tt.encoding)
			if tt.shouldFail {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				// The error must NOT be the bech32 "malformed secret key: mixed case" error —
				// that would indicate the wrong parser branch was taken.
				assert.NotContains(t, err.Error(), "malformed secret key")
			} else {
				require.NoError(t, err)
				require.NotNil(t, id)
			}
		})
	}
}

// TestParseIdentityPluginNotMixedCase verifies that plugin identities (AGE-PLUGIN-*)
// are NOT rejected with the "malformed secret key: mixed case" error that
// age.ParseIdentities() would produce for these keys. This is the regression test
// for issue #3316.
func TestParseIdentityPluginNotMixedCase(t *testing.T) {
	pluginID := "AGE-PLUGIN-YUBIKEY-1GKZKJQYZL98RLMC67F9PJ"

	// The standard library age.ParseIdentities returns an error for plugin identities.
	_, stdErr := age.ParseIdentities(strings.NewReader(pluginID))
	require.Error(t, stdErr, "standard age.ParseIdentities should fail for plugin identities")

	// Our parseIdentity must NOT return the mixed-case error for plugin identities.
	id, err := parseIdentity(pluginID)
	require.NoError(t, err)
	require.NotNil(t, id)
	// And certainly not the bech32 mixed-case error.
	if err != nil {
		assert.NotContains(t, err.Error(), "malformed secret key")
		assert.NotContains(t, err.Error(), "mixed case")
	}
}

func TestParseIdentities(t *testing.T) {
	t.Run("single X25519 identity", func(t *testing.T) {
		input := "AGE-SECRET-KEY-1RLNPSS8EV69RL40DKHUFUPU9SNWHUYYJQQMF3ZXQ7S4F3PTPS8EQ2RWVNA\n"
		ids, err := parseIdentities(strings.NewReader(input))
		require.NoError(t, err)
		require.Len(t, ids, 1)
	})

	t.Run("plugin identity", func(t *testing.T) {
		input := "AGE-PLUGIN-YUBIKEY-1GKZKJQYZL98RLMC67F9PJ\n"
		ids, err := parseIdentities(strings.NewReader(input))
		require.NoError(t, err)
		require.Len(t, ids, 1)
	})

	t.Run("comments and blank lines are skipped", func(t *testing.T) {
		input := "# this is a comment\n\nAGE-SECRET-KEY-1RLNPSS8EV69RL40DKHUFUPU9SNWHUYYJQQMF3ZXQ7S4F3PTPS8EQ2RWVNA\n"
		ids, err := parseIdentities(strings.NewReader(input))
		require.NoError(t, err)
		require.Len(t, ids, 1)
	})

	t.Run("multiple identities", func(t *testing.T) {
		k1, err := age.GenerateX25519Identity()
		require.NoError(t, err)
		k2, err := age.GenerateX25519Identity()
		require.NoError(t, err)
		input := k1.String() + "\n" + k2.String() + "\n"
		ids, err := parseIdentities(strings.NewReader(input))
		require.NoError(t, err)
		require.Len(t, ids, 2)
	})

	t.Run("empty input returns error", func(t *testing.T) {
		_, err := parseIdentities(strings.NewReader(""))
		require.Error(t, err)
	})
}
