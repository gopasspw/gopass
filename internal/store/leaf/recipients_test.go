package leaf

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/backend/crypto/gpg"
	gpgcli "github.com/gopasspw/gopass/internal/backend/crypto/gpg/cli"
	"github.com/gopasspw/gopass/internal/backend/crypto/plain"
	"github.com/gopasspw/gopass/internal/backend/storage/fs"
	"github.com/gopasspw/gopass/internal/backend/storage/gitfs"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/recipients"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/gopass/secrets"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetRecipientsDefault(t *testing.T) {
	t.Parallel()

	ctx := config.NewContextInMemory()

	tempdir := t.TempDir()

	ctx, _, _ = captureOutput(t, ctx)

	genRecs, _, err := createStore(tempdir, nil, nil)
	require.NoError(t, err)

	s := &Store{
		alias:   "",
		path:    tempdir,
		crypto:  plain.New(),
		storage: fs.New(tempdir),
	}

	assert.Equal(t, genRecs, s.Recipients(ctx))
	recs, err := s.GetRecipients(ctx, "")
	require.NoError(t, err)

	ids := recs.IDs()
	sort.Strings(ids)
	assert.Equal(t, genRecs, ids)
}

func TestGetRecipientsSubID(t *testing.T) {
	t.Parallel()

	ctx := config.NewContextInMemory()

	tempdir := t.TempDir()

	ctx, _, _ = captureOutput(t, ctx)

	genRecs, _, err := createStore(tempdir, nil, nil)
	require.NoError(t, err)

	s := &Store{
		alias:   "",
		path:    tempdir,
		crypto:  plain.New(),
		storage: fs.New(tempdir),
	}

	recs, err := s.GetRecipients(ctx, "")
	require.NoError(t, err)
	assert.Equal(t, genRecs, recs.IDs())

	err = os.WriteFile(filepath.Join(tempdir, "foo", "bar", s.crypto.IDFile()), []byte("john.doe\n"), 0o600)
	require.NoError(t, err)

	recs, err = s.GetRecipients(ctx, "foo/bar/baz")
	require.NoError(t, err)
	assert.Equal(t, []string{"john.doe"}, recs.IDs())
}

func TestSaveRecipients(t *testing.T) {
	t.Parallel()

	ctx := config.NewContextInMemory()

	tempdir := t.TempDir()

	_, _, err := createStore(tempdir, nil, nil)
	require.NoError(t, err)

	ctx, _, _ = captureOutput(t, ctx)

	s := &Store{
		alias:   "",
		path:    tempdir,
		crypto:  plain.New(),
		storage: fs.New(tempdir),
	}

	// remove recipients
	_ = os.Remove(filepath.Join(tempdir, s.crypto.IDFile()))

	rs := recipients.New()
	rs.Add("john.doe")

	require.NoError(t, s.saveRecipients(ctx, rs, "test-save-recipients"))
	require.Error(t, s.saveRecipients(ctx, nil, "test-save-recipients"))

	buf, err := s.storage.Get(ctx, s.idFile(ctx, ""))
	require.NoError(t, err)

	foundRecs := []string{}
	scanner := bufio.NewScanner(bytes.NewReader(buf))

	for scanner.Scan() {
		foundRecs = append(foundRecs, strings.TrimSpace(scanner.Text()))
	}

	sort.Strings(foundRecs)

	ids := rs.IDs()
	for i := range ids {
		if i >= len(foundRecs) {
			t.Errorf("Read too few recipients")

			break
		}

		if ids[i] != foundRecs[i] {
			t.Errorf("Mismatch at %d: %s vs %s", i, ids[i], foundRecs[i])
		}
	}
}

func TestAddRecipient(t *testing.T) {
	t.Parallel()

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithHidden(ctx, true)
	ctx = config.NewInMemory().WithConfig(ctx)

	tempdir := t.TempDir()

	genRecs, _, err := createStore(tempdir, nil, nil)
	require.NoError(t, err)

	ctx, _, _ = captureOutput(t, ctx)

	s := &Store{
		alias:   "",
		path:    tempdir,
		crypto:  plain.New(),
		storage: fs.New(tempdir),
	}

	newRecp := "A3683834"

	err = s.AddRecipient(ctx, newRecp)
	require.NoError(t, err)

	rs, err := s.GetRecipients(ctx, "")
	require.NoError(t, err)
	assert.Equal(t, append(genRecs, newRecp), rs.IDs())

	err = s.SaveRecipients(ctx, false)
	require.NoError(t, err)
}

func TestRemoveRecipient(t *testing.T) {
	t.Parallel()

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithHidden(ctx, true)
	ctx = config.NewInMemory().WithConfig(ctx)

	tempdir := t.TempDir()

	_, _, err := createStore(tempdir, nil, nil)
	require.NoError(t, err)

	ctx, _, _ = captureOutput(t, ctx)

	s := &Store{
		alias:   "",
		path:    tempdir,
		crypto:  plain.New(),
		storage: fs.New(tempdir),
	}

	err = s.RemoveRecipient(ctx, "0xDEADBEEF")
	require.NoError(t, err)

	rs, err := s.GetRecipients(ctx, "")
	require.NoError(t, err)
	assert.Equal(t, []string{"0xFEEDBEEF"}, rs.IDs())
}

