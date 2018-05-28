package action

import (
	"bytes"
	"context"
	"fmt"

	"github.com/gopasspw/gopass/pkg/audit"
	"github.com/gopasspw/gopass/pkg/editor"
	"github.com/gopasspw/gopass/pkg/out"
	"github.com/gopasspw/gopass/pkg/pwgen"
	"github.com/gopasspw/gopass/pkg/store/secret"
	"github.com/gopasspw/gopass/pkg/store/sub"
	"github.com/gopasspw/gopass/pkg/tpl"

	"github.com/urfave/cli"
)

// Edit the content of a password file
func (s *Action) Edit(ctx context.Context, c *cli.Context) error {
	name := c.Args().First()
	if name == "" {
		return ExitError(ctx, ExitUsage, nil, "Usage: %s edit secret", s.Name)
	}

	ed := editor.Path(c)

	// get existing content or generate new one from a template
	var content []byte
	var changed bool
	if s.Store.Exists(ctx, name) {
		sec, err := s.Store.Get(ctx, name)
		if err != nil {
			return ExitError(ctx, ExitDecrypt, err, "failed to decrypt %s: %s", name, err)
		}
		content, err = sec.Bytes()
		if err != nil {
			return ExitError(ctx, ExitDecrypt, err, "failed to decode %s: %s", name, err)
		}
	} else if tmpl, found := s.Store.LookupTemplate(ctx, name); found {
		changed = true
		// load template if it exists
		content = []byte(pwgen.GeneratePassword(defaultLength, false))
		if nc, err := tpl.Execute(ctx, string(tmpl), name, content, s.Store); err == nil {
			content = nc
		} else {
			fmt.Fprintf(stdout, "failed to execute template: %s\n", err)
		}
	}

	// invoke the editor to let the user edit the content
	nContent, err := editor.Invoke(ctx, ed, content)
	if err != nil {
		return ExitError(ctx, ExitUnknown, err, "failed to invoke editor: %s", err)
	}

	// If content is equal, nothing changed, exiting
	if bytes.Equal(content, nContent) && !changed {
		return nil
	}

	nSec, err := secret.Parse(nContent)
	if err != nil {
		out.Red(ctx, "WARNING: Invalid YAML: %s", err)
	}

	// if the secret has a password, we check it's strength
	if pw := nSec.Password(); pw != "" {
		audit.Single(ctx, pw)
	}

	// write result (back) to store
	if err := s.Store.Set(sub.WithReason(ctx, fmt.Sprintf("Edited with %s", ed)), name, nSec); err != nil {
		return ExitError(ctx, ExitEncrypt, err, "failed to encrypt secret %s: %s", name, err)
	}
	return nil
}
