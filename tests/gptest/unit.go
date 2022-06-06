package gptest

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	aclip "github.com/atotto/clipboard"
	"github.com/stretchr/testify/assert"
)

const (
	gopassConfig = `autoclip: true
autoimport: true
cliptimeout: 45
notifications: true
parsing: true`
)

var (
	defaultEntries    = []string{"foo"}
	defaultRecipients = []string{"0xDEADBEEF"}
)

// Unit is a gopass unit test helper.
type Unit struct {
	t          *testing.T
	Entries    []string
	Recipients []string
	Dir        string
	env        map[string]string
}

// GPConfig returns the gopass config location.
func (u Unit) GPConfig() string {
	return filepath.Join(u.Dir, "config.yml")
}

// GPGHome returns the GnuPG homedir.
func (u Unit) GPGHome() string {
	return filepath.Join(u.Dir, ".gnupg")
}

// NewUnitTester creates a new unit test helper.
func NewUnitTester(t *testing.T) *Unit {
	t.Helper()

	aclip.Unsupported = true
	td, err := os.MkdirTemp("", "gopass-")
	assert.NoError(t, err)

	u := &Unit{
		t:          t,
		Entries:    defaultEntries,
		Recipients: defaultRecipients,
		Dir:        td,
	}
	u.env = map[string]string{
		"CHECKPOINT_DISABLE":        "true",
		"GNUPGHOME":                 u.GPGHome(),
		"GOPASS_CONFIG":             u.GPConfig(),
		"GOPASS_DISABLE_ENCRYPTION": "true",
		"GOPASS_HOMEDIR":            u.Dir,
		"NO_COLOR":                  "true",
		"GOPASS_NO_NOTIFY":          "true",
		"PAGER":                     "",
	}
	assert.NoError(t, setupEnv(u.env))
	assert.NoError(t, os.Mkdir(u.GPGHome(), 0o700))
	assert.NoError(t, u.initConfig())
	assert.NoError(t, u.InitStore(""))

	return u
}

func (u Unit) initConfig() error {
	err := os.WriteFile(
		u.GPConfig(),
		[]byte(gopassConfig+"\npath: "+u.StoreDir("")+"\nexportkeys: true\n"),
		0o600,
	)
	if err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// StoreDir returns the password store dir.
func (u Unit) StoreDir(mount string) string {
	if mount != "" {
		mount = "-" + mount
	}

	return filepath.Join(u.Dir, "password-store"+mount)
}

func (u Unit) recipients() []byte {
	return []byte(strings.Join(u.Recipients, "\n"))
}

// InitStore initializes the test store.
func (u Unit) InitStore(name string) error {
	dir := u.StoreDir(name)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("failed to create store dir %s: %w", dir, err)
	}

	fn := filepath.Join(dir, ".plain-id") // plain.IDFile
	_ = os.Remove(fn)

	if err := os.WriteFile(fn, u.recipients(), 0o600); err != nil {
		return fmt.Errorf("failed to write IDFile %s: %w", fn, err)
	}

	for _, p := range AllPathsToSlash(u.Entries) {
		fn := filepath.Join(dir, p+".txt") // plain.Ext
		_ = os.Remove(fn)

		if err := os.MkdirAll(filepath.Dir(fn), 0o700); err != nil {
			return fmt.Errorf("failed to create dir %s: %w", filepath.Dir(fn), err)
		}

		if err := os.WriteFile(fn, []byte("secret\nsecond\nthird"), 0o600); err != nil {
			return fmt.Errorf("failed to write file %s: %w", fn, err)
		}
	}

	return nil
}

// Remove removes the test store.
func (u *Unit) Remove() {
	teardownEnv(u.env)

	if u.Dir == "" {
		return
	}

	assert.NoError(u.t, os.RemoveAll(u.Dir))
	u.Dir = ""
}
