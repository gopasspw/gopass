package action

import (
	"context"
	"io"
	"os"
	"strings"

	"github.com/gopasspw/gopass/internal/action/exit"
	"github.com/gopasspw/gopass/internal/audit"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/tree"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/fsutil"
	"github.com/urfave/cli/v2"
)

// Audit validates passwords against common flaws.
func (s *Action) Audit(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)

	_ = s.rem.Reset("audit")
	out.Print(ctx, "Auditing passwords for common flaws ...")

	t, err := s.Store.Tree(ctx)
	if err != nil {
		return exit.Error(exit.List, err, "failed to get store tree: %s", err)
	}

	if filter := c.Args().First(); filter != "" {
		subtree, err := t.FindFolder(filter)
		if err != nil {
			return exit.Error(exit.Unknown, err, "failed to find subtree: %s", err)
		}
		debug.Log("subtree for %q: %+v", filter, subtree)
		t = subtree
	}

	list := t.List(tree.INF)

	if len(list) < 1 {
		out.Printf(ctx, "No secrets found")

		return nil
	}

	var excludes string
	st := s.Store.Storage(ctx, c.Args().First())
	if buf, err := st.Get(ctx, ".gopass-audit-ignore"); err == nil && buf != nil {
		excludes = string(buf)
	}
	nList := audit.FilterExcludes(excludes, list)
	if len(nList) < len(list) {
		out.Warningf(ctx, "Excluding %d secrets based on .gopass-audit-ignore", len(list)-len(nList))
	}

	a := audit.New(c.Context, s.Store)
	r, err := a.Batch(ctx, nList)
	if err != nil {
		return exit.Error(exit.Unknown, err, "failed to audit password store: %s", err)
	}

	if p := c.String("template"); p != "" && fsutil.IsFile(p) {
		r.Template = p
	}

	switch c.String("format") {
	case "html":
		return saveReport(ctx, r.RenderHTML, c.String("output-file"), "html")
	case "csv":
		return saveReport(ctx, r.RenderCSV, c.String("output-file"), "csv")
	default:
		var err error
		if c.Bool("full") {
			debug.Log("Printing full report")
			err = r.PrintResults(ctx)
		}
		if c.Bool("summary") {
			debug.Log("Printing summary")

			nerr := r.PrintSummary(ctx)
			// do not overwrite err if it is already set
			if err == nil {
				err = nerr
			}
		}
		if !c.Bool("full") && !c.Bool("summary") {
			out.Warning(ctx, "No output format specified. Use `--full` or `--summary` to specify.")
		}

		return err
	}
}

func saveReport(ctx context.Context, f func(io.Writer) error, path, suffix string) error {
	if path == "" {
		out.Noticef(ctx, "No output filename given. Will use a random file name. Use `--output-file` to specify.")
	}

	fn, err := writeReport(f, path)
	if err != nil {
		return exit.Error(exit.Unknown, err, "failed to write report to %s: %s", fn, err)
	}

	if !strings.HasSuffix(fn, "."+suffix) {
		nfn := fn + "." + suffix
		if err := os.Rename(fn, fn+"."+suffix); err != nil {
			return exit.Error(exit.IO, err, "failed to rename report to %s: %s", nfn, err)
		}
		fn = nfn
	}

	out.Noticef(ctx, "Wrote report to %s", fn)

	return nil
}

func writeReport(f func(io.Writer) error, path string) (string, error) {
	fh, err := openReport(path)
	if err != nil {
		return "", err
	}
	defer fh.Close() //nolint:errcheck

	if err := f(fh); err != nil {
		return "", err
	}

	return fh.Name(), nil
}

func openReport(path string) (*os.File, error) {
	if path == "" {
		return os.CreateTemp("", "gopass-report")
	}

	return os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
}
