package sub

import (
	"context"
	"fmt"

	"github.com/justwatchcom/gopass/pkg/backend"
	gpgcli "github.com/justwatchcom/gopass/pkg/backend/crypto/gpg/cli"
	gitcli "github.com/justwatchcom/gopass/pkg/backend/rcs/git/cli"
	"github.com/justwatchcom/gopass/pkg/backend/rcs/git/gogit"
	"github.com/justwatchcom/gopass/pkg/out"
)

func (s *Store) initRCSBackend(ctx context.Context) error {
	switch s.url.RCS {
	case backend.GoGit:
		out.Cyan(ctx, "WARNING: Using experimental RCS backend 'go-git'")
		git, err := gogit.Open(s.url.Path)
		if err != nil {
			out.Debug(ctx, "Failed to initialize RCS backend 'gogit': %s", err)
		} else {
			s.rcs = git
			out.Debug(ctx, "Using RCS Backend: go-git")
		}
	case backend.GitCLI:
		gpgBin, _ := gpgcli.Binary(ctx, "")
		git, err := gitcli.Open(s.url.Path, gpgBin)
		if err != nil {
			out.Debug(ctx, "Failed to initialize RCS backend 'gitcli': %s", err)
		} else {
			s.rcs = git
			out.Debug(ctx, "Using RCS Backend: gitcli")
		}
	case backend.Noop:
		// no-op
		out.Debug(ctx, "Using RCS Backend: noop")
	default:
		return fmt.Errorf("Unknown RCS Backend")
	}
	return nil
}
