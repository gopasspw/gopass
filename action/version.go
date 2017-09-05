package action

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/dominikschulz/github-releases/ghrel"
	"github.com/fatih/color"
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
			u <- color.YellowString("\nYour version (%s) of gopass is out of date!\nThe latest version is %s.\nYou can update by downloading from www.justwatch.com/gopass or via your package manager", s.version, r.Version().String())
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
		fmt.Println(color.RedString("Version check timed out"))
	}

	return nil
}
