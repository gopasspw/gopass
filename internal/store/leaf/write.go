package leaf

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/queue"
	"github.com/gopasspw/gopass/internal/store"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/gopass"
)

// Set encodes and writes the cipertext of one entry to disk.
func (s *Store) Set(ctx context.Context, name string, sec gopass.Byter) error {
	if strings.Contains(name, "//") {
		return fmt.Errorf("invalid secret name: %s", name)
	}

	if config.FromContext(ctx).GetM(s.alias, "core.readonly") == "true" {
		return fmt.Errorf("writing to %s is disabled by `core.readonly`.", s.alias)
	}

	p := s.Passfile(name)

	recipients, err := s.useableKeys(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to list useable keys for %q: %w", p, err)
	}

	// make sure the encryptor can decrypt later
	recipients = s.ensureOurKeyID(ctx, recipients)

	ciphertext, err := s.crypto.Encrypt(ctx, sec.Bytes(), recipients)
	if err != nil {
		debug.Log("Failed encrypt secret: %s", err)

		return store.ErrEncrypt
	}

	if err := s.storage.Set(ctx, p, ciphertext); err != nil {
		return fmt.Errorf("failed to write secret: %w", err)
	}

	// It is not possible to perform concurrent git add and git commit commands
	// so we need to skip this step when using concurrency and perform them
	// at the end of the batch processing.
	if IsNoGitOps(ctx) {
		debug.Log("sub.Set(%s) - skipping git ops (disabled)")

		return nil
	}

	if err := s.storage.Add(ctx, p); err != nil {
		if errors.Is(err, store.ErrGitNotInit) {
			return nil
		}

		return fmt.Errorf("failed to add %q to git: %w", p, err)
	}

	if !ctxutil.IsGitCommit(ctx) {
		return nil
	}

	// try to enqueue this task, if the queue is not available
	// it will return the task and we will execute it inline
	t := queue.GetQueue(ctx).Add(func(_ context.Context) (context.Context, error) {
		return nil, s.gitCommitAndPush(ctx, name)
	})

	ctx, err = t(ctx)
	return err
}

func (s *Store) gitCommitAndPush(ctx context.Context, name string) error {
	if err := s.storage.Commit(ctx, fmt.Sprintf("Save secret to %s: %s", name, ctxutil.GetCommitMessage(ctx))); err != nil {
		switch {
		case errors.Is(err, store.ErrGitNotInit):
			debug.Log("commitAndPush - skipping git commit - git not initialized")
		case errors.Is(err, store.ErrGitNothingToCommit):
			debug.Log("commitAndPush - skipping git commit - nothing to commit")
		default:
			return fmt.Errorf("failed to commit changes to git: %w", err)
		}
	}

	debug.Log("syncing with remote ...")

	if err := s.storage.Push(ctx, "", ""); err != nil {
		if errors.Is(err, store.ErrGitNotInit) {
			msg := "Warning: git is not initialized for this.storage. Ignoring auto-push option\n" +
				"Run: gopass git init"
			out.Errorf(ctx, msg)

			return nil
		}

		if errors.Is(err, store.ErrGitNoRemote) {
			msg := "Warning: git has no remote. Ignoring auto-push option\n" +
				"Run: gopass git remote add origin ..."
			debug.Log(msg)

			return nil
		}

		return fmt.Errorf("failed to push to git remote: %w", err)
	}

	debug.Log("synced with remote")

	return nil
}
