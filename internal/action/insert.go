package action

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/gopasspw/gopass/internal/action/exit"
	"github.com/gopasspw/gopass/internal/audit"
	"github.com/gopasspw/gopass/internal/editor"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/store"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/gopass"
	"github.com/gopasspw/gopass/pkg/gopass/secrets"
	"github.com/gopasspw/gopass/pkg/termio"
	"github.com/urfave/cli/v2"
)

// Insert a string as content to a secret file.
func (s *Action) Insert(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	echo := c.Bool("echo")
	multiline := c.Bool("multiline")
	force := c.Bool("force")
	appending := c.Bool("append")

	args, kvps := parseArgs(c)
	name := args.Get(0)
	key := args.Get(1)

	if name == "" {
		return exit.Error(exit.NoName, nil, "Usage: %s insert name", s.Name)
	}

	return s.insert(ctx, c, name, key, echo, multiline, force, appending, kvps)
}

func (s *Action) insert(ctx context.Context, c *cli.Context, name, key string, echo, multiline, force, appending bool, kvps map[string]string) error {
	var content []byte

	// if content is piped to stdin, read and save it.
	if ctxutil.IsStdin(ctx) {
		buf := &bytes.Buffer{}

		if written, err := io.Copy(buf, stdin); err != nil {
			return exit.Error(exit.IO, err, "failed to copy after %d bytes: %s", written, err)
		}

		content = buf.Bytes()
	}

	// update to a single YAML entry.
	if key != "" {
		return s.insertYAML(ctx, name, key, content, kvps)
	}

	if ctxutil.IsStdin(ctx) {
		if !force && !appending && s.Store.Exists(ctx, name) {
			return exit.Error(exit.Aborted, nil, "not overwriting your current secret")
		}

		return s.insertStdin(ctx, name, content, appending)
	}

	// don't check if it's force anyway.
	if !force && s.Store.Exists(ctx, name) && !termio.AskForConfirmation(ctx, fmt.Sprintf("An entry already exists for %s. Overwrite it?", name)) {
		return exit.Error(exit.Aborted, nil, "not overwriting your current secret")
	}

	// if multi-line input is requested start an editor.
	if multiline && ctxutil.IsInteractive(ctx) {
		return s.insertMultiline(ctx, c, name)
	}

	// if echo mode is requested use a simple string input function.
	if echo {
		ctx = termio.WithPassPromptFunc(ctx, func(ctx context.Context, prompt string) (string, error) {
			return termio.AskForString(ctx, prompt, "")
		})
	}

	pw, err := termio.AskForPassword(ctx, fmt.Sprintf("password for %s", name), true)
	if err != nil {
		return exit.Error(exit.IO, err, "failed to ask for password: %s", err)
	}

	return s.insertSingle(ctx, name, pw, kvps)
}

func (s *Action) insertStdin(ctx context.Context, name string, content []byte, appendTo bool) error {
	var sec gopass.Secret = secrets.ParseAKV(content)

	if appendTo && s.Store.Exists(ctx, name) {
		var err error
		sec, err = s.insertStdinAppend(ctx, name, content)
		if err != nil {
			return err
		}
	}

	if err := s.Store.Set(ctxutil.WithCommitMessage(ctx, "Read secret from STDIN"), name, sec); err != nil {
		if !errors.Is(err, store.ErrMeaninglessWrite) {
			return exit.Error(exit.Encrypt, err, "failed to set %q: %s", name, err)
		}
		out.Warningf(ctx, "No need to write: the secret is already there and with the right value")
	}

	return nil
}

func (s *Action) insertStdinAppend(ctx context.Context, name string, content []byte) (gopass.Secret, error) {
	eSec, err := s.Store.Get(ctx, name)
	if err != nil {
		return nil, exit.Error(exit.Decrypt, err, "failed to decrypt existing secret: %s", err)
	}

	secW, ok := eSec.(io.Writer)
	if !ok {
		return nil, fmt.Errorf("%T is not an io.Writer", eSec)
	}

	if _, err := secW.Write(content); err != nil {
		return nil, exit.Error(exit.Encrypt, err, "failed to write %q: %q", content, err)
	}

	debug.Log("wrote to secretWriter")

	return eSec, nil
}

