package noop

import (
	"context"
	"fmt"

	"github.com/justwatchcom/gopass/pkg/backend"

	"github.com/blang/semver"
)

// Noop is a no-op git backend
type Noop struct{}

// New creates a new Noop object
func New() *Noop {
	return &Noop{}
}

// Add does nothing
func (g *Noop) Add(ctx context.Context, args ...string) error {
	return nil
}

// Commit does nothing
func (g *Noop) Commit(ctx context.Context, msg string) error {
	return nil
}

// Push does nothing
func (g *Noop) Push(ctx context.Context, origin, branch string) error {
	return nil
}

// Pull does nothing
func (g *Noop) Pull(ctx context.Context, origin, branch string) error {
	return nil
}

// Cmd does nothing
func (g *Noop) Cmd(ctx context.Context, name string, args ...string) error {
	return nil
}

// Init does nothing
func (g *Noop) Init(context.Context, string, string) error {
	return nil
}

// InitConfig does nothing
func (g *Noop) InitConfig(context.Context, string, string) error {
	return nil
}

// Version returns an empty version
func (g *Noop) Version(context.Context) semver.Version {
	return semver.Version{}
}

// Name returns noop
func (g *Noop) Name() string {
	return "noop"
}

// AddRemote does nothing
func (g *Noop) AddRemote(ctx context.Context, remote, url string) error {
	return nil
}

// RemoveRemote does nothing
func (g *Noop) RemoveRemote(ctx context.Context, remote string) error {
	return nil
}

// Revisions is not implemented
func (g *Noop) Revisions(context.Context, string) ([]backend.Revision, error) {
	return nil, fmt.Errorf("not yet implemented for %s", g.Name())
}

// GetRevision is not implemented
func (g *Noop) GetRevision(context.Context, string, string) ([]byte, error) {
	return []byte(""), nil
}
