package leaf

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/recipients"
	"github.com/gopasspw/gopass/internal/store"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
)

const (
	keyDir    = ".public-keys"
	oldKeyDir = ".gpg-keys"
)

// Recipients returns the list of recipients of this store.
func (s *Store) Recipients(ctx context.Context) []string {
	rs, err := s.GetRecipients(ctx, "")
	if err != nil {
		out.Errorf(ctx, "failed to read recipient list: %s", err)
	}
	return rs
}

// RecipientsTree returns a mapping of secrets to recipients.
func (s *Store) RecipientsTree(ctx context.Context) map[string][]string {
	idfs := s.idFiles(ctx)
	out := make(map[string][]string, len(idfs))

	root := s.Recipients(ctx)
	for _, idf := range idfs {
		if strings.HasPrefix(idf, ".") {
			continue
		}
		srs, err := s.getRecipients(ctx, idf)
		if err != nil {
			debug.Log("failed to list recipients: %s", err)
			continue
		}
		if cmp.Equal(out[""], srs) {
			debug.Log("root recipients equal secret recipients from %s", idf)
			continue
		}
		dir := filepath.Dir(idf)
		debug.Log("adding recipients %+v for %s", srs, dir)
		out[dir] = srs
	}
	out[""] = root
	return out
}

// AddRecipient adds a new recipient to the list.
func (s *Store) AddRecipient(ctx context.Context, id string) error {
	rs, err := s.GetRecipients(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to read recipient list: %w", err)
	}

	debug.Log("new recipient: %q - existing: %+v", id, rs)
	for _, k := range rs {
		if k == id {
			return fmt.Errorf("recipient already in store")
		}
	}

	rs = append(rs, id)

	if err := s.saveRecipients(ctx, rs, "Added Recipient "+id); err != nil {
		return fmt.Errorf("failed to save recipients: %w", err)
	}

	out.Printf(ctx, "Reencrypting existing secrets. This may take some time ...")
	return s.reencrypt(ctxutil.WithCommitMessage(ctx, "Added Recipient "+id))
}

// SaveRecipients persists the current recipients on disk.
func (s *Store) SaveRecipients(ctx context.Context) error {
	rs, err := s.GetRecipients(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to get recipients: %w", err)
	}
	return s.saveRecipients(ctx, rs, "Save Recipients")
}

// SetRecipients will update the stored recipients and the associated checksum.
func (s *Store) SetRecipients(ctx context.Context, rs []string) error {
	return s.saveRecipients(ctx, rs, "Set Recipients")
}

// RemoveRecipient will remove the given recipient from the store
// but if this key is not available on this machine we
// just try to remove it literally.
func (s *Store) RemoveRecipient(ctx context.Context, id string) error {
	keys, err := s.crypto.FindRecipients(ctx, id)
	if err != nil {
		out.Warningf(ctx, "Warning: Failed to get GPG Key Info for %s: %s", id, err)
	}

	rs, err := s.GetRecipients(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to read recipient list: %w", err)
	}

	nk := make([]string, 0, len(rs)-1)
RECIPIENTS:
	for _, k := range rs {

		// First lets try a simple match of the stored ids
		if k == id {
			debug.Log("removing recipient based on id match %s", k)
			continue RECIPIENTS
		}

		// If we don't match immediately, we may need to loop through the recipient keys to try and match.
		// To do this though, we need to ensure that we also do a FindRecipients on the id name from the stored ids.
		recipientIds, err := s.crypto.FindRecipients(ctx, k)
		if err != nil {
			out.Warningf(ctx, "Warning: Failed to get GPG Key Info for %s: %s", k, err)
		}
		debug.Log("returned the following ids for recipient %s: %s", k, recipientIds)

		// if the key is available locally we can also match the id against
		// the fingerprint or failing that we can try against the recipientIds
		for _, key := range keys {
			if strings.HasSuffix(key, k) {
				debug.Log("removing recipient based on id suffix match: %s %s", key, k)
				continue RECIPIENTS
			}

			for _, recipientID := range recipientIds {
				if recipientID == key {
					debug.Log("removing recipient based on recipient id match %s", recipientID)
					continue RECIPIENTS
				}
			}
		}

		nk = append(nk, k)
	}

	if len(rs) == len(nk) {
		return fmt.Errorf("recipient not in store")
	}

	if err := s.saveRecipients(ctx, nk, "Removed Recipient "+id); err != nil {
		return fmt.Errorf("failed to save recipients: %w", err)
	}

	return s.reencrypt(ctxutil.WithCommitMessage(ctx, "Removed Recipient "+id))
}

