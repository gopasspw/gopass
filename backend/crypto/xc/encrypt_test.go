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

type fakeAgent struct {
	pw string
}

func (f *fakeAgent) Ping() error {
	return nil
}

func (f *fakeAgent) Remove(string) error {
	return nil
}

func (f *fakeAgent) Passphrase(string, string) (string, error) {
	return f.pw, nil
}

func TestEncryptSimple(t *testing.T) {
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

	skr := keyring.NewSecring()
	assert.NoError(t, skr.Set(k1))

	pkr := keyring.NewPubring(skr)

	xc := &XC{
		pubring: pkr,
		secring: skr,
		client:  &fakeAgent{passphrase},
	}

	buf, err := xc.Encrypt(ctx, []byte("foobar"), []string{k1.Fingerprint()})
	assert.NoError(t, err)

	recps, err := xc.RecipientIDs(ctx, buf)
	assert.NoError(t, err)
	assert.Equal(t, []string{k1.Fingerprint()}, recps)

	buf, err = xc.Decrypt(ctx, buf)
	assert.NoError(t, err)
	assert.Equal(t, "foobar", string(buf))
}

func TestEncryptMultiKeys(t *testing.T) {
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

	buf, err := xc.Encrypt(ctx, []byte("foobar"), []string{k1.Fingerprint()})
	assert.NoError(t, err)

	recps, err := xc.RecipientIDs(ctx, buf)
	assert.NoError(t, err)
	assert.Equal(t, []string{k1.Fingerprint()}, recps)

	buf, err = xc.Decrypt(ctx, buf)
	assert.NoError(t, err)
	assert.Equal(t, "foobar", string(buf))
}