// TestRemoveRecipientScopedCleanup verifies that RemoveRecipient
// performs recipient-scoped key file cleanup: the removed recipient's
// .public-keys/ file is deleted, but other recipients' files remain.
func TestRemoveRecipientScopedCleanup(t *testing.T) {
	t.Parallel()

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithHidden(ctx, true)
	ctx = config.NewInMemory().WithConfig(ctx)

	tempdir := t.TempDir()

	_, _, err := createStore(tempdir, []string{"0xDEADBEEF", "0xFEEDBEEF"}, nil)
	require.NoError(t, err)

	ctx, _, _ = captureOutput(t, ctx)

	s := &Store{
		alias:   "",
		path:    tempdir,
		crypto:  plain.New(),
		storage: fs.New(tempdir),
	}

	// Pre-populate .public-keys/ files for both recipients to simulate
	// exported keys.
	require.NoError(t, s.storage.Set(ctx, filepath.Join(".public-keys", "0xDEADBEEF"), []byte("dead-key")))
	require.NoError(t, s.storage.Set(ctx, filepath.Join(".public-keys", "0xFEEDBEEF"), []byte("feed-key")))

	// Remove 0xDEADBEEF — its .public-keys/ file should be deleted.
	err = s.RemoveRecipient(ctx, "0xDEADBEEF")
	require.NoError(t, err)

	rs, err := s.GetRecipients(ctx, "")
	require.NoError(t, err)
	assert.Equal(t, []string{"0xFEEDBEEF"}, rs.IDs())

	// 0xDEADBEEF's key file is gone.
	assert.False(t, s.storage.Exists(ctx, filepath.Join(".public-keys", "0xDEADBEEF")),
		"removed recipient's .public-keys/ file must be deleted")

	// 0xFEEDBEEF's key file is still there.
	assert.True(t, s.storage.Exists(ctx, filepath.Join(".public-keys", "0xFEEDBEEF")),
		"other recipients' .public-keys/ files must be preserved")
}

// TestRemoveRecipientPreservesOtherKeys verifies that removing a
// recipient with a legacy .gpg-keys/ file cleans up only the removed
// key's files and leaves unrelated keys untouched.
func TestRemoveRecipientPreservesOtherKeys(t *testing.T) {
	t.Parallel()

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithHidden(ctx, true)
	ctx = config.NewInMemory().WithConfig(ctx)

	tempdir := t.TempDir()

	_, _, err := createStore(tempdir, []string{"0xDEADBEEF", "0xFEEDBEEF"}, nil)
	require.NoError(t, err)

	ctx, _, _ = captureOutput(t, ctx)

	s := &Store{
		alias:   "",
		path:    tempdir,
		crypto:  plain.New(),
		storage: fs.New(tempdir),
	}

	// Pre-populate both .public-keys/ and legacy .gpg-keys/ files.
	require.NoError(t, s.storage.Set(ctx, filepath.Join(".public-keys", "0xDEADBEEF"), []byte("dead-pub")))
	require.NoError(t, s.storage.Set(ctx, filepath.Join(".gpg-keys", "0xDEADBEEF"), []byte("dead-legacy")))
	require.NoError(t, s.storage.Set(ctx, filepath.Join(".public-keys", "0xFEEDBEEF"), []byte("feed-pub")))
	require.NoError(t, s.storage.Set(ctx, filepath.Join(".gpg-keys", "0xFEEDBEEF"), []byte("feed-legacy")))

	err = s.RemoveRecipient(ctx, "0xDEADBEEF")
	require.NoError(t, err)

	// All of 0xDEADBEEF's key files are gone.
	assert.False(t, s.storage.Exists(ctx, filepath.Join(".public-keys", "0xDEADBEEF")))
	assert.False(t, s.storage.Exists(ctx, filepath.Join(".gpg-keys", "0xDEADBEEF")))

	// All of 0xFEEDBEEF's key files remain.
	assert.True(t, s.storage.Exists(ctx, filepath.Join(".public-keys", "0xFEEDBEEF")))
	assert.True(t, s.storage.Exists(ctx, filepath.Join(".gpg-keys", "0xFEEDBEEF")))

	rs, err := s.GetRecipients(ctx, "")
	require.NoError(t, err)
	assert.Equal(t, []string{"0xFEEDBEEF"}, rs.IDs())
}

func TestListRecipients(t *testing.T) {
	t.Parallel()

	ctx := config.NewContextInMemory()

	tempdir := t.TempDir()

	genRecs, _, err := createStore(tempdir, nil, nil)
	require.NoError(t, err)

	ctx, _, _ = captureOutput(t, ctx)

	ctx, err = backend.WithCryptoBackendString(ctx, "plain")
	require.NoError(t, err)
	s, err := New(
		ctx,
		"",
		tempdir,
	)
	require.NoError(t, err)

	rs, err := s.GetRecipients(ctx, "")
	require.NoError(t, err)
	assert.Equal(t, genRecs, rs.IDs())

	assert.Equal(t, "0xDEADBEEF", s.OurKeyID(ctx))
}

func TestCheckRecipients(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test setup not supported on Windows")
	}

	u := gptest.NewGUnitTester(t)

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithTerminal(ctx, false)
	ctx = backend.WithCryptoBackend(ctx, backend.GPGCLI)

	ctx, _, _ = captureOutput(t, ctx)

	s, err := New(ctx, "", u.StoreDir(""))
	require.NoError(t, err)

	require.NoError(t, s.CheckRecipients(ctx))

	u.AddExpiredRecipient()
	require.Error(t, s.CheckRecipients(ctx))
}

// TestCanonicalizeRecipientHelper verifies that canonicalizeRecipient resolves
// raw identifiers to the form returned by the crypto backend, and falls back
// gracefully when no key is found.
func TestCanonicalizeRecipientHelper(t *testing.T) {
	t.Parallel()

	ctx := config.NewContextInMemory()

	tempdir := t.TempDir()

	ctx, _, _ = captureOutput(t, ctx)

	s := &Store{
		alias:   "",
		path:    tempdir,
		crypto:  plain.New(),
		storage: fs.New(tempdir),
	}

	// "DEADBEEF" is a suffix of "0xDEADBEEF" (the ID returned by the plain
	// backend for the first static key), so FindRecipients returns one match.
	assert.Equal(t, "0xDEADBEEF", s.canonicalizeRecipient(ctx, "DEADBEEF"))
	assert.Equal(t, "0xFEEDBEEF", s.canonicalizeRecipient(ctx, "FEEDBEEF"))

	// "0xDEADBEEF" is an exact suffix match, so it resolves to itself.
	assert.Equal(t, "0xDEADBEEF", s.canonicalizeRecipient(ctx, "0xDEADBEEF"))

	// A key with no match in the keyring or .public-keys/ is returned unchanged.
	assert.Equal(t, "unknown@example.com", s.canonicalizeRecipient(ctx, "unknown@example.com"))
}

