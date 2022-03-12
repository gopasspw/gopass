package backend

import (
	"context"
	"fmt"
	"time"

	"github.com/gopasspw/gopass/pkg/debug"
)

// rcs is a revision control backend.
type rcs interface {
	Add(ctx context.Context, args ...string) error
	Commit(ctx context.Context, msg string) error
	Push(ctx context.Context, remote, location string) error
	Pull(ctx context.Context, remote, location string) error

	InitConfig(ctx context.Context, name, email string) error
	AddRemote(ctx context.Context, remote, location string) error
	RemoveRemote(ctx context.Context, remote string) error

	Revisions(ctx context.Context, name string) ([]Revision, error)
	GetRevision(ctx context.Context, name, revision string) ([]byte, error)

	Status(ctx context.Context) ([]byte, error)
	Compact(ctx context.Context) error
}

// Revision is a SCM revision.
type Revision struct {
	Hash        string
	AuthorName  string
	AuthorEmail string
	Date        time.Time
	Subject     string
	Body        string
}

// Revisions implements the sort interface.
type Revisions []Revision

func (r Revisions) Len() int {
	return len(r)
}

func (r Revisions) Less(i, j int) bool {
	return r[i].Date.After(r[j].Date)
}

func (r Revisions) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

// Clone clones an existing repository from a remote.
func Clone(ctx context.Context, id StorageBackend, repo, path string) (Storage, error) {
	if be, err := StorageRegistry.Get(id); err == nil {
		debug.Log("Cloning with %s", be.String())

		return be.Clone(ctx, repo, path)
	}

	return nil, fmt.Errorf("unknown backend %d: %w", id, ErrNotFound)
}
