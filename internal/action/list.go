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
	"github.com/gopasspw/gopass/internal/tree"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"golang.org/x/crypto/ssh/terminal"

	shellquote "github.com/kballard/go-shellquote"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

// List all secrets as a tree. If the filter argument is non-empty
// display only those that have this prefix
func (s *Action) List(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	filter := c.Args().First()
	flat := c.Bool("flat")
	stripPrefix := c.Bool("strip-prefix")
	folders := c.Bool("folders")

	// print the path if the argument is a direct hit
	if s.Store.Exists(ctx, filter) && !s.Store.IsDir(ctx, filter) {
		fmt.Println(filter)
		return nil
	}

	// we only support listing folders in flat mode currently
	if folders {
		flat = true
	}

	ctx = s.Store.WithConfig(ctx, filter)

	l, err := s.Store.Tree(ctx)
	if err != nil {
		return ExitError(ctx, ExitList, err, "failed to list store: %s", err)
	}

	//set limit to len of store to loop over all values later
	limit := l.Len()
	if c.IsSet("limit") {
		limit = c.Int("limit")
	}

	if filter == "" {
		return s.listAll(ctx, l, limit, flat, folders)
	}

	return s.listFiltered(ctx, l, limit, flat, folders, stripPrefix, filter)
}

func (s *Action) listFiltered(ctx context.Context, l tree.Tree, limit int, flat, folders, stripPrefix bool, filter string) error {
	subtree, err := l.FindFolder(filter)
	if err != nil {
		return ExitError(ctx, ExitNotFound, nil, "Entry '%s' not found", filter)
	}

	// SetRoot formats the root entry properly
	subtree.SetRoot(true)
	subtree.SetName(filter)
	if flat {
		sep := "/"
		if strings.HasSuffix(filter, "/") {
			sep = ""
		}
		listOver := subtree.List
		if folders {
			listOver = subtree.ListFolders
		}
		for _, e := range listOver(limit) {
			if stripPrefix {
				fmt.Fprintln(stdout, e)
				continue
			}
			fmt.Fprintln(stdout, filter+sep+e)
		}
		return nil
	}

	// we may need to redirect stdout for the pager support
	so, buf := redirectPager(ctx, subtree)

	fmt.Fprintln(so, subtree.Format(limit))
	if buf != nil {
		if err := s.pager(ctx, buf); err != nil {
			return ExitError(ctx, ExitUnknown, err, "failed to invoke pager: %s", err)
		}
	}
	return nil
}

// redirectPager returns a redirected io.Writer if the output would exceed
// the terminal size
func redirectPager(ctx context.Context, subtree tree.Tree) (io.Writer, *bytes.Buffer) {
	if ctxutil.IsNoPager(ctx) {
		return stdout, nil
	}
	_, rows, err := terminal.GetSize(0)
	if err != nil {
		return stdout, nil
	}
	if subtree == nil || subtree.Len() < rows {
		return stdout, nil
	}
	if pager := os.Getenv("PAGER"); pager == "" {
		return stdout, nil
	}
	color.NoColor = true
	buf := &bytes.Buffer{}
	return buf, buf
}

// listAll will unconditionally list all entries, used if no filter is given
func (s *Action) listAll(ctx context.Context, l tree.Tree, limit int, flat, folders bool) error {
	if flat {
		listOver := l.List
		if folders {
			listOver = l.ListFolders
		}
		for _, e := range listOver(limit) {
			fmt.Fprintln(stdout, e)
		}
		return nil
	}

	// we may need to redirect stdout for the pager support
	so, buf := redirectPager(ctx, l)

	fmt.Fprintln(so, l.Format(limit))
	if buf != nil {
		if err := s.pager(ctx, buf); err != nil {
			return ExitError(ctx, ExitUnknown, err, "failed to invoke pager: %s", err)
		}
	}
	return nil
}

// pager invokes the default pager with the given content
func (s *Action) pager(ctx context.Context, buf io.Reader) error {
	pager := os.Getenv("PAGER")
	if pager == "" {
		fmt.Fprintln(stdout, buf)
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
