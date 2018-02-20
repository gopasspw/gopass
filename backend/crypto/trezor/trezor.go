package trezor

import (
	"github.com/blang/semver"
	"context"
	"fmt"
	"github.com/rendaw/go-trezor"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/rendaw/go-trezor/messages"
	"github.com/rendaw/go-trezor-encrypt"
	"bytes"
	"encoding/gob"
	"regexp"
	"github.com/pkg/errors"
	"encoding/hex"
	"os"
	"path"
	"bufio"
)

const (
	// Ext is the extension used by this backend
	Ext = "trezor"
)

type Lookups struct {
	deviceIdLookup map[string][]byte
	keyIdLookup map[string]string
}

type TrezorCrypto struct {
	dir string
	lookups Lookups
}

func New(dir string) (*TrezorCrypto, error) {
	t := TrezorCrypto{
		dir: dir,
	}
	f, err := os.Open(path.Join(dir, "trezor_lookups.gob"))
	if err != nil {
		if os.IsNotExist(err) {
			t.lookups.deviceIdLookup = make(map[string][]byte)
			t.lookups.keyIdLookup= make(map[string]string)
			return &t, nil
		}
		return nil, errors.WithStack(err)
	}
	reader := bufio.NewReader(f)
	err = gob.NewDecoder(reader).Decode(&t.lookups)
	return &t, err
}

func (t *TrezorCrypto) Initialized(ctx context.Context) error {
	if len(t.dir) == 0 {
		return errors.Errorf("trezor crypto not initialized")
	}
	return nil
}

func (t *TrezorCrypto) Name() string {
	return "trezor"
}

func (t *TrezorCrypto) Version(ctx context.Context) semver.Version {
	return semver.Version{
		Patch: 1,
	}
}

func (t *TrezorCrypto) Ext() string {
	return Ext
}

func (t *TrezorCrypto) IDFile() string {
	return "trezor"
}

func locateTrezor(ctx context.Context, recipient string) (trezor.Transport, error) {
	matches := regexp.MustCompile(`\[([^\]]*)\]$`).FindStringSubmatch(recipient)
	if len(matches) == 0 {
		return nil, fmt.Errorf("couldn't find key fingerprint in recipient name")
	}
	deviceId := matches[1]
	devices, err := trezor.Enumerate()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	for _, device := range devices {
		found := false
		err = trezor_encrypt.TrezorDo(device, func(features messages.Features) error {
			if *features.DeviceId == deviceId {
				found = true
			}
			return nil
		})
		if err != nil {
			out.Print(ctx, "Couldn't access %s", device.String(), err)
			continue
		}
		if found {
			return device, nil
		}
	}
	return nil, fmt.Errorf("couldn't locate trezor for recipient %s; is it attached?", recipient)
}

type Encrypted struct {
	Identities map[string][]byte
	Ciphertext []byte
}

