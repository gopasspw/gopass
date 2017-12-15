// +build linux

package action

import (
	"os"
	"os/exec"

	"github.com/urfave/cli"
)

func getEditor(c *cli.Context) string {
	if ed := c.String("editor"); ed != "" {
		return ed
	}
	if ed := os.Getenv("EDITOR"); ed != "" {
		return ed
	}
	if p, err := exec.LookPath("editor"); err == nil {
		return p
	}
	// if neither EDITOR is set nor "editor" available we'll just assume that vi
	// is installed. If this fails the user will have to set $EDITOR
	return "vi"
}
