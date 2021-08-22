package action

import (
	"bytes"
	"fmt"

	"github.com/gopasspw/gopass/internal/audit"
	"github.com/gopasspw/gopass/internal/editor"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/gopass/secrets"
	"github.com/urfave/cli/v2"
)

func (s *Action) Merge(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	to := c.Args().First()
	from := c.Args().Tail()

	if to == "" {
		return ExitError(ExitUsage, nil, "usage: %s merge <to> <from> [<from>]", s.Name)
	}
	if len(from) < 1 {
		return ExitError(ExitUsage, nil, "usage: %s merge <to> <from> [<from>]", s.Name)
	}

	ed := editor.Path(c)
	if err := editor.Check(ctx, ed); err != nil {
		out.Warningf(ctx, "Failed to check editor config: %s", err)
	}

	content := &bytes.Buffer{}
	for _, k := range c.Args().Slice() {
		if !s.Store.Exists(ctx, k) {
			continue
		}
		sec, err := s.Store.Get(ctxutil.WithShowParsing(ctx, false), k)
		if err != nil {
			return ExitError(ExitDecrypt, err, "failed to decrypt: %s: %w", k, err)
		}
		_, err = content.WriteString("\n# Secret: " + k + "\n")
		if err != nil {
			return ExitError(ExitUnknown, err, "failed to write: %w", err)
		}
		_, err = content.Write(sec.Bytes())
		if err != nil {
			return ExitError(ExitUnknown, err, "failed to write: %w", err)
		}
	}

	// invoke the editor to let the user edit the content
	newContent, err := editor.Invoke(ctx, ed, content.Bytes())
	if err != nil {
		return ExitError(ExitUnknown, err, "failed to invoke editor: %s", err)
	}
	// If content is equal, nothing changed, exiting
	if bytes.Equal(content.Bytes(), newContent) {
		return nil
	}

	nSec := secrets.ParsePlain(newContent)

	// if the secret has a password, we check it's strength
	if pw := nSec.Password(); pw != "" {
		audit.Single(ctx, pw)
	}

	// write result (back) to store
	if err := s.Store.Set(ctxutil.WithCommitMessage(ctx, fmt.Sprintf("Merged %+v", c.Args().Slice())), to, nSec); err != nil {
		return ExitError(ExitEncrypt, err, "failed to encrypt secret %s: %s", to, err)
	}

	if !c.Bool("delete") {
		return nil
	}

	for _, old := range from {
		if err := s.Store.Delete(ctx, old); err != nil {
			return ExitError(ExitUnknown, err, "failed to delete %s: %w", old, err)
		}
	}
	return nil
}
