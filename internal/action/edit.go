package action

import (
	"bytes"
	"context"
	"fmt"
	"errors"

	"github.com/gopasspw/gopass/internal/action/exit"
	"github.com/gopasspw/gopass/internal/audit"
	"github.com/gopasspw/gopass/internal/editor"
	"github.com/gopasspw/gopass/internal/store"	
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/gopass/secrets"
	"github.com/gopasspw/gopass/pkg/pwgen"
	"github.com/gopasspw/gopass/pkg/termio"
	"github.com/urfave/cli/v2"
)

// Edit the content of a password file.
func (s *Action) Edit(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	name := c.Args().First()
	if name == "" {
		return exit.Error(exit.Usage, nil, "Usage: %s edit secret", s.Name)
	}

	return s.edit(ctx, c, name)
}

func (s *Action) edit(ctx context.Context, c *cli.Context, name string) error {
	ed := editor.Path(c)

	// get existing content or generate new one from a template.
	name, content, changed, err := s.editGetContent(ctx, name, c.Bool("create"))
	if err != nil {
		return err
	}

	// invoke the editor to let the user edit the content.
	newContent, err := editor.Invoke(ctx, ed, content)
	if err != nil {
		return exit.Error(exit.Unknown, err, "failed to invoke editor: %s", err)
	}

	return s.editUpdate(ctx, name, content, newContent, changed, ed)
}

func (s *Action) editUpdate(ctx context.Context, name string, content, nContent []byte, changed bool, ed string) error {
	// If content is equal, nothing changed, exiting.
	if bytes.Equal(content, nContent) && !changed {
		return nil
	}

	nSec := secrets.ParseAKV(nContent)

	// if the secret has a password, we check its strength.
	if pw := nSec.Password(); pw != "" {
		audit.Single(ctx, pw)
	}

	// write result (back) to store.
	if err := s.Store.Set(ctxutil.WithCommitMessage(ctx, fmt.Sprintf("Edited with %s", ed)), name, nSec); err != nil {
		if !errors.Is(err, store.ErrMeaninglessWrite) {
			return exit.Error(exit.Encrypt, err, "failed to encrypt secret %s: %s", name, err)
		}
		out.Warningf(ctx, "The new value of the password is equal to its current value. Are you sure you want this?")
	}

	return nil
}

func (s *Action) editGetContent(ctx context.Context, name string, create bool) (string, []byte, bool, error) {
	if !s.Store.Exists(ctx, name) && !create {
		var err error
		name, err = s.editFindName(ctx, name)
		if err != nil {
			return "", nil, false, err
		}
	}

	// edit existing entry.
	if s.Store.Exists(ctx, name) {
		// we make sure we are not parsing the content of the file when editing.
		sec, err := s.Store.Get(ctxutil.WithShowParsing(ctx, false), name)
		if err != nil {
			return name, nil, false, exit.Error(exit.Decrypt, err, "failed to decrypt %s: %s", name, err)
		}

		return name, sec.Bytes(), false, nil
	}

	if !create {
		out.Warningf(ctx, "Entry %s not found. Creating new secret ...", name)
	}

	// load template if it exists.
	pwLength, _ := defaultLengthFromEnv(ctx)
	if content, found := s.renderTemplate(ctx, name, []byte(pwgen.GeneratePassword(pwLength, false))); found {
		return name, content, true, nil
	}

	// new entry, no template.
	return name, nil, false, nil
}

func (s *Action) editFindName(ctx context.Context, name string) (string, error) {
	newName := ""
	// capture only the name of the selected secret.
	cb := func(ctx context.Context, c *cli.Context, selectedName string, recurse bool) error {
		newName = selectedName

		return nil
	}
	if err := s.find(ctx, nil, name, cb, false); err != nil {
		debug.Log("failed to find secret %s: %s", name, err)

		return name, nil
	}

	cont, err := termio.AskForBool(ctx, fmt.Sprintf("Secret does not exist %q. Found possible match in %q. Edit existing entry?", name, newName), true)
	if err != nil {
		return name, err
	}

	if cont {
		return newName, nil
	}

	return name, nil
}
