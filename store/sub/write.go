package sub

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/store"
	"github.com/pkg/errors"
)

// Set encodes and write the ciphertext of one entry to disk
func (s *Store) Set(name string, content []byte, reason string) error {
	return s.SetConfirm(name, content, reason, nil)
}

// SetPassword update a password in an already existing entry on the disk
func (s *Store) SetPassword(name string, password []byte) error {
	var err error
	body, err := s.GetBody(name)
	if err != nil && err != store.ErrNoBody {
		return errors.Wrapf(err, "failed to get existing secret")
	}
	first := append(password, '\n')
	return s.SetConfirm(name, append(first, body...), fmt.Sprintf("Updated password in %s", name), nil)
}

// SetConfirm encodes and writes the cipertext of one entry to disk. This
// method can be passed a callback to confirm the recipients immedeately
// before encryption.
func (s *Store) SetConfirm(name string, content []byte, reason string, cb store.RecipientCallback) error {
	p := s.passfile(name)

	if !strings.HasPrefix(p, s.path) {
		return store.ErrSneaky
	}

	if s.IsDir(name) {
		return errors.Errorf("a folder named %s already exists", name)
	}

	recipients, err := s.useableKeys()
	if err != nil {
		return errors.Wrapf(err, "failed to list useable keys")
	}

	// confirm recipients
	if cb != nil {
		newRecipients, err := cb(name, recipients)
		if err != nil {
			return errors.Wrapf(err, "user aborted")
		}
		recipients = newRecipients
	}

	if err := s.gpg.Encrypt(p, content, recipients); err != nil {
		return store.ErrEncrypt
	}

	if err := s.gitAdd(p); err != nil {
		if errors.Cause(err) == store.ErrGitNotInit {
			return nil
		}
		return errors.Wrapf(err, "failed to add '%s' to git", p)
	}

	if err := s.gitCommit(fmt.Sprintf("Save secret to %s: %s", name, reason)); err != nil {
		if errors.Cause(err) == store.ErrGitNotInit {
			return nil
		}
		return errors.Wrapf(err, "failed to commit changes to git")
	}

	if !s.autoSync {
		return nil
	}

	if err := s.gitPush("", ""); err != nil {
		if errors.Cause(err) == store.ErrGitNotInit {
			msg := "Warning: git is not initialized for this store. Ignoring auto-push option\n" +
				"Run: gopass git init"
			fmt.Println(color.RedString(msg))
			return nil
		}
		if errors.Cause(err) == store.ErrGitNoRemote {
			msg := "Warning: git has not remote. Ignoring auto-push option\n" +
				"Run: gopass git remote add origin ..."
			fmt.Println(color.YellowString(msg))
			return nil
		}
		return errors.Wrapf(err, "failed to push to git remote")
	}
	fmt.Println(color.GreenString("Pushed changes to git remote"))
	return nil
}
