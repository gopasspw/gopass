package leaf

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/gopasspw/gopass/internal/backend/crypto/age"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/recipients"
	"github.com/gopasspw/gopass/internal/store"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"

	"github.com/pkg/errors"
)

const (
	keyDir    = ".public-keys"
	oldKeyDir = ".gpg-keys"
)

// Recipients returns the list of recipients of this store
func (s *Store) Recipients(ctx context.Context) []string {
	rs, err := s.GetRecipients(ctx, "")
	if err != nil {
		out.Error(ctx, "failed to read recipient list: %s", err)
	}
	return rs
}

// RecipientsTree returns a mapping of secrets to recipients
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
		out[dir] = srs
	}
	out[""] = root
	return out
}

// AddRecipient adds a new recipient to the list
func (s *Store) AddRecipient(ctx context.Context, id string) error {
	rs, err := s.GetRecipients(ctx, "")
	if err != nil {
		return errors.Wrapf(err, "failed to read recipient list")
	}

	for _, k := range rs {
		if k == id {
			return errors.Errorf("Recipient already in store")
		}
	}

	rs = append(rs, id)

	if err := s.saveRecipients(ctx, rs, "Added Recipient "+id); err != nil {
		return errors.Wrapf(err, "failed to save recipients")
	}

	out.Cyan(ctx, "Reencrypting existing secrets. This may take some time ...")
	return s.reencrypt(ctxutil.WithCommitMessage(ctx, "Added Recipient "+id))
}

// SaveRecipients persists the current recipients on disk
func (s *Store) SaveRecipients(ctx context.Context) error {
	rs, err := s.GetRecipients(ctx, "")
	if err != nil {
		return errors.Wrapf(err, "failed to get recipients")
	}
	return s.saveRecipients(ctx, rs, "Save Recipients")
}

// SetRecipients will update the stored recipients and the associated checksum
func (s *Store) SetRecipients(ctx context.Context, rs []string) error {
	return s.saveRecipients(ctx, rs, "Set Recipients")
}

// RemoveRecipient will remove the given recipient from the store
// but if this key is not available on this machine we
// just try to remove it literally
func (s *Store) RemoveRecipient(ctx context.Context, id string) error {
	keys, err := s.crypto.FindRecipients(ctx, id)
	if err != nil {
		out.Cyan(ctx, "Warning: Failed to get GPG Key Info for %s: %s", id, err)
	}

	rs, err := s.GetRecipients(ctx, "")
	if err != nil {
		return errors.Wrapf(err, "failed to read recipient list")
	}

	nk := make([]string, 0, len(rs)-1)
RECIPIENTS:
	for _, k := range rs {
		if k == id {
			continue RECIPIENTS
		}
		// if the key is available locally we can also match the id against
		// the fingerprint
		for _, key := range keys {
			if strings.HasSuffix(key, k) {
				continue RECIPIENTS
			}
		}
		nk = append(nk, k)
	}

	if len(rs) == len(nk) {
		return errors.Errorf("recipient not in store")
	}

	if err := s.saveRecipients(ctx, nk, "Removed Recipient "+id); err != nil {
		return errors.Wrapf(err, "failed to save recipients")
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
// (if any)
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
// secret path
func (s *Store) GetRecipients(ctx context.Context, name string) ([]string, error) {
	return s.getRecipients(ctx, s.idFile(ctx, name))
}

func (s *Store) getRecipients(ctx context.Context, idf string) ([]string, error) {
	buf, err := s.storage.Get(ctx, idf)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get recipients from %s", idf)
	}

	rawRecps := recipients.Unmarshal(buf)
	finalRecps := make([]string, 0, len(rawRecps))
	for _, r := range rawRecps {
		fp := s.crypto.Fingerprint(ctx, r)
		if fp == "" {
			fp = r
		}
		finalRecps = append(finalRecps, fp)
	}
	sort.Strings(finalRecps)
	return finalRecps, nil
}

// ExportMissingPublicKeys will export any possibly missing public keys to the
// stores .public-keys directory
func (s *Store) ExportMissingPublicKeys(ctx context.Context, rs []string) (bool, error) {
	// do not export any keys for age, where public key == key id
	if _, ok := s.crypto.(*age.Age); ok {
		debug.Log("not exporting public keys for age")
		return false, nil
	}
	ok := true
	exported := false
	for _, r := range rs {
		if r == "" {
			continue
		}
		path, err := s.exportPublicKey(ctx, r)
		if err != nil {
			ok = false
			out.Error(ctx, "failed to export public key for '%s': %s", r, err)
			continue
		}
		if path == "" {
			continue
		}
		// at least one key has been exported
		exported = true
		if err := s.storage.Add(ctx, path); err != nil {
			if errors.Cause(err) == store.ErrGitNotInit {
				continue
			}
			ok = false
			out.Error(ctx, "failed to add public key for '%s' to git: %s", r, err)
			continue
		}
		if err := s.storage.Commit(ctx, fmt.Sprintf("Exported Public Keys %s", r)); err != nil && err != store.ErrGitNothingToCommit {
			ok = false
			out.Error(ctx, "Failed to git commit: %s", err)
			continue
		}
	}
	if !ok {
		return exported, errors.New("some keys failed")
	}
	return exported, nil
}

// Save all Recipients in memory to the .gpg-id file on disk.
func (s *Store) saveRecipients(ctx context.Context, rs []string, msg string) error {
	if len(rs) < 1 {
		return errors.New("can not remove all recipients")
	}

	idf := s.idFile(ctx, "")
	buf := recipients.Marshal(rs)
	if err := s.storage.Set(ctx, idf, buf); err != nil {
		return errors.Wrapf(err, "failed to write recipients file")
	}

	if err := s.storage.Add(ctx, idf); err != nil {
		if err != store.ErrGitNotInit {
			return errors.Wrapf(err, "failed to add file '%s' to git", idf)
		}
	}

	if err := s.storage.Commit(ctx, msg); err != nil {
		if err != store.ErrGitNotInit && err != store.ErrGitNothingToCommit {
			return errors.Wrapf(err, "failed to commit changes to git")
		}
	}

	// save all recipients public keys to the repo
	if ctxutil.IsExportKeys(ctx) {
		if _, err := s.ExportMissingPublicKeys(ctx, rs); err != nil {
			out.Error(ctx, "Failed to export missing public keys: %s", err)
		}
	}

	// push to remote repo
	if err := s.storage.Push(ctx, "", ""); err != nil {
		if errors.Cause(err) == store.ErrGitNotInit {
			return nil
		}
		if errors.Cause(err) == store.ErrGitNoRemote {
			msg := "Warning: git has no remote. Ignoring auto-push option\n" +
				"Run: gopass git remote add origin ..."
			debug.Log(msg)
			return nil
		}
		return errors.Wrapf(err, "failed to push changes to git")
	}

	return nil
}
