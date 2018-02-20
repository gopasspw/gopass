package cli

import (
	"bufio"
	"bytes"
	"context"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/justwatchcom/gopass/backend/crypto/gpg"
	"github.com/justwatchcom/gopass/utils/out"
)

var (
	reUIDComment        = regexp.MustCompile(`([^(<]+)\s+(\([^)]+\))\s+<([^>]+)>`)
	reUID               = regexp.MustCompile(`([^(<]+)\s+<([^>]+)>`)
	reUIDNoEmailComment = regexp.MustCompile(`([^(<]+)\s+(\([^)]+\))`)
	// defaultArgs contains the default GPG args for non-interactive use. Note: Do not use '--batch'
	// as this will disable (necessary) passphrase questions!
	defaultArgs = []string{"--quiet", "--yes", "--compress-algo=none", "--no-encrypt-to", "--no-auto-check-trustdb"}
	// Ext is the file extension used by this backend
	Ext = "gpg"
	// IDFile is the name of the recipients file used by this backend
	IDFile = ".gpg-id"
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
func New(ctx context.Context, cfg Config) (*GPG, error) {
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

	bin, err := Binary(ctx, cfg.Binary)
	if err != nil {
		return nil, err
	}
	g.binary = bin

	return g, nil
}

// RecipientIDs returns a list of recipient IDs for a given file
func (g *GPG) RecipientIDs(ctx context.Context, buf []byte) ([]string, error) {
	_ = os.Setenv("LANGUAGE", "C")
	recp := make([]string, 0, 5)

	args := []string{"--batch", "--list-only", "--list-packets", "--no-default-keyring", "--secret-keyring", "/dev/null"}
	cmd := exec.CommandContext(ctx, g.binary, args...)
	cmd.Stdin = bytes.NewReader(buf)
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
func (g *GPG) Encrypt(ctx context.Context, plaintext []byte, recipients []string) ([]byte, error) {
	args := append(g.args, "--encrypt")
	if gpg.IsAlwaysTrust(ctx) {
		// changing the trustmodel is possibly dangerous. A user should always
		// explicitly opt-in to do this
		args = append(args, "--trust-model=always")
	}
	for _, r := range recipients {
		args = append(args, "--recipient", r)
	}

	buf := &bytes.Buffer{}

	cmd := exec.CommandContext(ctx, g.binary, args...)
	cmd.Stdin = bytes.NewReader(plaintext)
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr

	out.Debug(ctx, "gpg.Encrypt: %s %+v", cmd.Path, cmd.Args)
	err := cmd.Run()
	return buf.Bytes(), err
}

// Decrypt will try to decrypt the given file
func (g *GPG) Decrypt(ctx context.Context, ciphertext []byte) ([]byte, error) {
	args := append(g.args, "--decrypt")
	cmd := exec.CommandContext(ctx, g.binary, args...)
	cmd.Stdin = bytes.NewReader(ciphertext)

	out.Debug(ctx, "gpg.Decrypt: %s %+v", cmd.Path, cmd.Args)
	return cmd.Output()
}

// Initialized always returns nil
func (g *GPG) Initialized(ctx context.Context) error {
	return nil
}

// Name returns gpg
func (g *GPG) Name() string {
	return "gpg"
}

// Ext returns gpg
func (g *GPG) Ext() string {
	return Ext
}

// IDFile returns .gpg-id
func (g *GPG) IDFile() string {
	return IDFile
}
