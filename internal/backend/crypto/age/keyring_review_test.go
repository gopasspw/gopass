package age

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"filippo.io/age/plugin"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/termio"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeyringRecipientReview(t *testing.T) {
	a := newTestAge(t)
	protector := keyringTestIdentity(t)
	recovery := keyringTestIdentity(t)
	software := keyringTestIdentity(t)
	bootstrap := filepath.Join(t.TempDir(), "unlock.txt")
	writeBootstrapIdentity(t, bootstrap, protector)
	ctx := protectedKeyringContext(t, bootstrap, protector.Recipient().String())
	oldOutput := termio.Stderr
	var output bytes.Buffer
	termio.Stderr = &output
	t.Cleanup(func() { termio.Stderr = oldOutput })

	// Automatic yes must still read the explicit refusal.
	termio.Stdin = strings.NewReader("n\n")
	err := a.saveIdentities(ctxutil.WithAlwaysYes(ctx, true), []string{software.String()}, true)
	require.ErrorIs(t, err, termio.ErrAborted)
	_, err = os.Stat(a.identity)
	require.True(t, os.IsNotExist(err))
	assert.Contains(t, output.String(), protector.Recipient().String())
	assert.Contains(t, output.String(), "Any corresponding private identity")

	require.ErrorContains(t, a.saveIdentities(ctxutil.WithInteractive(ctx, false), []string{software.String()}, true), "manual review")
	require.ErrorContains(t, a.saveIdentities(ctxutil.WithTerminal(ctx, false), []string{software.String()}, true), "manual review")
	termio.Stdin = strings.NewReader("y\n")
	require.NoError(t, a.saveIdentities(ctx, []string{software.String()}, true))
	assert.Equal(t, software.String(), readProtectedKeyring(t, a.identity, protector))
	review := readKeyringReview(a.identity)
	require.NotNil(t, review)
	assert.Equal(t, []string{protector.Recipient().String()}, review.Recipients)
	if runtime.GOOS != "windows" {
		info, err := os.Stat(a.identity + ".review.json")
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
	}

	// Same recipients permit noninteractive writes and refresh the ciphertext hash.
	output.Reset()
	require.NoError(t, a.saveIdentities(ctxutil.WithInteractive(ctx, false), []string{software.String()}, false))
	require.NotNil(t, readKeyringReview(a.identity))
	assert.Empty(t, output.String())

	changed := protectedKeyringContext(t, bootstrap, protector.Recipient().String()+","+recovery.Recipient().String())
	before, err := os.ReadFile(a.identity)
	require.NoError(t, err)
	termio.Stdin = strings.NewReader("n\n")
	require.ErrorIs(t, a.reencryptIdentities(changed), termio.ErrAborted)
	after, err := os.ReadFile(a.identity)
	require.NoError(t, err)
	assert.Equal(t, before, after)
	assert.Contains(t, output.String(), "Added: "+recovery.Recipient().String())
	termio.Stdin = strings.NewReader("y\n")
	require.NoError(t, a.reencryptIdentities(changed))
	assert.Equal(t, software.String(), readProtectedKeyring(t, a.identity, recovery))

	// Order, whitespace and duplicates do not change the approved set.
	reordered := protectedKeyringContext(t, bootstrap, " "+recovery.Recipient().String()+", "+protector.Recipient().String()+","+protector.Recipient().String())
	output.Reset()
	require.NoError(t, a.reencryptIdentities(ctxutil.WithInteractive(reordered, false)))
	assert.Empty(t, output.String())

	output.Reset()
	termio.Stdin = strings.NewReader("y\n")
	require.NoError(t, a.reencryptIdentities(ctx))
	assert.Contains(t, output.String(), "Removed: "+recovery.Recipient().String())
}

func TestKeyringReviewInvalidRecords(t *testing.T) {
	for _, mode := range []string{"missing", "malformed", "version", "hash", "replaced", "unsorted"} {
		t.Run(mode, func(t *testing.T) {
			a := newTestAge(t)
			protector := keyringTestIdentity(t)
			software := keyringTestIdentity(t)
			bootstrap := filepath.Join(t.TempDir(), "unlock.txt")
			writeBootstrapIdentity(t, bootstrap, protector)
			ctx := protectedKeyringContext(t, bootstrap, protector.Recipient().String())
			require.NoError(t, a.saveIdentities(ctx, []string{software.String()}, true))
			filename := a.identity + ".review.json"
			record := readKeyringReview(a.identity)
			require.NotNil(t, record)
			switch mode {
			case "missing":
				require.NoError(t, os.Remove(filename))
			case "malformed":
				require.NoError(t, os.WriteFile(filename, []byte("invalid"), 0o600))
			case "replaced":
				ciphertext, err := a.encrypt([]byte(software.String()), protector.Recipient())
				require.NoError(t, err)
				require.NoError(t, writeEncryptedKeyring(a.identity, ciphertext))
			default:
				switch mode {
				case "version":
					record.Version++
				case "hash":
					record.CiphertextSHA256 = "stale"
				case "unsorted":
					record.Recipients = []string{"z", "a"}
				}
				data, err := json.Marshal(record)
				require.NoError(t, err)
				require.NoError(t, os.WriteFile(filename, data, 0o600))
			}
			require.Nil(t, readKeyringReview(a.identity))
			before, err := os.ReadFile(a.identity)
			require.NoError(t, err)
			require.ErrorContains(t, a.saveIdentities(ctxutil.WithInteractive(ctx, false), []string{software.String()}, false), "manual review")
			termio.Stdin = strings.NewReader("n\n")
			require.ErrorIs(t, a.saveIdentities(ctx, []string{software.String()}, false), termio.ErrAborted)
			after, err := os.ReadFile(a.identity)
			require.NoError(t, err)
			assert.Equal(t, before, after)
			termio.Stdin = strings.NewReader("y\n")
			require.NoError(t, a.saveIdentities(ctx, []string{software.String()}, false))
			require.NotNil(t, readKeyringReview(a.identity))
		})
	}
}

