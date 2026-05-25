package leaf

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/gopasspw/gopass/internal/backend"
	_ "github.com/gopasspw/gopass/internal/backend/crypto"
	"github.com/gopasspw/gopass/internal/backend/crypto/gpg"
	_ "github.com/gopasspw/gopass/internal/backend/storage"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/gopass/secrets"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSet(t *testing.T) {
	ctx := gpg.WithAlwaysTrust(config.NewContextInMemory(), true)

	s, err := createSubStore(t)
	require.NoError(t, err)

	sec := secrets.NewAKV()
	sec.SetPassword("foo")
	_, err = sec.Write([]byte("bar"))
	require.NoError(t, err)
	require.NoError(t, s.Set(ctx, "zab/zab", sec))

	require.Error(t, s.Set(ctx, "../../../../../etc/passwd", sec))

	require.NoError(t, s.Set(ctx, "zab", sec))
}

// TestSetWarnsAboutInvalidRecipient verifies that when a recipient has no
// useable key (e.g. because it is expired or not present in the keyring),
// gopass emits a visible warning instead of silently omitting that recipient.
func TestSetWarnsAboutInvalidRecipient(t *testing.T) {
	buf := &bytes.Buffer{}
	oldStderr := out.Stderr
	out.Stderr = buf
	t.Cleanup(func() { out.Stderr = oldStderr })

	dir := t.TempDir()
	sd := filepath.Join(dir, "sub")

	// 0xDEADBEEF is in the plain backend's static key list; 0xBADKEY is not.
	_, _, err := createStore(sd, []string{"0xDEADBEEF", "0xBADKEY"}, nil)
	require.NoError(t, err)

	t.Setenv("GOPASS_HOMEDIR", dir)
	t.Setenv("CHECKPOINT_DISABLE", "true")
	t.Setenv("GOPASS_NO_NOTIFY", "true")
	t.Setenv("GOPASS_DISABLE_ENCRYPTION", "true")
	t.Setenv("GNUPGHOME", filepath.Join(dir, ".gnupg"))
	require.NoError(t, os.Unsetenv("PAGER"))

	ctx := gpg.WithAlwaysTrust(config.NewContextInMemory(), true)
	ctx, err = backend.WithCryptoBackendString(ctx, "plain")
	require.NoError(t, err)
	ctx, err = backend.WithStorageBackendString(ctx, "fs")
	require.NoError(t, err)

	s, err := New(ctx, "", sd)
	require.NoError(t, err)

	sec := secrets.NewAKV()
	sec.SetPassword("hunter2")
	require.NoError(t, s.Set(ctx, "test/entry", sec))

	// A warning about the recipient with no useable key must have been printed.
	assert.Contains(t, buf.String(), "0xBADKEY")
}
