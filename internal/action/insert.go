package action

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/gopasspw/gopass/internal/secrets"
	"io"
	"strings"

	"github.com/gopasspw/gopass/internal/debug"

	"github.com/gopasspw/gopass/internal/audit"
	"github.com/gopasspw/gopass/internal/editor"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/termio"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/gopass"
	"github.com/gopasspw/gopass/pkg/gopass/secret"
	"github.com/gopasspw/gopass/pkg/gopass/secret/secparse"

	"github.com/urfave/cli/v2"
)

// Insert a string as content to a secret file
func (s *Action) Insert(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	echo := c.Bool("echo")
	multiline := c.Bool("multiline")
	force := c.Bool("force")
	append := c.Bool("append")

	args, kvps := parseArgs(c)
	name := args.Get(0)
	key := args.Get(1)

	if name == "" {
		return ExitError(ExitNoName, nil, "Usage: %s insert name", s.Name)
	}

	return s.insert(ctx, c, name, key, echo, multiline, force, append, kvps)
}

func (s *Action) insert(ctx context.Context, c *cli.Context, name, key string, echo, multiline, force, append bool, kvps map[string]string) error {
	var content []byte

	// if content is piped to stdin, read and save it
	if ctxutil.IsStdin(ctx) {
		buf := &bytes.Buffer{}

		if written, err := io.Copy(buf, stdin); err != nil {
			return ExitError(ExitIO, err, "failed to copy after %d bytes: %s", written, err)
		}

		content = buf.Bytes()
	}

	// update to a single YAML entry
	if key != "" {
		debug.Log("inserting a single key: ", key)
		return s.insertYAML(ctx, name, key, content, kvps)
	}

	if ctxutil.IsStdin(ctx) {
		if !force && !append && s.Store.Exists(ctx, name) {
			return ExitError(ExitAborted, nil, "not overwriting your current secret")
		}
		return s.insertStdin(ctx, name, content, append)
	}

	// don't check if it's force anyway
	if !force && s.Store.Exists(ctx, name) && !termio.AskForConfirmation(ctx, fmt.Sprintf("An entry already exists for %s. Overwrite it?", name)) {
		return ExitError(ExitAborted, nil, "not overwriting your current secret")
	}

	// if multi-line input is requested start an editor
	if multiline && ctxutil.IsInteractive(ctx) {
		debug.Log("inserting multi-line input")
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
		return ExitError(ExitIO, err, "failed to ask for password: %s", err)
	}

	return s.insertSingle(ctx, name, pw, kvps)
}

func (s *Action) insertStdin(ctx context.Context, name string, content []byte, appendTo bool) error {
	debug.Log("calling insertStdin on name ", name)
	var sec *secret.MIME
	if appendTo && s.Store.Exists(ctx, name) {
		eSec, err := s.Store.Get(ctx, name)
		if err != nil {
			return ExitError(ExitDecrypt, err, "failed to decrypt existing secret: %s", err)
		}
		sec = eSec.MIME()
		sec.Write(content)
	} else {
		debug.Log("creating a new secret ", name)
		plain, err := secparse.Parse(content)
		if err != nil {
			return ExitError(ExitAborted, err, "failed to parse secret from stdin: %s", err)
		}
		switch plain.(type) {
		// if we parsed it as a KV, we can easily convert it to Mime if Mime is enabled.
		case *secrets.KV:
			content = checkMime(content)
			tmp, err := secparse.Parse(content)
			if err == nil {
				plain = tmp
			}
		}
		sec = plain.MIME()
	}
	if err := s.Store.Set(ctxutil.WithCommitMessage(ctx, "Read secret from STDIN"), name, sec); err != nil {
		return ExitError(ExitEncrypt, err, "failed to set '%s': %s", name, err)
	}
	return nil
}

func (s *Action) insertSingle(ctx context.Context, name, pw string, kvps map[string]string) error {
	sec := secret.New()
	if s.Store.Exists(ctx, name) {
		gs, err := s.Store.Get(ctx, name)
		if err != nil {
			return ExitError(ExitDecrypt, err, "failed to decrypt existing secret: %s", err)
		}
		sec = gs.MIME()
	} else {
		if content, found := s.renderTemplate(ctx, name, []byte(pw)); found {
			nSec, err := secparse.Parse(content)
			if err == nil {
				sec = nSec.MIME()
			}
		}
	}

	setMetadata(sec, kvps)

	// we only update the pw if the kvps were not set or if it's non-empty, because otherwise we were updating the kvps
	if pw != "" || len(kvps) == 0 {
		sec.Set("password", pw)
		audit.Single(ctx, pw)
	}

	if err := s.Store.Set(ctxutil.WithCommitMessage(ctx, "Inserted user supplied password"), name, sec); err != nil {
		return ExitError(ExitEncrypt, err, "failed to write secret '%s': %s", name, err)
	}
	return nil
}

func (s *Action) insertYAML(ctx context.Context, name, key string, content []byte, kvps map[string]string) error {
	if ctxutil.IsInteractive(ctx) {
		pw, err := termio.AskForString(ctx, name+":"+key, "")
		if err != nil {
			return ExitError(ExitIO, err, "failed to ask for user input: %s", err)
		}
		content = []byte(pw)
	}

	var sec gopass.Secret
	if s.Store.Exists(ctx, name) {
		var err error
		sec, err = s.Store.Get(ctx, name)
		if err != nil {
			return ExitError(ExitEncrypt, err, "failed to set key '%s' of '%s': %s", key, name, err)
		}
	} else {
		sec = secret.New()
	}
	setMetadata(sec, kvps)
	sec.Set(key, string(content))
	if err := s.Store.Set(ctxutil.WithCommitMessage(ctx, "Inserted YAML value from STDIN"), name, sec); err != nil {
		return ExitError(ExitEncrypt, err, "failed to set key '%s' of '%s': %s", key, name, err)
	}
	return nil
}

func (s *Action) insertMultiline(ctx context.Context, c *cli.Context, name string) error {
	buf := []byte{}
	if s.Store.Exists(ctx, name) {
		var err error
		sec, err := s.Store.Get(ctx, name)
		if err != nil {
			return ExitError(ExitDecrypt, err, "failed to decrypt existing secret: %s", err)
		}
		buf = sec.Bytes()
	}
	ed := editor.Path(c)
	content, err := editor.Invoke(ctx, ed, buf)
	content = checkMime(content)
	if err != nil {
		return ExitError(ExitUnknown, err, "failed to start editor: %s", err)
	}
	sec, err := secparse.Parse(content)
	if err != nil {
		out.Error(ctx, "WARNING: Invalid YAML: %s", err)
	}
	if err := s.Store.Set(ctxutil.WithCommitMessage(ctx, fmt.Sprintf("Inserted user supplied password with %s", ed)), name, sec); err != nil {
		return ExitError(ExitEncrypt, err, "failed to store secret '%s': %s", name, err)
	}
	return nil
}

func checkMime(content []byte) []byte {
	// we return the content if not in WriteMime mode
	if !secret.WriteMIME {
		return content
	}
	scanner := bufio.NewScanner(bytes.NewReader(content))
	if !scanner.Scan() {
		debug.Log("checkMime reached unexpected end of content on first Scan")
		return content
	}
	magic := scanner.Text()
	// if there is already the magic, no need to add it
	if magic == secret.Ident {
		return content
	}
	// if there is no magic, let's add it
	prepend := []byte(secret.Ident + "\n")
	if strings.HasSuffix(strings.ToLower(magic), "password:") {
		return append(prepend, content...)
	}

	prepend = append(prepend, []byte("Password: ")...)
	return append(prepend, content...)
}
