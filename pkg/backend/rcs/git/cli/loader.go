package cli

import (
	"context"

	"github.com/gopasspw/gopass/pkg/backend"
	gpgcli "github.com/gopasspw/gopass/pkg/backend/crypto/gpg/cli"
)

const (
	name = "gitcli"
)

func Init() {
	backend.RegisterRCS(backend.GitCLI, name, &loader{})
}

type loader struct{}

// Open implements backend.RCSLoader
func (l loader) Open(ctx context.Context, path string) (backend.RCS, error) {
	gpgBin, _ := gpgcli.Binary(ctx, "")
	return Open(path, gpgBin)
}

// Clone implements backend.RCSLoader
func (l loader) Clone(ctx context.Context, repo, path string) (backend.RCS, error) {
	return Clone(ctx, repo, path)
}

// Init implements backend.RCSLoader
func (l loader) Init(ctx context.Context, path, username, email string) (backend.RCS, error) {
	return Init(ctx, path, username, email)
}

func (l loader) String() string {
	return name
}
