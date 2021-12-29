package action

import (
	"context"
	"fmt"

	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/termio"
	"github.com/urfave/cli/v2"
)

// Delete a secret file with its content.
func (s *Action) Delete(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	force := c.Bool("force")
	recursive := c.Bool("recursive")

	name := c.Args().First()
	if name == "" {
		return ExitError(ExitUsage, nil, "Usage: %s rm name", s.Name)
	}

	if !recursive && s.Store.IsDir(ctx, name) && !s.Store.Exists(ctx, name) {
		return ExitError(ExitUsage, nil, "Cannot remove %q: Is a directory. Use 'gopass rm -r %s' to delete", name, name)
	}

	// specifying a key is optional.
	key := c.Args().Get(1)

	if recursive && key != "" {
		return ExitError(ExitUsage, nil, "Can not use -r with a key. Invoke delete either with a key or with -r")
	}

	if !force { // don't check if it's force anyway.
		recStr := ""
		if recursive {
			recStr = "recursively "
		}
		if (s.Store.Exists(ctx, name) || s.Store.IsDir(ctx, name)) && key == "" && !termio.AskForConfirmation(ctx, fmt.Sprintf("☠ Are you sure you would like to %sdelete %s?", recStr, name)) {
			return nil
		}
	}

	if recursive && key == "" {
		debug.Log("pruning %q", name)
		if err := s.Store.Prune(ctx, name); err != nil {
			return ExitError(ExitUnknown, err, "failed to prune %q: %s", name, err)
		}
		debug.Log("pruned %q", name)
		return nil
	}

	// deletes a single key from a YAML doc.
	if key != "" {
		debug.Log("removing key %q from %q", key, name)
		return s.deleteKeyFromYAML(ctx, name, key)
	}

	debug.Log("removing entry %q", name)
	if err := s.Store.Delete(ctx, name); err != nil {
		return ExitError(ExitIO, err, "Can not delete %q: %s", name, err)
	}
	return nil
}

// deleteKeyFromYAML deletes a single key from YAML.
func (s *Action) deleteKeyFromYAML(ctx context.Context, name, key string) error {
	sec, err := s.Store.Get(ctx, name)
	if err != nil {
		return ExitError(ExitIO, err, "Can not delete key %q from %q: %s", key, name, err)
	}
	sec.Del(key)
	if err := s.Store.Set(ctxutil.WithCommitMessage(ctx, "Updated Key"), name, sec); err != nil {
		return ExitError(ExitIO, err, "Can not delete key %q from %q: %s", key, name, err)
	}
	return nil
}
