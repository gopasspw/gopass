package action

import (
	"context"
	"os"
	"path/filepath"

	"github.com/justwatchcom/gopass/config"
	"github.com/justwatchcom/gopass/store/sub"
	"github.com/justwatchcom/gopass/utils/fsutil"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/urfave/cli"
)

// Fsck checks the store integrity
func (s *Action) Fsck(ctx context.Context, c *cli.Context) error {
	if c.IsSet("check") {
		ctx = sub.WithFsckCheck(ctx, c.Bool("check"))
	}
	if c.IsSet("force") {
		ctx = sub.WithFsckForce(ctx, c.Bool("force"))
	}
	// make sure config is in the right place
	// we may have loaded it from one of the fallback locations
	if err := s.cfg.Save(); err != nil {
		return exitError(ctx, ExitConfig, err, "failed to save config: %s", err)
	}
	// clean up any previous config locations
	oldCfg := filepath.Join(config.Homedir(), ".gopass.yml")
	if fsutil.IsFile(oldCfg) {
		if err := os.Remove(oldCfg); err != nil {
			out.Red(ctx, "Failed to remove old gopass config %s: %s", oldCfg, err)
		}
	}

	if _, err := s.Store.Fsck(ctx, ""); err != nil {
		return exitError(ctx, ExitFsck, err, "fsck found errors")
	}
	return nil
}
