package mock

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/blang/semver"
	"github.com/justwatchcom/gopass/backend/gpg"
	"github.com/pkg/errors"
)

var staticPrivateKeyList = gpg.KeyList{
	gpg.Key{
		KeyType:      "rsa",
		KeyLength:    2048,
		Validity:     "u",
		CreationDate: time.Now(),
		Fingerprint:  "000000000000000000000000DEADBEEF",
		Identities: map[string]gpg.Identity{
			"Dead Beef <dead.beef@example.com>": gpg.Identity{
				Name:         "Dead Beef",
				Email:        "dead.beef@example.com",
				CreationDate: time.Now(),
			},
		},
	},
}

// Mocker is a no-op GPG mock
type Mocker struct{}

// New creates a new GPG mock
func New() *Mocker {
	return &Mocker{}
}

// ListPublicKeys does nothing
func (m *Mocker) ListPublicKeys(context.Context) (gpg.KeyList, error) {
	return gpg.KeyList{}, nil
}

// FindPublicKeys does nothing
func (m *Mocker) FindPublicKeys(context.Context, ...string) (gpg.KeyList, error) {
	return gpg.KeyList{}, nil
}

// ListPrivateKeys does nothing
func (m *Mocker) ListPrivateKeys(context.Context) (gpg.KeyList, error) {
	return staticPrivateKeyList, nil
}

// FindPrivateKeys does nothing
func (m *Mocker) FindPrivateKeys(context.Context, ...string) (gpg.KeyList, error) {
	return staticPrivateKeyList, nil
}

// GetRecipients does nothing
func (m *Mocker) GetRecipients(context.Context, string) ([]string, error) {
	return []string{}, nil
}

// Encrypt writes the input to disk unaltered
func (m *Mocker) Encrypt(ctx context.Context, path string, content []byte, recipients []string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return errors.Wrapf(err, "failed to create dir '%s'", path)
	}
	return ioutil.WriteFile(path, content, 0600)
}

// Decrypt read the file from disk unaltered
func (m *Mocker) Decrypt(ctx context.Context, path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

// ExportPublicKey does nothing
func (m *Mocker) ExportPublicKey(context.Context, string, string) error {
	return nil
}

// ImportPublicKey does nothing
func (m *Mocker) ImportPublicKey(context.Context, string) error {
	return nil
}

// Version returns dummy version info
func (m *Mocker) Version(context.Context) semver.Version {
	return semver.Version{}
}

// Binary always returns 'gpg'
func (m *Mocker) Binary() string {
	return "gpg"
}

// Sign writes the hashsum to the given file
func (m *Mocker) Sign(ctx context.Context, in string, sigf string) error {
	buf, err := ioutil.ReadFile(in)
	if err != nil {
		return err
	}
	sum := sha256.New()
	_, _ = sum.Write(buf)
	hexsum := fmt.Sprintf("%X", sum.Sum(nil))
	return ioutil.WriteFile(sigf, []byte(hexsum), 0644)
}

// Verify does a pseudo-verification
func (m *Mocker) Verify(ctx context.Context, sigf string, in string) error {
	sigb, err := ioutil.ReadFile(sigf)
	if err != nil {
		return err
	}

	buf, err := ioutil.ReadFile(in)
	if err != nil {
		return err
	}
	sum := sha256.New()
	_, _ = sum.Write(buf)
	hexsum := fmt.Sprintf("%X", sum.Sum(nil))

	if string(sigb) != hexsum {
		return fmt.Errorf("hashsum mismatch")
	}

	return nil
}
