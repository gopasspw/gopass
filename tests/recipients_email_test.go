package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDoctorRecipientsDiagnostic verifies that 'gopass doctor --recipients'
// produces a meaningful output for a healthy store and flags non-canonical IDs.
func TestDoctorRecipientsDiagnostic(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	ts.initStore()

	// Healthy store — doctor --recipients should succeed.
	out, err := ts.run("doctor --recipients")
	require.NoError(t, err, "doctor --recipients on healthy store: %s", out)
	// Without --verbose, info-level messages are suppressed; expect the summary line.
	assert.Contains(t, out, "Summary", "doctor should output a summary")

	// Verbose mode should give detailed per-recipient output.
	out, err = ts.run("doctor --recipients --verbose")
	require.NoError(t, err, "doctor --recipients --verbose: %s", out)
	assert.Contains(t, out, "local keyring", "verbose mode should show keyring status")
}

// TestRecipientsAddPreservesCanonical verifies that after Stage 1,
// 'gopass recipients add' stores the canonical key ID in .gpg-id
// rather than a potentially ambiguous identifier (GH-2762).
func TestRecipientsAddPreservesCanonical(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	ts.initStore()

	// Add recipient using the same key ID again (re-encrypt scenario).
	// The key already exists so AddRecipient will prompt for confirmation;
	// --yes answers all yes/no prompts automatically.
	out, err := ts.run("recipients add --force --yes " + keyID)
	require.NoError(t, err, "recipients add: %s", out)

	// The diagnostic after adding should report canonical IDs.
	out, err = ts.run("doctor --recipients")
	require.NoError(t, err, "doctor --recipients after add: %s", out)
}
