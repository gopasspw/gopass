package action

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/config"
	"github.com/justwatchcom/gopass/store/sub"
	"github.com/justwatchcom/gopass/utils/fsutil"
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
		return s.exitError(ctx, ExitConfig, err, "failed to save config: %s", err)
	}
	// clean up any previous config locations
	oldCfg := filepath.Join(config.Homedir(), ".gopass.yml")
	if fsutil.IsFile(oldCfg) {
		if err := os.Remove(oldCfg); err != nil {
			fmt.Println(color.RedString("Failed to remove old gopass config %s: %s", oldCfg, err))
		}
	}

	if _, err := s.Store.Fsck(ctx, ""); err != nil {
		return s.exitError(ctx, ExitFsck, err, "fsck found errors")
	}
	return nil
}
