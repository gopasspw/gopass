package action

import (
	"os"
	"path/filepath"

	"github.com/gopasspw/gopass/internal/tree"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/store/leaf"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/fsutil"
	"github.com/gopasspw/gopass/pkg/termio"
	"github.com/urfave/cli/v2"
)

// Fsck checks the store integrity
func (s *Action) Fsck(c *cli.Context) error {
	s.rem.Reset("fsck")

	ctx := ctxutil.WithGlobalFlags(c)
	if c.IsSet("decrypt") {
		ctx = leaf.WithFsckDecrypt(ctx, c.Bool("decrypt"))
	}
	out.Printf(ctx, "Checking store integrity ...")
	// make sure config is in the right place
	// we may have loaded it from one of the fallback locations
	if err := s.cfg.Save(); err != nil {
		return ExitError(ExitConfig, err, "failed to save config: %s", err)
	}

	// clean up any previous config locations
	oldCfg := filepath.Join(config.Homedir(), ".gopass.yml")
	if fsutil.IsFile(oldCfg) {
		if err := os.Remove(oldCfg); err != nil {
			out.Errorf(ctx, "Failed to remove old gopass config %s: %s", oldCfg, err)
		}
	}

	// display progress bar
	t, err := s.Store.Tree(ctx)
	if err != nil {
		return ExitError(ExitUnknown, err, "failed to list stores: %s", err)
	}

	pwList := t.List(tree.INF)

	bar := termio.NewProgressBar(int64(len(pwList)*2), ctxutil.IsHidden(ctx))
	ctx = ctxutil.WithProgressCallback(ctx, func() {
		bar.Inc()
	})
	ctx = out.AddPrefix(ctx, "\n")

	// the main work in done by the sub stores
	if err := s.Store.Fsck(ctx, c.Args().Get(0)); err != nil {
		return ExitError(ExitFsck, err, "fsck found errors: %s", err)
	}
	bar.Done()
	return nil
}
