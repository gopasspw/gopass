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

	TryAdd(ctx context.Context, args ...string) error
	TryCommit(ctx context.Context, msg string) error
	TryPush(ctx context.Context, remote, location string) error

	InitConfig(ctx context.Context, name, email string) error
	AddRemote(ctx context.Context, remote, location string) error
	RemoveRemote(ctx context.Context, remote string) error

	Revisions(ctx context.Context, name string) ([]Revision, error)
	GetRevision(ctx context.Context, name, revision string) ([]byte, error)

	Status(ctx context.Context) ([]byte, error)
	Compact(ctx context.Context) error
}

// UncommittedFileStager is an optional capability for resuming a write that
// reached the worktree or index but was not committed. StageUncommitted reports
// whether the given files need a commit. It must leave clean files alone and
// refuse to replace an index entry that has diverged from the worktree.
type UncommittedFileStager interface {
	// CheckUnstaged requires that a file to be replaced has no staged version
	// different from HEAD. The worktree file need not exist yet.
	CheckUnstaged(ctx context.Context, names ...string) error
	StageUncommitted(ctx context.Context, names ...string) (bool, error)
}

// StagingFile names an exact path. If Expected is non-nil, it must match the
// current worktree bytes (used for shared metadata checked by a caller).
type StagingFile struct {
	Name     string
	Expected []byte
}

// StagingPlanner lets a wrapper check all files before changing the index.
type StagingPlanner interface {
	PlanStaging(ctx context.Context, files ...StagingFile) (StagingPlan, error)
}

// StagingPlan has no side effects until Apply is called. Pending includes
// files already staged with the same contents as the worktree.
type StagingPlan struct {
	Pending bool
	Apply   func(context.Context) error
}

// StagingFiles converts a list of exact paths to staging requests.
func StagingFiles(names ...string) []StagingFile {
	files := make([]StagingFile, len(names))
	for i, name := range names {
		files[i].Name = name
	}

	return files
}

// Stage applies a plan if any files need to be added.
func (p StagingPlan) Stage(ctx context.Context) (bool, error) {
	if p.Apply != nil {
		if err := p.Apply(ctx); err != nil {
			return false, err
		}
	}

	return p.Pending, nil
}

// CommittedFileReader reads a file from HEAD, returning os.ErrNotExist if the
// file or HEAD does not exist.
type CommittedFileReader interface {
	CommittedFile(ctx context.Context, name string) ([]byte, error)
}

// ConflictingIndexError reports a file the user must resolve before recovery.
func ConflictingIndexError(name string) error {
	return fmt.Errorf("cannot recover %q: Git index and worktree diverge; resolve the staged changes and retry", name)
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
