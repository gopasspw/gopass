package age

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/blang/semver"
	"github.com/google/go-github/github"
	"github.com/gopasspw/gopass/pkg/out"
)

const (
	// Ext is the file extension for age encrypted secrets
	Ext = "age"
	// IDFile is the name for age recipients
	IDFile = ".age-ids"
)

// Age is an age backend
type Age struct {
	binary  string
	keyring string
	ghc     *github.Client
	ghCache *ghCache
}

// New creates a new Age backend
func New() (*Age, error) {
	ucd, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	return &Age{
		binary:  "age",
		ghc:     github.NewClient(nil),
		ghCache: &ghCache{},
		keyring: filepath.Join(ucd, "gopass", "age.txt"),
	}, nil
}

// Initialized returns nil
func (a *Age) Initialized(ctx context.Context) error {
	if a == nil {
		return fmt.Errorf("Age not initialized")
	}

	return nil
}

// Name returns age
func (a *Age) Name() string {
	return "age"
}

// Version return 1.0.0
func (a *Age) Version(ctx context.Context) semver.Version {
	return semver.Version{
		Patch: 1,
	}
}

// Ext returns the extension
func (a *Age) Ext() string {
	return Ext
}

// IDFile return the recipients file
func (a *Age) IDFile() string {
	return IDFile
}

func (a *Age) filterRecipients(ctx context.Context, recipients []string) ([]string, error) {
	out := make([]string, 0, len(recipients))
	for _, r := range recipients {
		if strings.HasPrefix(r, "age1") {
			out = append(out, r)
			continue
		}
		if strings.HasPrefix(r, "ssh-ed25519 ") {
			out = append(out, r)
			continue
		}
		if strings.HasPrefix(r, "ssh-rsa ") {
			out = append(out, r)
			continue
		}
		if strings.HasPrefix(r, "github:") {
			pks, err := a.getPublicKeysGithub(ctx, strings.TrimPrefix(r, "github:"))
			if err != nil {
				return out, err
			}
			out = append(out, pks...)
		}
	}
	return out, nil
}

// Encrypt will encrypt the given payload
func (a *Age) Encrypt(ctx context.Context, plaintext []byte, recipients []string) ([]byte, error) {
	args := []string{"--recipient", a.pkself(ctx)}
	recp, err := a.filterRecipients(ctx, recipients)
	if err != nil {
		return nil, err
	}
	for _, r := range recp {
		args = append(args, "--recipient", r)
	}

	buf := &bytes.Buffer{}

	cmd := exec.CommandContext(ctx, a.binary, args...)
	cmd.Stdin = bytes.NewReader(plaintext)
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr

	out.Debug(ctx, "age.Encrypt: %s %+v", cmd.Path, cmd.Args)
	err = cmd.Run()
	return buf.Bytes(), err
}

// Decrypt will attempt to decrypt the given payload
func (a *Age) Decrypt(ctx context.Context, ciphertext []byte) ([]byte, error) {
	args := []string{"--decrypt"}
	for _, k := range a.listPrivateKeyFiles(ctx) {
		args = append(args, "--identity")
		args = append(args, k)
	}

	cmd := exec.CommandContext(ctx, a.binary, args...)
	cmd.Stdin = bytes.NewReader(ciphertext)
	cmd.Stderr = os.Stderr

	out.Debug(ctx, "age.Decrypt: %s %+v", cmd.Path, cmd.Args)
	return cmd.Output()
}

// CreatePrivateKey will create a new native private key
func (a *Age) CreatePrivateKey(ctx context.Context) error {
	buf := &bytes.Buffer{}
	cmd := exec.CommandContext(ctx, "age-keygen")
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(a.keyring), 0700); err != nil {
		return err
	}

	fh, err := os.OpenFile(a.keyring, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer fh.Close()

	out.Debug(ctx, "age.CreatePrivateKey: %s", buf.String())

	_, err = fh.Write(buf.Bytes())
	return err
}

// CreatePrivateKeyBatch is TODO
func (a *Age) CreatePrivateKeyBatch(ctx context.Context, name, email, pw string) error {
	return a.CreatePrivateKey(ctx)
}

