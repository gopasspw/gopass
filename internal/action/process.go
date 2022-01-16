package action

import (
	"io/ioutil"

	"github.com/gopasspw/gopass/internal/action/exit"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/tpl"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/urfave/cli/v2"
)

// Process is a command to process a template and replace secrets contained in it.
func (s *Action) Process(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	file := c.Args().First()
	if file == "" {
		return exit.Error(exit.Usage, nil, "Usage: %s process <FILE>", s.Name)
	}

	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return exit.Error(exit.IO, err, "Failed to read file: %s", file)
	}

	obuf, err := tpl.Execute(ctx, string(buf), file, nil, s.Store)
	if err != nil {
		return exit.Error(exit.IO, err, "Failed to process file: %s", file)
	}

	out.Print(ctx, string(obuf))

	return nil
}
