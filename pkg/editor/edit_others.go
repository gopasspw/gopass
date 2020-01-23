// +build !linux,!windows

package editor

import (
	"os"

	"gopkg.in/urfave/cli.v1"
)

// Path return the name/path of the preferred editor
func Path(c *cli.Context) string {
	if c != nil {
		if ed := c.String("editor"); ed != "" {
			return ed
		}
	}
	if ed := os.Getenv("EDITOR"); ed != "" {
		return ed
	}
	// given, this is a very opinionated default, but this should be available
	// on virtually any UNIX system and the user can still set EDITOR to get
	// his favorite one
	return "vi"
}
