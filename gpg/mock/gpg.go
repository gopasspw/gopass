package mock

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/blang/semver"
	"github.com/justwatchcom/gopass/gpg"
	"github.com/pkg/errors"
)

// Mocker is a no-op GPG mock
type Mocker struct{}

// New creates a new GPG mock
func New() *Mocker {
	return &Mocker{}
}

// ListPublicKeys does nothing
func (m *Mocker) ListPublicKeys() (gpg.KeyList, error) {
	return gpg.KeyList{}, nil
}

// FindPublicKeys does nothing
func (m *Mocker) FindPublicKeys(...string) (gpg.KeyList, error) {
	return gpg.KeyList{}, nil
}

// ListPrivateKeys does nothing
func (m *Mocker) ListPrivateKeys() (gpg.KeyList, error) {
	return gpg.KeyList{}, nil
}

// FindPrivateKeys does nothing
func (m *Mocker) FindPrivateKeys(...string) (gpg.KeyList, error) {
	return gpg.KeyList{}, nil
}

// GetRecipients does nothing
func (m *Mocker) GetRecipients(string) ([]string, error) {
	return []string{}, nil
}

// Encrypt writes the input to disk unaltered
func (m *Mocker) Encrypt(path string, content []byte, recipients []string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return errors.Wrapf(err, "failed to create dir '%s'", path)
	}
	return ioutil.WriteFile(path, content, 0600)
}

// Decrypt read the file from disk unaltered
func (m *Mocker) Decrypt(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

// ExportPublicKey does nothing
func (m *Mocker) ExportPublicKey(string, string) error {
	return nil
}

// ImportPublicKey does nothing
func (m *Mocker) ImportPublicKey(string) error {
	return nil
}

// Version returns dummy version info
func (m *Mocker) Version() semver.Version {
	return semver.Version{}
}
