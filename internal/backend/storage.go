package backend

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/debug"
)

// ErrNotSupported is returned by backends for unsupported calls.
var ErrNotSupported = fmt.Errorf("not supported")

// StorageBackend is a type of storage backend.
type StorageBackend int

const (
	// FS is a filesystem-backed storage.
	FS StorageBackend = iota
	// GitFS is a filesystem-backed storage with Git.
	GitFS
	// FossilFS is a filesystem-backed storage with Fossil.
	FossilFS
	// JJ is a filesystem-backed storage with Jujutsu.
	JJFS
	// CryptFS is a filename encrypting storage.
	CryptFS
	// GoGit is a filesystem-backed storage with Git using the pure-Go go-git
	// implementation instead of the git binary.
	GoGit
)

func (s StorageBackend) String() string {
	if be, err := StorageRegistry.BackendName(s); err == nil {
		return be
	}

	return ""
}

// envGitBackend returns the raw value of the GOPASS_GIT_BACKEND environment
// variable. It is a thin wrapper so the docs lint test (TestEnvVarsInDocs) can
// see the os.Getenv("GOPASS_GIT_BACKEND") call.
func envGitBackend() string {
	return os.Getenv("GOPASS_GIT_BACKEND")
}

// GitBackendFromEnv returns the git backend named by the GOPASS_GIT_BACKEND
// environment variable, if any. It accepts "gogit" for the pure-Go backend and
// "gitfs"/"cli"/"cmd" for the git CLI backend.
func GitBackendFromEnv() (StorageBackend, bool) {
	switch strings.ToLower(strings.TrimSpace(envGitBackend())) {
	case "gogit":
		return GoGit, true
	case "gitfs", "cli", "cmd":
		return GitFS, true
	default:
		return FS, false
	}
}

// ResolveStorageBackend downgrades a requested GitFS backend to the pure-Go
// go-git backend when no git binary is available. This keeps a .git-backed
// store versioned on git-less systems instead of degrading to the unversioned
// fs backend. Any other requested backend is returned unchanged.
func ResolveStorageBackend(ctx context.Context, requested StorageBackend) StorageBackend {
	if requested != GitFS {
		return requested
	}

	if _, err := exec.LookPath("git"); err != nil {
		if _, rerr := StorageRegistry.Get(GoGit); rerr == nil {
			debug.Log("git binary not found; falling back to pure-Go go-git backend")
			out.Warningf(ctx, "git binary not found in PATH, falling back to the pure-Go go-git backend")

			return GoGit
		}
	}

	return requested
}

// Storage is an storage backend.
type Storage interface {
	fmt.Stringer
	rcs
	Get(ctx context.Context, name string) ([]byte, error)
	Set(ctx context.Context, name string, value []byte) error
	Delete(ctx context.Context, name string) error
	Exists(ctx context.Context, name string) bool
	Move(ctx context.Context, from, to string, del bool) error
	List(ctx context.Context, prefix string) ([]string, error)
	IsDir(ctx context.Context, name string) bool
	Prune(ctx context.Context, prefix string) error
	Link(ctx context.Context, from, to string) error

	Name() string
	Path() string
	Version(context.Context) semver.Version
	Fsck(context.Context) error
}

// DetectStorage tries to detect the storage backend being used.
func DetectStorage(ctx context.Context, path string) (Storage, error) {
	// The call to HasStorageBackend is important since GetStorageBackend will always return FS
	// if nothing is found in the context.
	if HasStorageBackend(ctx) {
		if st, err := detectRequestedStorage(ctx, path); err == nil {
			return st, nil
		}
	}

	// Check if a backend is explicitly selected via GOPASS_GIT_BACKEND. This
	// takes precedence over the persisted storage.backend config value.
	if envBE, ok := GitBackendFromEnv(); ok {
		if st, err := detectStorageByBackend(ctx, ResolveStorageBackend(ctx, envBE), path); err == nil {
			return st, nil
		}
	}

	// Check if a backend is explicitly configured via the config file.
	if name := config.String(ctx, "storage.backend"); name != "" {
		if key, err := StorageRegistry.Backend(name); err == nil {
			if st, err := detectStorageByBackend(ctx, ResolveStorageBackend(ctx, key), path); err == nil {
				return st, nil
			}
		}
	}

	// Nothing requested in the context. Try to detect the backend.
	for _, be := range StorageRegistry.Prioritized() {
		debug.V(1).Log("Trying storage backend %q for %q", be, path)
		if err := be.Handles(ctx, path); err != nil {
			debug.Log("failed to use %s for %s: %s", be, path, err)

			continue
		}
		debug.Log("Detected storage backend %q for %q", be, path)

		return be.New(ctx, path)
	}

	// fallback to FS
	be, err := StorageRegistry.Get(FS)
	if err != nil {
		return nil, err
	}
	debug.Log("Using default fallback %q for %q", be, path)

	return be.Init(ctx, path)
}

// detectRequestedStorage tries the backend explicitly set in the context,
// falling back to FS when it cannot be opened. A requested GitFS backend is
// downgraded to the pure-Go go-git backend when no git binary is available.
func detectRequestedStorage(ctx context.Context, path string) (Storage, error) {
	requested := ResolveStorageBackend(ctx, GetStorageBackend(ctx))
	be, err := StorageRegistry.Get(requested)
	if err != nil {
		return nil, err
	}

	debug.V(1).Log("Trying requested storage backend %q for %q", be, path)
	if st, err := be.New(ctx, path); err == nil {
		debug.Log("Successfully loaded requested storage backend %q for %q", be, path)

		return st, nil
	} else {
		debug.Log("Failed to use requested storage backend %q for %s: %q", be, path, err)
	}

	// fallback to FS
	fsBE, err := StorageRegistry.Get(FS)
	if err != nil {
		return nil, err
	}
	debug.Log("Using fallback %q for %q", fsBE, path)

	return fsBE.Init(ctx, path)
}

// detectStorageByBackend looks up a storage backend by key and tries to open path with it.
func detectStorageByBackend(ctx context.Context, key StorageBackend, path string) (Storage, error) {
	be, err := StorageRegistry.Get(key)
	if err != nil {
		return nil, err
	}

	st, err := be.New(ctx, path)
	if err != nil {
		debug.Log("Failed to use configured storage backend %q for %q: %s", key, path, err)

		return nil, err
	}

	return st, nil
}

// NewStorage initializes an existing storage backend.
func NewStorage(ctx context.Context, id StorageBackend, path string) (Storage, error) {
	if be, err := StorageRegistry.Get(id); err == nil {
		debug.Log("Using storage backend %q for %q", be, path)

		return be.New(ctx, path)
	}

	return nil, fmt.Errorf("unknown backend %q: %w", path, ErrNotFound)
}

// InitStorage initilizes a new storage location.
func InitStorage(ctx context.Context, id StorageBackend, path string) (Storage, error) {
	if be, err := StorageRegistry.Get(id); err == nil {
		debug.Log("Using %s for %s", be, path)

		return be.Init(ctx, path)
	}

	return nil, fmt.Errorf("unknown backend %q: %w", path, ErrNotFound)
}
