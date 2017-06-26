package action

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/termutil"
	shellquote "github.com/kballard/go-shellquote"
	"github.com/urfave/cli"
)

// List all secrets as a tree
func (s *Action) List(c *cli.Context) error {
	filter := c.Args().First()
	flat := c.Bool("flat")
	stripPrefix := c.Bool("strip-prefix")
	limit := c.Int("limit")

	l, err := s.Store.Tree()
	if err != nil {
		return err
	}

	var out io.Writer
	var buf *bytes.Buffer
	out = os.Stdout

	if filter == "" {
		if flat {
			for _, e := range l.List(limit) {
				fmt.Fprintln(out, e)
			}
			return nil
		}
		if rows, _ := termutil.GetTermsize(); l.Len() > rows {
			color.NoColor = true
			buf = &bytes.Buffer{}
			out = buf
		}
		fmt.Fprintln(out, l.Format(limit))
		if buf != nil {
			return s.pager(buf)
		}
		return nil
	}

	subtree := l.FindFolder(filter)
	if subtree == nil {
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
				fmt.Fprintln(out, e)
				continue
			}
			fmt.Fprintln(out, filter+sep+e)
		}
		return nil
	}
	if rows, _ := termutil.GetTermsize(); subtree.Len() > rows {
		color.NoColor = true
		buf = &bytes.Buffer{}
		out = buf
	}
	fmt.Fprintln(out, subtree.Format(limit))
	if buf != nil {
		return s.pager(buf)
	}
	return nil
}

func (s *Action) pager(buf io.Reader) error {
	pager := os.Getenv("PAGER")
	if pager == "" {
		pager = "pager"
	}

	args, err := shellquote.Split(pager)
	if err != nil {
		return err
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = buf
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
