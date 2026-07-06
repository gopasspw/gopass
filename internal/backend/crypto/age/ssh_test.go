package age

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetSSHIdentitiesCustomSshKeyPathTilde is a regression test for the bug
// where a custom age.ssh-key-path with a leading ~/ was silently ignored:
// fsutil.IsDir("~/custom-ssh") is false (no shell expansion), so the custom
// path was skipped and only the default ~/.ssh discovery ran.
//
// Setup: GOPASS_HOMEDIR points at a temp dir, the custom dir lives at
// <td>/custom-ssh, and there is NO <td>/.ssh and no GOPASS_SSH_DIR. So the only
// way getSSHIdentities returns non-error is if the ~/custom-ssh path is
// expanded to <td>/custom-ssh and recognized as a directory.
//
// Before the fix this returned ErrNoSSHDir; after the fix the dir is found and
// searched (no error).
func TestGetSSHIdentitiesCustomSshKeyPathTilde(t *testing.T) {
	td := t.TempDir()
	t.Setenv("GOPASS_HOMEDIR", td)
	t.Setenv("GOPASS_SSH_DIR", "")

	require.NoError(t, os.MkdirAll(filepath.Join(td, "custom-ssh"), 0o700))

	// The package-level sshCache is never reset elsewhere; clear it so this test
	// observes a fresh filesystem read, and clear it again afterward so a
	// populated cache cannot leak to later tests. Safe because no test in this
	// package touches sshCache in parallel. (same-package access to the var.)
	sshCacheMu.Lock()
	sshCache = nil
	sshCacheMu.Unlock()
	t.Cleanup(func() {
		sshCacheMu.Lock()
		sshCache = nil
		sshCacheMu.Unlock()
	})

	a, err := New(t.Context(), "~/custom-ssh")
	require.NoError(t, err)

	// Before the fix this returned ErrNoSSHDir (literal ~ is not a dir). After
	// the fix the expanded dir is found and searched: assert no error and that a
	// (possibly empty) identities map is returned.
	ids, err := a.getSSHIdentities(t.Context())
	require.NoError(t, err)
	assert.NotNil(t, ids)
}