func (t *TrezorCrypto) Encrypt(ctx context.Context, plaintext []byte, recipients []string) ([]byte, error) {
	if len(recipients) != 1 {
		return nil, fmt.Errorf("Trezor supports only one recipient")
	}
	device, err := locateTrezor(ctx, recipients[0])
	if err != nil {
		return nil, errors.WithStack(err)
	}
	keyIdentity, err := t.identifyKey(ctx, device)
	if err != nil {
		return nil, err
	}
	var ciphertext []byte
	err = trezor_encrypt.TrezorDo(device, func(features messages.Features) error {
		var err error
		ciphertext, err = trezor_encrypt.EncryptWithDevice(
			device, true, "Authorize gopass to encrypt", "Authorize GOPASS", plaintext)
		return err
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	identities := make(map[string][]byte)
	identities["key1"] = keyIdentity
	err = enc.Encode(Encrypted{
		Identities: identities,
		Ciphertext: ciphertext,
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return buf.Bytes(), nil
}

func (t *TrezorCrypto) identifyKey(ctx context.Context, device trezor.Transport) ([]byte, error) {
	var key []byte
	err := trezor_encrypt.TrezorDo(device, func(features messages.Features) error {
			var err error
			key, err = trezor_encrypt.GetPublicKey(device, "gopass: Read public key")
			if err != nil {
				return errors.WithStack(err)
			}
			_ = t.writeIdentityMapping(ctx, key, *features.DeviceId)
			return nil
	})
	if err != nil {
		return nil, err
	}
	return key, nil
}

func (t *TrezorCrypto) Decrypt(ctx context.Context, ciphertext []byte) ([]byte, error) {
	buf := bytes.NewBuffer(ciphertext)
	var encrypted Encrypted
	err := gob.NewDecoder(buf).Decode(&encrypted)
	if err != nil {
		return nil, err
	}
	device, err := t.locateTrezorByKey(ctx, encrypted.Identities["key1"])
	if err != nil {
		return nil, err
	}
	var plaintext []byte
	err = trezor_encrypt.TrezorDo(device, func(features messages.Features) error {
		var err error
		plaintext, err = trezor_encrypt.EncryptWithDevice(
			device, false, "Authorize gopass to decrypt", "Authorize GOPASS", encrypted.Ciphertext)
		return err
	})
	return plaintext, err
}

func (t *TrezorCrypto) locateTrezorByKey(ctx context.Context, identity []byte) (trezor.Transport, error) {
	deviceId, foundLookup := t.lookups.keyIdLookup[string(identity)]
	devices, err := trezor.Enumerate()
	if err != nil {
		return nil, err
	}
	for _, device := range devices {
		var found bool
		err = trezor_encrypt.TrezorDo(device, func(features messages.Features) error {
			if !foundLookup {
				key, err := trezor_encrypt.GetPublicKey(device, "gopass: Read public key")
				if err != nil {
					return err
				}
				if bytes.Equal(key, identity) {
					_ = t.writeIdentityMapping(ctx, identity, *features.DeviceId)
					found = true
				}
			} else if *features.DeviceId == deviceId {
				found = true
			}
			return nil
		})
		if found {
			return device, nil
		}
	}
	return nil, fmt.Errorf("couldn't locate trezor with id[:8] %s", hex.EncodeToString(identity[:8]))
}

func (t *TrezorCrypto) writeIdentityMapping(ctx context.Context, keyId []byte, deviceId string) error {
	t.lookups.keyIdLookup[string(keyId)] = deviceId
	t.lookups.deviceIdLookup[string(keyId)] = keyId
	tempPath := path.Join(t.dir, fmt.Sprintf(".trezor_lookups.gob.%s", os.Getpid()))
	f, err := os.Create(tempPath)
	if err != nil {
		return  err
	}
	writer := bufio.NewWriter(f)
	enc := gob.NewEncoder(writer)
	err = enc.Encode(t.lookups)
	if err != nil {
		_ = os.Remove(tempPath)
		return  err
	}
	err = f.Close()
	if err != nil {
		_ = os.Remove(tempPath)
		return  err
	}
	err = os.Rename(tempPath, path.Join(t.dir, "trezor_lookups.gob"))
	if err != nil {
		_ = os.Remove(tempPath)
		return err
	}
	return nil
}

func (t *TrezorCrypto) RecipientIDs(ctx context.Context, ciphertext []byte) ([]string, error) {
	return []string{}, nil
}

func (t *TrezorCrypto) ImportPublicKey(ctx context.Context, key []byte) error {
	return fmt.Errorf("trezor doesn't currently support importing public keys")
}

func (t *TrezorCrypto) ExportPublicKey(ctx context.Context, id string) ([]byte, error) {
	return nil, nil
}

func (t *TrezorCrypto) ListPublicKeyIDs(ctx context.Context) ([]string, error) {
	return nil, fmt.Errorf("trezor doesn't currently support listing public key ids")
}

func (t *TrezorCrypto) ListPrivateKeyIDs(ctx context.Context) ([]string, error) {
	var ids []string
	devices, err := trezor.Enumerate()
	if err != nil {
		return nil, err
	}
	for _, device := range devices {
		err = trezor_encrypt.TrezorDo(device, func(features messages.Features) error {
			ids = append(ids, fmt.Sprintf("%s [%s]", features.GetLabel(), *features.DeviceId))
			return nil
		})
		if err != nil {
			out.Print(ctx, "Couldn't access %s", device.String(), err)
			continue
		}
	}
	return ids, nil
}

func (t *TrezorCrypto) FindPublicKeys(ctx context.Context, needles ...string) ([]string, error) {
	/*
	var ids []string
	devices, err := trezor.Enumerate()
	if err != nil {
		return nil, err
	}
	for _, device := range devices {
		for _, needle := range needles {
			path := strings.Split(needle, "] ")[0][1:]
			if device.Info.Path == path {
				ids = append(ids, needle)
			}
		}
	}
	*/
	return needles, nil
}

func (t *TrezorCrypto) FindPrivateKeys(ctx context.Context, needles ...string) ([]string, error) {
	return needles, nil
}

func (t *TrezorCrypto) FormatKey(ctx context.Context, id string) string {
	return id
}

func (t *TrezorCrypto) NameFromKey(ctx context.Context, id string) string {
	return id
}

func (t *TrezorCrypto) EmailFromKey(ctx context.Context, id string) string {
	return id
}

func (t *TrezorCrypto) ReadNamesFromKey(ctx context.Context, buf []byte) ([]string, error) {
	return nil, fmt.Errorf("trezor doesn't currently support read names from key")
}

func (t *TrezorCrypto) CreatePrivateKeyBatch(ctx context.Context, name, email, passphrase string) error {
	return fmt.Errorf("trezor doesn't currently support creating private key batches")
}

func (t *TrezorCrypto) CreatePrivateKey(ctx context.Context) error {
	return fmt.Errorf("trezor doesn't currently support creating private keys")
}
