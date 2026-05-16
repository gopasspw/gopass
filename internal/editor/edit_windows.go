//go:build windows

package editor

import (
	"context"
	"os"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/urfave/cli/v3"
)

// Path return the name/path of the preferred editor
func Path(ctx context.Context, cmd *cli.Command) string {
	if cmd != nil {
		if ed := cmd.String("editor"); ed != "" {
			return ed
		}
	}
	if ed := config.String(ctx, "edit.editor"); ed != "" {
		return ed
	}
	if ed := os.Getenv("EDITOR"); ed != "" {
		return ed
	}
	return "notepad.exe"
}
