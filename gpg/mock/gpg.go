package mock

import (
	"context"
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
func (m *Mocker) ListPublicKeys(context.Context) (gpg.KeyList, error) {
	return gpg.KeyList{}, nil
}

// FindPublicKeys does nothing
func (m *Mocker) FindPublicKeys(context.Context, ...string) (gpg.KeyList, error) {
	return gpg.KeyList{}, nil
}

// ListPrivateKeys does nothing
func (m *Mocker) ListPrivateKeys(context.Context) (gpg.KeyList, error) {
	return gpg.KeyList{}, nil
}

// FindPrivateKeys does nothing
func (m *Mocker) FindPrivateKeys(context.Context, ...string) (gpg.KeyList, error) {
	return gpg.KeyList{}, nil
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
