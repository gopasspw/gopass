package action

import (
	"bytes"
	"context"
	"fmt"

	"github.com/gopasspw/gopass/internal/audit"
	"github.com/gopasspw/gopass/internal/editor"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/gopass/secrets"
	"github.com/gopasspw/gopass/pkg/pwgen"
	"github.com/gopasspw/gopass/pkg/termio"

	"github.com/urfave/cli/v2"
)

// Edit the content of a password file
func (s *Action) Edit(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	name := c.Args().First()
	if name == "" {
		return ExitError(ExitUsage, nil, "Usage: %s edit secret", s.Name)
	}

	return s.edit(ctx, c, name)
}

func (s *Action) edit(ctx context.Context, c *cli.Context, name string) error {
	ed := editor.Path(c)
	if err := editor.Check(ctx, ed); err != nil {
		out.Warningf(ctx, "Failed to check editor config: %s", err)
	}

	// get existing content or generate new one from a template
	name, content, changed, err := s.editGetContent(ctx, name, c.Bool("create"))
	if err != nil {
		return err
	}

	// invoke the editor to let the user edit the content
	newContent, err := editor.Invoke(ctx, ed, content)
	if err != nil {
		return ExitError(ExitUnknown, err, "failed to invoke editor: %s", err)
	}
	return s.editUpdate(ctx, name, content, newContent, changed, ed)
}

func (s *Action) editUpdate(ctx context.Context, name string, content, nContent []byte, changed bool, ed string) error {
	// If content is equal, nothing changed, exiting
	if bytes.Equal(content, nContent) && !changed {
		return nil
	}

	nSec := secrets.ParsePlain(nContent)

	// if the secret has a password, we check it's strength
	if pw := nSec.Password(); pw != "" {
		audit.Single(ctx, pw)
	}

	// write result (back) to store
	if err := s.Store.Set(ctxutil.WithCommitMessage(ctx, fmt.Sprintf("Edited with %s", ed)), name, nSec); err != nil {
		return ExitError(ExitEncrypt, err, "failed to encrypt secret %s: %s", name, err)
	}
	return nil
}

func (s *Action) editGetContent(ctx context.Context, name string, create bool) (string, []byte, bool, error) {
	if !s.Store.Exists(ctx, name) && !create {
		newName := ""
		// capture only the name of the selected secret
		cb := func(ctx context.Context, c *cli.Context, name string, recurse bool) error {
			newName = name
			return nil
		}
		if err := s.find(ctx, nil, name, cb, false); err == nil {
			cont, err := termio.AskForBool(ctx, fmt.Sprintf("Secret does not exist %q. Found possible match in %q. Edit existing entry?", name, newName), true)
			if err != nil {
				return "", nil, false, err
			}
			if cont {
				name = newName
			}
		}
	}

	// edit existing entry
	if s.Store.Exists(ctx, name) {
		// we make sure we are not parsing the content of the file when editing
		sec, err := s.Store.Get(ctxutil.WithShowParsing(ctx, false), name)
		if err != nil {
			return name, nil, false, ExitError(ExitDecrypt, err, "failed to decrypt %s: %s", name, err)
		}
		return name, sec.Bytes(), false, nil
	}

	if !create {
		out.Warningf(ctx, "Entry %s not found. Creating new secret ...", name)
	}

	// load template if it exists
	if content, found := s.renderTemplate(ctx, name, []byte(pwgen.GeneratePassword(defaultLength, false))); found {
		return name, content, true, nil
	}

	// new entry, no template
	return name, nil, false, nil
}
