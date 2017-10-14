package action

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/dominikschulz/github-releases/ghrel"
	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/utils/out"
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

		if s.version.String() == "0.0.0+HEAD" {
			// chan not check version against HEAD
			u <- ""
			return
		}

		r, err := ghrel.FetchLatestStableRelease(gitHubOrg, gitHubRepo)
		if err != nil {
			u <- color.RedString("\nError checking latest version: %s", err)
			return
		}

		if s.version.LT(r.Version()) {
			notice := fmt.Sprintf("\nYour version (%s) of gopass is out of date!\nThe latest version is %s.\n", s.version, r.Version().String())
			notice += "You can update by downloading from www.justwatch.com/gopass"
			if err := s.isUpdateable(ctx); err == nil {
				notice += " by running 'gopass update' "
			}
			notice += "or via your package manager"
			u <- color.YellowString(notice)
		}
		u <- ""
	}(version)

	cli.VersionPrinter(c)

	fmt.Printf("  GPG: %s\n", s.Store.GPGVersion(ctx).String())
	fmt.Printf("  Git: %s\n", s.Store.GitVersion(ctx).String())

	select {
	case vi := <-version:
		if vi != "" {
			fmt.Println(vi)
		}
	case <-time.After(2 * time.Second):
		out.Red(ctx, "Version check timed out")
	case <-ctx.Done():
		return s.exitError(ctx, ExitAborted, nil, "user aborted")
	}

	return nil
}
