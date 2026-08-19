//go:build canary

package updater

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGPGVerifyIn6Months tests that the signature is still valid in 6 months.
// This is supposed to act as a canary so we don't forget to roll the key
// before it expires. See README.md for details.
func TestGPGVerifyIn6Months(t *testing.T) {
	t.Parallel()

	ok, err := gpgVerifyAt(testData, testSignature, func() time.Time { return time.Now().AddDate(0, 6, 0) })
	require.NoError(t, err, "If TestGPGVerify succeeds but this test fails the self-updater key is about to expire. Please open an issue to update the key. Thank you.")
	assert.True(t, ok)
}
