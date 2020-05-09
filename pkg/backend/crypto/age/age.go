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
	"github.com/gopasspw/gopass/pkg/fsutil"
	"github.com/gopasspw/gopass/pkg/out"
	"github.com/gopasspw/gopass/pkg/termio"
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
		keyring: filepath.Join(ucd, "gopass", "age.age"),
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

	return a.encrypt(ctx, plaintext, args...)
}

func (a *Age) encrypt(ctx context.Context, plaintext []byte, args ...string) ([]byte, error) {
	buf := &bytes.Buffer{}

	cmd := exec.CommandContext(ctx, a.binary, args...)
	cmd.Stdin = bytes.NewReader(plaintext)
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr

	out.Debug(ctx, "age.encrypt: %s %+v", cmd.Path, cmd.Args)
	err := cmd.Run()
	return buf.Bytes(), err
}

func (a *Age) encryptFile(ctx context.Context, filename string, plaintext []byte) error {
	buf, err := a.encrypt(ctx, plaintext, "-p")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, buf, 0600)
}

// Decrypt will attempt to decrypt the given payload
func (a *Age) Decrypt(ctx context.Context, ciphertext []byte) ([]byte, error) {
	args := []string{}
	for _, k := range a.listPrivateKeyFiles(ctx) {
		if k == "native-keyring" {
			out.Debug(ctx, "age.Decrypt - decrypting native keyring for file decrypt")
			td, err := ioutil.TempDir("", "gpa")
			if err != nil {
				return nil, err
			}
			defer os.RemoveAll(td)
			fn := filepath.Join(td, "keys.txt")
			if err := a.decryptFileTo(ctx, a.keyring, fn); err != nil {
				return nil, err
			}
			k = fn
		}
		args = append(args, "--identity")
		args = append(args, k)
	}

	return a.decrypt(ctx, ciphertext, args...)
}

func (a *Age) decrypt(ctx context.Context, ciphertext []byte, args ...string) ([]byte, error) {
	args = append(args, "--decrypt")
	cmd := exec.CommandContext(ctx, a.binary, args...)
	cmd.Stdin = bytes.NewReader(ciphertext)
	cmd.Stderr = os.Stderr

	out.Debug(ctx, "age.Decrypt: %s %+v", cmd.Path, cmd.Args)
	return cmd.Output()
}

func (a *Age) decryptFile(ctx context.Context, filename string) ([]byte, error) {
	ciphertext, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return a.decrypt(ctx, ciphertext)
}

func (a *Age) decryptFileTo(ctx context.Context, src, dst string) error {
	buf, err := a.decryptFile(ctx, src)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(dst, buf, 0600)
}

// CreatePrivateKeyBatch will create a new native private key
func (a *Age) CreatePrivateKeyBatch(ctx context.Context, name, email, pw string) error {
	buf := &bytes.Buffer{}
	cmd := exec.CommandContext(ctx, "age-keygen")
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr

	// decrypt and prepend existing keyring content
	if fsutil.IsFile(a.keyring) {
		b2, err := a.decryptFile(ctx, a.keyring)
		if err != nil {
			return err
		}
		if _, err := buf.Write(b2); err != nil {
			return err
		}
	}

	// write gopass metadata of new entry
	fmt.Fprintf(buf, "# gopass-age-keypair\n")
	fmt.Fprintf(buf, "# Name: %s\n", name)
	fmt.Fprintf(buf, "# Email: %s\n", email)

	// create new keypair
	if err := cmd.Run(); err != nil {
		return err
	}

	out.Debug(ctx, "age.CreatePrivateKey: %s", buf.String())

	if err := os.MkdirAll(filepath.Dir(a.keyring), 0700); err != nil {
		return err
	}

	// encrypt final keyring
	return a.encryptFile(ctx, a.keyring, buf.Bytes())
}

// CreatePrivateKey is TODO
func (a *Age) CreatePrivateKey(ctx context.Context) error {
	out.Print(ctx, "Generating new Age keypair ...")
	name, err := termio.AskForString(ctx, "What is your name?", "")
	if err != nil {
		return err
	}

	email, err := termio.AskForString(ctx, "What is your email?", "")
	if err != nil {
		return err
	}

	return a.CreatePrivateKeyBatch(ctx, name, email, "unused")
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
	idSet := map[string]struct{}{}
	for _, v := range keys {
		idSet[v] = struct{}{}
	}
	ids := make([]string, 0, len(keys))
	for k := range idSet {
		ids = append(ids, k)
	}
	return ids
}

// ExportPublicKey is not implemented
func (a *Age) ExportPublicKey(ctx context.Context, id string) ([]byte, error) {
	return []byte(id), nil
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
	out.Debug(ctx, "age.getNativeKeypairs - decrypting keyring")
	buf, err := a.decryptFile(ctx, a.keyring)
	if err != nil {
		return nil, err
	}
	var pub string
	keys := map[string]string{}
	lines := strings.Split(string(buf), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "# public key: age1") {
			pub = strings.TrimPrefix(line, "# public key: ")
			continue
		}
		if strings.HasPrefix(line, "AGE-SECRET-KEY") {
			if pub != "" {
				keys[pub] = "native-keyring"
				pub = ""
			}
		}
	}
	return keys, nil
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
