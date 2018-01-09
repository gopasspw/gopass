package gptest

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	gopassConfig = `askformore: false
autoimport: true
autosync: true
cliptimeout: 45
noconfirm: true
safecontent: true`
)

var (
	defaultEntries    = []string{"foo"}
	defaultRecipients = []string{"0xDEADBEEF"}
)

type Unit struct {
	t          *testing.T
	Entries    []string
	Recipients []string
	Dir        string
	env        map[string]string
}

func (u Unit) GPConfig() string {
	return filepath.Join(u.Dir, ".gopass.yml")
}

func (u Unit) GPGHome() string {
	return filepath.Join(u.Dir, ".gnupg")
}

func NewUnitTester(t *testing.T) *Unit {
	td, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)

	u := &Unit{
		t:          t,
		Entries:    defaultEntries,
		Recipients: defaultRecipients,
		Dir:        td,
	}
	u.env = map[string]string{
		"CHECKPOINT_DISABLE": "true",
		"GNUPGHOME":          u.GPGHome(),
		"GOPASS_CONFIG":      u.GPConfig(),
		"GOPASS_HOMEDIR":     u.Dir,
		"GOPASS_NOCOLOR":     "true",
		"GOPASS_NO_NOTIFY":   "true",
		"PAGER":              "",
	}
	assert.NoError(t, setupEnv(u.env))
	assert.NoError(t, os.Mkdir(u.GPGHome(), 0600))
	assert.NoError(t, u.initConfig())
	assert.NoError(t, u.InitStore(""))

	return u
}

func (u Unit) initConfig() error {
	return ioutil.WriteFile(u.GPConfig(), []byte(gopassConfig+"\npath: "+u.StoreDir("")+"\n"), 0600)
}

func (u Unit) StoreDir(mount string) string {
	if mount != "" {
		mount = "-" + mount
	}
	return filepath.Join(u.Dir, ".password-store"+mount)
}

func (u Unit) InitStore(name string) error {
	dir := u.StoreDir(name)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	if err := ioutil.WriteFile(filepath.Join(dir, ".gpg-id"), []byte(strings.Join(u.Recipients, "\n")), 0600); err != nil {
		return err
	}
	for _, p := range AllPathsToSlash(u.Entries) {
		if err := ioutil.WriteFile(filepath.Join(dir, p+".gpg"), []byte("secret"), 0600); err != nil {
			return err
		}
	}
	return nil
}

func (u *Unit) Remove() {
	teardownEnv(u.env)
	if u.Dir == "" {
		return
	}
	assert.NoError(u.t, os.RemoveAll(u.Dir))
	u.Dir = ""
}
