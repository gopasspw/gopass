package xc

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/gopasspw/gopass/internal/backend/crypto/xc/keyring"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateIdentity(t *testing.T) {
	ctx := context.Background()

	td, err := ioutil.TempDir("", "gopass-")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(td)
	}()
	assert.NoError(t, os.Setenv("GOPASS_CONFIG", filepath.Join(td, ".gopass.yml")))
	assert.NoError(t, os.Setenv("GOPASS_HOMEDIR", td))

	passphrase := "test"
	skr, err := keyring.LoadSecring(filepath.Join(td, "skr"))
	require.NoError(t, err)
	require.NotNil(t, skr)
	pkr, err := keyring.LoadPubring(filepath.Join(td, "pkr"), skr)
	require.NoError(t, err)
	require.NotNil(t, pkr)

	xc := &XC{
		pubring: pkr,
		secring: skr,
		client:  &fakeAgent{passphrase},
	}

	assert.NoError(t, xc.GenerateIdentity(ctx, "foo", "bar@example.org", passphrase))

	pubKeys, err := xc.ListRecipients(ctx)
	require.NoError(t, err)

	privKeys, err := xc.ListIdentities(ctx)
	require.NoError(t, err)

	assert.Equal(t, 1, len(pubKeys))
	assert.Equal(t, len(pubKeys), len(privKeys))

	id := pubKeys[0]
	assert.Contains(t, xc.FormatKey(ctx, id, ""), "foo <bar@example.org>")

	pubKeys, err = xc.FindRecipients(ctx, id)
	assert.NoError(t, err)
	assert.Equal(t, []string{id}, pubKeys)

	privKeys, err = xc.FindIdentities(ctx, id)
	assert.NoError(t, err)
	assert.Equal(t, []string{id}, privKeys)

	assert.NoError(t, xc.RemoveKey(id))
}
