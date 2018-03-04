package sub

import (
	"context"
	"fmt"
	"strings"

	"github.com/justwatchcom/gopass/store"
	"github.com/justwatchcom/gopass/store/secret"
	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/pkg/errors"
)

// Set encodes and writes the cipertext of one entry to disk. This
// method can be passed a callback to confirm the recipients immediately
// before encryption.
func (s *Store) Set(ctx context.Context, name string, sec *secret.Secret) error {
	if strings.Contains(name, "//") {
		return errors.Errorf("invalid secret name: %s", name)
	}

	p := s.passfile(name)

	if s.IsDir(ctx, name) {
		return errors.Errorf("a folder named %s already exists", name)
	}

	recipients, err := s.useableKeys(ctx, name)
	if err != nil {
		return errors.Wrapf(err, "failed to list useable keys for '%s'", p)
	}

	// confirm recipients
	newRecipients, err := GetRecipientFunc(ctx)(ctx, name, recipients)
	if err != nil {
		return errors.Wrapf(err, "user aborted")
	}
	recipients = newRecipients

	buf, err := sec.Bytes()
	if err != nil {
		return errors.Wrapf(err, "failed to encode secret")
	}

	ciphertext, err := s.crypto.Encrypt(ctx, buf, recipients)
	if err != nil {
		out.Debug(ctx, "Failed encrypt secret: %s", err)
		return store.ErrEncrypt
	}

	if err := s.store.Set(ctx, p, ciphertext); err != nil {
		return errors.Wrapf(err, "failed to write secret")
	}

	if err := s.sync.Add(ctx, p); err != nil {
		if errors.Cause(err) == store.ErrGitNotInit {
			return nil
		}
		return errors.Wrapf(err, "failed to add '%s' to git", p)
	}

	if !ctxutil.IsGitCommit(ctx) {
		return nil
	}

	return s.gitCommitAndPush(ctx, name)
}

func (s *Store) gitCommitAndPush(ctx context.Context, name string) error {
	if err := s.sync.Commit(ctx, fmt.Sprintf("Save secret to %s: %s", name, GetReason(ctx))); err != nil {
		if errors.Cause(err) == store.ErrGitNotInit {
			return nil
		}
		return errors.Wrapf(err, "failed to commit changes to git")
	}

	if !IsAutoSync(ctx) {
		return nil
	}

	if err := s.sync.Push(ctx, "", ""); err != nil {
		if errors.Cause(err) == store.ErrGitNotInit {
			msg := "Warning: git is not initialized for this store. Ignoring auto-push option\n" +
				"Run: gopass git init"
			out.Red(ctx, msg)
			return nil
		}
		if errors.Cause(err) == store.ErrGitNoRemote {
			msg := "Warning: git has no remote. Ignoring auto-push option\n" +
				"Run: gopass git remote add origin ..."
			out.Yellow(ctx, msg)
			return nil
		}
		return errors.Wrapf(err, "failed to push to git remote")
	}
	out.Green(ctx, "Pushed changes to git remote")
	return nil
}
