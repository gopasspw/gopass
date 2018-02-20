package xc

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/justwatchcom/gopass/backend/crypto/xc/keyring"
	"github.com/stretchr/testify/assert"
)

func TestExportKey(t *testing.T) {
	ctx := context.Background()

	td, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(td)
	}()
	assert.NoError(t, os.Setenv("GOPASS_CONFIG", filepath.Join(td, ".gopass.yml")))
	assert.NoError(t, os.Setenv("GOPASS_HOMEDIR", td))

	passphrase := "test"

	k1, err := keyring.GenerateKeypair(passphrase)
	assert.NoError(t, err)
	k2, err := keyring.GenerateKeypair(passphrase)
	assert.NoError(t, err)
	k3, err := keyring.GenerateKeypair(passphrase)
	assert.NoError(t, err)
	k3.Identity.Name = "foobar"

	skr := keyring.NewSecring()
	assert.NoError(t, skr.Set(k1))

	pkr := keyring.NewPubring(skr)
	assert.NoError(t, pkr.Set(&k2.PublicKey))
	assert.NoError(t, pkr.Set(&k3.PublicKey))

	xc := &XC{
		pubring: pkr,
		secring: skr,
		client:  &fakeAgent{passphrase},
	}

	_, err = xc.ExportPublicKey(ctx, k1.Fingerprint())
	assert.NoError(t, err)

	_, err = xc.ExportPublicKey(ctx, k2.Fingerprint())
	assert.NoError(t, err)

	buf, err := xc.ExportPublicKey(ctx, k3.Fingerprint())
	assert.NoError(t, err)

	names, err := xc.ReadNamesFromKey(ctx, buf)
	assert.NoError(t, err)
	assert.Equal(t, []string{"foobar"}, names)

	_, err = xc.ExportPublicKey(ctx, "foobar")
	assert.Error(t, err)

	_, err = xc.ExportPrivateKey(ctx, k1.Fingerprint())
	assert.NoError(t, err)

	_, err = xc.ExportPrivateKey(ctx, k2.Fingerprint())
	assert.Error(t, err)
}