func (s *Store) ensureOurKeyID(ctx context.Context, rs []string) []string {
	ourID := s.OurKeyID(ctx)
	if ourID == "" {
		return rs
	}
	for _, r := range rs {
		if r == ourID {
			return rs
		}
	}
	rs = append(rs, ourID)
	return rs
}

// OurKeyID returns the key fingprint this user can use to access the store
// (if any).
func (s *Store) OurKeyID(ctx context.Context) string {
	for _, r := range s.Recipients(ctx) {
		kl, err := s.crypto.FindIdentities(ctx, r)
		if err != nil || len(kl) < 1 {
			continue
		}
		return kl[0]
	}
	return ""
}

// GetRecipients will load all Recipients from the .gpg-id file for the given
// secret path.
func (s *Store) GetRecipients(ctx context.Context, name string) ([]string, error) {
	return s.getRecipients(ctx, s.idFile(ctx, name))
}

func (s *Store) getRecipients(ctx context.Context, idf string) ([]string, error) {
	buf, err := s.storage.Get(ctx, idf)
	if err != nil {
		return nil, fmt.Errorf("failed to get recipients from %q: %w", idf, err)
	}

	return recipients.Unmarshal(buf), nil
}

type keyExporter interface {
	ExportPublicKey(ctx context.Context, id string) ([]byte, error)
}

// ExportMissingPublicKeys will export any possibly missing public keys to the
// stores .public-keys directory.
func (s *Store) ExportMissingPublicKeys(ctx context.Context, rs []string) (bool, error) {
	exp, ok := s.crypto.(keyExporter)
	if !ok {
		debug.Log("not exporting public keys for %T", s.crypto)
		return false, nil
	}

	var failed, exported bool
	for _, r := range rs {
		if r == "" {
			continue
		}
		path, err := s.exportPublicKey(ctx, exp, r)
		if err != nil {
			failed = true
			out.Errorf(ctx, "failed to export public key for %q: %s", r, err)
			continue
		}
		if path == "" {
			continue
		}
		// at least one key has been exported
		exported = true
		if err := s.storage.Add(ctx, path); err != nil {
			if errors.Is(err, store.ErrGitNotInit) {
				continue
			}
			failed = true
			out.Errorf(ctx, "failed to add public key for %q to git: %s", r, err)
			continue
		}
		if err := s.storage.Commit(ctx, fmt.Sprintf("Exported Public Keys %s", r)); err != nil && err != store.ErrGitNothingToCommit {
			failed = true
			out.Errorf(ctx, "Failed to git commit: %s", err)
			continue
		}
	}
	if failed {
		return exported, fmt.Errorf("some keys failed")
	}
	return exported, nil
}

// Save all Recipients in memory to the recipients file on disk.
func (s *Store) saveRecipients(ctx context.Context, rs []string, msg string) error {
	if len(rs) < 1 {
		return fmt.Errorf("can not remove all recipients")
	}

	idf := s.idFile(ctx, "")
	buf := recipients.Marshal(rs)
	if err := s.storage.Set(ctx, idf, buf); err != nil {
		return fmt.Errorf("failed to write recipients file: %w", err)
	}

	if err := s.storage.Add(ctx, idf); err != nil {
		if err != store.ErrGitNotInit {
			return fmt.Errorf("failed to add file %q to git: %w", idf, err)
		}
	}

	if err := s.storage.Commit(ctx, msg); err != nil {
		if err != store.ErrGitNotInit && err != store.ErrGitNothingToCommit {
			return fmt.Errorf("failed to commit changes to git: %w", err)
		}
	}

	// save all recipients public keys to the repo
	if ctxutil.IsExportKeys(ctx) {
		if _, err := s.ExportMissingPublicKeys(ctx, rs); err != nil {
			out.Errorf(ctx, "Failed to export missing public keys: %s", err)
		}
	}

	// push to remote repo
	if err := s.storage.Push(ctx, "", ""); err != nil {
		if errors.Is(err, store.ErrGitNotInit) {
			return nil
		}
		if errors.Is(err, store.ErrGitNoRemote) {
			msg := "Warning: git has no remote. Ignoring auto-push option\n" +
				"Run: gopass git remote add origin ..."
			debug.Log(msg)
			return nil
		}
		return fmt.Errorf("failed to push changes to git: %w", err)
	}

	return nil
}
