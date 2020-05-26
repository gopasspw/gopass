// Package plain implements a plaintext backend
// TODO(2.x) DEPRECATED and slated for removal in the 2.0.0 release.
package plain

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/gopasspw/gopass/internal/backend/crypto/gpg"

	"github.com/blang/semver"
)

var staticPrivateKeyList = gpg.KeyList{
	gpg.Key{
		KeyType:      "rsa",
		KeyLength:    2048,
		Validity:     "u",
		CreationDate: time.Now(),
		Fingerprint:  "000000000000000000000000DEADBEEF",
		Identities: map[string]gpg.Identity{
			"Dead Beef <dead.beef@example.com>": {
				Name:         "Dead Beef",
				Email:        "dead.beef@example.com",
				CreationDate: time.Now(),
			},
		},
	},
	gpg.Key{
		KeyType:      "rsa",
		KeyLength:    2048,
		Validity:     "u",
		CreationDate: time.Now(),
		Fingerprint:  "000000000000000000000000FEEDBEEF",
		Identities: map[string]gpg.Identity{
			"Feed Beef <feed.beef@example.com>": {
				Name:         "Feed Beef",
				Email:        "feed.beef@example.com",
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

// ListRecipients does nothing
func (m *Mocker) ListRecipients(context.Context) ([]string, error) {
	return staticPrivateKeyList.Recipients(), nil
}

// FindPublicKeys does nothing
func (m *Mocker) FindPublicKeys(ctx context.Context, keys ...string) ([]string, error) {
	rs := staticPrivateKeyList.Recipients()
	res := make([]string, 0, len(rs))
	for _, r := range rs {
		for _, needle := range keys {
			if strings.HasSuffix(r, needle) {
				res = append(res, r)
			}
		}
	}
	return res, nil
}

// ListIdentities does nothing
func (m *Mocker) ListIdentities(context.Context) ([]string, error) {
	return staticPrivateKeyList.Recipients(), nil
}

// FindPrivateKeys does nothing
func (m *Mocker) FindPrivateKeys(ctx context.Context, keys ...string) ([]string, error) {
	return m.FindPublicKeys(ctx, keys...)
}

// RecipientIDs does nothing
func (m *Mocker) RecipientIDs(context.Context, []byte) ([]string, error) {
	return staticPrivateKeyList.Recipients(), nil
}

// Encrypt writes the input to disk unaltered
func (m *Mocker) Encrypt(ctx context.Context, content []byte, recipients []string) ([]byte, error) {
	return content, nil
}

// Decrypt read the file from disk unaltered
func (m *Mocker) Decrypt(ctx context.Context, ciphertext []byte) ([]byte, error) {
	return ciphertext, nil
}

// ExportPublicKey does nothing
func (m *Mocker) ExportPublicKey(context.Context, string) ([]byte, error) {
	return nil, nil
}

// ImportPublicKey does nothing
func (m *Mocker) ImportPublicKey(context.Context, []byte) error {
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

// CreatePrivateKey is not implemented
func (m *Mocker) CreatePrivateKey(ctx context.Context) error {
	return fmt.Errorf("not yet implemented")
}

// CreatePrivateKeyBatch is not implemented
func (m *Mocker) CreatePrivateKeyBatch(ctx context.Context, name, email, pw string) error {
	return fmt.Errorf("not yet implemented")
}

// EmailFromKey returns nothing
func (m *Mocker) EmailFromKey(context.Context, string) string {
	return ""
}

// NameFromKey returns nothing
func (m *Mocker) NameFromKey(context.Context, string) string {
	return ""
}

// FormatKey returns the id
func (m *Mocker) FormatKey(ctx context.Context, id string) string {
	return id
}

// Fingerprint returns the full-length native fingerprint
func (m *Mocker) Fingerprint(ctx context.Context, id string) string {
	return id
}

// Initialized returns nil
func (m *Mocker) Initialized(context.Context) error {
	return nil
}

// Name returns plain
func (m *Mocker) Name() string {
	return Name
}

// Ext returns gpg
func (m *Mocker) Ext() string {
	return Ext
}

const (
	// Name is the name of this backend
	Name = "plain"
	// Ext is the file extension used by this backend
	Ext = "txt"
	// IDFile is the name of the recipients file used by this backend
	IDFile = ".plain-id"
)

// IDFile returns .gpg-id
func (m *Mocker) IDFile() string {
	return IDFile
}

// ReadNamesFromKey does nothing
func (m *Mocker) ReadNamesFromKey(ctx context.Context, buf []byte) ([]string, error) {
	return []string{"unsupported"}, nil
}