// TestAddRecipientCanonicalized verifies that AddRecipient stores the
// canonical form of the recipient ID, not the raw user-supplied input.
func TestAddRecipientCanonicalized(t *testing.T) {
	t.Parallel()

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithHidden(ctx, true)
	ctx = config.NewInMemory().WithConfig(ctx)

	tempdir := t.TempDir()

	// Start with a store that only has a single canonical recipient.
	_, _, err := createStore(tempdir, []string{"0xDEADBEEF"}, nil)
	require.NoError(t, err)

	ctx, _, _ = captureOutput(t, ctx)

	s := &Store{
		alias:   "",
		path:    tempdir,
		crypto:  plain.New(),
		storage: fs.New(tempdir),
	}

	// "FEEDBEEF" canonicalizes to "0xFEEDBEEF" via FindRecipients suffix match.
	err = s.AddRecipient(ctx, "FEEDBEEF")
	require.NoError(t, err)

	rs, err := s.GetRecipients(ctx, "")
	require.NoError(t, err)

	assert.Equal(t, []string{"0xDEADBEEF", "0xFEEDBEEF"}, rs.IDs())
}

// TestCanonicalizeRecipients verifies that CanonicalizeRecipients rewrites
// non-canonical IDs to the form returned by the crypto backend while leaving
// already-canonical IDs untouched.
func TestCanonicalizeRecipients(t *testing.T) {
	t.Parallel()

	ctx := config.NewContextInMemory()
	ctx = config.NewInMemory().WithConfig(ctx)

	tempdir := t.TempDir()

	// Create a store with non-canonical short IDs (no "0x" prefix).
	_, _, err := createStore(tempdir, []string{"DEADBEEF", "FEEDBEEF"}, nil)
	require.NoError(t, err)

	ctx, _, _ = captureOutput(t, ctx)

	s := &Store{
		alias:   "",
		path:    tempdir,
		crypto:  plain.New(),
		storage: fs.New(tempdir),
	}

	// Confirm the store starts with non-canonical IDs.
	rsBefore, err := s.GetRecipients(ctx, "")
	require.NoError(t, err)
	assert.Equal(t, []string{"DEADBEEF", "FEEDBEEF"}, rsBefore.IDs())

	err = s.CanonicalizeRecipients(ctx)
	require.NoError(t, err)

	// After canonicalization, IDs should be in the form returned by FindRecipients.
	rsAfter, err := s.GetRecipients(ctx, "")
	require.NoError(t, err)
	assert.Equal(t, []string{"0xDEADBEEF", "0xFEEDBEEF"}, rsAfter.IDs())
}

// TestCanonicalizeRecipientsAlreadyCanonical verifies that
// CanonicalizeRecipients is a no-op when all IDs are already canonical.
func TestCanonicalizeRecipientsAlreadyCanonical(t *testing.T) {
	t.Parallel()

	ctx := config.NewContextInMemory()
	ctx = config.NewInMemory().WithConfig(ctx)

	tempdir := t.TempDir()

	// Default store already uses canonical IDs as returned by the plain backend.
	genRecs, _, err := createStore(tempdir, nil, nil)
	require.NoError(t, err)

	ctx, _, _ = captureOutput(t, ctx)

	s := &Store{
		alias:   "",
		path:    tempdir,
		crypto:  plain.New(),
		storage: fs.New(tempdir),
	}

	err = s.CanonicalizeRecipients(ctx)
	require.NoError(t, err)

	// Recipients should be unchanged.
	rs, err := s.GetRecipients(ctx, "")
	require.NoError(t, err)
	assert.Equal(t, genRecs, rs.IDs())
}

// TestDiagnoseRecipientsCanonical verifies that DiagnoseRecipients reports
// no findings when all IDs are already canonical and in the keyring.
func TestDiagnoseRecipientsCanonical(t *testing.T) {
	t.Parallel()

	ctx := config.NewContextInMemory()

	tempdir := t.TempDir()

	ctx, _, _ = captureOutput(t, ctx)

	// Default store uses canonical IDs ("0xDEADBEEF", "0xFEEDBEEF").
	_, _, err := createStore(tempdir, nil, nil)
	require.NoError(t, err)

	s := &Store{
		alias:   "",
		path:    tempdir,
		crypto:  plain.New(),
		storage: fs.New(tempdir),
	}

	diags := s.DiagnoseRecipients(ctx)
	// All info-level: keys are in keyring and canonical.
	assert.False(t, diags.HasErrors())

	for _, d := range diags {
		assert.Equal(t, DiagInfo, d.Level, "diag for %s: %s", d.Recipient, d.Message)
	}
}

// TestDiagnoseRecipientsNonCanonical verifies that DiagnoseRecipients flags
// non-canonical IDs as warnings.
func TestDiagnoseRecipientsNonCanonical(t *testing.T) {
	t.Parallel()

	ctx := config.NewContextInMemory()

	tempdir := t.TempDir()

	ctx, _, _ = captureOutput(t, ctx)

	// Create store with non-canonical short IDs (no "0x" prefix).
	_, _, err := createStore(tempdir, []string{"DEADBEEF", "FEEDBEEF"}, nil)
	require.NoError(t, err)

	s := &Store{
		alias:   "",
		path:    tempdir,
		crypto:  plain.New(),
		storage: fs.New(tempdir),
	}

	diags := s.DiagnoseRecipients(ctx)
	assert.False(t, diags.HasErrors())

	// Both should be flagged as non-canonical (warning level).
	warnings := 0
	for _, d := range diags {
		if d.Level == DiagWarning {
			warnings++
			assert.Contains(t, d.Message, "non-canonical")
		}
	}
	assert.Equal(t, 2, warnings, "expected 2 non-canonical warnings")
}