func TestKeyringReviewWriteFailure(t *testing.T) {
	a := newTestAge(t)
	protector := keyringTestIdentity(t)
	software := keyringTestIdentity(t)
	bootstrap := filepath.Join(t.TempDir(), "unlock.txt")
	writeBootstrapIdentity(t, bootstrap, protector)
	ctx := protectedKeyringContext(t, bootstrap, protector.Recipient().String())
	require.NoError(t, os.MkdirAll(a.identity+".review.json", 0o700))
	require.ErrorContains(t, a.saveIdentities(ctx, []string{software.String()}, true), "keyring updated, but failed to save recipient review")
	assert.Equal(t, software.String(), readProtectedKeyring(t, a.identity, protector))
	require.Nil(t, readKeyringReview(a.identity))
	require.ErrorContains(t, a.saveIdentities(ctxutil.WithInteractive(ctx, false), []string{software.String()}, false), "manual review")
}

func TestPassphraseKeyringDoesNotReviewRecipients(t *testing.T) {
	a := newTestAge(t)
	ctx := protectedKeyringContext(t, "", "")
	software := keyringTestIdentity(t)
	oldOutput := termio.Stderr
	var output bytes.Buffer
	termio.Stderr = &output
	t.Cleanup(func() { termio.Stderr = oldOutput })
	termio.Stdin = strings.NewReader("n\n")
	require.NoError(t, a.saveIdentities(ctxutil.WithInteractive(ctx, false), []string{software.String()}, true))
	assert.Empty(t, output.String())
	_, err := os.Stat(a.identity + ".review.json")
	assert.True(t, os.IsNotExist(err))
}

func TestKeyringReviewVerificationFailurePreservesFiles(t *testing.T) {
	a := newTestAge(t)
	protector := keyringTestIdentity(t)
	unrelated := keyringTestIdentity(t)
	software := keyringTestIdentity(t)
	bootstrap := filepath.Join(t.TempDir(), "unlock.txt")
	writeBootstrapIdentity(t, bootstrap, protector)
	ctx := protectedKeyringContext(t, bootstrap, protector.Recipient().String())
	require.NoError(t, a.saveIdentities(ctx, []string{software.String()}, true))
	ciphertext, err := os.ReadFile(a.identity)
	require.NoError(t, err)
	review, err := os.ReadFile(a.identity + ".review.json")
	require.NoError(t, err)
	changed := protectedKeyringContext(t, bootstrap, unrelated.Recipient().String())
	require.ErrorContains(t, a.saveIdentities(changed, []string{software.String()}, false), "cannot unlock the new identity keyring")
	after, err := os.ReadFile(a.identity)
	require.NoError(t, err)
	assert.Equal(t, ciphertext, after)
	after, err = os.ReadFile(a.identity + ".review.json")
	require.NoError(t, err)
	assert.Equal(t, review, after)

	// A syntactically valid recipient whose plugin executable is absent fails
	// during encryption, after review but before replacing either file.
	missingPlugin := plugin.EncodeRecipient("gopassmissingreview", []byte("test"))
	missingCtx := protectedKeyringContext(t, bootstrap, missingPlugin)
	require.Error(t, a.saveIdentities(missingCtx, []string{software.String()}, false))
	after, err = os.ReadFile(a.identity)
	require.NoError(t, err)
	assert.Equal(t, ciphertext, after)
	after, err = os.ReadFile(a.identity + ".review.json")
	require.NoError(t, err)
	assert.Equal(t, review, after)

	invalid := protectedKeyringContext(t, bootstrap, "age1invalid")
	require.Error(t, a.saveIdentities(invalid, []string{software.String()}, false))
	after, err = os.ReadFile(a.identity)
	require.NoError(t, err)
	assert.Equal(t, ciphertext, after)
}
