package gptest

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gopasspw/clipboard"
	"github.com/stretchr/testify/require"
)

const (
	gopassConfig = `[generate]
	autoclip = true
[core]
	autoimport = true
	cliptimeout = 45
	notifications = true
	nopager = true
` // it's important the [core] subsection is the last one here
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
	return filepath.Join(u.Dir, ".config", "gopass", "config")
}

// GPGHome returns the GnuPG homedir.
func (u Unit) GPGHome() string {
	return filepath.Join(u.Dir, ".gnupg")
}

// NewUnitTester creates a new unit test helper.
func NewUnitTester(t *testing.T) *Unit {
	t.Helper()

	clipboard.ForceUnsupported = true

	td := t.TempDir()
	u := &Unit{
		t:          t,
		Entries:    defaultEntries,
		Recipients: defaultRecipients,
		Dir:        td,
	}

	u.env = map[string]string{
		"CHECKPOINT_DISABLE":        "true",
		"GNUPGHOME":                 u.GPGHome(),
		"GOPASS_CONFIG_NOSYSTEM":    "true",
		"GOPASS_CONFIG_NO_MIGRATE":  "true",
		"GOPASS_DISABLE_ENCRYPTION": "true",
		"GOPASS_HOMEDIR":            u.Dir,
		"NO_COLOR":                  "true",
		"GOPASS_NO_NOTIFY":          "true",
		"PAGER":                     "",
		"GIT_AUTHOR_NAME":           "gopass-tests",
		"GIT_AUTHOR_EMAIL":          "tests@gopass.pw",
	}
	setupEnv(t, u.env)

	require.NoError(t, os.Mkdir(u.GPGHome(), 0o700))
	// we need to init store before init config, so that the right folders exist
	require.NoError(t, u.InitStore(""), "init store")
	require.NoError(t, u.initConfig(), "pre-populate config")

	return u
}

func (u Unit) initConfig() error {
	if err := os.MkdirAll(filepath.Dir(u.GPConfig()), 0o755); err != nil {
		return fmt.Errorf("failed to initialize the test config at %q: %w", u.GPConfig(), err)
	}

	err := os.WriteFile(
		u.GPConfig(),
		[]byte(gopassConfig+"\texportkeys = true\n[mounts]\n\tpath = "+u.StoreDir("")+"\n"),
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
