package action

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/gopasspw/gopass/internal/action/exit"
	"github.com/gopasspw/gopass/internal/tree"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/urfave/cli/v2"
)

// Env implements the env subcommand. It populates the environment of a subprocess with
// a set of environment variables corresponding to the secret subtree specified on the
// command line.
func (s *Action) Env(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	name := c.Args().First()
	args := c.Args().Tail()

	if len(args) == 0 {
		return exit.Error(exit.Usage, nil, "Missing subcommand to execute")
	}

	if !s.Store.Exists(ctx, name) && !s.Store.IsDir(ctx, name) {
		return exit.Error(exit.NotFound, nil, "Secret %s not found", name)
	}

	keys := make([]string, 0, 1)
	if s.Store.IsDir(ctx, name) {
		debug.Log("%q is a dir, adding it's entries", name)

		l, err := s.Store.Tree(ctx)
		if err != nil {
			return exit.Error(exit.List, err, "failed to list store: %s", err)
		}

		subtree, err := l.FindFolder(name)
		if err != nil {
			return exit.Error(exit.NotFound, nil, "Entry %q not found", name)
		}

		for _, e := range subtree.List(tree.INF) {
			debug.Log("found key: %s", e)
			keys = append(keys, e)
		}
	} else {
		keys = append(keys, name)
	}

	env := make([]string, 0, 1)
	for _, key := range keys {
		debug.Log("exporting to environment key: %s", key)
		sec, err := s.Store.Get(ctx, key)
		if err != nil {
			return fmt.Errorf("failed to get entry for env prefix %q: %w", name, err)
		}
		env = append(env, fmt.Sprintf("%s=%s", strings.ToUpper(path.Base(key)), sec.Password()))
	}

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Env = append(os.Environ(), env...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
