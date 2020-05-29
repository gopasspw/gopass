package backend

import (
	"context"
	"sort"
	"time"

	"github.com/blang/semver"
	"github.com/gopasspw/gopass/internal/debug"
	"github.com/pkg/errors"
)

// RCSBackend is a remote-sync backend
type RCSBackend int

const (
	// Noop is a no-op mock backend
	Noop RCSBackend = iota
	// GitCLI is a git-cli based sync backend
	GitCLI
	// OnDiskRCS is the OnDisk storage backend in disguise as an RCS backend
	OnDiskRCS
)

func (s RCSBackend) String() string {
	return rcsNameFromBackend(s)
}

// RCS is a revision control backend
type RCS interface {
	Add(ctx context.Context, args ...string) error
	Commit(ctx context.Context, msg string) error
	Push(ctx context.Context, remote, location string) error
	Pull(ctx context.Context, remote, location string) error

	Name() string
	Version(ctx context.Context) semver.Version

	InitConfig(ctx context.Context, name, email string) error
	AddRemote(ctx context.Context, remote, location string) error
	RemoveRemote(ctx context.Context, remote string) error

	Revisions(ctx context.Context, name string) ([]Revision, error)
	GetRevision(ctx context.Context, name, revision string) ([]byte, error)

	Status(ctx context.Context) ([]byte, error)
	Compact(ctx context.Context) error
}

// Revision is a SCM revision
type Revision struct {
	Hash        string
	AuthorName  string
	AuthorEmail string
	Date        time.Time
	Subject     string
	Body        string
}

// RegisterRCS registers a new RCS backend with the backend registry.
func RegisterRCS(id RCSBackend, name string, loader RCSLoader) {
	rcsRegistry[id] = loader
	rcsNameToBackendMap[name] = id
	rcsBackendToNameMap[id] = name
}

// DetectRCS tried to detect the RCS backend being used
func DetectRCS(ctx context.Context, path string) (RCS, error) {
	if HasRCSBackend(ctx) {
		if be, found := rcsRegistry[GetRCSBackend(ctx)]; found {
			rcs, err := be.Open(ctx, path)
			if err == nil {
				return rcs, nil
			}
			rcs, err = be.InitRCS(ctx, path)
			if err == nil {
				return rcs, nil
			}
			return rcsRegistry[Noop].InitRCS(ctx, path)
		}
	}
	bes := make([]RCSBackend, 0, len(rcsRegistry))
	for id := range rcsRegistry {
		bes = append(bes, id)
	}
	sort.Slice(bes, func(i, j int) bool {
		return rcsRegistry[bes[i]].Priority() < rcsRegistry[bes[j]].Priority()
	})
	for _, id := range bes {
		be := rcsRegistry[id]
		debug.Log("Trying %s for %s", be, path)
		if err := be.Handles(path); err != nil {
			debug.Log("failed to use RCS %s for %s", id, path)
			continue
		}
		debug.Log("Using %s for %s", be, path)
		return be.Open(ctx, path)
	}
	debug.Log("No supported RCS found for %s. using NOOP", path)
	return rcsRegistry[Noop].InitRCS(ctx, path)
}

// CloneRCS clones an existing repository from a remote.
func CloneRCS(ctx context.Context, id RCSBackend, repo, path string) (RCS, error) {
	if be, found := rcsRegistry[id]; found {
		debug.Log("Cloning with %s", be.String())
		return be.Clone(ctx, repo, path)
	}
	return nil, errors.Wrapf(ErrNotFound, "unknown backend: %d", id)
}

// InitRCS initializes a new repository.
func InitRCS(ctx context.Context, id RCSBackend, path string) (RCS, error) {
	if be, found := rcsRegistry[id]; found {
		return be.InitRCS(ctx, path)
	}
	return nil, errors.Wrapf(ErrNotFound, "unknown backend: %d", id)
}
