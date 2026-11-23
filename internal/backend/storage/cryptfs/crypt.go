// Package cryptfs implements a filename encrypting storage backend.
package cryptfs

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/blang/semver/v4"
	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/backend/crypto/age"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/store"
	"github.com/gopasspw/gopass/pkg/debug"
)

const (
	name = "cryptfs"
	// mappingFile is the file that contains the name mapping.
	mappingFile = ".gopass-mapping"
)

// Crypt is a storage backend that encrypts filenames.
type Crypt struct {
	sub      backend.Storage
	crypto   *age.Age
	path     string
	mux      sync.RWMutex
	mappings map[string]string
}

// newCrypt creates a new cryptfs backend.
func newCrypt(ctx context.Context, sub backend.Storage) (*Crypt, error) {
	a, err := age.New(ctx, false, "")
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
	debug.Log("Loaded %d mappings", len(c.mappings))

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

// LinkTarget returns the target of a linked secret, if it can be determined.
func (c *Crypt) LinkTarget(_ context.Context, name string) (string, bool, error) {
	c.mux.RLock()
	defer c.mux.RUnlock()

	h, ok := c.mappings[name]
	if !ok {
		return "", false, nil
	}

	if h == c.hash(name) {
		return "", false, nil
	}

	for candidate, candidateHash := range c.mappings {
		if candidateHash == h && candidate != name && c.hash(candidate) == h {
			return candidate, true, nil
		}
	}

	for candidate, candidateHash := range c.mappings {
		if candidateHash == h && candidate != name {
			return candidate, true, nil
		}
	}

	return "", true, nil
}

// Version returns the version of the backend.
func (c *Crypt) Version(ctx context.Context) semver.Version {
	return semver.Version{Major: 1}
}

// Fsck performs a consistency check on the backend.
func (c *Crypt) Fsck(ctx context.Context) error {
	c.mux.RLock()
	defer c.mux.RUnlock()

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

	return c.sub.Fsck(ctx)
}

func (c *Crypt) Prune(ctx context.Context, prefix string) error {
	return c.sub.Prune(ctx, prefix)
}

// Link creates a symlink.
func (c *Crypt) Link(ctx context.Context, from, to string) error {
	c.mux.RLock()
	defer c.mux.RUnlock()

	h, ok := c.mappings[from]
	if !ok {
		return os.ErrNotExist
	}
	if _, ok := c.mappings[to]; ok {
		return fmt.Errorf("destination %s already exists", to)
	}

	c.mappings[to] = h

	return c.saveMappings(ctx)
}

// rcs methods.
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
	rel = strings.TrimSuffix(rel, ext)

	return filepath.ToSlash(rel), nil
}

func (c *Crypt) Add(ctx context.Context, files ...string) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	hashedFiles := make([]string, 0, len(files))
	for _, file := range files {
		name, err := c.pathToName(ctx, file)
		if err != nil {
			// not in our store?
			debug.Log("Failed to get name for %s: %s", file, err)

			name = file
		}

		debug.Log("Mapping file %s to name %s", file, name)

		h, ok := c.mappings[name]
		if !ok {
			// could be a directory or a path that git understands (like '.')
			// pass it through and hope for the best.
			// This is not perfect, but should work for the common cases.
			hashedFiles = append(hashedFiles, file)

			debug.Log("No mapping for %s found, passing through", name)

			continue
		}

		debug.Log("Mapping name %s to hash %s found", name, h)
		hashedFile := filepath.Join(c.path, h)
		hashedFiles = append(hashedFiles, hashedFile)
	}
	// always add the mapping file
	hashedFiles = append(hashedFiles, filepath.Join(c.path, mappingFile))

	debug.Log("Adding files to the git index: %+v", hashedFiles)

	return c.sub.Add(ctx, hashedFiles...)
}

func (c *Crypt) Commit(ctx context.Context, msg string) error {
	return c.sub.Commit(ctx, msg)
}

func (c *Crypt) TryAdd(ctx context.Context, files ...string) error {
	err := c.Add(ctx, files...)
	if err == nil {
		return nil
	}
	if errors.Is(err, store.ErrGitNotInit) {
		debug.Log("Git not initialized. Ignoring.")

		return nil
	}

	return err
}

