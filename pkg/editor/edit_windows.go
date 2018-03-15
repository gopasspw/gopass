// +build windows

package editor

import (
	"os"

	"github.com/urfave/cli"
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
	return "notepad.exe"
}
