package gptest

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/armor"
	aclip "github.com/atotto/clipboard"
	"github.com/gopasspw/gopass/tests/can"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var gpgDefaultRecipients = []string{"BE73F104"}

// GUnit is a gopass unit test helper.
type GUnit struct {
	t          *testing.T
	Entries    []string
	Recipients []string
	Dir        string
	env        map[string]string
}

// GPConfig returns the gopass config location.
func (u GUnit) GPConfig() string {
	return filepath.Join(u.Dir, ".config", "gopass", "config")
}

// GPGHome returns the GnuPG homedir.
func (u GUnit) GPGHome() string {
	return filepath.Join(u.Dir, ".gnupg")
}

// NewGUnitTester creates a new unit test helper.
func NewGUnitTester(t *testing.T) *GUnit {
	t.Helper()

	aclip.Unsupported = true

	td := t.TempDir()
	u := &GUnit{
		t:          t,
		Entries:    defaultEntries,
		Recipients: gpgDefaultRecipients,
		Dir:        td,
	}

	u.env = map[string]string{
		"CHECKPOINT_DISABLE":       "true",
		"GNUPGHOME":                u.GPGHome(),
		"GOPASS_CONFIG_NOSYSTEM":   "true",
		"GOPASS_CONFIG_NO_MIGRATE": "true",
		"GOPASS_HOMEDIR":           u.Dir,
		"NO_COLOR":                 "true",
		"GOPASS_NO_NOTIFY":         "true",
		"PAGER":                    "",
		"GIT_AUTHOR_NAME":          "gopass-tests",
		"GIT_AUTHOR_EMAIL":         "tests@gopass.pw",
	}
	setupEnv(t, u.env)

	require.NoError(t, os.Mkdir(u.GPGHome(), 0o700))
	assert.NoError(t, u.initConfig())
	assert.NoError(t, u.InitStore(""))

	return u
}

func (u GUnit) initConfig() error {
	if err := os.MkdirAll(filepath.Dir(u.GPConfig()), 0o755); err != nil {
		return err
	}
	err := os.WriteFile(
		u.GPConfig(),
		[]byte(gopassConfig+"\texportkeys = false\n[mounts]\npath = "+u.StoreDir("")+"\n"),
		0o600,
	)
	if err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// StoreDir returns the password store dir.
func (u GUnit) StoreDir(mount string) string {
	if mount != "" {
		mount = "-" + mount
	}

	return filepath.Join(u.Dir, "password-store"+mount)
}

func (u GUnit) recipients() []byte {
	return []byte(strings.Join(u.Recipients, "\n"))
}

// InitStore initializes the test store.
func (u GUnit) InitStore(name string) error {
	dir := u.StoreDir(name)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("failed to create store dir %s: %w", dir, err)
	}

	fn := filepath.Join(dir, ".gpg-id") // gpgcli.IDFile
	_ = os.Remove(fn)

	if err := os.WriteFile(fn, u.recipients(), 0o600); err != nil {
		return fmt.Errorf("failed to write IDFile %s: %w", fn, err)
	}

	if err := can.WriteTo(u.GPGHome()); err != nil {
		return err
	}

	// write embedded public keys to the store so we can import them
	el := can.EmbeddedKeyRing()
	for _, e := range el {
		tfn := filepath.Join(dir, ".public-keys", e.PrimaryKey.KeyIdShortString())
		if err := os.MkdirAll(filepath.Dir(tfn), 0o700); err != nil {
			return fmt.Errorf("failed to create public-keys dir %s: %w", filepath.Dir(tfn), err)
		}
		fh, err := os.Create(tfn)
		if err != nil {
			return fmt.Errorf("failed to create public-keys file %s: %w", tfn, err)
		}
		defer fh.Close() //nolint:errcheck

		wc, err := armor.Encode(fh, openpgp.PublicKeyType, nil)
		if err != nil {
			return err
		}
		if err := e.Serialize(wc); err != nil {
			return err
		}
		if err := wc.Close(); err != nil {
			return err
		}
	}

	return nil
}