func (s *Action) insertSingle(ctx context.Context, name, pw string, kvps map[string]string) error {
	sec, err := s.insertGetSecret(ctx, name, pw)
	if err != nil {
		return err
	}

	setMetadata(sec, kvps)

	// we only update the pw if the kvps were not set or if it's non-empty, because otherwise we were updating the kvps.
	if pw != "" || len(kvps) == 0 {
		sec.SetPassword(pw)
		audit.Single(ctx, pw)
	}

	if err := s.Store.Set(ctxutil.WithCommitMessage(ctx, "Inserted user supplied password"), name, sec); err != nil {
		if !errors.Is(err, store.ErrMeaninglessWrite) {
			return exit.Error(exit.Encrypt, err, "failed to write secret %q: %s", name, err)
		}
		out.Warningf(ctx, "No need to write: the secret is already there and with the right value")
	}

	return nil
}

func (s *Action) insertGetSecret(ctx context.Context, name, pw string) (gopass.Secret, error) {
	if s.Store.Exists(ctx, name) {
		sec, err := s.Store.Get(ctx, name)
		if err != nil {
			return nil, exit.Error(exit.Decrypt, err, "failed to decrypt existing secret: %s", err)
		}

		return sec, nil
	}

	content, found := s.renderTemplate(ctx, name, []byte(pw))
	// no template found
	if !found {
		return secrets.New(), nil
	}

	// render template into a new secret
	sec := secrets.NewAKV()
	if _, err := sec.Write(content); err != nil {
		debug.Log("failed to handle template: %s", err)

		return secrets.New(), nil
	}

	return sec, nil
}

// insertYAML will overwrite existing keys.
func (s *Action) insertYAML(ctx context.Context, name, key string, content []byte, kvps map[string]string) error {
	debug.Log("insertYAML: %s - %s -> %s", name, key, content)
	if ctxutil.IsInteractive(ctx) {
		pw, err := termio.AskForString(ctx, name+":"+key, "")
		if err != nil {
			return exit.Error(exit.IO, err, "failed to ask for user input: %s", err)
		}
		content = []byte(pw)
	}

	var sec gopass.Secret
	if s.Store.Exists(ctx, name) {
		var err error
		sec, err = s.Store.Get(ctx, name)
		if err != nil {
			return exit.Error(exit.Encrypt, err, "failed to set key %q of %q: %s", key, name, err)
		}
		debug.Log("using existing secret %s", name)
	} else {
		sec = secrets.New()
		debug.Log("creating new secret %s", name)
	}

	setMetadata(sec, kvps)

	debug.Log("setting %s to %s", key, string(content))
	if err := sec.Set(key, string(content)); err != nil {
		return exit.Error(exit.Usage, err, "failed set key %q of %q: %q", key, name, err)
	}

	if err := s.Store.Set(ctxutil.WithCommitMessage(ctx, "Inserted YAML value from STDIN"), name, sec); err != nil {
		if !errors.Is(err, store.ErrMeaninglessWrite) {
			return exit.Error(exit.Encrypt, err, "failed to set key %q of %q: %s", key, name, err)
		}
		out.Warningf(ctx, "No need to write: the secret is already there and with the right value")
	}

	return nil
}

func (s *Action) insertMultiline(ctx context.Context, c *cli.Context, name string) error {
	buf := []byte{}
	if s.Store.Exists(ctx, name) {
		var err error
		sec, err := s.Store.Get(ctx, name)
		if err != nil {
			return exit.Error(exit.Decrypt, err, "failed to decrypt existing secret: %s", err)
		}
		buf = sec.Bytes()
	}
	ed := editor.Path(c)
	content, err := editor.Invoke(ctx, ed, buf)
	if err != nil {
		return exit.Error(exit.Unknown, err, "failed to start editor: %s", err)
	}

	sec := secrets.NewAKV()
	n, err := sec.Write(content)
	if err != nil || n < 0 {
		out.Errorf(ctx, "WARNING: Invalid secret: %s of len %d", err, n)
	}

	if err := s.Store.Set(ctxutil.WithCommitMessage(ctx, fmt.Sprintf("Inserted user supplied password with %s", ed)), name, sec); err != nil {
		if !errors.Is(err, store.ErrMeaninglessWrite) {
			return exit.Error(exit.Encrypt, err, "failed to store secret %q: %s", name, err)
		}
		out.Warningf(ctx, "No need to write: the secret is already there and with the right value")
	}

	return nil
}