// CheckUnstaged protects target ciphertext and any mapping that will be rewritten.
func (c *Crypt) CheckUnstaged(ctx context.Context, names ...string) error {
	stager, ok := c.sub.(backend.UncommittedFileStager)
	if !ok {
		return nil
	}
	c.mux.RLock()
	defer c.mux.RUnlock()

	if _, err := c.mappingSnapshot(ctx); err != nil {
		return err
	}
	hashes := make([]string, 0, len(names)+1)
	var newMapping bool
	for _, name := range names {
		h, found := c.mappings[name]
		if !found {
			h = c.hash(name)
			newMapping = true
		}
		if err := c.checkUniqueMapping(name, h); err != nil {
			return err
		}
		hashes = append(hashes, h)
	}
	if newMapping {
		hashes = append(hashes, mappingFile)
	}

	return stager.CheckUnstaged(ctx, hashes...)
}

// StageUncommitted recovers the ciphertext and any missing name mappings.
func (c *Crypt) StageUncommitted(ctx context.Context, names ...string) (bool, error) {
	planner, ok := c.sub.(backend.StagingPlanner)
	if !ok {
		return false, nil
	}
	reader, ok := c.sub.(backend.CommittedFileReader)
	if !ok {
		return false, fmt.Errorf("cannot verify committed cryptfs name mappings")
	}
	c.mux.RLock()
	defer c.mux.RUnlock()

	targets := make(map[string]bool, len(names))
	files := make([]backend.StagingFile, 0, len(names)+1)
	for _, name := range names {
		h, ok := c.mappings[name]
		if !ok {
			return false, fmt.Errorf("missing cryptfs mapping for %q", name)
		}
		if err := c.checkUniqueMapping(name, h); err != nil {
			return false, err
		}
		targets[name] = true
		files = append(files, backend.StagingFile{Name: h})
	}
	if len(files) == 0 {
		return false, nil
	}

	snapshot, err := c.mappingSnapshot(ctx)
	if err != nil {
		return false, err
	}
	committed, err := c.committedMappings(ctx, reader)
	if err != nil {
		return false, err
	}
	needsMapping, unrelated, err := c.mappingChanges(committed, targets)
	if err != nil {
		return false, err
	}
	plan, err := planner.PlanStaging(ctx, files...)
	if err != nil || (!plan.Pending && !needsMapping) {
		return false, err
	}
	// The mapping is a single encrypted file. A commit for a recovered
	// passkey must not sweep in pending names outside this import.
	if unrelated {
		return false, fmt.Errorf("cannot recover cryptfs name mappings with unrelated pending changes; resolve the mapping changes and retry")
	}
	files = append(files, backend.StagingFile{Name: mappingFile, Expected: snapshot})
	plan, err = planner.PlanStaging(ctx, files...)
	if err != nil {
		return false, err
	}
	if needsMapping && !plan.Pending {
		return false, fmt.Errorf("cryptfs name mappings differ from HEAD but could not be staged")
	}

	return plan.Stage(ctx)
}

func (c *Crypt) mappingChanges(committed map[string]string, targets map[string]bool) (bool, bool, error) {
	var needsMapping, unrelated bool
	for name, h := range c.mappings {
		if committed[name] != h {
			if targets[name] && committed[name] != "" {
				return false, false, fmt.Errorf("cannot recover a changed cryptfs mapping for %q; resolve the mapping and retry", name)
			}
			needsMapping = needsMapping || targets[name]
			unrelated = unrelated || !targets[name]
		}
	}
	for name := range committed {
		if _, present := c.mappings[name]; !present {
			unrelated = true

			break
		}
	}

	return needsMapping, unrelated, nil
}

func (c *Crypt) checkUniqueMapping(name, hash string) error {
	decoded, err := hex.DecodeString(hash)
	if err != nil || len(decoded) != sha256.Size || strings.ToLower(hash) != hash {
		return fmt.Errorf("invalid cryptfs ciphertext name for %q", name)
	}
	for other, h := range c.mappings {
		if strings.EqualFold(h, hash) && other != name {
			return fmt.Errorf("cryptfs names %q and %q share ciphertext; resolve the mapping and retry", name, other)
		}
	}

	return nil
}

