package main

import (
	"io"
	"fmt"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/gopass"
	"github.com/urfave/cli/v2"
)


type gc struct {
	gp gopass.Store
}

// Get outputs the password for given path on stdout
func (s *gc) Get(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	ctx = ctxutil.WithNoNetwork(ctx, true)
	path := c.Args().Get(0)
	secret, err := s.gp.Get(ctx, path, "latest")
	if err != nil {
		return err
	}
	password := secret.Get("password")
	i, err := io.WriteString(out.Stdout, password+"\n")
	_ = i
	if err != nil {
		return fmt.Errorf("could not write to stdout: %s", err)
	}
	return nil
}
