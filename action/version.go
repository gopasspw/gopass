package action

import (
	"fmt"
	"os"

	"github.com/dominikschulz/github-releases/ghrel"
	"github.com/fatih/color"
	"github.com/urfave/cli"
)

// Version prints the gopass version
func (s *Action) Version(c *cli.Context) error {
	cli.VersionPrinter(c)

	gv := s.Store.GPGVersion()
	fmt.Printf("  GPG: %d.%d.%d\n", gv.Major, gv.Minor, gv.Patch)
	gmaj, gmin, gpa := s.Store.GitVersion()
	fmt.Printf("  Git: %d.%d.%d\n", gmaj, gmin, gpa)

	if disabled := os.Getenv("CHECKPOINT_DISABLE"); disabled != "" {
		return nil
	}

	r, err := ghrel.FetchLatestStableRelease("justwatchcom", "gopass")
	if err != nil {
		fmt.Println(color.RedString("\nError checking latest version: %s", err))
		os.Exit(1)
	}

	if r.Name != s.version {
		fmt.Println(color.YellowString("\nYour version of gopass is out of date!\nThe latest version is %s.\nYou can update by downloading from www.justwatch.com/gopass", r.Name))
	}

	return nil
}
