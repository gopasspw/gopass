package age

import (
	"fmt"
	"os"
	"path/filepath"
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

// newTestAge creates an Age instance whose identity file lives under a temp
// directory and uses a fixed test passphrase so no interactive prompt is
// needed during tests.
func newTestAge(t *testing.T) *Age {
	t.Helper()
	td := t.TempDir()
	t.Setenv("GOPASS_HOMEDIR", td)
	ctx := t.Context()
	a, err := New(ctx, "")
	require.NoError(t, err)
	// Override the identity path to a known temp location.
	a.identity = filepath.Join(td, "age", "identities")
	// Use a fixed passphrase so no UI interaction is needed.
	a.pwCallback = func(_ string, _ bool) ([]byte, error) {
		return []byte("test-passphrase"), nil
	}
	a.pwPurgeCallback = func(_ string) {}

	return a
}

// TestAddIdentityToNewFile verifies that addIdentity works when no identity
// file exists yet (i.e. creates a new file correctly).
func TestAddIdentityToNewFile(t *testing.T) {
	ctx := t.Context()
	a := newTestAge(t)

	id, err := age.GenerateX25519Identity()
	require.NoError(t, err)

	require.NoError(t, a.addIdentity(ctx, id))

	// The file should now exist and contain exactly the one identity.
	ids, err := a.Identities(ctx)
	require.NoError(t, err)
	require.Len(t, ids, 1)
	rec := IdentityToRecipient(ids[0])
	require.NotNil(t, rec)
	assert.Equal(t, id.Recipient().String(), fmt.Sprintf("%s", rec))
}

// TestAddIdentityDoesNotParseExistingPluginLines verifies Option A: when a
// plugin-format line is already present in the identity file, adding a new
// native identity does NOT re-invoke the plugin binary (because we only do a
// raw text append, not a parse-all-then-serialize cycle).
//
// We simulate this by writing a plugin-format raw line directly into the
// encrypted identity file via saveIdentities, then adding a new key. If the
// old code path were still active it would call parseIdentity on the plugin
// line, which calls plugin.NewIdentity() and would fail without the binary.
// With Option A, the plugin line is copied verbatim.
func TestAddIdentityPreservesPluginLineWithoutInvokingPlugin(t *testing.T) {
	ctx := t.Context()
	a := newTestAge(t)

	// Plant a plugin-format line (gopass custom format: identity|recipient).
	// This is the serialized form that saveIdentities / identitiesToString
	// would produce for a wrappedIdentity.
	pluginRaw := "AGE-PLUGIN-YUBIKEY-1GKZKJQYZL98RLMC67F9PJ|age1yubikey1qt2r3tfk7wvlykudm7ew28dqqm3h8ln9zfsxsq4lcd2w8rh4n4hhz46ur24"
	require.NoError(t, a.saveIdentities(ctx, []string{pluginRaw}, true))

	// Now add a real native identity. Option A must NOT try to call
	// plugin.NewIdentity() on the existing pluginRaw line.
	newID, err := age.GenerateX25519Identity()
	require.NoError(t, err)
	require.NoError(t, a.addIdentity(ctx, newID))

	// Read back the raw file and confirm both lines are present.
	raw, err := a.loadIdentityFile(ctx)
	require.NoError(t, err)

	assert.Contains(t, raw, pluginRaw, "plugin line must be preserved verbatim")
	assert.Contains(t, raw, newID.String(), "new native identity must be appended")
}

// TestAddMultipleIdentitiesAccumulate verifies that calling addIdentity
// multiple times accumulates all identities in the file, each as its own line.
func TestAddMultipleIdentitiesAccumulate(t *testing.T) {
	ctx := t.Context()
	a := newTestAge(t)

	id1, err := age.GenerateX25519Identity()
	require.NoError(t, err)
	id2, err := age.GenerateX25519Identity()
	require.NoError(t, err)
	id3, err := age.GenerateX25519Identity()
	require.NoError(t, err)

	require.NoError(t, a.addIdentity(ctx, id1))
	require.NoError(t, a.addIdentity(ctx, id2))
	require.NoError(t, a.addIdentity(ctx, id3))

	ids, err := a.Identities(ctx)
	require.NoError(t, err)
	require.Len(t, ids, 3)
}

// TestLoadIdentityFileNotExist verifies that loadIdentityFile returns an
// os.ErrNotExist-wrapped error when the identity file has not been created yet.
func TestLoadIdentityFileNotExist(t *testing.T) {
	a := newTestAge(t)
	ctx := t.Context()

	_, err := a.loadIdentityFile(ctx)
	require.Error(t, err)
	assert.ErrorIs(t, err, os.ErrNotExist,
		"loadIdentityFile should surface an os.ErrNotExist-compatible error")
}
