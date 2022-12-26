package leaf

import (
	"context"
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/recipients"
	"github.com/gopasspw/gopass/internal/set"
	"github.com/gopasspw/gopass/internal/store"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/termio"
)

const (
	keyDir    = ".public-keys"
	oldKeyDir = ".gpg-keys"
)

// ErrInvalidHash indicates an outdated value of `recipients.hash`.
var ErrInvalidHash = fmt.Errorf("recipients.hash invalid")

// InvalidRecipientsError is a custom error type that contains a
// list of invalid recipients with their check failures.
type InvalidRecipientsError struct {
	Invalid map[string]error
}

func (e InvalidRecipientsError) Error() string {
	var sb strings.Builder

	sb.WriteString("Invalid Recipients: ")
	for _, k := range set.SortedKeys(e.Invalid) {
		sb.WriteString(k)
		sb.WriteString(": ")
		sb.WriteString(e.Invalid[k].Error())
		sb.WriteString(", ")
	}

	return sb.String()
}

// IsError returns true if this multi error contains any underlying errors.
func (e InvalidRecipientsError) IsError() bool {
	return len(e.Invalid) > 0
}

// Recipients returns the list of recipients of this store.
func (s *Store) Recipients(ctx context.Context) []string {
	rs, err := s.GetRecipients(ctx, "")
	if err != nil {
		out.Errorf(ctx, "failed to read recipient list: %s", err)
		out.Notice(ctx, "Please review the recipients list and confirm any changes with 'gopass recipients ack'")
	}

	return rs.IDs()
}

// RecipientsTree returns a mapping of secrets to recipients.
// Note: Usually that is one set of recipients per store, but we
// offer limited support of different recipients per sub-directory
// so this is why we are here.
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
		out[dir] = srs.IDs()
	}

	out[""] = root

	return out
}

// CheckRecipients makes sure all existing recipients are valid.
func (s *Store) CheckRecipients(ctx context.Context) error {
	rs, err := s.GetRecipients(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to read recipient list: %w", err)
	}

	er := InvalidRecipientsError{
		Invalid: make(map[string]error, len(rs.IDs())),
	}
	for _, k := range rs.IDs() {
		validKeys, err := s.crypto.FindRecipients(ctx, k)
		if err != nil {
			debug.Log("no GPG key info (unexpected) for %s: %s", k, err)
			er.Invalid[k] = err

			continue
		}

		if len(validKeys) < 1 {
			debug.Log("no valid keys (expired?) for %s", k)
			er.Invalid[k] = fmt.Errorf("no valid keys (expired?)")

			continue
		}

		debug.Log("valid keys found for %s", k)
	}

	if er.IsError() {
		return er
	}

	return nil
}

// AddRecipient adds a new recipient to the list.
func (s *Store) AddRecipient(ctx context.Context, id string) error {
	rs, err := s.GetRecipients(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to read recipient list: %w", err)
	}

	debug.Log("new recipient: %q - existing: %+v", id, rs)

	idAlreadyInStore := rs.Has(id)
	if idAlreadyInStore {
		if !termio.AskForConfirmation(ctx, fmt.Sprintf("key %q already in store. Do you want to re-encrypt with public key? This is useful if you changed your public key (e.g. added subkeys).", id)) {
			return nil
		}
	} else {
		rs.Add(id)

		if err := s.saveRecipients(ctx, rs, "Added Recipient "+id); err != nil {
			return fmt.Errorf("failed to save recipients: %w", err)
		}
	}

	out.Printf(ctx, "Reencrypting existing secrets. This may take some time ...")

	commitMsg := "Recipient " + id
	if idAlreadyInStore {
		commitMsg = "Re-encrypted Store for " + commitMsg
	} else {
		commitMsg = "Added " + commitMsg
	}

	return s.reencrypt(ctxutil.WithCommitMessage(ctx, commitMsg))
}

// SaveRecipients persists the current recipients on disk. Setting ack to true
// will acknowledge an invalid hash and allow updating it.
func (s *Store) SaveRecipients(ctx context.Context, ack bool) error {
	rs, err := s.GetRecipients(ctx, "")
	if err != nil {
		if !errors.Is(err, ErrInvalidHash) || (errors.Is(err, ErrInvalidHash) && !ack) {
			return fmt.Errorf("failed to get recipients: %w", err)
		}
	}

	return s.saveRecipients(ctx, rs, "Save Recipients")
}

// SetRecipients will update the stored recipients.
func (s *Store) SetRecipients(ctx context.Context, rs *recipients.Recipients) error {
	return s.saveRecipients(ctx, rs, "Set Recipients")
}