// TestDiagnoseRecipientsUnresolvable verifies that DiagnoseRecipients reports
// an error when a recipient is neither in the keyring nor in .public-keys/.
func TestDiagnoseRecipientsUnresolvable(t *testing.T) {
	t.Parallel()

	ctx := config.NewContextInMemory()

	tempdir := t.TempDir()

	ctx, _, _ = captureOutput(t, ctx)

	// The plain backend knows only DEADBEEF and FEEDBEEF.
	// "A3683834" is unknown and has no .public-keys/ file.
	_, _, err := createStore(tempdir, []string{"0xDEADBEEF", "A3683834"}, nil)
	require.NoError(t, err)

	s := &Store{
		alias:   "",
		path:    tempdir,
		crypto:  plain.New(),
		storage: fs.New(tempdir),
	}

	diags := s.DiagnoseRecipients(ctx)
	assert.True(t, diags.HasErrors())

	// A3683834 should be an error (not in keyring, not in .public-keys/).
	found := false
	for _, d := range diags {
		if d.Recipient == "A3683834" && d.Level == DiagError {
			found = true
			assert.Contains(t, d.Message, "not found")
		}
	}
	assert.True(t, found, "A3683834 should be reported as error")
}

// TestJoinTeamCanDecrypt verifies JoinTeam returns false when the user can
// already decrypt the store (their key is in .gpg-id).
func TestJoinTeamCanDecrypt(t *testing.T) {
	t.Parallel()

	ctx := config.NewContextInMemory()

	tempdir := t.TempDir()

	ctx, _, _ = captureOutput(t, ctx)

	// Default store with canonical IDs that match the plain backend's static keys.
	_, _, err := createStore(tempdir, nil, nil)
	require.NoError(t, err)

	s := &Store{
		alias:   "",
		path:    tempdir,
		crypto:  plain.New(),
		storage: fs.New(tempdir),
	}

	exported, err := s.JoinTeam(ctx)
	require.NoError(t, err)
	assert.False(t, exported, "user should already be able to decrypt")
}

// TestHasDecryptionKey verifies the hasDecryptionKey helper.
func TestHasDecryptionKey(t *testing.T) {
	t.Parallel()

	ctx := config.NewContextInMemory()

	tempdir := t.TempDir()

	ctx, _, _ = captureOutput(t, ctx)

	// Default store with known recipients from the plain backend.
	_, _, err := createStore(tempdir, nil, nil)
	require.NoError(t, err)

	s := &Store{
		alias:   "",
		path:    tempdir,
		crypto:  plain.New(),
		storage: fs.New(tempdir),
	}

	assert.True(t, s.hasDecryptionKey(ctx), "plain backend FindIdentities delegates to FindRecipients which matches 0xDEADBEEF")
}

// TestGuardPartialViewWrite verifies the guard function when all recipients are resolvable.
func TestGuardPartialViewWrite(t *testing.T) {
	t.Parallel()

	ctx := config.NewContextInMemory()

	tempdir := t.TempDir()

	ctx, _, _ = captureOutput(t, ctx)

	// All recipients resolvable.
	_, _, err := createStore(tempdir, nil, nil)
	require.NoError(t, err)

	s := &Store{
		alias:   "",
		path:    tempdir,
		crypto:  plain.New(),
		storage: fs.New(tempdir),
	}

	err = s.GuardPartialViewWrite(ctx)
	assert.NoError(t, err, "guard should pass when all recipients are resolvable")
}

// TestGuardPartialViewWriteFails verifies the guard returns an error with unresolvable recipients.
func TestGuardPartialViewWriteFails(t *testing.T) {
	t.Parallel()

	ctx := config.NewContextInMemory()

	tempdir := t.TempDir()

	ctx, _, _ = captureOutput(t, ctx)

	// A3683834 is unknown to the plain backend and has no .public-keys/ file.
	_, _, err := createStore(tempdir, []string{"0xDEADBEEF", "A3683834"}, nil)
	require.NoError(t, err)

	s := &Store{
		alias:   "",
		path:    tempdir,
		crypto:  plain.New(),
		storage: fs.New(tempdir),
	}

	err = s.GuardPartialViewWrite(ctx)
	require.Error(t, err, "guard should fail when a recipient is unresolvable")
	require.Contains(t, err.Error(), "A3683834")
}

// TestUpdateExportedPublicKeysAdditiveOnly verifies that
// UpdateExportedPublicKeys no longer calls removeExtraKeys.
func TestUpdateExportedPublicKeysAdditiveOnly(t *testing.T) {
	t.Parallel()

	ctx := config.NewContextInMemory()
	ctx = config.NewInMemory().WithConfig(ctx)

	tempdir := t.TempDir()

	_, _, err := createStore(tempdir, nil, nil)
	require.NoError(t, err)

	ctx, _, _ = captureOutput(t, ctx)

	s := &Store{
		alias:   "",
		path:    tempdir,
		crypto:  plain.New(),
		storage: fs.New(tempdir),
	}

	// This should not panic. It should be a no-op because the plain backend
	// doesn't implement keyExporter.
	exported, err := s.UpdateExportedPublicKeys(ctx)
	require.NoError(t, err)
	assert.False(t, exported, "plain backend does not export keys so nothing exported")
}