// mappingSnapshot verifies what is on disk, not just the in-memory mapping.
func (c *Crypt) mappingSnapshot(ctx context.Context) ([]byte, error) {
	ciphertext, err := c.sub.Get(ctx, mappingFile)
	if errors.Is(err, os.ErrNotExist) && len(c.mappings) == 0 {
		return []byte{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read cryptfs name mappings: %w", err)
	}
	mappings, err := c.decodeMappings(ctx, ciphertext)
	if err != nil {
		return nil, err
	}
	if !maps.Equal(mappings, c.mappings) {
		return nil, fmt.Errorf("cryptfs name mappings changed on disk; reopen the store and retry")
	}

	return ciphertext, nil
}

func (c *Crypt) committedMappings(ctx context.Context, reader backend.CommittedFileReader) (map[string]string, error) {
	ciphertext, err := reader.CommittedFile(ctx, mappingFile)
	if errors.Is(err, os.ErrNotExist) {
		return map[string]string{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read committed cryptfs name mappings: %w", err)
	}

	return c.decodeMappings(ctx, ciphertext)
}

func (c *Crypt) decodeMappings(ctx context.Context, ciphertext []byte) (map[string]string, error) {
	plaintext, err := c.crypto.Decrypt(ctx, ciphertext)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt cryptfs name mappings: %w", err)
	}
	var mappings map[string]string
	if err := json.Unmarshal(plaintext, &mappings); err != nil {
		return nil, fmt.Errorf("failed to parse cryptfs name mappings: %w", err)
	}

	return mappings, nil
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
	c.mux.RLock()
	defer c.mux.RUnlock()

	h, ok := c.mappings[name]
	if !ok {
		return nil, os.ErrNotExist
	}

	return c.sub.Revisions(ctx, h)
}

func (c *Crypt) GetRevision(ctx context.Context, name, revision string) ([]byte, error) {
	c.mux.RLock()
	defer c.mux.RUnlock()

	h, ok := c.mappings[name]
	if !ok {
		return nil, os.ErrNotExist
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
	c.mux.RLock()
	defer c.mux.RUnlock()

	for k := range c.mappings {
		if strings.HasPrefix(k, name+"/") {
			return true
		}
	}

	return false
}

// List returns a list of all secrets.
func (c *Crypt) List(ctx context.Context, prefix string) ([]string, error) {
	c.mux.RLock()
	defer c.mux.RUnlock()

	list := make([]string, 0, len(c.mappings))

	if !strings.HasSuffix(prefix, "/") && prefix != "" {
		prefix += "/"
	}

	for k := range c.mappings {
		if !strings.HasPrefix(k, prefix) {
			continue
		}
		list = append(list, k)
	}
	sort.Strings(list)

	return list, nil
}

// Get returns the content of a secret.
func (c *Crypt) Get(ctx context.Context, name string) ([]byte, error) {
	c.mux.RLock()
	defer c.mux.RUnlock()

	h, ok := c.mappings[name]
	if !ok {
		if c.sub.Exists(ctx, name) {
			return c.sub.Get(ctx, name)
		}

		return nil, os.ErrNotExist
	}

	debug.Log("Reading content for %s from %s", name, h)

	return c.sub.Get(ctx, h)
}

// Set sets the content of a secret.
func (c *Crypt) Set(ctx context.Context, name string, value []byte) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	h, ok := c.mappings[name]
	if !ok {
		h = c.hash(name)
	}

	debug.Log("Writing content for %s to %s", name, h)
	if err := c.sub.Set(ctx, h, value); err != nil {
		return err
	}
	if ok {
		return nil
	}
	c.mappings[name] = h
	debug.Log("New mapping: %s -> %s", name, h)
	if err := c.saveMappings(ctx); err != nil {
		delete(c.mappings, name)

		return err
	}

	return nil
}

// Delete removes a secret.
func (c *Crypt) Delete(ctx context.Context, name string) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	h, ok := c.mappings[name]
	if !ok {
		return os.ErrNotExist
	}
	if err := c.sub.Delete(ctx, h); err != nil {
		return err
	}
	delete(c.mappings, name)

	return c.saveMappings(ctx)
}

// Exists returns true if a secret exists.
func (c *Crypt) Exists(ctx context.Context, name string) bool {
	if c.sub.Exists(ctx, name) {
		return true
	}

	c.mux.RLock()
	defer c.mux.RUnlock()

	_, ok := c.mappings[name]

	return ok
}

// Move moves a secret.
func (c *Crypt) Move(ctx context.Context, from, to string, del bool) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	fromH, ok := c.mappings[from]
	if !ok {
		return os.ErrNotExist
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
