package action

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/gopasspw/gopass/internal/action/exit"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/updater"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/protect"
	"github.com/urfave/cli/v2"
)

// Version prints the gopass version.
func (s *Action) Version(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	version := make(chan string, 1)
	go s.checkVersion(ctx, version)

	cli.VersionPrinter(c)

	select {
	case vi := <-version:
		if vi != "" {
			fmt.Fprintln(stdout, vi)
		}
	case <-time.After(2 * time.Second):
		out.Errorf(ctx, "Version check timed out")
	case <-ctx.Done():
		return exit.Error(exit.Aborted, nil, "user aborted")
	}

	return nil
}

func (s *Action) checkVersion(ctx context.Context, u chan string) {
	msg := ""
	defer func() {
		u <- msg
	}()

	if disabled := os.Getenv("CHECKPOINT_DISABLE"); disabled != "" {
		debug.Log("remote version check disabled by CHECKPOINT_DISABLE")

		return
	}

	if cfg := config.FromContext(ctx); cfg.IsSet("updater.check") && !cfg.GetBool("updater.check") {
		debug.Log("remote version check disabled by updater.check = false")

		return
	}

	// force checking for updates, mainly for testing.
	force := os.Getenv("GOPASS_FORCE_CHECK") != ""

	if !force && strings.HasSuffix(s.version.String(), "+HEAD") {
		// chan not check version against HEAD.
		debug.Log("remote version check disabled for dev version")

		return
	}

	if !force && protect.ProtectEnabled {
		// chan not check version
		// against pledge(2)'d OpenBSD.
		debug.Log("remote version check disabled for pledge(2)'d version")

		return
	}

	r, err := updater.FetchLatestRelease(ctx)
	if err != nil {
		msg = color.RedString("\nError checking latest version: %s", err)

		return
	}

	if s.version.GTE(r.Version) {
		_ = s.rem.Reset("update")
		debug.Log("gopass is up-to-date (local: %q, GitHub: %q)", s.version, r.Version)

		return
	}

	notice := fmt.Sprintf("\nYour version (%s) of gopass is out of date!\nThe latest version is %s.\n", s.version, r.Version.String())
	notice += "You can update by downloading from https://www.gopass.pw/#install"
	if err := updater.IsUpdateable(ctx); err == nil {
		notice += " by running 'gopass update'"
	}
	notice += " or via your package manager"
	msg = color.YellowString(notice)
}
