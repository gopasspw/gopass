package main

import (
	"fmt"
	"io"
	"os"

	"github.com/gopasspw/gopass/internal/action"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/gopass"
	"github.com/urfave/cli/v2"
)

var (
	// Stdout is exported for tests
	Stdout io.Writer = os.Stdout
)

type gc struct {
	gp gopass.Store
}

// Get outputs the password for given path on stdout
func (s *gc) Get(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	ctx = ctxutil.WithNoNetwork(ctx, true)
	path := c.Args().Get(0)
	key := c.Args().Get(1)
	if key != "" {
		ctx = action.WithKey(ctx, key)
	}
	secret, err := s.gp.Get(ctx, path, "latest")
	if err != nil {
		return err
	}

	if key == "" {
		fmt.Fprintln(Stdout, secret.Password())
		return nil
	}

	val, found := secret.Get(key)
	if found {
		fmt.Fprintln(Stdout, val)
	}

	return nil
}
