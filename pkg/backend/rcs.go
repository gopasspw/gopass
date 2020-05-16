package backend

import (
	"context"
	"time"

	"github.com/blang/semver"
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
