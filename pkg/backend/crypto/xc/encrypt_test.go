package xc

import (
	"bytes"
	"context"
	"crypto/rand"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/gopasspw/gopass/pkg/backend/crypto/xc/keyring"
	"github.com/gopasspw/gopass/pkg/backend/crypto/xc/xcpb"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeAgent struct {
	pw string
}

func (f *fakeAgent) Ping(context.Context) error {
	return nil
}

func (f *fakeAgent) Remove(context.Context, string) error {
	return nil
}

func (f *fakeAgent) Passphrase(context.Context, string, string) (string, error) {
	return f.pw, nil
}

func TestEncryptSimple(t *testing.T) {
	ctx := context.Background()

	td, err := ioutil.TempDir("", "gopass-")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(td)
	}()
	assert.NoError(t, os.Setenv("GOPASS_CONFIG", filepath.Join(td, ".gopass.yml")))
	assert.NoError(t, os.Setenv("GOPASS_HOMEDIR", td))

	passphrase := "test"

	k1, err := keyring.GenerateKeypair(passphrase)
	require.NoError(t, err)

	skr := keyring.NewSecring()
	assert.NoError(t, skr.Set(k1))

	pkr := keyring.NewPubring(skr)

	xc := &XC{
		pubring: pkr,
		secring: skr,
		client:  &fakeAgent{passphrase},
	}

	buf, err := xc.Encrypt(ctx, []byte("foobar"), []string{k1.Fingerprint()})
	require.NoError(t, err)

	recps, err := xc.RecipientIDs(ctx, buf)
	require.NoError(t, err)
	assert.Equal(t, []string{k1.Fingerprint()}, recps)

	buf, err = xc.Decrypt(ctx, buf)
	require.NoError(t, err)
	assert.Equal(t, "foobar", string(buf))
}

func TestEncryptMultiKeys(t *testing.T) {
	ctx := context.Background()

	td, err := ioutil.TempDir("", "gopass-")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(td)
	}()
	assert.NoError(t, os.Setenv("GOPASS_CONFIG", filepath.Join(td, ".gopass.yml")))
	assert.NoError(t, os.Setenv("GOPASS_HOMEDIR", td))

	passphrase := "test"

	k1, err := keyring.GenerateKeypair(passphrase)
	require.NoError(t, err)
	k2, err := keyring.GenerateKeypair(passphrase)
	require.NoError(t, err)
	k3, err := keyring.GenerateKeypair(passphrase)
	require.NoError(t, err)

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
	require.NoError(t, err)

	recps, err := xc.RecipientIDs(ctx, buf)
	require.NoError(t, err)
	assert.Equal(t, []string{k1.Fingerprint()}, recps)

	buf, err = xc.Decrypt(ctx, buf)
	require.NoError(t, err)
	assert.Equal(t, "foobar", string(buf))
}

func TestEncryptChunks(t *testing.T) {
	ctx := context.Background()

	td, err := ioutil.TempDir("", "gopass-")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(td)
	}()
	assert.NoError(t, os.Setenv("GOPASS_CONFIG", filepath.Join(td, ".gopass.yml")))
	assert.NoError(t, os.Setenv("GOPASS_HOMEDIR", td))

	passphrase := "test"

	k1, err := keyring.GenerateKeypair(passphrase)
	require.NoError(t, err)

	skr := keyring.NewSecring()
	assert.NoError(t, skr.Set(k1))

	pkr := keyring.NewPubring(skr)

	xc := &XC{
		pubring: pkr,
		secring: skr,
		client:  &fakeAgent{passphrase},
	}

	plaintext := &bytes.Buffer{}
	p := make([]byte, 1024)
	for i := 0; i < 10*(chunkSizeMax/1024); i++ {
		_, _ = rand.Read(p)
		plaintext.Write(p)
	}
	assert.Equal(t, 163840, plaintext.Len())

	ciphertext, err := xc.Encrypt(ctx, plaintext.Bytes(), []string{k1.Fingerprint()})
	require.NoError(t, err)

	// check recipients
	recps, err := xc.RecipientIDs(ctx, ciphertext)
	require.NoError(t, err)
	assert.Equal(t, []string{k1.Fingerprint()}, recps)

	// check number of chunks
	msg := &xcpb.Message{}
	assert.NoError(t, proto.Unmarshal(ciphertext, msg))
	assert.Equal(t, 10, len(msg.Chunks))

	// check decryption works and yields exactly the input
	plainagain, err := xc.Decrypt(ctx, ciphertext)
	assert.NoError(t, err)
	assert.Equal(t, plaintext.String(), string(plainagain))

	// reorder some chunks
	msg = &xcpb.Message{}
	assert.NoError(t, proto.Unmarshal(ciphertext, msg))
	assert.Equal(t, 10, len(msg.Chunks))

	msg.Chunks[0], msg.Chunks[1] = msg.Chunks[1], msg.Chunks[0]

	ciphertext, err = proto.Marshal(msg)
	assert.NoError(t, err)

	// check decryption fails
	_, err = xc.Decrypt(ctx, ciphertext)
	assert.Error(t, err)
}

func TestEncryptCompress(t *testing.T) {
	ctx := context.Background()

	td, err := ioutil.TempDir("", "gopass-")
	require.NoError(t, err)
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

	data := &bytes.Buffer{}
	for i := 0; i < 1024*1024; i++ {
		data.WriteString("aaaaaaaa")
	}

	buf, err := xc.Encrypt(ctx, data.Bytes(), []string{k1.Fingerprint()})
	assert.NoError(t, err)

	recps, err := xc.RecipientIDs(ctx, buf)
	assert.NoError(t, err)
	assert.Equal(t, []string{k1.Fingerprint()}, recps)

	// check compress flag
	msg := &xcpb.Message{}
	assert.NoError(t, proto.Unmarshal(buf, msg))
	assert.Equal(t, true, msg.Compressed)

	buf, err = xc.Decrypt(ctx, buf)
	assert.NoError(t, err)
	assert.Equal(t, data.String(), string(buf))
}
