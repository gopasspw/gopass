package mock

import (
	"context"

	"github.com/blang/semver"
)

// Git is a no-op git backend
type Git struct{}

// New creates a new Git object
func New() *Git {
	return &Git{}
}

// Add does nothing
func (g *Git) Add(ctx context.Context, args ...string) error {
	return nil
}

// Commit does nothing
func (g *Git) Commit(ctx context.Context, msg string) error {
	return nil
}

// Push does nothing
func (g *Git) Push(ctx context.Context, origin, branch string) error {
	return nil
}

// Pull does nothing
func (g *Git) Pull(ctx context.Context, origin, branch string) error {
	return nil
}

// Cmd does nothing
func (g *Git) Cmd(ctx context.Context, name string, args ...string) error {
	return nil
}

// Init does nothing
func (g *Git) Init(context.Context, string, string) error {
	return nil
}

// InitConfig does nothing
func (g *Git) InitConfig(context.Context, string, string) error {
	return nil
}

// Version returns an empty version
func (g *Git) Version(context.Context) semver.Version {
	return semver.Version{}
}

// Name returns git-mock
func (g *Git) Name() string {
	return "git-mock"
}

// AddRemote does nothing
func (g *Git) AddRemote(ctx context.Context, remote, url string) error {
	return nil
}
