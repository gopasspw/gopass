// +build windows

package action

import (
	"os"

	"github.com/urfave/cli"
)

func getEditor(c *cli.Context) string {
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
