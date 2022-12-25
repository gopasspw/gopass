package action

import (
	"bytes"
	"fmt"
	"time"
	"errors"

	"github.com/gopasspw/gopass/internal/action/exit"
	"github.com/gopasspw/gopass/internal/audit"
	"github.com/gopasspw/gopass/internal/editor"
	"github.com/gopasspw/gopass/internal/store"
	"github.com/gopasspw/gopass/internal/queue"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/gopass/secrets"
	"github.com/urfave/cli/v2"
)

// Merge implements the merge subcommand that allows merging multiple entries.
func (s *Action) Merge(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	to := c.Args().First()
	from := c.Args().Tail()

	if to == "" {
		return exit.Error(exit.Usage, nil, "usage: %s merge <to> <from> [<from>]", s.Name)
	}

	if len(from) < 1 {
		return exit.Error(exit.Usage, nil, "usage: %s merge <to> <from> [<from>]", s.Name)
	}

	ed := editor.Path(c)

	content := &bytes.Buffer{}
	for _, k := range c.Args().Slice() {
		if !s.Store.Exists(ctx, k) {
			continue
		}
		sec, err := s.Store.Get(ctxutil.WithShowParsing(ctx, false), k)
		if err != nil {
			return exit.Error(exit.Decrypt, err, "failed to decrypt: %s: %s", k, err)
		}

		_, err = content.WriteString("\n# Secret: " + k + "\n")
		if err != nil {
			return exit.Error(exit.Unknown, err, "failed to write: %s", err)
		}

		_, err = content.Write(sec.Bytes())
		if err != nil {
			return exit.Error(exit.Unknown, err, "failed to write: %s", err)
		}
	}

	newContent := content.Bytes()
	if !c.Bool("force") {
		var err error
		// invoke the editor to let the user edit the content
		newContent, err = editor.Invoke(ctx, ed, content.Bytes())
		if err != nil {
			return exit.Error(exit.Unknown, err, "failed to invoke editor: %s", err)
		}

		// If content is equal, nothing changed, exiting
		if bytes.Equal(content.Bytes(), newContent) {
			return nil
		}
	}

	nSec := secrets.ParseAKV(newContent)

	// if the secret has a password, we check it's strength
	if pw := nSec.Password(); pw != "" && !c.Bool("force") {
		audit.Single(ctx, pw)
	}

	// write result (back) to store
	if err := s.Store.Set(ctxutil.WithCommitMessage(ctx, fmt.Sprintf("Merged %+v", c.Args().Slice())), to, nSec); err != nil {
		if errors.Is(err, store.ErrMeaninglessWrite) {
			out.Warningf(ctx, "No need to write: the secret is already there and with the right value")
		} else {
			return exit.Error(exit.Encrypt, err, "failed to encrypt secret %s: %s", to, err)
		}		
	}

	if !c.Bool("delete") {
		return nil
	}

	// wait until the previous commit is done
	// This wouldn't be necessary if we could handle merging and deleting
	// in a single commit, but then we'd need to expose additional implementation
	// details of the underlying VCS. Or create some kind of transaction on top
	// of the Git wrapper.
	if err := queue.GetQueue(ctx).Idle(time.Minute); err != nil {
		return err
	}

	for _, old := range from {
		if !s.Store.Exists(ctx, old) {
			continue
		}
		debug.Log("deleting merged entry %s", old)
		if err := s.Store.Delete(ctx, old); err != nil {
			return exit.Error(exit.Unknown, err, "failed to delete %s: %s", old, err)
		}
	}

	return nil
}
