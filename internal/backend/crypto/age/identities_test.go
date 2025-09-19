package agecrypto

import (
	"fmt"
	"testing"

	"filippo.io/age"
	"filippo.io/age/plugin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseIdentity(t *testing.T) {
	tests := []struct {
		name         string
		encoding     string
		expectedType any
		shouldFail   bool
	}{
		{
			"plugin id",
			"AGE-PLUGIN-YUBIKEY-1GKZKJQYZL98RLMC67F9PJ",
			&wrappedIdentity{},
			false,
		},
		{
			"native age id",
			"AGE-SECRET-KEY-1RLNPSS8EV69RL40DKHUFUPU9SNWHUYYJQQMF3ZXQ7S4F3PTPS8EQ2RWVNA",
			&age.X25519Identity{},
			false,
		},
		{
			"invalid age id",
			"AGE-NONSECRET-KEY-TEST",
			nil,
			true,
		},
		{
			"invalid bech32 plugin",
			"AGE-PLUGIN-YUBIKEY-1GKKJQYZL98RLM7FJ",
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tID, err := parseIdentity(tt.encoding)
			if tt.shouldFail {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.IsType(t, tt.expectedType, tID)
			}
		})
	}
}

func TestIdentityAndRecipient(t *testing.T) {
	testID, err := age.GenerateX25519Identity()
	require.NoError(t, err)

	pluginID, err := plugin.NewIdentity("AGE-PLUGIN-YUBIKEY-1GKZKJQYZL98RLMC67F9PJ", nil)
	require.NoError(t, err)
	pluginRec, err := plugin.NewRecipient("age1yubikey1qt2r3tfk7wvlykudm7ew28dqqm3h8ln9zfsxsq4lcd2w8rh4n4hhz46ur24", nil)
	require.NoError(t, err)
	wrRec := &wrappedRecipient{
		rec:      pluginRec,
		encoding: "age0yubikey1qt2r3tfk7wvlykudm7ew28dqqm3h8ln9zfsxsq4lcd2w8rh4n4hhz46ur24",
	}
	tests := []struct {
		name string
		id   age.Identity
		want age.Recipient
	}{
		{
			"native identity",
			testID,
			testID.Recipient(),
		},
		{
			"wrapped native id",
			&wrappedIdentity{
				id:       testID,
				rec:      testID.Recipient(),
				encoding: testID.String(),
			},
			testID.Recipient(),
		},
		{
			"wrapped plugin id",
			&wrappedIdentity{
				id:       pluginID,
				rec:      wrRec,
				encoding: "AGE-PLUGIN-YUBIKEY-1GKZKJQYZL98RLMC67F9PJ",
			},
			wrRec,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, IdentityToRecipient(tt.id), "IdentityToRecipient(%v)", tt.id)
			// ensure recipient strings aren't equal identity strings
			assert.NotEqual(t, fmt.Sprintf("%s", tt.id), fmt.Sprintf("%s", tt.want))
			// ensure Parsing works on the String of the id:
			_, err = parseIdentity(fmt.Sprintf("%s", tt.id))
			require.NoError(t, err)
		})
	}
}
