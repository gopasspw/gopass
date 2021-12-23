package action

import (
	"context"
	"fmt"

	"github.com/gopasspw/gopass/internal/create"
	"github.com/gopasspw/gopass/internal/cui"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/clipboard"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/urfave/cli/v2"
)

// Create displays the password creation wizard
func (s *Action) Create(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)

	out.Printf(ctx, "🌟 Welcome to the secret creation wizard (gopass create)!")
	out.Printf(ctx, "🧪 Hint: Use 'gopass edit -c' for more control!")

	wiz, err := create.New(ctx, s.Store.Storage(ctx, c.String("store")))
	if err != nil {
		return err
	}
	acts := wiz.Actions(s.Store, s.createPrintOrCopy)

	act, sel := cui.GetSelection(ctx, "Please select the type of secret you would like to create", acts.Selection())
	switch act {
	case "default":
		fallthrough
	case "show":
		return acts.Run(ctx, c, sel)
	default:
		return ExitError(ExitAborted, nil, "user aborted")
	}
}

// createPrintOrCopy will display the created password (or copy to clipboard)
func (s *Action) createPrintOrCopy(ctx context.Context, c *cli.Context, name, password string, genPw bool) error {
	if !genPw {
		return nil
	}

	if c.Bool("print") {
		fmt.Fprintf(out.Stdout, "The generated password for %s is:\n%s\n", name, password)
		return nil
	}

	if err := clipboard.CopyTo(ctx, name, []byte(password), s.cfg.ClipTimeout); err != nil {
		return ExitError(ExitIO, err, "failed to copy to clipboard: %s", err)
	}
	return nil
}
