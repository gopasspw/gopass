// Package cryptfs implements a filename encrypting storage backend.
package cryptfs

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"path/filepath"

	"fmt"
	"sort"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/backend/crypto/age"
	"github.com/gopasspw/gopass/internal/out"
)

const (
	name = "cryptfs"
	// mappingFile is the file that contains the name mapping.
	mappingFile = ".gopass-mapping.age"
)

// Crypt is a storage backend that encrypts filenames.
type Crypt struct {
	sub      backend.Storage
	crypto   *age.Age
	path     string
	mappings map[string]string
}

// newCrypt creates a new cryptfs backend.
func newCrypt(ctx context.Context, sub backend.Storage) (*Crypt, error) {
	a, err := age.New(ctx, "")
	if err != nil {
		return nil, err
	}

	c := &Crypt{
		sub:      sub,
		crypto:   a,
		path:     sub.Path(),
		mappings: make(map[string]string),
	}

	if err := c.loadMappings(ctx); err != nil {
		out.Warningf(ctx, "Failed to load mappings: %s", err)
	}

	return c, nil
}

func (c *Crypt) hash(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

func (c *Crypt) loadMappings(ctx context.Context) error {
	if !c.sub.Exists(ctx, mappingFile) {
		return nil
	}

	ciphertext, err := c.sub.Get(ctx, mappingFile)
	if err != nil {
		return err
	}

	plaintext, err := c.crypto.Decrypt(ctx, ciphertext)
	if err != nil {
		return err
	}

	return json.Unmarshal(plaintext, &c.mappings)
}

func (c *Crypt) saveMappings(ctx context.Context) error {
	plaintext, err := json.MarshalIndent(c.mappings, "", "  ")
	if err != nil {
		return err
	}

	recipientsFile := c.crypto.IDFile()
	content, err := c.sub.Get(ctx, recipientsFile)
	if err != nil {
		return fmt.Errorf("failed to read recipients file %s: %w", recipientsFile, err)
	}

	recipients := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(recipients) == 0 || (len(recipients) == 1 && recipients[0] == "") {
		return fmt.Errorf("no recipients found in %s", recipientsFile)
	}

	ciphertext, err := c.crypto.Encrypt(ctx, plaintext, recipients)
	if err != nil {
		return err
	}

	return c.sub.Set(ctx, mappingFile, ciphertext)
}

// String implements fmt.Stringer.
func (c *Crypt) String() string {
	return name
}

// Name returns the name of the backend.
func (c *Crypt) Name() string {
	return name
}

// Path returns the path of the backend.
func (c *Crypt) Path() string {
	return c.path
}

// Version returns the version of the backend.
func (c *Crypt) Version(ctx context.Context) semver.Version {
	return semver.Version{Major: 1}
}

// Fsck performs a consistency check on the backend.
func (c *Crypt) Fsck(ctx context.Context) error {
	return c.sub.Fsck(ctx)
}

func (c *Crypt) Prune(ctx context.Context, prefix string) error {
	// list all files in sub-storage
	allFiles, err := c.sub.List(ctx, "")
	if err != nil {
		return err
	}

	// create a set of mapped hashes
	mappedHashes := make(map[string]struct{})
	for _, h := range c.mappings {
		mappedHashes[h] = struct{}{}
	}

	// find orphans and delete them
	for _, file := range allFiles {
		if file == mappingFile {
			continue
		}
		if c.sub.IsDir(ctx, file) {
			continue
		}
		if _, ok := mappedHashes[file]; !ok {
			if err := c.sub.Delete(ctx, file); err != nil {
				out.Warningf(ctx, "Failed to prune orphaned file %s: %s", file, err)
			}
		}
	}
	return c.sub.Prune(ctx, prefix)
}

// Link creates a symlink.
func (c *Crypt) Link(ctx context.Context, from, to string) error {
	// not yet supported
	return backend.ErrNotSupported
}

// rcs methods
func (c *Crypt) getCryptoExt(ctx context.Context) string {
	cryptoID := backend.GetCryptoBackend(ctx)
	loader, err := backend.CryptoRegistry.Get(cryptoID)
	if err != nil {
		// fallback to gpg
		return ".gpg"
	}
	crypto, err := loader.New(ctx)
	if err != nil {
		// fallback to gpg
		return ".gpg"
	}
	return crypto.Ext()
}

func (c *Crypt) pathToName(ctx context.Context, p string) (string, error) {
	rel, err := filepath.Rel(c.path, p)
	if err != nil {
		return "", fmt.Errorf("failed to get relative path for %s: %w", p, err)
	}
	ext := c.getCryptoExt(ctx)
	if strings.HasSuffix(rel, ext) {
		rel = strings.TrimSuffix(rel, ext)
	}
	return filepath.ToSlash(rel), nil
}

func (c *Crypt) Add(ctx context.Context, files ...string) error {
	hashedFiles := make([]string, 0, len(files))
	for _, file := range files {
		name, err := c.pathToName(ctx, file)
		if err != nil {
			// not in our store? pass through.
			hashedFiles = append(hashedFiles, file)
			continue
		}

		h, ok := c.mappings[name]
		if !ok {
			// could be a directory or a path that git understands (like '.')
			// pass it through and hope for the best.
			// This is not perfect, but should work for the common cases.
			hashedFiles = append(hashedFiles, file)
			continue
		}
		hashedFile := filepath.Join(c.path, h)
		hashedFiles = append(hashedFiles, hashedFile)
	}
	// always add the mapping file
	hashedFiles = append(hashedFiles, filepath.Join(c.path, mappingFile))

	return c.sub.Add(ctx, hashedFiles...)
}

func (c *Crypt) Commit(ctx context.Context, msg string) error {
	return c.sub.Commit(ctx, msg)
}

func (c *Crypt) TryAdd(ctx context.Context, files ...string) error {
	return c.sub.TryAdd(ctx, files...)
}

func (c *Crypt) TryCommit(ctx context.Context, msg string) error {
	return c.sub.TryCommit(ctx, msg)
}

func (c *Crypt) Push(ctx context.Context, remote, branch string) error {
	return c.sub.Push(ctx, remote, branch)
}

func (c *Crypt) Pull(ctx context.Context, remote, branch string) error {
	return c.sub.Pull(ctx, remote, branch)
}

func (c *Crypt) TryPush(ctx context.Context, remote, branch string) error {
	return c.sub.TryPush(ctx, remote, branch)
}

func (c *Crypt) Revisions(ctx context.Context, name string) ([]backend.Revision, error) {
	h, ok := c.mappings[name]
	if !ok {
		return nil, backend.ErrNotFound
	}
	return c.sub.Revisions(ctx, h)
}

func (c *Crypt) GetRevision(ctx context.Context, name, revision string) ([]byte, error) {
	h, ok := c.mappings[name]
	if !ok {
		return nil, backend.ErrNotFound
	}
	return c.sub.GetRevision(ctx, h, revision)
}

func (c *Crypt) Status(ctx context.Context) ([]byte, error) {
	return c.sub.Status(ctx)
}

func (c *Crypt) Compact(ctx context.Context) error {
	return c.sub.Compact(ctx)
}

func (c *Crypt) InitConfig(ctx context.Context, name, email string) error {
	return c.sub.InitConfig(ctx, name, email)
}

func (c *Crypt) AddRemote(ctx context.Context, remote, url string) error {
	return c.sub.AddRemote(ctx, remote, url)
}

func (c *Crypt) RemoveRemote(ctx context.Context, remote string) error {
	return c.sub.RemoveRemote(ctx, remote)
}

// IsDir returns true if the given path is a directory.
func (c *Crypt) IsDir(ctx context.Context, name string) bool {
	for k := range c.mappings {
		if strings.HasPrefix(k, name+"/") {
			return true
		}
	}
	return false
}

// List returns a list of all secrets.
func (c *Crypt) List(ctx context.Context, prefix string) ([]string, error) {
	var list []string
	seen := make(map[string]struct{})

	if !strings.HasSuffix(prefix, "/") && prefix != "" {
		prefix += "/"
	}

	for k := range c.mappings {
		if !strings.HasPrefix(k, prefix) {
			continue
		}
		// remove prefix
		sub := strings.TrimPrefix(k, prefix)
		// take the first path component
		parts := strings.SplitN(sub, "/", 2)
		entry := parts[0]
		if len(parts) > 1 {
			entry += "/"
		}
		if _, ok := seen[entry]; !ok {
			list = append(list, entry)
			seen[entry] = struct{}{}
		}
	}
	sort.Strings(list)
	return list, nil
}

// Get returns the content of a secret.
func (c *Crypt) Get(ctx context.Context, name string) ([]byte, error) {
	h, ok := c.mappings[name]
	if !ok {
		return nil, backend.ErrNotFound
	}
	return c.sub.Get(ctx, h)
}

// Set sets the content of a secret.
func (c *Crypt) Set(ctx context.Context, name string, value []byte) error {
	h, ok := c.mappings[name]
	if !ok {
		h = c.hash(name)
		c.mappings[name] = h
	}
	if err := c.sub.Set(ctx, h, value); err != nil {
		return err
	}
	return c.saveMappings(ctx)
}

// Delete removes a secret.
func (c *Crypt) Delete(ctx context.Context, name string) error {
	h, ok := c.mappings[name]
	if !ok {
		return backend.ErrNotFound
	}
	if err := c.sub.Delete(ctx, h); err != nil {
		return err
	}
	delete(c.mappings, name)
	return c.saveMappings(ctx)
}

// Exists returns true if a secret exists.
func (c *Crypt) Exists(ctx context.Context, name string) bool {
	_, ok := c.mappings[name]
	return ok
}

// Move moves a secret.
func (c *Crypt) Move(ctx context.Context, from, to string, del bool) error {
	fromH, ok := c.mappings[from]
	if !ok {
		return backend.ErrNotFound
	}
	if _, ok := c.mappings[to]; ok {
		return fmt.Errorf("destination %s already exists", to)
	}

	// get content
	content, err := c.sub.Get(ctx, fromH)
	if err != nil {
		return err
	}

	// set new
	toH := c.hash(to)
	if err := c.sub.Set(ctx, toH, content); err != nil {
		return err
	}

	// update mapping
	delete(c.mappings, from)
	c.mappings[to] = toH
	if err := c.saveMappings(ctx); err != nil {
		// try to rollback
		c.mappings[from] = fromH
		delete(c.mappings, to)
		return err
	}

	// delete old
	if err := c.sub.Delete(ctx, fromH); err != nil {
		// this is not ideal, we have two copies now.
		// Fsck should detect this.
		out.Warningf(ctx, "Failed to delete old file %s after move: %s", fromH, err)
	}

	return nil
}