// RemoveRecipient will remove the given recipient from the store
// but if this key is not available on this machine we
// just try to remove it literally.
func (s *Store) RemoveRecipient(ctx context.Context, key string) error {
	rs, err := s.GetRecipients(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to read recipient list: %w", err)
	}

	var removed int
RECIPIENTS:
	for _, k := range rs.IDs() { //nolint:whitespace

		// First lets try a simple match of the stored ids
		if k == key {
			debug.Log("removing recipient based on id match %s", k)
			if rs.Remove(k) {
				removed++
			}

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
		if strings.HasSuffix(key, k) {
			debug.Log("removing recipient based on id suffix match: %s %s", key, k)
			if rs.Remove(k) {
				removed++
			}

			continue RECIPIENTS
		}

		for _, recipientID := range recipientIds {
			if recipientID == key {
				debug.Log("removing recipient based on recipient id match %s", recipientID)
				if rs.Remove(k) {
					removed++
				}

				continue RECIPIENTS
			}
		}
	}

	if removed < 1 {
		return fmt.Errorf("recipient not in store")
	}

	if err := s.saveRecipients(ctx, rs, "Removed Recipient "+key); err != nil {
		return fmt.Errorf("failed to save recipients: %w", err)
	}

	return s.reencrypt(ctxutil.WithCommitMessage(ctx, "Removed Recipient "+key))
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
func (s *Store) GetRecipients(ctx context.Context, name string) (*recipients.Recipients, error) {
	return s.getRecipients(ctx, s.idFile(ctx, name))
}

func (s *Store) getRecipients(ctx context.Context, idf string) (*recipients.Recipients, error) {
	buf, err := s.storage.Get(ctx, idf)
	if err != nil {
		return recipients.New(), fmt.Errorf("failed to get recipients from %q: %w", idf, err)
	}

	rs := recipients.Unmarshal(buf)

	// check recipients hash
	if !config.Bool(ctx, "recipients.check") {
		return rs, nil
	}

	cfg := config.FromContext(ctx)
	cfgHash := cfg.GetM(s.alias, "recipients.hash")
	rsHash := rs.Hash()
	if rsHash != cfgHash {
		return rs, fmt.Errorf("Config: %s - Recipients file: %s: %w", cfgHash, rsHash, ErrInvalidHash)
	}

	return rs, nil
}


// UpdateExportedPublicKeys will export any possibly missing public keys to the
// stores .public-keys directory.
func (s *Store) UpdateExportedPublicKeys(ctx context.Context, rs []string) (bool, error) {
	exp, ok := s.crypto.(keyExporter)
	if !ok {
		debug.Log("not exporting public keys for %T", s.crypto)

		return false, nil
	}

	recipients := make(map[string]bool, len(rs))
	for _, r := range rs {
		recipients[r] = true
	}

	// add any missing keys
	var failed, exported bool
	for r := range recipients {
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
	}

	// remove any extra key files
	keys, err := s.storage.List(ctx, keyDir)
	if err != nil {
		failed = true

		out.Errorf(ctx, "Failed to list keys: %s", err)
	}

	debug.Log("Checking %q for extra keys that need to be removed", keys)
	for _, key := range keys {
		// do not use filepath, that would break on Windows. storage.List normalizes all paths
		// returned to normal (forward) slashes. Even on Windows.
		key := path.Base(key)

		if recipients[key] {
			debug.Log("Key %s found. Not removing", key)

			continue
		}

		debug.Log("Remvoing extra key %s", key)

		if err := s.storage.Delete(ctx, filepath.Join(keyDir, key)); err != nil {
			out.Errorf(ctx, "Failed to remove extra key %q: %s", key, err)

			continue
		}

		if err := s.storage.Add(ctx, filepath.Join(keyDir, key)); err != nil {
			out.Errorf(ctx, "Failed to mark extra key for removal %q: %s", key, err)

			continue
		}

		// to ensure the commit
		exported = true
		debug.Log("Removed extra key %s", key)
	}

	if exported {
		if err := s.storage.Commit(ctx, fmt.Sprintf("Updated exported Public Keys")); err != nil && !errors.Is(err, store.ErrGitNothingToCommit) {
			failed = true

			out.Errorf(ctx, "Failed to git commit: %s", err)
		}
	}

	if failed {
		return exported, fmt.Errorf("some keys failed")
	}

	return exported, nil
}

type recipientMarshaler interface {
	IDs() []string
	Marshal() []byte
	Hash() string
}

// Save all Recipients in memory to the recipients file on disk.
func (s *Store) saveRecipients(ctx context.Context, rs recipientMarshaler, msg string) error {
	if rs == nil {
		return fmt.Errorf("need valid recipients")
	}
	if len(rs.IDs()) < 1 {
		return fmt.Errorf("can not remove all recipients")
	}

	idf := s.idFile(ctx, "")

	buf := rs.Marshal()
	if err := s.storage.Set(ctx, idf, buf); err != nil {
		if !errors.Is(err, store.ErrMeaninglessWrite) {
			return fmt.Errorf("failed to write recipients file: %w", err)
		}
		return nil // No need to overwrite recipients file
	}

	if err := s.storage.Add(ctx, idf); err != nil {
		if !errors.Is(err, store.ErrGitNotInit) {
			return fmt.Errorf("failed to add file %q to git: %w", idf, err)
		}
	}

	if err := s.storage.Commit(ctx, msg); err != nil {
		if !errors.Is(err, store.ErrGitNotInit) && !errors.Is(err, store.ErrGitNothingToCommit) {
			return fmt.Errorf("failed to commit changes to git: %w", err)
		}
	}

	// save recipients hash
	if err := config.FromContext(ctx).Set(s.alias, "recipients.hash", rs.Hash()); err != nil {
		out.Errorf(ctx, "Failed to update recipients.hash: %s", err)
	}

	// save all recipients public keys to the repo
	if config.Bool(ctx, "core.exportkeys") {
		debug.Log("updating exported keys")
		if _, err := s.UpdateExportedPublicKeys(ctx, rs.IDs()); err != nil {
			out.Errorf(ctx, "Failed to export missing public keys: %s", err)
		}
	} else {
		debug.Log("updating exported keys not requested")
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
