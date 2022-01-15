package gitfs

import (
	"os"
	"os/exec"
	"strings"

	"github.com/gopasspw/gopass/internal/action/exit"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/urfave/cli/v2"
)

// Commands returns the commands that are available for the gitfs backend.
// TODO: maybe we just want to add the Before action when populating the final
// command slice (unless it's non-nil so backends can override it). A similar
// approach could be taken with the Action function. We could wrap it, parse
// "global" flags like store and put that into the context. A bit hacky
// but on the other hand less ugly wrt. the function signature.
func (l loader) Commands(i func(*cli.Context) error, s func(string) (string, error)) []*cli.Command {
	return []*cli.Command{
		{
			Name:  "git",
			Usage: "Run a git command inside a password store: gopass git [--store=<store>] <git-command>",
			Description: "" +
				"If the password store is a git repository, execute a git command " +
				"specified by git-command-args.",
			Hidden: true,
			Before: i,
			Action: func(c *cli.Context) error {
				ctx := ctxutil.WithGlobalFlags(c)
				store := c.String("store")

				path, err := s(store)
				if err != nil {
					return exit.Error(exit.Unknown, err, "failed to get sub store %s: %s", store, err)
				}

				args := c.Args().Slice()
				out.Noticef(ctx, "Running 'git %s' in %s...", strings.Join(args, " "), path)
				cmd := exec.CommandContext(ctx, "git", args...)
				cmd.Dir = path
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				cmd.Stdin = os.Stdin

				return cmd.Run()
			},
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "store",
					Usage: "Store to operate on",
				},
			},
		},
	}
}
