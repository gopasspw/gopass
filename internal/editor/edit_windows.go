//go:build windows
// +build windows

package editor

import (
	"os"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/urfave/cli/v2"
)

// Path return the name/path of the preferred editor
func Path(c *cli.Context) string {
	if c != nil {
		if ed := c.String("editor"); ed != "" {
			return ed
		}
	}
	if ed := config.String(c.Context, "edit.editor"); ed != "" {
		return ed
	}
	if ed := os.Getenv("EDITOR"); ed != "" {
		return ed
	}
	return "notepad.exe"
}
