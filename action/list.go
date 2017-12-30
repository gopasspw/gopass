package action

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/justwatchcom/gopass/utils/termutil"
	"github.com/justwatchcom/gopass/utils/tree"
	shellquote "github.com/kballard/go-shellquote"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

// List all secrets as a tree
func (s *Action) List(ctx context.Context, c *cli.Context) error {
	filter := c.Args().First()
	flat := c.Bool("flat")
	stripPrefix := c.Bool("strip-prefix")
	limit := c.Int("limit")

	l, err := s.Store.Tree()
	if err != nil {
		return exitError(ctx, ExitList, err, "failed to list store: %s", err)
	}

	if filter == "" {
		return s.listAll(ctx, l, limit, flat)
	}
	return s.listFiltered(ctx, l, limit, flat, stripPrefix, filter)
}

func (s *Action) listFiltered(ctx context.Context, l tree.Tree, limit int, flat, stripPrefix bool, filter string) error {
	subtree, err := l.FindFolder(filter)
	if err != nil {
		out.Red(ctx, "Entry '%s' not found", filter)
		return nil
	}

	subtree.SetRoot(true)
	subtree.SetName(filter)
	if flat {
		sep := "/"
		if strings.HasSuffix(filter, "/") {
			sep = ""
		}
		for _, e := range subtree.List(limit) {
			if stripPrefix {
				fmt.Println(e)
				continue
			}
			fmt.Println(filter + sep + e)
		}
		return nil
	}

	var stdout io.Writer
	var buf *bytes.Buffer
	// we may need to redirect stdout for the pager support
	stdout = os.Stdout

	if rows, _ := termutil.GetTermsize(); subtree.Len() > rows {
		color.NoColor = true
		buf = &bytes.Buffer{}
		stdout = buf
	}

	fmt.Fprintln(stdout, subtree.Format(limit))
	if buf != nil {
		if err := s.pager(ctx, buf); err != nil {
			return exitError(ctx, ExitUnknown, err, "failed to invoke pager: %s", err)
		}
	}
	return nil
}

func (s *Action) listAll(ctx context.Context, l tree.Tree, limit int, flat bool) error {
	if flat {
		for _, e := range l.List(limit) {
			fmt.Println(e)
		}
		return nil
	}

	var stdout io.Writer
	var buf *bytes.Buffer
	// we may need to redirect stdout for the pager support
	stdout = os.Stdout

	if rows, _ := termutil.GetTermsize(); l.Len() > rows && !ctxutil.IsNoPager(ctx) {
		color.NoColor = true
		buf = &bytes.Buffer{}
		stdout = buf
	}
	fmt.Fprintln(stdout, l.Format(limit))
	if buf != nil {
		if err := s.pager(ctx, buf); err != nil {
			return exitError(ctx, ExitUnknown, err, "failed to invoke pager: %s", err)
		}
	}
	return nil
}

func (s *Action) pager(ctx context.Context, buf io.Reader) error {
	pager := os.Getenv("PAGER")
	if pager == "" {
		fmt.Println(buf)
		return nil
	}

	args, err := shellquote.Split(pager)
	if err != nil {
		return errors.Wrapf(err, "failed to split pager command")
	}

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Stdin = buf
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
