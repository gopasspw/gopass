package gpg

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	fileMode = 0600
	dirPerm  = 0700
)

var (
	reUIDComment = regexp.MustCompile(`([^(<]+)\s+(\([^)]+\))\s+<([^>]+)>`)
	reUID        = regexp.MustCompile(`([^(<]+)\s+<([^>]+)>`)
	// defaultArgs contains the default GPG args for non-interactive use. Note: Do not use '--batch'
	// as this will disable (necessary) passphrase questions!
	defaultArgs = []string{"--quiet", "--yes", "--compress-algo=none", "--no-encrypt-to", "--no-auto-check-trustdb"}
)

// GPG is a gpg wrapper
type GPG struct {
	binary      string
	args        []string
	debug       bool
	pubKeys     KeyList
	privKeys    KeyList
	alwaysTrust bool
}

// Config is the gpg wrapper config
type Config struct {
	Binary      string
	Args        []string
	Debug       bool
	AlwaysTrust bool
}

// New creates a new GPG wrapper
func New(cfg Config) *GPG {
	// ensure created files don't have group or world perms set
	// this setting should be inherited by sub-processes
	umask(077)

	for _, b := range []string{cfg.Binary, "gpg2", "gpg1", "gpg"} {
		if p, err := exec.LookPath(b); err == nil {
			cfg.Binary = p
			break
		}
	}
	if len(cfg.Args) < 1 {
		cfg.Args = defaultArgs
	}

	g := &GPG{
		binary:      cfg.Binary,
		args:        cfg.Args,
		debug:       cfg.Debug,
		alwaysTrust: cfg.AlwaysTrust,
	}
	return g
}

// listKey lists all keys of the given type and matching the search strings
func (g *GPG) listKeys(typ string, search ...string) (KeyList, error) {
	args := []string{"--with-colons", "--with-fingerprint", "--fixed-list-mode", "--list-" + typ + "-keys"}
	args = append(args, search...)
	cmd := exec.Command(g.binary, args...)
	if g.debug {
		fmt.Printf("[DEBUG] gpg.listKeys: %s %+v\n", cmd.Path, cmd.Args)
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		if bytes.Contains(out, []byte("secret key not available")) {
			return KeyList{}, nil
		}
		return KeyList{}, err
	}

	return g.parseColons(bytes.NewBuffer(out)), nil
}

// ListPublicKeys returns a parsed list of GPG public keys
func (g *GPG) ListPublicKeys() (KeyList, error) {
	if g.pubKeys == nil {
		kl, err := g.listKeys("public")
		if err != nil {
			return nil, err
		}
		g.pubKeys = kl
	}
	return g.pubKeys, nil
}

// FindPublicKeys searches for the given public keys
func (g *GPG) FindPublicKeys(search ...string) (KeyList, error) {
	// TODO use cache
	return g.listKeys("public", search...)
}

// ListPrivateKeys returns a parsed list of GPG secret keys
func (g *GPG) ListPrivateKeys() (KeyList, error) {
	if g.privKeys == nil {
		kl, err := g.listKeys("secret")
		if err != nil {
			return nil, err
		}
		g.privKeys = kl
	}
	return g.privKeys, nil
}

// FindPrivateKeys searches for the given private keys
func (g *GPG) FindPrivateKeys(search ...string) (KeyList, error) {
	// TODO use cache
	return g.listKeys("secret", search...)
}

// GetRecipients returns a list of recipient IDs for a given file
func (g *GPG) GetRecipients(file string) ([]string, error) {
	_ = os.Setenv("LANGUAGE", "C")
	recp := make([]string, 0, 5)

	args := []string{"--batch", "--list-only", "--list-packets", "--no-default-keyring", "--secret-keyring", "/dev/null", file}
	cmd := exec.Command(g.binary, args...)
	if g.debug {
		fmt.Printf("[DEBUG] gpg.GetRecipients: %s %+v\n", cmd.Path, cmd.Args)
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return []string{}, err
	}

	scanner := bufio.NewScanner(bytes.NewBuffer(out))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if g.debug {
			fmt.Printf("[DEBUG] gpg Output: %s\n", line)
		}
		if !strings.HasPrefix(line, ":pubkey enc packet:") {
			continue
		}
		m := splitPacket(line)
		if keyid, found := m["keyid"]; found {
			recp = append(recp, keyid)
		}
	}

	return recp, nil
}

// Encrypt will encrypt the given content for the recipients. If alwaysTrust is true
// the trust-model will be set to always as to avoid (annoying) "unuseable public key"
// errors when encrypting.
func (g *GPG) Encrypt(path string, content []byte, recipients []string) error {
	if err := os.MkdirAll(filepath.Dir(path), dirPerm); err != nil {
		return err
	}

	args := append(g.args, "--encrypt", "--output", path)
	if g.alwaysTrust {
		// changing the trustmodel is possibly dangerous. A user should always
		// explicitly opt-in to do this
		args = append(args, "--trust-model=always")
	}
	for _, r := range recipients {
		args = append(args, "--recipient", r)
	}

	cmd := exec.Command(g.binary, args...)
	if g.debug {
		fmt.Printf("[DEBUG] gpg.Encrypt: %s %+v\n", cmd.Path, cmd.Args)
	}
	cmd.Stdin = bytes.NewReader(content)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// Decrypt will try to decrypt the given file
func (g *GPG) Decrypt(path string) ([]byte, error) {
	args := append(g.args, "--decrypt", path)
	cmd := exec.Command(g.binary, args...)
	if g.debug {
		fmt.Printf("[DEBUG] gpg.Decrypt: %s %+v\n", cmd.Path, cmd.Args)
	}
	return cmd.Output()
}

// ExportPublicKey will export the named public key to the location given
func (g *GPG) ExportPublicKey(id, filename string) error {
	args := append(g.args, "--armor", "--export", id)
	cmd := exec.Command(g.binary, args...)
	if g.debug {
		fmt.Printf("[DEBUG] gpg.ExportPublicKey: %s %+v\n", cmd.Path, cmd.Args)
	}
	out, err := cmd.Output()
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, out, fileMode)
}

// ImportPublicKey will import a key from the given location
func (g *GPG) ImportPublicKey(filename string) error {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	args := append(g.args, "--import")
	cmd := exec.Command(g.binary, args...)
	if g.debug {
		fmt.Printf("[DEBUG] gpg.ImportPublicKey: %s %+v\n", cmd.Path, cmd.Args)
	}
	cmd.Stdin = bytes.NewReader(buf)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}
	// clear key cache
	g.privKeys = nil
	g.pubKeys = nil
	return nil
}
