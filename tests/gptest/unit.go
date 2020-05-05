package gptest

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/crypto/sha3"

	"github.com/gopasspw/gopass/pkg/backend"

	aclip "github.com/atotto/clipboard"
	"github.com/stretchr/testify/assert"
)

const (
	gopassConfig = `root:
  askformore: false
  autoclip: true
  autoimport: true
  autosync: true
  cliptimeout: 45
  noconfirm: false
  notifications: true
  safecontent: false`
)

var (
	defaultEntries    = []string{"foo"}
	defaultRecipients = []string{"0xDEADBEEF"}
)

// Unit is a gopass unit test helper
type Unit struct {
	t          *testing.T
	Entries    []string
	Recipients []string
	Dir        string
	env        map[string]string
}

// GPConfig returns the gopass config location
func (u Unit) GPConfig() string {
	return filepath.Join(u.Dir, ".gopass.yml")
}

// GPGHome returns the gopass homedir
func (u Unit) GPGHome() string {
	return filepath.Join(u.Dir, ".gnupg")
}

// NewUnitTester creates a new unit test helper
func NewUnitTester(t *testing.T) *Unit {
	aclip.Unsupported = true
	td, err := ioutil.TempDir("", "gopass-")
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
		"GOPASS_EXPERIMENTAL_GOGIT": "",
		"GOPASS_HOMEDIR":            u.Dir,
		"GOPASS_NOCOLOR":            "true",
		"GOPASS_NO_NOTIFY":          "true",
		"PAGER":                     "",
	}
	assert.NoError(t, setupEnv(u.env))
	assert.NoError(t, os.Mkdir(u.GPGHome(), 0700))
	assert.NoError(t, u.initConfig())
	assert.NoError(t, u.InitStore(""))

	return u
}

func (u Unit) initConfig() error {
	url := backend.FromPath(u.StoreDir(""))
	url.Crypto = backend.Plain
	url.RCS = backend.Noop
	return ioutil.WriteFile(
		u.GPConfig(),
		[]byte(gopassConfig+"\n  path: "+url.String()+"\n"+"  recipient_hash:\n    .gpg-id: "+fmt.Sprintf("%x", sha3.New512().Sum(u.recipients()))),
		0600,
	)
}

// StoreDir returns the password store dir
func (u Unit) StoreDir(mount string) string {
	if mount != "" {
		mount = "-" + mount
	}
	return filepath.Join(u.Dir, ".password-store"+mount)
}

func (u Unit) recipients() []byte {
	return []byte(strings.Join(u.Recipients, "\n"))
}

// InitStore initializes the test store
func (u Unit) InitStore(name string) error {
	dir := u.StoreDir(name)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	fn := filepath.Join(dir, ".gpg-id")
	_ = os.Remove(fn)
	if err := ioutil.WriteFile(fn, u.recipients(), 0600); err != nil {
		return err
	}
	for _, p := range AllPathsToSlash(u.Entries) {
		fn := filepath.Join(dir, p+".txt")
		_ = os.Remove(fn)
		if err := os.MkdirAll(filepath.Dir(fn), 0700); err != nil {
			return err
		}
		if err := ioutil.WriteFile(fn, []byte("secret\nsecond\nhird"), 0600); err != nil {
			return err
		}
	}
	return nil
}

// Remove removes the test store
func (u *Unit) Remove() {
	teardownEnv(u.env)
	if u.Dir == "" {
		return
	}
	assert.NoError(u.t, os.RemoveAll(u.Dir))
	u.Dir = ""
}
