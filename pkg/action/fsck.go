package action

import (
	"context"
	"os"
	"path/filepath"

	"github.com/gopasspw/gopass/pkg/config"
	"github.com/gopasspw/gopass/pkg/fsutil"
	"github.com/gopasspw/gopass/pkg/out"

	"github.com/urfave/cli"
)

// Fsck checks the store integrity
func (s *Action) Fsck(ctx context.Context, c *cli.Context) error {
	// make sure config is in the right place
	// we may have loaded it from one of the fallback locations
	if err := s.cfg.Save(); err != nil {
		return ExitError(ctx, ExitConfig, err, "failed to save config: %s", err)
	}
	// clean up any previous config locations
	oldCfg := filepath.Join(config.Homedir(), ".gopass.yml")
	if fsutil.IsFile(oldCfg) {
		if err := os.Remove(oldCfg); err != nil {
			out.Red(ctx, "Failed to remove old gopass config %s: %s", oldCfg, err)
		}
	}

	// the main work in done by the sub stores
	if err := s.Store.Fsck(ctx, c.Args().Get(0)); err != nil {
		return ExitError(ctx, ExitFsck, err, "fsck found errors: %s", err)
	}
	return nil
}
