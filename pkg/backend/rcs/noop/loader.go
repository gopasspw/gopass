package noop

import (
	"context"

	"github.com/gopasspw/gopass/pkg/backend"
)

const (
	name = "noop"
)

func Init() {
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
func (l loader) Init(ctx context.Context, path, username, email string) (backend.RCS, error) {
	return New(), nil
}

func (l loader) String() string {
	return name
}
