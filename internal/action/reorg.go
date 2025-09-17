package action

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/gopasspw/gopass/internal/action/exit"
	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/editor"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/termio"
	"github.com/urfave/cli/v2"
)

// Reorg is the action that allows to reorganize a part of the store.
func (s *Action) Reorg(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)

	prefix := c.Args().Get(0)

	// list secrets
	secrets, err := s.Store.List(ctx, 0)
	if err != nil {
		return exit.Error(exit.List, err, "failed to list secrets: %s", err)
	}

	// filter by prefix
	var initialSecrets []string
	if prefix != "" {
		for _, secret := range secrets {
			if strings.HasPrefix(secret, prefix) {
				initialSecrets = append(initialSecrets, secret)
			}
		}
	} else {
		initialSecrets = secrets
	}

	if len(initialSecrets) == 0 {
		out.Printf(ctx, "No secrets found to reorganize.")

		return nil
	}

	// get initial content
	initialContent := []byte(strings.Join(initialSecrets, "\n") + "\n")

	// open editor
	editorPath := editor.Path(c)
	modifiedContent, err := editor.Invoke(ctx, editorPath, initialContent)
	if err != nil {
		return exit.Error(exit.Unknown, err, "failed to invoke editor: %s", err)
	}

	// parse modified secrets
	var modifiedSecrets []string
	scanner := bufio.NewScanner(bytes.NewReader(modifiedContent))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			modifiedSecrets = append(modifiedSecrets, line)
		}
	}

	return s.ReorgAfterEdit(ctx, initialSecrets, modifiedSecrets)
}

// ReorgAfterEdit performs the reorganization after the user has edited the list of secrets.
func (s *Action) ReorgAfterEdit(ctx context.Context, initialSecrets, modifiedSecrets []string) error {
	if len(initialSecrets) != len(modifiedSecrets) {
		return exit.Error(exit.Usage, nil, "number of secrets must not be changed. Aborting.")
	}

	// calculate moves
	moves := make(map[string]string)
	for i, oldSecret := range initialSecrets {
		newSecret := modifiedSecrets[i]
		if oldSecret != newSecret {
			moves[oldSecret] = newSecret
		}
	}

	if len(moves) == 0 {
		out.Printf(ctx, "No changes detected.")

		return nil
	}

	// validate moves
	if err := s.validateMoves(ctx, moves); err != nil {
		return exit.Error(exit.Usage, err, "failed to validate moves: %s", err)
	}

	// display diff and ask for confirmation
	out.Printf(ctx, "The following moves will be performed:")
	for from, to := range moves {
		out.Printf(ctx, "  - %s -> %s", from, to)
	}

	if !termio.AskForConfirmation(ctx, "Do you want to proceed?") {
		return exit.Error(exit.Aborted, nil, "user aborted")
	}

	// disable automatic commits
	ctx = ctxutil.WithGitCommit(ctx, false)

	// execute moves
	for from, to := range moves {
		if err := s.Store.Move(ctx, from, to); err != nil {
			return exit.Error(exit.Unknown, err, "failed to move %s to %s: %s", from, to, err)
		}
		out.Printf(ctx, "Moved %s to %s", from, to)
	}

	// get storage from the first move
	var storage backend.Storage
	for from := range moves {
		storage = s.Store.Storage(ctx, from)

		break
	}
	if storage == nil {
		return exit.Error(exit.Git, nil, "failed to get storage backend")
	}

	// commit changes
	if err := storage.TryCommit(ctx, "Reorganized secrets"); err != nil {
		return exit.Error(exit.Git, err, "failed to commit changes: %s", err)
	}

	out.Printf(ctx, "Successfully reorganized secrets.")

	return nil
}

func (s *Action) validateMoves(ctx context.Context, moves map[string]string) error {
	destinations := make(map[string]string, len(moves))
	for from, to := range moves {
		// check for duplicate destinations
		if existingFrom, found := destinations[to]; found {
			return fmt.Errorf("duplicate destination %q for %q and %q", to, from, existingFrom)
		}
		destinations[to] = from

		// check for cross-mount moves
		fromMount := s.Store.MountPoint(from)
		toMount := s.Store.MountPoint(to)
		if fromMount != toMount {
			return fmt.Errorf("moving secrets across mounts is not supported: %s (%s) -> %s (%s)", from, fromMount, to, toMount)
		}
	}

	return nil
}
