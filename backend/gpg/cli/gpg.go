package cli

import (
	"bufio"
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/blang/semver"
	"github.com/justwatchcom/gopass/backend/gpg"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/pkg/errors"
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
	binary   string
	args     []string
	pubKeys  gpg.KeyList
	privKeys gpg.KeyList
}

// Config is the gpg wrapper config
type Config struct {
	Binary string
	Args   []string
	Umask  int
}

// New creates a new GPG wrapper
func New(cfg Config) (*GPG, error) {
	// ensure created files don't have group or world perms set
	// this setting should be inherited by sub-processes
	umask(cfg.Umask)

	// make sure GPG_TTY is set (if possible)
	if gt := os.Getenv("GPG_TTY"); gt == "" {
		if t := tty(); t != "" {
			_ = os.Setenv("GPG_TTY", t)
		}
	}

	g := &GPG{
		binary: "gpg",
		args:   append(defaultArgs, cfg.Args...),
	}
	if err := g.detectBinary(cfg.Binary); err != nil {
		return nil, err
	}

	return g, nil
}

// Binary returns the GPG binary location
func (g *GPG) Binary() string {
	return g.binary
}

// listKey lists all keys of the given type and matching the search strings
func (g *GPG) listKeys(ctx context.Context, typ string, search ...string) (gpg.KeyList, error) {
	args := []string{"--with-colons", "--with-fingerprint", "--fixed-list-mode", "--list-" + typ + "-keys"}
	args = append(args, search...)
	cmd := exec.CommandContext(ctx, g.binary, args...)
	cmd.Stderr = nil

	out.Debug(ctx, "gpg.listKeys: %s %+v\n", cmd.Path, cmd.Args)
	cmdout, err := cmd.Output()
	if err != nil {
		if bytes.Contains(cmdout, []byte("secret key not available")) {
			return gpg.KeyList{}, nil
		}
		return gpg.KeyList{}, err
	}

	return g.parseColons(bytes.NewBuffer(cmdout)), nil
}

// ListPublicKeys returns a parsed list of GPG public keys
func (g *GPG) ListPublicKeys(ctx context.Context) (gpg.KeyList, error) {
	if g.pubKeys == nil {
		kl, err := g.listKeys(ctx, "public")
		if err != nil {
			return nil, err
		}
		g.pubKeys = kl
	}
	return g.pubKeys, nil
}

// FindPublicKeys searches for the given public keys
func (g *GPG) FindPublicKeys(ctx context.Context, search ...string) (gpg.KeyList, error) {
	// TODO use cache
	return g.listKeys(ctx, "public", search...)
}

// ListPrivateKeys returns a parsed list of GPG secret keys
func (g *GPG) ListPrivateKeys(ctx context.Context) (gpg.KeyList, error) {
	if g.privKeys == nil {
		kl, err := g.listKeys(ctx, "secret")
		if err != nil {
			return nil, err
		}
		g.privKeys = kl
	}
	return g.privKeys, nil
}

// FindPrivateKeys searches for the given private keys
func (g *GPG) FindPrivateKeys(ctx context.Context, search ...string) (gpg.KeyList, error) {
	// TODO use cache
	return g.listKeys(ctx, "secret", search...)
}

