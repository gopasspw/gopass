package action

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/urfave/cli/v2"
)

// Env implements an env subcommand the populates the env of an subprocess with a set of secrets
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
		subtree.SetRoot(true)
		subtree.SetName(name)
		for _, e := range subtree.List(0) {
			en := path.Join(name, e)
			log.Println(en)
			keys = append(keys, en)
		}
	} else {
		keys = append(keys, name)
	}

	env := make([]string, 0, 1)
	for _, key := range keys {
		log.Println(key)
		sec, err := s.Store.Get(ctx, key)
		if err != nil {
			return err
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
