package xc

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/gopasspw/gopass/pkg/backend/crypto/xc/keyring"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreatePrivateKeyBatch(t *testing.T) {
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

	assert.NoError(t, xc.CreatePrivateKeyBatch(ctx, "foo", "bar@example.org", passphrase))

	pubKeys, err := xc.ListPublicKeyIDs(ctx)
	require.NoError(t, err)

	privKeys, err := xc.ListPrivateKeyIDs(ctx)
	require.NoError(t, err)

	assert.Equal(t, 1, len(pubKeys))
	assert.Equal(t, len(pubKeys), len(privKeys))

	id := pubKeys[0]
	assert.Contains(t, xc.FormatKey(ctx, id), "foo <bar@example.org>")
	assert.Equal(t, "foo", xc.NameFromKey(ctx, id))
	assert.Equal(t, "bar@example.org", xc.EmailFromKey(ctx, id))

	pubKeys, err = xc.FindPublicKeys(ctx, id)
	assert.NoError(t, err)
	assert.Equal(t, []string{id}, pubKeys)

	privKeys, err = xc.FindPrivateKeys(ctx, id)
	assert.NoError(t, err)
	assert.Equal(t, []string{id}, privKeys)

	assert.NoError(t, xc.RemoveKey(id))
}

func TestCreatePrivateKey(t *testing.T) {
	ctx := context.Background()

	var x *XC

	assert.Error(t, x.CreatePrivateKey(ctx))
}