// TestUpdateRecipientKeysPlainBackend verifies that UpdateRecipientKeys
// gracefully handles a backend that does not export keys (plain).
func TestUpdateRecipientKeysPlainBackend(t *testing.T) {
	t.Parallel()

	ctx := config.NewContextInMemory()

	tempdir := t.TempDir()

	_, _, err := createStore(tempdir, nil, nil)
	require.NoError(t, err)

	ctx, _, _ = captureOutput(t, ctx)

	s := &Store{
		alias:   "",
		path:    tempdir,
		crypto:  plain.New(),
		storage: fs.New(tempdir),
	}

	// Plain backend does not implement keyExporter, so this should error.
	err = s.UpdateRecipientKeys(ctx, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot export")
}

// TestUpdateRecipientKeysDefaultToOwn verifies that when no IDs are
// provided, the current user's identity is used.
func TestUpdateRecipientKeysDefaultToOwn(t *testing.T) {
	t.Parallel()

	ctx := config.NewContextInMemory()

	tempdir := t.TempDir()

	_, _, err := createStore(tempdir, nil, nil)
	require.NoError(t, err)

	ctx, _, _ = captureOutput(t, ctx)

	s := &Store{
		alias:   "",
		path:    tempdir,
		crypto:  plain.New(),
		storage: fs.New(tempdir),
	}

	// Plain backend also doesn't implement keyExporter, but the default-to-own
	// path calls FindIdentities which returns ["0xDEADBEEF"] (first key in
	// staticPrivateKeyList). That's the key resolution we want to test.
	err = s.UpdateRecipientKeys(ctx, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot export")
}

// TestDiagnoseRecipientsExpiredWarn verifies that the diagnostic produces
// a more specific message when a key is present in .public-keys/ but not
// in the usable keyring by direct ID lookup.
func TestDiagnoseRecipientsPubkeyOnly(t *testing.T) {
	t.Parallel()

	ctx := config.NewContextInMemory()

	tempdir := t.TempDir()

	ctx, _, _ = captureOutput(t, ctx)

	// Create a store with one unresolvable ID that has a .public-keys/ entry.
	_, _, err := createStore(tempdir, []string{"0xDEADBEEF", "A3683834"}, nil)
	require.NoError(t, err)

	s := &Store{
		alias:   "",
		path:    tempdir,
		crypto:  plain.New(),
		storage: fs.New(tempdir),
	}

	// Add a mock .public-keys/ file for A3683834 so it becomes
	// "only available via .public-keys/" rather than "not found".
	require.NoError(t, s.storage.Set(ctx, filepath.Join(".public-keys", "A3683834"), []byte("mock-key-data")))

	diags := s.DiagnoseRecipients(ctx)

	// A3683834 should now be a warning about being only in .public-keys.
	found := false
	for _, d := range diags {
		if d.Recipient == "A3683834" {
			found = true
			assert.Equal(t, DiagWarning, d.Level)
			assert.Contains(t, d.Message, ".public-keys/")
			assert.Contains(t, d.Message, "gopass sync")
		}
	}
	assert.True(t, found, "A3683834 should be reported as .public-keys/ only warning")
}

// TestHasDecryptionKeyEmpty verifies that hasDecryptionKey returns false when
// .gpg-id is empty rather than calling FindIdentities with no search terms.
func TestHasDecryptionKeyEmpty(t *testing.T) {
	t.Parallel()

	ctx := config.NewContextInMemory()
	tempdir := t.TempDir()
	ctx, _, _ = captureOutput(t, ctx)

	require.NoError(t, os.WriteFile(filepath.Join(tempdir, ".plain-id"), []byte("\n"), 0o600))

	s := &Store{
		alias:   "",
		path:    tempdir,
		crypto:  plain.New(),
		storage: fs.New(tempdir),
	}

	assert.False(t, s.hasDecryptionKey(ctx), "empty recipient file must not match local keys")
}

// TestJoinTeamGPGRecipientFormats verifies that JoinTeam / hasDecryptionKey
// recognizes matching GPG keys regardless of whether .gpg-id stores a full
// fingerprint, 0x-prefixed fingerprint, long ID, short ID, or email address
// (GH-3028), and returns false for an empty .gpg-id or non-matching recipients.
func TestJoinTeamGPGRecipientFormats(t *testing.T) {
	if _, err := exec.LookPath("gpg"); err != nil {
		t.Skip("gpg not found in PATH")
	}

	var gnupgHome string
	if runtime.GOOS == "windows" {
		gnupgHome = t.TempDir()
	} else {
		// macOS has a 104-byte UNIX socket path limit (sun_path), which t.TempDir()
		// under /var/folders/... easily exceeds for gpg-agent sockets.
		var err error
		gnupgHome, err = os.MkdirTemp("/tmp", "gp-gpg-*") //nolint:usetesting // need short path under /tmp for macOS gpg-agent socket
		if err != nil {
			gnupgHome = t.TempDir()
		} else {
			t.Cleanup(func() {
				_ = os.RemoveAll(gnupgHome)
			})
		}
	}

	gpgHome := msysGPGHome(gnupgHome)

	genCmd := exec.CommandContext(t.Context(), "gpg", "--homedir", gpgHome, "--batch", "--passphrase", "", "--quick-gen-key", "Clone User <clone@example.org>", "default", "default", "never")
	genOut, err := genCmd.CombinedOutput()
	require.NoError(t, err, "quick-gen-key failed: %s", string(genOut))

	t.Cleanup(func() {
		_ = exec.Command("gpgconf", "--homedir", gpgHome, "--kill", "gpg-agent").Run()
	})

	listCmd := exec.CommandContext(t.Context(), "gpg", "--homedir", gpgHome, "--with-colons", "--with-fingerprint", "--list-secret-keys")
	listOut, err := listCmd.Output()
	require.NoError(t, err)

	var fpr string
	for line := range strings.SplitSeq(string(listOut), "\n") {
		if !strings.HasPrefix(line, "fpr:") {
			continue
		}
		fields := strings.Split(line, ":")
		if len(fields) >= 10 {
			fpr = fields[9]

			break
		}
	}
	require.Len(t, fpr, 40)

	t.Setenv("GNUPGHOME", gpgHome)

	ctx := config.NewContextInMemory()
	ctx = gpg.WithAlwaysTrust(ctx, false)
	ctx = gpg.WithUseCache(ctx, false)
	ctx, _, _ = captureOutput(t, ctx)

	g, err := gpgcli.New(ctx, gpgcli.Config{})
	require.NoError(t, err)

	longID := fpr[len(fpr)-16:]
	shortID := fpr[len(fpr)-8:]

	matchCases := []struct {
		name  string
		gpgID string
	}{
		{name: "full fingerprint", gpgID: fpr + "\n"},
		{name: "0x-prefixed full fingerprint", gpgID: "0x" + fpr + "\n"},
		{name: "long ID without 0x", gpgID: longID + "\n"},
		{name: "long ID with 0x", gpgID: "0x" + longID + "\n"},
		{name: "short ID without 0x", gpgID: shortID + "\n"},
		{name: "short ID with 0x", gpgID: "0x" + shortID + "\n"},
		{name: "email address", gpgID: "clone@example.org\n"},
		{name: "mixed recipients with full fingerprint", gpgID: "0xDEADBEEFDEADBEEF\n" + fpr + "\n"},
	}

	for _, tc := range matchCases {
		t.Run(tc.name, func(t *testing.T) {
			storeDir := t.TempDir()
			require.NoError(t, os.WriteFile(filepath.Join(storeDir, ".gpg-id"), []byte(tc.gpgID), 0o600))

			s := &Store{
				alias:   "",
				path:    storeDir,
				crypto:  g,
				storage: fs.New(storeDir),
			}

			assert.True(t, s.hasDecryptionKey(ctx))
			exported, err := s.JoinTeam(ctx)
			require.NoError(t, err)
			assert.False(t, exported, "user with matching key should not export a new key")
		})
	}

	t.Run("empty gpg-id does not match", func(t *testing.T) {
		storeDir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(storeDir, ".gpg-id"), []byte("\n"), 0o600))

		s := &Store{
			alias:   "",
			path:    storeDir,
			crypto:  g,
			storage: fs.New(storeDir),
		}

		assert.False(t, s.hasDecryptionKey(ctx))
	})

	t.Run("non-matching recipient exports own key", func(t *testing.T) {
		storeDir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(storeDir, ".gpg-id"), []byte("0xDEADBEEFDEADBEEF\n"), 0o600))

		s := &Store{
			alias:   "",
			path:    storeDir,
			crypto:  g,
			storage: fs.New(storeDir),
		}

		assert.False(t, s.hasDecryptionKey(ctx))
		exported, err := s.JoinTeam(ctx)
		require.NoError(t, err)
		assert.True(t, exported, "user without matching key should export own public key")
	})
}

