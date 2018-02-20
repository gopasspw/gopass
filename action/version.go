package action

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dominikschulz/github-releases/ghrel"
	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/justwatchcom/gopass/utils/protect"
	"github.com/urfave/cli"
)

const (
	gitHubOrg  = "justwatchcom"
	gitHubRepo = "gopass"
)

// Version prints the gopass version
func (s *Action) Version(ctx context.Context, c *cli.Context) error {
	version := make(chan string, 1)
	go func(u chan string) {
		if disabled := os.Getenv("CHECKPOINT_DISABLE"); disabled != "" {
			u <- ""
			return
		}

		if strings.HasSuffix(s.version.String(), "+HEAD") || protect.ProtectEnabled {
			// chan not check version against HEAD or
			// against pledge(2)'d OpenBSD
			u <- ""
			return
		}

		var r ghrel.Release
		var err error
		if len(s.version.Pre) > 0 {
			r, err = ghrel.FetchLatestRelease(gitHubOrg, gitHubRepo)
		} else {
			r, err = ghrel.FetchLatestStableRelease(gitHubOrg, gitHubRepo)
		}
		if err != nil {
			u <- color.RedString("\nError checking latest version: %s", err)
			return
		}

		if s.version.LT(r.Version()) {
			notice := fmt.Sprintf("\nYour version (%s) of gopass is out of date!\nThe latest version is %s.\n", s.version, r.Version().String())
			notice += "You can update by downloading from www.justwatch.com/gopass"
			if err := s.isUpdateable(ctx); err == nil {
				notice += " by running 'gopass update'"
			}
			notice += " or via your package manager"
			u <- color.YellowString(notice)
		}
		u <- ""
	}(version)

	cli.VersionPrinter(c)

	// report all used crypto, sync and fs backends
	for _, mp := range append(s.Store.MountPoints(), "") {
		crypto := s.Store.Crypto(ctx, mp)
		fmt.Fprintf(stdout, "[%s] Crypto: %s %s\n", mp, crypto.Name(), crypto.Version(ctx))
		sync := s.Store.Sync(ctx, mp)
		fmt.Fprintf(stdout, "[%s] Sync: %s %s\n", mp, sync.Name(), sync.Version(ctx))
		storer := s.Store.Store(ctx, mp)
		fmt.Fprintf(stdout, "[%s] Store: %s %s\n", mp, storer.Name(), storer.Version())
	}

	select {
	case vi := <-version:
		if vi != "" {
			fmt.Fprintln(stdout, vi)
		}
	case <-time.After(2 * time.Second):
		out.Red(ctx, "Version check timed out")
	case <-ctx.Done():
		return exitError(ctx, ExitAborted, nil, "user aborted")
	}

	return nil
}
