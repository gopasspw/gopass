package action

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/store/sub"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/fsutil"
	"github.com/muesli/goprogressbar"
	"github.com/urfave/cli/v2"
)

// Fsck checks the store integrity
func (s *Action) Fsck(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	if c.IsSet("decrypt") {
		ctx = sub.WithFsckDecrypt(ctx, c.Bool("decrypt"))
	}
	out.Print(ctx, "Checking store integrity ...")
	// make sure config is in the right place
	// we may have loaded it from one of the fallback locations
	if err := s.cfg.Save(); err != nil {
		return ExitError(ctx, ExitConfig, err, "failed to save config: %s", err)
	}

	// clean up any previous config locations
	oldCfg := filepath.Join(config.Homedir(), ".gopass.yml")
	if fsutil.IsFile(oldCfg) {
		if err := os.Remove(oldCfg); err != nil {
			out.Error(ctx, "Failed to remove old gopass config %s: %s", oldCfg, err)
		}
	}

	// display progress bar
	t, err := s.Store.Tree(ctx)
	if err != nil {
		return ExitError(ctx, ExitUnknown, err, "failed to list stores: %s", err)
	}

	pwList := t.List(0)

	bar := &goprogressbar.ProgressBar{
		Total: int64(len(pwList) * 2),
		Width: 120,
	}
	if out.IsHidden(ctx) {
		old := goprogressbar.Stdout
		goprogressbar.Stdout = ioutil.Discard
		defer func() {
			goprogressbar.Stdout = old
		}()
	}

	ctx = ctxutil.WithProgressCallback(ctx, func() {
		bar.Current++
		bar.Text = fmt.Sprintf("%d of %d objects checked", bar.Current, bar.Total)
		bar.LazyPrint()
	})
	ctx = out.AddPrefix(ctx, "\n")

	// the main work in done by the sub stores
	if err := s.Store.Fsck(ctx, c.Args().Get(0)); err != nil {
		return ExitError(ctx, ExitFsck, err, "fsck found errors: %s", err)
	}
	out.Print(ctx, "")
	return nil
}