// GetRecipients returns a list of recipient IDs for a given file
func (g *GPG) GetRecipients(ctx context.Context, file string) ([]string, error) {
	_ = os.Setenv("LANGUAGE", "C")
	recp := make([]string, 0, 5)

	args := []string{"--batch", "--list-only", "--list-packets", "--no-default-keyring", "--secret-keyring", "/dev/null", file}
	cmd := exec.CommandContext(ctx, g.binary, args...)
	out.Debug(ctx, "gpg.GetRecipients: %s %+v", cmd.Path, cmd.Args)

	cmdout, err := cmd.CombinedOutput()
	if err != nil {
		return []string{}, err
	}

	scanner := bufio.NewScanner(bytes.NewBuffer(cmdout))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		out.Debug(ctx, "gpg Output: %s", line)
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
// the trust-model will be set to always as to avoid (annoying) "unusable public key"
// errors when encrypting.
func (g *GPG) Encrypt(ctx context.Context, path string, content []byte, recipients []string) error {
	if err := os.MkdirAll(filepath.Dir(path), dirPerm); err != nil {
		return errors.Wrapf(err, "failed to create dir '%s'", path)
	}

	args := append(g.args, "--encrypt", "--output", path)
	if gpg.IsAlwaysTrust(ctx) {
		// changing the trustmodel is possibly dangerous. A user should always
		// explicitly opt-in to do this
		args = append(args, "--trust-model=always")
	}
	for _, r := range recipients {
		args = append(args, "--recipient", r)
	}

	cmd := exec.CommandContext(ctx, g.binary, args...)
	cmd.Stdin = bytes.NewReader(content)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	out.Debug(ctx, "gpg.Encrypt: %s %+v", cmd.Path, cmd.Args)
	return cmd.Run()
}

// Decrypt will try to decrypt the given file
func (g *GPG) Decrypt(ctx context.Context, path string) ([]byte, error) {
	args := append(g.args, "--decrypt", path)
	cmd := exec.CommandContext(ctx, g.binary, args...)

	out.Debug(ctx, "gpg.Decrypt: %s %+v", cmd.Path, cmd.Args)
	return cmd.Output()
}

// ExportPublicKey will export the named public key to the location given
func (g *GPG) ExportPublicKey(ctx context.Context, id, filename string) error {
	args := append(g.args, "--armor", "--export", id)
	cmd := exec.CommandContext(ctx, g.binary, args...)

	out.Debug(ctx, "gpg.ExportPublicKey: %s %+v", cmd.Path, cmd.Args)
	out, err := cmd.Output()
	if err != nil {
		return errors.Wrapf(err, "failed to run command '%s %+v'", cmd.Path, cmd.Args)
	}

	if len(out) < 1 {
		return errors.Errorf("Key not found")
	}

	return ioutil.WriteFile(filename, out, fileMode)
}

// ImportPublicKey will import a key from the given location
func (g *GPG) ImportPublicKey(ctx context.Context, filename string) error {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return errors.Wrapf(err, "failed to read file '%s'", filename)
	}

	args := append(g.args, "--import")
	cmd := exec.CommandContext(ctx, g.binary, args...)
	cmd.Stdin = bytes.NewReader(buf)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	out.Debug(ctx, "gpg.ImportPublicKey: %s %+v", cmd.Path, cmd.Args)
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "failed to run command: '%s %+v'", cmd.Path, cmd.Args)
	}

	// clear key cache
	g.privKeys = nil
	g.pubKeys = nil
	return nil
}

// Version will returns GPG version information
func (g *GPG) Version(ctx context.Context) semver.Version {
	v := semver.Version{}

	cmd := exec.CommandContext(ctx, g.binary, "--version")
	out, err := cmd.Output()
	if err != nil {
		return v
	}

	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "gpg ") {
			p := strings.Fields(line)
			sv, err := semver.Parse(p[len(p)-1])
			if err != nil {
				continue
			}
			return sv
		}
	}
	return v
}

// CreatePrivateKeyBatch will create a new GPG keypair in batch mode
func (g *GPG) CreatePrivateKeyBatch(ctx context.Context, name, email, passphrase string) error {
	buf := &bytes.Buffer{}
	// https://git.gnupg.org/cgi-bin/gitweb.cgi?p=gnupg.git;a=blob;f=doc/DETAILS;h=de0f21ccba60c3037c2a155156202df1cd098507;hb=refs/heads/STABLE-BRANCH-1-4#l716
	_, _ = buf.WriteString(`%echo Generating a RSA/RSA key pair
Key-Type: RSA
Key-Length: 2048
Subkey-Type: RSA
Subkey-Length: 2048
Expire-Date: 0
`)
	_, _ = buf.WriteString("Name-Real: " + name + "\n")
	_, _ = buf.WriteString("Name-Email: " + email + "\n")
	_, _ = buf.WriteString("Passphrase: " + passphrase + "\n")

	args := []string{"--batch", "--gen-key"}
	cmd := exec.CommandContext(ctx, g.binary, args...)
	cmd.Stdin = bytes.NewReader(buf.Bytes())
	cmd.Stdout = nil
	cmd.Stderr = nil

	out.Debug(ctx, "gpg.CreatePrivateKeyBatch: %s %+v", cmd.Path, cmd.Args)
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "failed to run command: '%s %+v'", cmd.Path, cmd.Args)
	}
	g.privKeys = nil
	g.pubKeys = nil
	return nil
}

// CreatePrivateKey will create a new GPG key in interactive mode
func (g *GPG) CreatePrivateKey(ctx context.Context) error {
	args := []string{"--gen-key"}
	cmd := exec.CommandContext(ctx, g.binary, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	out.Debug(ctx, "gpg.CreatePrivateKey: %s %+v", cmd.Path, cmd.Args)
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "failed to run command: '%s %+v'", cmd.Path, cmd.Args)
	}

	g.privKeys = nil
	g.pubKeys = nil
	return nil
}
