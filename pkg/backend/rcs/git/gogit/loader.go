package gogit

import (
	"context"

	"github.com/gopasspw/gopass/pkg/backend"
	"github.com/gopasspw/gopass/pkg/out"
)

func init() {
	backend.RegisterRCS(backend.GoGit, "gogit", &loader{})
}

type loader struct{}

// Open implements backend.RCSLoader
func (l loader) Open(ctx context.Context, path string) (backend.RCS, error) {
	out.Cyan(ctx, "WARNING: Using experimental RCS backend 'gogit' for '%s'", path)
	return Open(path)
}

// Clone implements backend.RCSLoader
func (l loader) Clone(ctx context.Context, repo, path string) (backend.RCS, error) {
	return Clone(ctx, repo, path)
}

// Init implements backend.RCSLoader
func (l loader) Init(ctx context.Context, path, username, email string) (backend.RCS, error) {
	return Init(ctx, path)
}
