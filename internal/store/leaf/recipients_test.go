package leaf

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/backend/crypto/plain"
	"github.com/gopasspw/gopass/internal/backend/storage/fs"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/recipients"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetRecipientsDefault(t *testing.T) {
	t.Parallel()

	ctx := config.NewContextInMemory()

	tempdir := t.TempDir()

	obuf := &bytes.Buffer{}
	out.Stdout = obuf

	defer func() {
		out.Stdout = os.Stdout
	}()

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

	obuf := &bytes.Buffer{}
	out.Stdout = obuf

	defer func() {
		out.Stdout = os.Stdout
	}()

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

	obuf := &bytes.Buffer{}
	out.Stdout = obuf

	defer func() {
		out.Stdout = os.Stdout
	}()

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

	obuf := &bytes.Buffer{}
	out.Stdout = obuf

	defer func() {
		out.Stdout = os.Stdout
	}()

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

	obuf := &bytes.Buffer{}
	out.Stdout = obuf

	defer func() {
		out.Stdout = os.Stdout
	}()

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

	obuf := &bytes.Buffer{}
	out.Stdout = obuf

	defer func() {
		out.Stdout = os.Stdout
	}()

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

	obuf := &bytes.Buffer{}
	out.Stdout = obuf

	defer func() {
		out.Stdout = os.Stdout
	}()

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

	obuf := &bytes.Buffer{}
	out.Stdout = obuf

	defer func() {
		out.Stdout = os.Stdout
	}()

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

	obuf := &bytes.Buffer{}
	out.Stdout = obuf

	defer func() {
		out.Stdout = os.Stdout
	}()

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

	obuf := &bytes.Buffer{}
	out.Stdout = obuf

	defer func() {
		out.Stdout = os.Stdout
	}()

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

	obuf := &bytes.Buffer{}
	out.Stdout = obuf

	defer func() {
		out.Stdout = os.Stdout
	}()

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

	obuf := &bytes.Buffer{}
	out.Stdout = obuf

	defer func() {
		out.Stdout = os.Stdout
	}()

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

	obuf := &bytes.Buffer{}
	out.Stdout = obuf

	defer func() {
		out.Stdout = os.Stdout
	}()

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

	obuf := &bytes.Buffer{}
	out.Stdout = obuf

	defer func() {
		out.Stdout = os.Stdout
	}()

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

	obuf := &bytes.Buffer{}
	out.Stdout = obuf

	defer func() {
		out.Stdout = os.Stdout
	}()

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

	obuf := &bytes.Buffer{}
	out.Stdout = obuf

	defer func() {
		out.Stdout = os.Stdout
	}()

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

	obuf := &bytes.Buffer{}
	out.Stdout = obuf

	defer func() {
		out.Stdout = os.Stdout
	}()

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

	obuf := &bytes.Buffer{}
	out.Stdout = obuf

	defer func() {
		out.Stdout = os.Stdout
	}()

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

	obuf := &bytes.Buffer{}
	out.Stdout = obuf

	defer func() {
		out.Stdout = os.Stdout
	}()

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

	obuf := &bytes.Buffer{}
	out.Stdout = obuf

	defer func() {
		out.Stdout = os.Stdout
	}()

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

	obuf := &bytes.Buffer{}
	out.Stdout = obuf

	defer func() {
		out.Stdout = os.Stdout
	}()

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

	obuf := &bytes.Buffer{}
	out.Stdout = obuf

	defer func() {
		out.Stdout = os.Stdout
	}()

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

	obuf := &bytes.Buffer{}
	out.Stdout = obuf

	defer func() {
		out.Stdout = os.Stdout
	}()

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

	obuf := &bytes.Buffer{}
	out.Stdout = obuf

	defer func() {
		out.Stdout = os.Stdout
	}()

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
