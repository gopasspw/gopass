package action

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/justwatchcom/gopass/pkg/audit"
	"github.com/justwatchcom/gopass/pkg/ctxutil"
	"github.com/justwatchcom/gopass/pkg/editor"
	"github.com/justwatchcom/gopass/pkg/out"
	"github.com/justwatchcom/gopass/pkg/store"
	"github.com/justwatchcom/gopass/pkg/store/secret"
	"github.com/justwatchcom/gopass/pkg/store/sub"
	"github.com/justwatchcom/gopass/pkg/termio"

	"github.com/urfave/cli"
)

// Insert a string as content to a secret file
func (s *Action) Insert(ctx context.Context, c *cli.Context) error {
	echo := c.Bool("echo")
	multiline := c.Bool("multiline")
	force := c.Bool("force")

	args, kvps := parseArgs(c)
	name := args.Get(0)
	key := args.Get(1)

	if name == "" {
		return ExitError(ctx, ExitNoName, nil, "Usage: %s insert name", s.Name)
	}

	return s.insert(ctx, c, name, key, echo, multiline, force, kvps)
}

func (s *Action) insert(ctx context.Context, c *cli.Context, name, key string, echo, multiline, force bool, kvps map[string]string) error {
	// if force mode is requested we mock the recipient func to just return anything that goes in
	if force {
		ctx = sub.WithRecipientFunc(ctx, func(ctx context.Context, msg string, rs []string) ([]string, error) {
			return rs, nil
		})
	}

	var content []byte

	// if content is piped to stdin, read and save it
	if ctxutil.IsStdin(ctx) {
		buf := &bytes.Buffer{}

		if written, err := io.Copy(buf, os.Stdin); err != nil {
			return ExitError(ctx, ExitIO, err, "failed to copy after %d bytes: %s", written, err)
		}

		content = buf.Bytes()
	}

	// update to a single YAML entry
	if key != "" {
		return s.insertYAML(ctx, name, key, content, kvps)
	}

	if ctxutil.IsStdin(ctx) {
		return s.insertStdin(ctx, name, content)
	}

	// don't check if it's force anyway
	if !force && s.Store.Exists(ctx, name) && !termio.AskForConfirmation(ctx, fmt.Sprintf("An entry already exists for %s. Overwrite it?", name)) {
		return ExitError(ctx, ExitAborted, nil, "not overwriting your current secret")
	}

	// if multi-line input is requested start an editor
	if multiline && ctxutil.IsInteractive(ctx) {
		return s.insertMultiline(ctx, c, name)
	}

	// if echo mode is requested use a simple string input function
	if echo {
		ctx = termio.WithPassPromptFunc(ctx, func(ctx context.Context, prompt string) (string, error) {
			return termio.AskForString(ctx, prompt, "")
		})
	}

	pw, err := termio.AskForPassword(ctx, name)
	if err != nil {
		return ExitError(ctx, ExitIO, err, "failed to ask for password: %s", err)
	}

	return s.insertSingle(ctx, name, pw, kvps)
}

func (s *Action) insertStdin(ctx context.Context, name string, content []byte) error {
	sec, err := secret.Parse(content)
	if err != nil {
		out.Red(ctx, "WARNING: Invalid YAML: %s", err)
	}
	if err := s.Store.Set(sub.WithReason(ctx, "Read secret from STDIN"), name, sec); err != nil {
		return ExitError(ctx, ExitEncrypt, err, "failed to set '%s': %s", name, err)
	}
	return nil
}

func (s *Action) insertSingle(ctx context.Context, name, pw string, kvps map[string]string) error {
	var sec store.Secret
	if s.Store.Exists(ctx, name) {
		var err error
		sec, err = s.Store.Get(ctx, name)
		if err != nil {
			return ExitError(ctx, ExitDecrypt, err, "failed to decrypt existing secret: %s", err)
		}
	} else {
		sec = &secret.Secret{}
	}
	setMetadata(sec, kvps)
	sec.SetPassword(pw)
	audit.Single(ctx, sec.Password())

	if err := s.Store.Set(sub.WithReason(ctx, "Inserted user supplied password"), name, sec); err != nil {
		return ExitError(ctx, ExitEncrypt, err, "failed to write secret '%s': %s", name, err)
	}
	return nil
}

func (s *Action) insertYAML(ctx context.Context, name, key string, content []byte, kvps map[string]string) error {
	if ctxutil.IsInteractive(ctx) {
		pw, err := termio.AskForString(ctx, name+":"+key, "")
		if err != nil {
			return ExitError(ctx, ExitIO, err, "failed to ask for user input: %s", err)
		}
		content = []byte(pw)
	}

	var sec store.Secret
	if s.Store.Exists(ctx, name) {
		var err error
		sec, err = s.Store.Get(ctx, name)
		if err != nil {
			return ExitError(ctx, ExitEncrypt, err, "failed to set key '%s' of '%s': %s", key, name, err)
		}
	} else {
		sec = &secret.Secret{}
	}
	setMetadata(sec, kvps)
	if err := sec.SetValue(key, string(content)); err != nil {
		return ExitError(ctx, ExitEncrypt, err, "failed to set key '%s' of '%s': %s", key, name, err)
	}
	if err := s.Store.Set(sub.WithReason(ctx, "Inserted YAML value from STDIN"), name, sec); err != nil {
		return ExitError(ctx, ExitEncrypt, err, "failed to set key '%s' of '%s': %s", key, name, err)
	}
	return nil
}

func (s *Action) insertMultiline(ctx context.Context, c *cli.Context, name string) error {
	buf := []byte{}
	if s.Store.Exists(ctx, name) {
		var err error
		sec, err := s.Store.Get(ctx, name)
		if err != nil {
			return ExitError(ctx, ExitDecrypt, err, "failed to decrypt existing secret: %s", err)
		}
		buf, err = sec.Bytes()
		if err != nil {
			return ExitError(ctx, ExitUnknown, err, "failed to encode secret: %s", err)
		}
	}
	ed := editor.Path(c)
	content, err := editor.Invoke(ctx, ed, buf)
	if err != nil {
		return ExitError(ctx, ExitUnknown, err, "failed to start editor: %s", err)
	}
	sec, err := secret.Parse(content)
	if err != nil {
		out.Red(ctx, "WARNING: Invalid YAML: %s", err)
	}
	if err := s.Store.Set(sub.WithReason(ctx, fmt.Sprintf("Inserted user supplied password with %s", ed)), name, sec); err != nil {
		return ExitError(ctx, ExitEncrypt, err, "failed to store secret '%s': %s", name, err)
	}
	return nil
}