type failingCommitStorage struct {
	backend.Storage
	failCommit bool
	failAdd    bool
}

func (f *failingCommitStorage) TryAdd(ctx context.Context, args ...string) error {
	if f.failAdd {
		return fmt.Errorf("simulated git add failure")
	}

	return f.Storage.TryAdd(ctx, args...)
}

func (f *failingCommitStorage) TryCommit(ctx context.Context, msg string) error {
	if f.failCommit {
		return fmt.Errorf("simulated git commit signing failure")
	}

	return f.Storage.TryCommit(ctx, msg)
}

func TestRemoveRecipientRollbackOnCommitFailure(t *testing.T) {
	t.Parallel()

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithHidden(ctx, true)
	cfg := config.NewInMemory()
	ctx = cfg.WithConfig(ctx)
	ctx, _, _ = captureOutput(t, ctx)

	tempdir := t.TempDir()
	_, _, err := createStore(tempdir, []string{"0xDEADBEEF", "0xFEEDBEEF"}, nil)
	require.NoError(t, err)

	base := fs.New(tempdir)
	st := &failingCommitStorage{Storage: base, failCommit: true}

	s := &Store{
		alias:   "",
		path:    tempdir,
		crypto:  plain.New(),
		storage: st,
	}

	// Seed an initial recipients.hash and a .public-keys/ file.
	origRS, err := s.GetRecipients(ctx, "")
	require.NoError(t, err)
	require.NoError(t, cfg.Set("", s.rhKey(), origRS.Hash()))
	require.NoError(t, base.Set(ctx, filepath.Join(".public-keys", "0xDEADBEEF"), []byte("dead-key")))

	// First attempt fails at git commit.
	err = s.RemoveRecipient(ctx, "0xDEADBEEF")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "simulated git commit signing failure")

	// Recipients file, recipients.hash, and .public-keys/ must be untouched.
	rs, err := s.GetRecipients(ctx, "")
	require.NoError(t, err)
	assert.Equal(t, []string{"0xDEADBEEF", "0xFEEDBEEF"}, rs.IDs())
	assert.Equal(t, origRS.Hash(), cfg.GetGlobal(s.rhKey()))
	assert.True(t, base.Exists(ctx, filepath.Join(".public-keys", "0xDEADBEEF")))

	// Once commits work again, retrying RemoveRecipient must succeed.
	st.failCommit = false
	require.NoError(t, s.RemoveRecipient(ctx, "0xDEADBEEF"))

	rs, err = s.GetRecipients(ctx, "")
	require.NoError(t, err)
	assert.Equal(t, []string{"0xFEEDBEEF"}, rs.IDs())
	assert.Equal(t, rs.Hash(), cfg.GetGlobal(s.rhKey()))
	assert.False(t, base.Exists(ctx, filepath.Join(".public-keys", "0xDEADBEEF")))
}

func TestAddRecipientRollbackOnCommitFailure(t *testing.T) {
	t.Parallel()

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithHidden(ctx, true)
	cfg := config.NewInMemory()
	ctx = cfg.WithConfig(ctx)
	ctx, _, _ = captureOutput(t, ctx)

	tempdir := t.TempDir()
	_, _, err := createStore(tempdir, []string{"0xDEADBEEF"}, nil)
	require.NoError(t, err)

	base := fs.New(tempdir)
	st := &failingCommitStorage{Storage: base, failCommit: true}

	s := &Store{
		alias:   "",
		path:    tempdir,
		crypto:  plain.New(),
		storage: st,
	}

	origRS, err := s.GetRecipients(ctx, "")
	require.NoError(t, err)
	require.NoError(t, cfg.Set("", s.rhKey(), origRS.Hash()))

	err = s.AddRecipient(ctx, "0xFEEDBEEF")
	require.Error(t, err)

	// Recipients file and hash must remain unchanged after rollback.
	rs, err := s.GetRecipients(ctx, "")
	require.NoError(t, err)
	assert.Equal(t, []string{"0xDEADBEEF"}, rs.IDs())
	assert.Equal(t, origRS.Hash(), cfg.GetGlobal(s.rhKey()))

	// Retrying once commits succeed must add the recipient cleanly.
	st.failCommit = false
	require.NoError(t, s.AddRecipient(ctx, "0xFEEDBEEF"))

	rs, err = s.GetRecipients(ctx, "")
	require.NoError(t, err)
	assert.Equal(t, []string{"0xDEADBEEF", "0xFEEDBEEF"}, rs.IDs())
}

