// +build !linux,!windows

package action

import (
	"os"

	"github.com/urfave/cli"
)

func getEditor(c *cli.Context) string {
	if ed := c.String("editor"); ed != "" {
		return ed
	}
	if ed := os.Getenv("EDITOR"); ed != "" {
		return ed
	}
	// given, this is a very opinionated default, but this should be available
	// on virtually any UNIX system and the user can still set EDITOR to get
	// his favorite one
	return "vi"
}
