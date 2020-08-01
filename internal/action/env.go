package action

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/gopasspw/gopass/internal/debug"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/urfave/cli/v2"
)

// Env implements the env subcommand. It populates the environment of a subprocess with
// a set of environment variables corresponding to the secret subtree specified on the
// command line.
func (s *Action) Env(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	name := c.Args().First()
	args := c.Args().Tail()

	if !s.Store.Exists(ctx, name) && !s.Store.IsDir(ctx, name) {
		return ExitError(ExitNotFound, nil, "Secret %s not found", name)
	}

	keys := make([]string, 0, 1)
	if s.Store.IsDir(ctx, name) {
		l, err := s.Store.Tree(ctx)
		if err != nil {
			return ExitError(ExitList, err, "failed to list store: %s", err)
		}
		subtree, err := l.FindFolder(name)
		if err != nil {
			return ExitError(ExitNotFound, nil, "Entry '%s' not found", name)
		}
		subtree.SetName(name)
		for _, e := range subtree.List(0) {
			en := path.Join(name, e)
			debug.Log("found key: %s", en)
			keys = append(keys, en)
		}
	} else {
		keys = append(keys, name)
	}

	env := make([]string, 0, 1)
	for _, key := range keys {
		debug.Log("exporting to environment key: %s", key)
		sec, err := s.Store.Get(ctx, key)
		if err != nil {
			return err
		}
		env = append(env, fmt.Sprintf("%s=%s", strings.ToUpper(path.Base(key)), sec.Get("password")))
	}

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Env = append(os.Environ(), env...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