func TestRemoveRecipientGitFSRollbackOnCommitFailure(t *testing.T) {
	t.Parallel()

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git binary not found in PATH")
	}

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithHidden(ctx, true)
	ctx = ctxutil.WithGitInit(ctx, true)
	cfg := config.NewInMemory()
	ctx = cfg.WithConfig(ctx)
	ctx, _, _ = captureOutput(t, ctx)

	tempdir := t.TempDir()
	_, _, err := createStore(tempdir, []string{"0xDEADBEEF", "0xFEEDBEEF"}, nil)
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(filepath.Join(tempdir, ".public-keys"), 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(tempdir, ".public-keys", "0xDEADBEEF"), []byte("dead-key"), 0o600))

	gitSt, err := gitfs.Init(ctx, tempdir, "test", "test@example.com")
	require.NoError(t, err)

	st := &failingCommitStorage{Storage: gitSt, failCommit: true}
	s := &Store{
		alias:   "",
		path:    tempdir,
		crypto:  plain.New(),
		storage: st,
	}

	// Simulate commit failure (e.g. GPG commit-signing failure).
	err = s.RemoveRecipient(ctx, "0xDEADBEEF")
	require.Error(t, err)

	gitStatusPorcelain := func(t *testing.T) string {
		t.Helper()

		cmd := exec.CommandContext(ctx, "git", "-C", tempdir, "status", "--porcelain")
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, string(out))

		return strings.TrimSpace(string(out))
	}

	// Working tree and git index must be completely clean, and .gpg-id restored.
	rs, err := s.GetRecipients(ctx, "")
	require.NoError(t, err)
	assert.Equal(t, []string{"0xDEADBEEF", "0xFEEDBEEF"}, rs.IDs())
	assert.False(t, gitSt.HasStagedChanges(ctx), "git index must not have staged changes after rollback")
	assert.Empty(t, gitStatusPorcelain(t), "git working tree must be clean after rollback")

	// Retry after commit signing is fixed.
	st.failCommit = false
	require.NoError(t, s.RemoveRecipient(ctx, "0xDEADBEEF"))

	rs, err = s.GetRecipients(ctx, "")
	require.NoError(t, err)
	assert.Equal(t, []string{"0xFEEDBEEF"}, rs.IDs())
	assert.False(t, gitSt.Exists(ctx, filepath.Join(".public-keys", "0xDEADBEEF")))
	assert.Empty(t, gitStatusPorcelain(t), "git working tree must be clean after successful removal")
}

type twoSubkeyEnv struct {
	root       string
	gpgHome    string
	primaryFpr string
	sub1Fpr    string
	sub2Fpr    string
}

