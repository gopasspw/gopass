package action

import (
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/fsutil"
	"github.com/urfave/cli"
)

// Fsck checks the store integrity
func (s *Action) Fsck(c *cli.Context) error {
	check := c.Bool("check")
	force := c.Bool("force")
	if check {
		force = false
	}
	// make sure config is in the right place
	// we may have loaded it from one of the fallback locations
	if err := s.Store.Config().Save(); err != nil {
		return err
	}
	// clean up any previous config locations
	oldCfg := filepath.Join(os.Getenv("HOME"), ".gopass.yml")
	if fsutil.IsFile(oldCfg) {
		if err := os.Remove(oldCfg); err != nil {
			color.Red("Failed to remove old gopass config %s: %s", oldCfg, err)
		}
	}
	_, err := s.Store.Fsck("", check, force)
	return err
}
