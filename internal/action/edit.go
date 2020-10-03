package action

import (
	"bytes"
	"context"
	"fmt"

	"github.com/gopasspw/gopass/internal/audit"
	"github.com/gopasspw/gopass/internal/editor"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/termio"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/gopass/secret"
	"github.com/gopasspw/gopass/pkg/gopass/secret/secparse"
	"github.com/gopasspw/gopass/pkg/pwgen"

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
		out.Red(ctx, "Failed to check editor config: %s", err)
	}

	// get existing content or generate new one from a template
	name, content, changed, err := s.editGetContent(ctx, name, c.Bool("create"))
	if err != nil {
		return err
	}

	fromMIME := false
	if bytes.HasPrefix(content, []byte(secret.Ident)) {
		fromMIME = true
		content = bytes.TrimPrefix(content, []byte(secret.Ident+"\n"))
	}

	// preserve intermediate state across retries
	tmpContent := content
	for {
		// invoke the editor to let the user edit the content
		newContent, err := editor.Invoke(ctx, ed, tmpContent)
		if err != nil {
			return ExitError(ExitUnknown, err, "failed to invoke editor: %s", err)
		}
		// retain edited content for next try
		tmpContent = newContent
		if fromMIME {
			newContent = append([]byte(secret.Ident+"\n"), newContent...)
			if _, err := secparse.Parse(newContent); err != nil {
				out.Red(ctx, "ERROR: Cannot parse Gopass native Secret: %s", err)
				out.Yellow(ctx, "Hint: Does your Key-Value section end with a new line?")
				cont, err := termio.AskForBool(ctx, "Retry?", true)
				if err != nil {
					return err
				}
				if !cont {
					return fmt.Errorf("malformed secret")
				}
				continue
			}
		}
		return s.editUpdate(ctx, name, content, newContent, changed, ed)
	}
}

func (s *Action) editUpdate(ctx context.Context, name string, content, nContent []byte, changed bool, ed string) error {
	// If content is equal, nothing changed, exiting
	if bytes.Equal(content, nContent) && !changed {
		return nil
	}

	nSec, err := secparse.Parse(nContent)
	if err != nil {
		out.Error(ctx, "WARNING: Invalid Secret: %s", err)
	}

	// if the secret has a password, we check it's strength
	if pw := nSec.Get("password"); pw != "" {
		audit.Single(ctx, pw)
	}

	// write result (back) to store
	if err := s.Store.Set(ctxutil.WithCommitMessage(ctx, fmt.Sprintf("Edited with %s", ed)), name, nSec); err != nil {
		return ExitError(ExitEncrypt, err, "failed to encrypt secret %s: %s", name, err)
	}
	return nil
}

func (s *Action) editGetContent(ctx context.Context, name string, create bool) (string, []byte, bool, error) {
	if !s.Store.Exists(ctx, name) {
		newName := ""
		// capture only the name of the selected secret
		cb := func(ctx context.Context, c *cli.Context, name string, recurse bool) error {
			newName = name
			return nil
		}
		if err := s.find(ctxutil.WithFuzzySearch(ctx, false), nil, name, cb); err == nil {
			name = newName
		}
	}

	// edit existing entry
	if s.Store.Exists(ctx, name) {
		sec, err := s.Store.Get(ctx, name)
		if err != nil {
			return name, nil, false, ExitError(ExitDecrypt, err, "failed to decrypt %s: %s", name, err)
		}
		return name, sec.Bytes(), false, nil
	}

	if !create {
		out.Yellow(ctx, "Entry %s not found. Creating new secret ...", name)
	}

	// load template if it exists
	if content, found := s.renderTemplate(ctx, name, []byte(pwgen.GeneratePassword(defaultLength, false))); found {
		return name, content, true, nil
	}

	// new entry, no template
	return name, nil, false, nil
}
