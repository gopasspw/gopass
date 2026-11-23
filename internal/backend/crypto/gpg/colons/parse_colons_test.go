package colons

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseColonIdentity(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		in      string
		uid     string
		name    string
		comment string
		email   string
	}{
		{
			in:      "uid:-::::1460666077::780A2FDD0570B3E52E5B1E24EBB406B68526CAFD::ThisIsNotAnAlias:",
			uid:     "ThisIsNotAnAlias",
			name:    "ThisIsNotAnAlias",
			comment: "",
			email:   "",
		},
		{
			in:      "uid:::::1441103821::AEFC3F5B6CAD79A946D7F0FF83BB8B7E10B578CA::John Doe <john.doe@example.com>:",
			uid:     "John Doe <john.doe@example.com>",
			name:    "John Doe",
			comment: "",
			email:   "john.doe@example.com",
		},
		{
			in:      "uid:::::1441103821::AEFC3F5B6CAD79A946D7F0FF83BB8B7E10B578CA::John Doe (user) <john.doe@example.com>:",
			uid:     "John Doe (user) <john.doe@example.com>",
			name:    "John Doe",
			comment: "user",
			email:   "john.doe@example.com",
		},
		{
			in:      "uid:::::1441103821::AEFC3F5B6CAD79A946D7F0FF83BB8B7E10B578CA::John Doe (user):",
			uid:     "John Doe (user)",
			name:    "John Doe",
			comment: "user",
			email:   "",
		},
	} {
		t.Run(tc.in, func(t *testing.T) {
			t.Parallel()

			gi := parseColonIdentity(strings.Split(tc.in, ":"))
			assert.Equal(t, tc.uid, gi.UID)
			assert.Equal(t, tc.name, gi.Name)
			assert.Equal(t, tc.comment, gi.Comment)
			assert.Equal(t, tc.email, gi.Email)
		})
	}
}

func TestParseSubKeys(t *testing.T) {
	t.Parallel()

	input := strings.Join([]string{
		"tru::1:1795431639:0:3:1:5",
		"pub:u:255:22:A4467D31995B04EA:1795431639:::u:::scESC:::::ed25519:::0:",
		"fpr:::::::::E00B91241B8C95802B0782ECA4467D31995B04EA:",
		"uid:u::::1795431639::5FF5D9003D4880F850B6F88D44B7A295707886A2::Subkey Test <subkey@example.com>::::::::::0:",
		"sub:u:255:18:24E8B312C3A2051E:1795431639::::::e:::::cv25519::",
		"fpr:::::::::D0918DBF23B72AF4ADAF3DF324E8B312C3A2051E:",
		"sub:e:255:18:FEFBDF96D56494C3:1500000000:1600000000:::::e:::::cv25519::",
		"fpr:::::::::82AD5B824BB983C4667E076CFEFBDF96D56494C3:",
		"sub:u:255:22:1122334455667788:1795431639::::::s:::::ed25519::",
		"fpr:::::::::AABBCCDDEEFF0011223344551122334455667788:",
	}, "\n")

	kl := Parse(strings.NewReader(input))
	require.Len(t, kl, 1)

	k := kl[0]
	assert.Equal(t, "E00B91241B8C95802B0782ECA4467D31995B04EA", k.Fingerprint)
	require.Len(t, k.SubKeyList, 3)

	// SubKeys map should contain both 16-char IDs and 40-char fingerprints.
	assert.Contains(t, k.SubKeys, "24E8B312C3A2051E")
	assert.Contains(t, k.SubKeys, "D0918DBF23B72AF4ADAF3DF324E8B312C3A2051E")
	assert.Contains(t, k.SubKeys, "82AD5B824BB983C4667E076CFEFBDF96D56494C3")

	// Valid encryption subkey matches full fingerprint, lowercase 0x ID, and trailing !.
	for _, query := range []string{
		"D0918DBF23B72AF4ADAF3DF324E8B312C3A2051E!",
		"d0918dbf23b72af4adaf3df324e8b312c3a2051e!",
		"0x24e8b312c3a2051e!",
		"24E8B312C3A2051E",
	} {
		sk, ok := k.MatchSubKey(query)
		require.True(t, ok, "query=%q", query)
		assert.Equal(t, "D0918DBF23B72AF4ADAF3DF324E8B312C3A2051E", sk.Fingerprint)
		assert.True(t, sk.IsUseable(false))
		assert.True(t, sk.IsUseable(true))
	}

	// Expired encryption subkey is never useable, even with alwaysTrust=true.
	skExp, ok := k.MatchSubKey("82AD5B824BB983C4667E076CFEFBDF96D56494C3!")
	require.True(t, ok)
	assert.False(t, skExp.IsUseable(false))
	assert.False(t, skExp.IsUseable(true))

	// Signing-only subkey is not useable for encryption.
	skSign, ok := k.MatchSubKey("AABBCCDDEEFF0011223344551122334455667788!")
	require.True(t, ok)
	assert.True(t, skSign.Caps.Sign)
	assert.False(t, skSign.Caps.Encrypt)
	assert.False(t, skSign.IsUseable(false))
	assert.False(t, skSign.IsUseable(true))
}
