package cryptfs

import (
	"context"
	"fmt"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/fsutil"
)

func init() {
	backend.StorageRegistry.Register(backend.CryptFS, "cryptfs", &loader{})
}

type loader struct{}

// New returns a new cryptfs storage backend.
func (l *loader) New(ctx context.Context, path string) (backend.Storage, error) {
	subID, err := getSubStorage(ctx)
	if err != nil {
		return nil, err
	}
	sub, err := backend.NewStorage(ctx, subID, path)
	if err != nil {
		return nil, err
	}
	return newCrypt(ctx, sub)
}

// Init initializes a new cryptfs storage backend.
func (l *loader) Init(ctx context.Context, path string) (backend.Storage, error) {
	subID, err := getSubStorage(ctx)
	if err != nil {
		return nil, err
	}
	sub, err := backend.InitStorage(ctx, subID, path)
	if err != nil {
		return nil, err
	}
	c, err := newCrypt(ctx, sub)
	if err != nil {
		return nil, err
	}
	if err := c.saveMappings(ctx); err != nil {
		out.Warningf(ctx, "Failed to save initial mapping: %s", err)
	}
	return c, nil
}

// Handles returns true if this backend handles the given path.
func (l *loader) Handles(ctx context.Context, path string) error {
	if fsutil.IsFile(path + "/" + mappingFile) {
		return nil
	}
	return fmt.Errorf("no mapping file found")
}

// String returns the name of this backend.
func (l *loader) String() string {
	return name
}

// Priority returns the priority of this backend.
func (l *loader) Priority() int {
	return 50
}

// Clone clones an existing repository and initializes the cryptfs backend.
func (l *loader) Clone(ctx context.Context, repo, path string) (backend.Storage, error) {
	subID, err := getSubStorage(ctx)
	if err != nil {
		return nil, err
	}
	subLoader, err := backend.StorageRegistry.Get(subID)
	if err != nil {
		return nil, err
	}
	sub, err := subLoader.Clone(ctx, repo, path)
	if err != nil {
		return nil, err
	}
	return newCrypt(ctx, sub)
}

func getSubStorage(ctx context.Context) (backend.StorageBackend, error) {
	subStoreName := config.String(ctx, "cryptfs.substorage")
	if subStoreName == "" {
		subStoreName = "gitfs"
	}
	id, err := backend.StorageRegistry.Backend(subStoreName)
	if err != nil {
		debug.Log("Failed to get backend ID for %q: %s", subStoreName, err)
	}
	return id, err
}
