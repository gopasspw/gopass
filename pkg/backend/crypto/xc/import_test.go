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

func TestImportKey(t *testing.T) {
	ctx := context.Background()

	td, err := ioutil.TempDir("", "gopass-")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(td)
	}()
	assert.NoError(t, os.Setenv("GOPASS_CONFIG", filepath.Join(td, ".gopass.yml")))
	assert.NoError(t, os.Setenv("GOPASS_HOMEDIR", td))

	passphrase := "test"

	// XC #1
	x1k1, err := keyring.GenerateKeypair(passphrase)
	assert.NoError(t, err)
	x1k2, err := keyring.GenerateKeypair(passphrase)
	assert.NoError(t, err)
	x1k3, err := keyring.GenerateKeypair(passphrase)
	assert.NoError(t, err)

	x1skrfn := filepath.Join(td, "x1skr")
	x1skr, err := keyring.LoadSecring(x1skrfn)
	require.NoError(t, err)
	assert.NoError(t, x1skr.Set(x1k1))

	x1pkrfn := filepath.Join(td, "x1pkr")
	x1pkr, err := keyring.LoadPubring(x1pkrfn, x1skr)
	require.NoError(t, err)
	assert.NoError(t, x1pkr.Set(&x1k2.PublicKey))
	assert.NoError(t, x1pkr.Set(&x1k3.PublicKey))

	xc1 := &XC{
		pubring: x1pkr,
		secring: x1skr,
		client:  &fakeAgent{passphrase},
	}

	// XC #2
	x2k1, err := keyring.GenerateKeypair(passphrase)
	require.NoError(t, err)
	x2k2, err := keyring.GenerateKeypair(passphrase)
	require.NoError(t, err)
	x2k3, err := keyring.GenerateKeypair(passphrase)
	require.NoError(t, err)

	x2skrfn := filepath.Join(td, "x2skr")
	x2skr, err := keyring.LoadSecring(x2skrfn)
	require.NoError(t, err)
	assert.NoError(t, x2skr.Set(x2k1))

	x2pkrfn := filepath.Join(td, "x2pkr")
	x2pkr, err := keyring.LoadPubring(x2pkrfn, x2skr)
	require.NoError(t, err)
	assert.NoError(t, x2pkr.Set(&x2k2.PublicKey))
	assert.NoError(t, x2pkr.Set(&x2k3.PublicKey))

	xc2 := &XC{
		pubring: x2pkr,
		secring: x2skr,
		client:  &fakeAgent{passphrase},
	}

	// export & import public key from X1 -> X2
	buf, err := xc1.ExportPublicKey(ctx, x1k1.Fingerprint())
	require.NoError(t, err)

	assert.NoError(t, xc2.ImportPublicKey(ctx, buf))
	assert.Equal(t, true, x2pkr.Contains(x1k1.Fingerprint()))

	// export & import private key from X2 -> X1
	buf, err = xc2.ExportPrivateKey(ctx, x2k1.Fingerprint())
	require.NoError(t, err)

	assert.NoError(t, xc1.ImportPrivateKey(ctx, buf))
	assert.Equal(t, true, x1pkr.Contains(x2k1.Fingerprint()))
	assert.Equal(t, true, x1skr.Contains(x2k1.Fingerprint()))
}
