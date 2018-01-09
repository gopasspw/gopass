package backend

import (
	"context"

	"github.com/blang/semver"
)

// SyncBackend is a remote-sync backend
type SyncBackend int

const (
	// GitMock is a no-op mock backend
	GitMock SyncBackend = iota
	// GitCLI is a git-cli based sync backend
	GitCLI
	// GoGit is an src-d/go-git.v4 based sync backend
	GoGit
)

// Sync is a sync backend
type Sync interface {
	Add(ctx context.Context, args ...string) error
	Commit(ctx context.Context, msg string) error
	Push(ctx context.Context, remote, location string) error
	Pull(ctx context.Context, remote, location string) error

	Name() string
	Version(ctx context.Context) semver.Version

	InitConfig(ctx context.Context, name, email string) error
	AddRemote(ctx context.Context, remote, location string) error
}