// ListPrivateKeyIDs is TODO
func (a *Age) ListPrivateKeyIDs(ctx context.Context) ([]string, error) {
	native, err := a.getNativeKeypairs(ctx)
	if err != nil {
		return nil, err
	}
	ssh, err := a.getSSHKeypairs()
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(native)+len(ssh))
	for k := range native {
		ids = append(ids, k)
	}
	for k := range ssh {
		ids = append(ids, k)
	}
	return ids, nil
}

func (a *Age) listPrivateKeyFiles(ctx context.Context) []string {
	keys, err := a.getAllKeypairs(ctx)
	if err != nil {
		out.Debug(ctx, "Error fetching keys: %s", err)
	}
	ids := make([]string, 0, len(keys))
	for _, v := range keys {
		ids = append(ids, v)
	}
	return ids
}

// ExportPublicKey is not implemented
func (a *Age) ExportPublicKey(ctx context.Context, id string) ([]byte, error) {
	return nil, fmt.Errorf("not implemented")
}

// map[public key]filename
func (a *Age) getSSHKeypairs() (map[string]string, error) {
	uhd, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	files, err := ioutil.ReadDir(filepath.Join(uhd, ".ssh"))
	if err != nil {
		return nil, err
	}
	keys := make(map[string]string, len(files))
	for _, file := range files {
		fn := file.Name()
		if !strings.HasSuffix(fn, ".pub") {
			continue
		}
		pfn := strings.TrimSuffix(fn, ".pub")
		_, err := os.Stat(pfn)
		if err != nil {
			continue
		}
		pbuf, err := ioutil.ReadFile(fn)
		if err != nil {
			continue
		}
		p := strings.Split(string(pbuf), " ")
		if len(p) < 2 {
			continue
		}
		keys[strings.Join(p[0:1], " ")] = pfn
	}
	return keys, nil
}

func (a *Age) pkself(ctx context.Context) string {
	keys, _ := a.getNativeKeypairs(ctx)
	for k := range keys {
		return k
	}
	return ""
}

// map[public key]filename
func (a *Age) getNativeKeypairs(ctx context.Context) (map[string]string, error) {
	_, err := os.Stat(a.keyring)
	if os.IsNotExist(err) {
		out.Debug(ctx, "No native age key found. Generating ...")
		if err := a.CreatePrivateKey(ctx); err != nil {
			return nil, err
		}
	}
	buf, err := ioutil.ReadFile(a.keyring)
	if err != nil {
		return nil, err
	}
	var pub, priv string
	lines := strings.Split(string(buf), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "# public key: age1") {
			pub = strings.TrimPrefix(line, "# public key: ")
			continue
		}
		if strings.HasPrefix(line, "AGE-SECRET-KEY") {
			priv = a.keyring
			break
		}
	}
	return map[string]string{
		pub: priv,
	}, nil
}

func (a *Age) getAllKeypairs(ctx context.Context) (map[string]string, error) {
	native, err := a.getNativeKeypairs(ctx)
	if err != nil {
		return nil, err
	}
	ssh, err := a.getSSHKeypairs()
	if err != nil {
		return nil, err
	}

	keys := make(map[string]string, len(native)+len(ssh))
	for k, v := range native {
		keys[k] = v
	}
	for k, v := range ssh {
		keys[k] = v
	}
	return keys, nil
}

// FindPublicKeys it TODO
func (a *Age) FindPublicKeys(ctx context.Context, keys ...string) ([]string, error) {
	nk, err := a.getAllKeypairs(ctx)
	if err != nil {
		return nil, err
	}
	matches := make([]string, 0, len(nk))
	for _, k := range keys {
		if _, found := nk[k]; found {
			matches = append(matches, k)
		}
	}
	return matches, nil
}

// FindPrivateKeys is TODO
func (a *Age) FindPrivateKeys(ctx context.Context, keys ...string) ([]string, error) {
	return a.FindPublicKeys(ctx, keys...)
}
