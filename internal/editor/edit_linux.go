//go:build linux

package editor

import (
	"os"
	"os/exec"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/urfave/cli/v2"
)

// Path return the name/path of the preferred editor.
func Path(c *cli.Context) string {
	if c != nil {
		if ed := c.String("editor"); ed != "" {
			debug.Log("Using editor from command line: %s", ed)

			return ed
		}
	}
	if ed := config.String(c.Context, "edit.editor"); ed != "" {
		debug.Log("Using editor from config: %s", ed)

		return ed
	}
	if ed := os.Getenv("EDITOR"); ed != "" {
		debug.Log("Using editor from $EDITOR: %s", ed)

		return ed
	}
	if p, err := exec.LookPath("editor"); err == nil {
		debug.Log("Using editor from $PATH: %s", p)

		return p
	}
	// if neither EDITOR is set nor "editor" available we'll just assume that vi
	// is installed. If this fails the user will have to set `$EDITOR`.
	debug.Log("Using default editor: %s", "vi")

	return "vi"
}
