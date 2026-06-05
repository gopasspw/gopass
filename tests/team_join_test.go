package tests

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTeamJoinPreservesPublicKeys is a reproducer for GH-2620: after cloning
// a store (or equivalent join operation), the .public-keys/ directory must
// retain all recipient keys — not just the cloning member's own key.
//
// This test creates a multi-recipient store, simulates a "join" by reading
// .public-keys/ from the store, and verifies all expected keys are present.
func TestTeamJoinPreservesPublicKeys(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	// Create a second GPG key to represent a team member.
	key2 := createGPGKey(t, ts, "Teammate", "teammate@example.com")

	// Enable exportkeys in the config so .public-keys/ is populated during init.
	cfgPath := ts.gopassConfig()
	cfgData, err := os.ReadFile(cfgPath)
	require.NoError(t, err)
	cfgData = bytes.ReplaceAll(cfgData, []byte("exportkeys = false"), []byte("exportkeys = true"))
	require.NoError(t, os.WriteFile(cfgPath, cfgData, 0o600))

	// Init the store with both keys as recipients.
	out, err := ts.run("init --crypto=gpgcli --storage=fs " + keyID + " " + key2)
	require.NoError(t, err, "init: %s", out)

	storePath := ts.storeDir("root")

	// Verify .gpg-id contains both keys.
	gpgID, err := os.ReadFile(filepath.Join(storePath, ".gpg-id"))
	require.NoError(t, err)
	assert.Contains(t, string(gpgID), keyID)
	assert.Contains(t, string(gpgID), key2)

	// .public-keys/ should contain both keys.
	pubKeysDir := filepath.Join(storePath, ".public-keys")
	files, err := os.ReadDir(pubKeysDir)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(files), 2, ".public-keys/ must have entries for both recipients")

	// Verify the diagnostic tool reports both recipients are healthy.
	out, err = ts.run("doctor --recipients --verbose")
	require.NoError(t, err, "doctor --recipients: %s", out)
	assert.Contains(t, out, "local keyring")
}
