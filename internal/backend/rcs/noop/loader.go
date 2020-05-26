package noop

import (
	"context"

	"github.com/gopasspw/gopass/internal/backend"
)

const (
	name = "noop"
)

func init() {
	backend.RegisterRCS(backend.Noop, name, &loader{})
}

type loader struct{}

// Open implements backend.RCSLoader
func (l loader) Open(ctx context.Context, path string) (backend.RCS, error) {
	return New(), nil
}

// Clone implements backend.RCSLoader
func (l loader) Clone(ctx context.Context, repo, path string) (backend.RCS, error) {
	return New(), nil
}

// Init implements backend.RCSLoader
func (l loader) InitRCS(ctx context.Context, path string) (backend.RCS, error) {
	return New(), nil
}

func (l loader) Handles(_ string) error {
	return nil
}

func (l loader) Priority() int {
	return 1000
}
func (l loader) String() string {
	return name
}
