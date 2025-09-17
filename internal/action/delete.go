package action

import (
	"context"
	"errors"
	"fmt"

	"github.com/gopasspw/gopass/internal/action/exit"
	"github.com/gopasspw/gopass/internal/hook"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/store"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/termio"
	"github.com/urfave/cli/v2"
)

// Delete a secret file with its content.
func (s *Action) Delete(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	recursive := c.Bool("recursive")

	name := c.Args().First()
	if name == "" {
		return exit.Error(exit.Usage, nil, "Usage: %s rm name", s.Name)
	}

	if recursive {
		if len(c.Args().Tail()) > 1 {
			return exit.Error(exit.Usage, nil, "Deleting multiple keys is not supported in recursive mode")
		}

		return s.deleteRecursive(ctx, name, c.Bool("force"))
	}

	if s.Store.IsDir(ctx, name) && !s.Store.Exists(ctx, name) {
		return exit.Error(exit.Usage, nil, "Cannot remove %q: Is a directory. Use 'gopass rm -r %s' to delete", name, name)
	}
	// specifying a key is optional.
	key := c.Args().Get(1)

	// multiple secrets, so not a key
	if len(c.Args().Tail()) > 1 {
		key = ""
	}

	// Check for custom commit message
	commitMsg := fmt.Sprintf("Deleted %s", name)
	if key != "" {
		commitMsg = fmt.Sprintf("Deleted key %s from %s", key, name)
	}
	if c.IsSet("commit-message") {
		commitMsg = c.String("commit-message")
	}
	if c.Bool("interactive-commit") {
		commitMsg = ""
	}
	ctx = ctxutil.WithCommitMessage(ctx, commitMsg)

	// multiple secrets, so not a key
	if len(c.Args().Tail()) > 1 {
		key = ""
	}

	names := append([]string{name}, c.Args().Tail()...)

	if key != "" && s.Store.Exists(ctx, key) {
		return exit.Error(exit.Unsupported, nil, "Key %q clashes with a secret of this name, use 'gopass edit %s' to delete", key, name)
	}

	if !s.Store.Exists(ctx, name) {
		return exit.Error(exit.NotFound, nil, "Secret %q does not exist", name)
	}

	if !c.Bool("force") { // don't check if it's force anyway.
		qStr := fmt.Sprintf("☠ Are you sure you would like to delete %q?", names)
		if key != "" {
			qStr = fmt.Sprintf("☠ Are you sure you would like to delete %q from %q?", key, name)
		}
		if (s.Store.Exists(ctx, name) || s.Store.IsDir(ctx, name)) && key == "" && !termio.AskForConfirmation(ctx, qStr) {
			return nil
		}
	}

	// deletes a single key from a YAML doc.
	if key != "" {
		debug.Log("removing key %q from %q", key, name)

		return s.deleteKeyFromYAML(ctx, name, key)
	}

	for _, name := range names {
		debug.Log("removing entry %q", name)
		if err := s.Store.Delete(ctx, name); err != nil {
			return exit.Error(exit.IO, err, "Can not delete %q: %s", name, err)
		}

		if err := hook.InvokeRoot(ctx, "delete.post-hook", name, s.Store); err != nil {
			return exit.Error(exit.Hook, err, "Hook failed for %s: %s", name, err)
		}
	}

	return nil
}

func (s *Action) deleteRecursive(ctx context.Context, name string, force bool) error {
	if !force { // don't check if it's force anyway.
		if (s.Store.Exists(ctx, name) || s.Store.IsDir(ctx, name)) && !termio.AskForConfirmation(ctx, fmt.Sprintf("☠ Are you sure you would like to recursively delete %q?", name)) {
			return nil
		}
	}

	debug.Log("pruning %q", name)
	if err := s.Store.Prune(ctx, name); err != nil {
		return exit.Error(exit.Unknown, err, "failed to prune %q: %s", name, err)
	}
	debug.Log("pruned %q", name)

	return nil
}

// deleteKeyFromYAML deletes a single key from YAML.
func (s *Action) deleteKeyFromYAML(ctx context.Context, name, key string) error {
	sec, err := s.Store.Get(ctx, name)
	if err != nil {
		return exit.Error(exit.IO, err, "Can not delete key %q from %q: %s", key, name, err)
	}

	sec.Del(key)

	if err := s.Store.Set(ctx, name, sec); err != nil {
		if !errors.Is(err, store.ErrMeaninglessWrite) {
			return exit.Error(exit.IO, err, "Can not delete key %q from %q: %s", key, name, err)
		}
		out.Warningf(ctx, "No need to write: the YAML file does't seem to have the key to be deleted")
	}

	return hook.Invoke(ctx, "delete.post-hook", name, key)
}