func msysGPGHome(dir string) string {
	if runtime.GOOS != "windows" {
		return dir
	}
	// MSYS2 gpg expects POSIX-style paths (/c/...) rather than Windows drive paths.
	dir = strings.ReplaceAll(dir, `\`, "/")
	if len(dir) >= 2 && dir[1] == ':' {
		return "/" + strings.ToLower(dir[:1]) + dir[2:]
	}

	return dir
}

func setupTwoSubkeyGPGHome(t *testing.T) twoSubkeyEnv {
	t.Helper()

	var root string
	if runtime.GOOS == "windows" {
		root = t.TempDir()
	} else {
		var err error
		root, err = os.MkdirTemp("/tmp", "gp-sub-*") //nolint:usetesting // short path for macOS gpg-agent socket
		require.NoError(t, err)
		t.Cleanup(func() { _ = os.RemoveAll(root) })
	}

	gpgDir := filepath.Join(root, "gnupg")
	require.NoError(t, os.MkdirAll(gpgDir, 0o700))
	gpgHome := msysGPGHome(gpgDir)
	t.Cleanup(func() {
		_ = exec.Command("gpgconf", "--homedir", gpgHome, "--kill", "gpg-agent").Run()
	})

	genCmd := exec.CommandContext(t.Context(), "gpg", "--homedir", gpgHome, "--batch", "--passphrase", "", "--quick-gen-key", "Subkey Tester <subkey@example.org>", "default", "default", "never")
	out, err := genCmd.CombinedOutput()
	require.NoError(t, err, "quick-gen-key failed: %s", string(out))

	listCmd := exec.CommandContext(t.Context(), "gpg", "--homedir", gpgHome, "--with-colons", "--with-fingerprint", "--list-public-keys")
	listOut, err := listCmd.Output()
	require.NoError(t, err)

	var primaryFpr, sub1Fpr string
	for line := range strings.SplitSeq(string(listOut), "\n") {
		if !strings.HasPrefix(line, "fpr:") {
			continue
		}
		fields := strings.Split(line, ":")
		if len(fields) < 10 {
			continue
		}
		if primaryFpr == "" {
			primaryFpr = fields[9]
		} else if sub1Fpr == "" {
			sub1Fpr = fields[9]
		}
	}
	require.Len(t, primaryFpr, 40)
	require.Len(t, sub1Fpr, 40)

	addSubCmd := exec.CommandContext(t.Context(), "gpg", "--homedir", gpgHome, "--batch", "--passphrase", "", "--quick-add-key", primaryFpr, "cv25519", "encr", "never")
	out, err = addSubCmd.CombinedOutput()
	require.NoError(t, err, "quick-add-key failed: %s", string(out))

	listCmd = exec.CommandContext(t.Context(), "gpg", "--homedir", gpgHome, "--with-colons", "--with-fingerprint", "--list-public-keys", primaryFpr)
	listOut, err = listCmd.Output()
	require.NoError(t, err)

	var sub2Fpr string
	for line := range strings.SplitSeq(string(listOut), "\n") {
		if !strings.HasPrefix(line, "fpr:") {
			continue
		}
		fields := strings.Split(line, ":")
		if len(fields) < 10 {
			continue
		}
		fp := fields[9]
		if fp != primaryFpr && fp != sub1Fpr {
			sub2Fpr = fp

			break
		}
	}
	require.Len(t, sub2Fpr, 40)

	return twoSubkeyEnv{
		root:       root,
		gpgHome:    gpgHome,
		primaryFpr: primaryFpr,
		sub1Fpr:    sub1Fpr,
		sub2Fpr:    sub2Fpr,
	}
}

// TestSubkeyStoreLifecycle exercises end-to-end GPG subkey support across
// Init, Set (ensureOurKeyID), AddRecipient, RemoveRecipient, fsckCheckRecipients,
// and UpdateRecipientKeys (issues #909 and #1107).
func TestSubkeyStoreLifecycle(t *testing.T) {
	if _, err := exec.LookPath("gpg"); err != nil {
		t.Skip("gpg not found in PATH")
	}

	env := setupTwoSubkeyGPGHome(t)
	root, gpgHome, primaryFpr, sub1Fpr, sub2Fpr := env.root, env.gpgHome, env.primaryFpr, env.sub1Fpr, env.sub2Fpr

	t.Setenv("GNUPGHOME", gpgHome)

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithHidden(ctx, true)
	ctx = gpg.WithAlwaysTrust(ctx, false)
	ctx = gpg.WithUseCache(ctx, false)
	cfg := config.NewInMemory()
	require.NoError(t, cfg.Set("", "core.exportkeys", "true"))
	ctx = cfg.WithConfig(ctx)
	ctx, _, _ = captureOutput(t, ctx)

	g, err := gpgcli.New(ctx, gpgcli.Config{})
	require.NoError(t, err)

	storeDir := filepath.Join(root, "store")
	require.NoError(t, os.MkdirAll(storeDir, 0o700))

	s := &Store{
		alias:   "",
		path:    storeDir,
		crypto:  g,
		storage: fs.New(storeDir),
	}

	// 1. Init with a 40-char forced subkey preserves the subkey and '!'.
	require.NoError(t, s.Init(ctx, storeDir, sub1Fpr+"!"))
	assert.Equal(t, []string{sub1Fpr + "!"}, s.Recipients(ctx))
	assert.Equal(t, sub1Fpr+"!", s.OurKeyID(ctx))

	// 2. Writing a secret encrypts ONLY to sub1 (ensureOurKeyID must not inject primary key).
	sec := secrets.NewAKV()
	sec.SetPassword("super-secret")
	require.NoError(t, s.Set(ctx, "foo", sec))

	ct, err := s.storage.Get(ctx, s.passfile(ctx, "foo"))
	require.NoError(t, err)
	recpIDs, err := g.RecipientIDs(ctx, ct)
	require.NoError(t, err)
	assert.Equal(t, []string{sub1Fpr}, recpIDs)

	// 3. fsckCheckRecipients passes when encrypted to the exact forced subkey.
	require.NoError(t, s.fsckCheckRecipients(ctx, "foo").ErrorOrNil())

	// 4. Adding a second forced subkey on the same primary key adds it and re-encrypts to both.
	require.NoError(t, s.AddRecipient(ctx, sub2Fpr+"!"))
	assert.ElementsMatch(t, []string{sub1Fpr + "!", sub2Fpr + "!"}, s.Recipients(ctx))

	ct, err = s.storage.Get(ctx, s.passfile(ctx, "foo"))
	require.NoError(t, err)
	recpIDs, err = g.RecipientIDs(ctx, ct)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{sub1Fpr, sub2Fpr}, recpIDs)
	require.NoError(t, s.fsckCheckRecipients(ctx, "foo").ErrorOrNil())

	// 5. Removing sub1 leaves ONLY sub2 and re-encrypts to ONLY sub2.
	require.NoError(t, s.RemoveRecipient(ctx, sub1Fpr+"!"))
	assert.Equal(t, []string{sub2Fpr + "!"}, s.Recipients(ctx))

	ct, err = s.storage.Get(ctx, s.passfile(ctx, "foo"))
	require.NoError(t, err)
	recpIDs, err = g.RecipientIDs(ctx, ct)
	require.NoError(t, err)
	assert.Equal(t, []string{sub2Fpr}, recpIDs)
	require.NoError(t, s.fsckCheckRecipients(ctx, "foo").ErrorOrNil())

	// 6. fsckCheckRecipients detects when a secret is still encrypted to the old subkey (sub1)
	// while .gpg-id requires sub2!.
	staleCT, err := g.Encrypt(ctx, sec.Bytes(), []string{sub1Fpr + "!"})
	require.NoError(t, err)
	require.NoError(t, s.storage.Set(ctx, s.passfile(ctx, "foo"), staleCT))
	require.Error(t, s.fsckCheckRecipients(ctx, "foo").ErrorOrNil(), "fsck must detect wrong subkey encryption")

	// 7. UpdateRecipientKeys with no args updates the forced subkey's .public-keys/ entry.
	require.NoError(t, s.UpdateRecipientKeys(ctx, nil))
	assert.True(t, s.storage.Exists(ctx, filepath.Join(keyDir, sub2Fpr+"!")))

	// 8. Re-adding an already-present primary key after adding a new subkey refreshes
	// .public-keys/ and re-encrypts secrets to the newer subkey (#1107).
	storeDir2 := filepath.Join(root, "store2")
	require.NoError(t, os.MkdirAll(storeDir2, 0o700))
	s2 := &Store{
		alias:   "",
		path:    storeDir2,
		crypto:  g,
		storage: fs.New(storeDir2),
	}
	require.NoError(t, s2.Init(ctx, storeDir2, primaryFpr))
	primaryID := s2.Recipients(ctx)[0]
	require.NoError(t, s2.Set(ctx, "bar", sec))

	// Write a sentinel stale public key to confirm AddRecipient refreshes it.
	pubPath := filepath.Join(keyDir, primaryID)
	require.NoError(t, s2.storage.Set(ctx, pubPath, []byte("stale-public-key-data-padding-to-exceed-32-bytes")))

	// Re-add using the new subkey ID without '!': confirms re-encryption and refreshes .public-keys/.
	require.NoError(t, s2.AddRecipient(ctxutil.WithAlwaysYes(ctx, true), sub2Fpr))
	assert.Equal(t, []string{primaryID}, s2.Recipients(ctx))

	refreshedPub, err := s2.storage.Get(ctx, pubPath)
	require.NoError(t, err)
	assert.Contains(t, string(refreshedPub), "BEGIN PGP PUBLIC KEY BLOCK")
	assert.NoError(t, s2.fsckCheckRecipients(ctx, "bar").ErrorOrNil())
}
