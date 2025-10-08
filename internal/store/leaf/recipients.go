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
	"github.com/gopasspw/gopass/internal/store"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/set"
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

// AllRecipients returns a list of all recipients of this store,
// including all sub-stores.
func (s *Store) AllRecipients(ctx context.Context) *recipients.Recipients {
	rs := recipients.New()
	for _, recs := range s.RecipientsTree(ctx) {
		for _, r := range recs {
			rs.Add(r)
		}
	}

	return rs
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
		if !errors.Is(err, ErrInvalidHash) || !ack {
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
		debug.V(1).Log("testing key: %q", k)
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

func (s *Store) ensureOurKeyID(ctx context.Context, recp []string) []string {
	kl, _ := s.crypto.FindIdentities(ctx, recp...)
	if len(kl) > 0 {
		debug.Log("one of our key is already in the recipient list, not changing it")

		return recp
	}

	ourID := s.OurKeyID(ctx)
	if ourID == "" {
		debug.Log("no owner key found, couldn't add it to the recipients list")

		return recp
	}
	debug.Log("adding our key to the recipient list")
	recp = append(recp, ourID)

	return recp
}

// OurKeyID returns the key fingprint this user can use to access the store
// (if any).
func (s *Store) OurKeyID(ctx context.Context) string {
	recp := s.Recipients(ctx)

	debug.Log("getting our key ID from store for recipients %v", recp)

	kl, err := s.crypto.FindIdentities(ctx, recp...)
	if err != nil || len(kl) < 1 {
		debug.Log("WARNING: no owner key found in %v", recp)
		out.Warning(ctx, "No owner key found. Make sure your key is fully trusted.")

		return ""
	}

	return kl[0]
}

// GetRecipients will load all Recipients from the .gpg-id file for the given
// secret path.
func (s *Store) GetRecipients(ctx context.Context, name string) (*recipients.Recipients, error) {
	return s.getRecipients(ctx, s.idFile(ctx, name))
}

func (s *Store) getRecipients(ctx context.Context, idf string) (*recipients.Recipients, error) {
	buf, err := s.storage.Get(ctx, idf)
	if err != nil {
		return recipients.New(), fmt.Errorf("failed to get recipients from IDFile %q: %w", idf, err)
	}

	rs := recipients.Unmarshal(buf)

	cfg, _ := config.FromContext(ctx)
	// check recipients hash, global config takes precedence here for security reasons
	if cfg.GetGlobal("recipients.check") != "true" && !config.AsBool(cfg.GetM(s.alias, "recipients.check")) {
		return rs, nil
	}

	// we do NOT support local recipients.hash keys since they could be remotely changed
	cfgHash := cfg.GetGlobal(s.rhKey())
	rsHash := rs.Hash()
	if rsHash != cfgHash {
		return rs, fmt.Errorf("config hash %q= %q - Recipients file %q = %q: %w", s.rhKey(), cfgHash, idf, rsHash, ErrInvalidHash)
	}

	return rs, nil
}

// UpdateExportedPublicKeys will export any possibly missing public keys to the
// stores .public-keys directory.
func (s *Store) UpdateExportedPublicKeys(ctx context.Context) (bool, error) {
	exp, ok := s.crypto.(keyExporter)
	if !ok {
		debug.Log("not exporting public keys for %T", s.crypto)

		return false, nil
	}

	recipients := make(map[string]bool, s.AllRecipients(ctx).Len())
	for _, r := range s.AllRecipients(ctx).IDs() {
		recipients[r] = true
	}

	// add any missing keys
	failed, exported := s.addMissingKeys(ctx, exp, recipients)

	// remove any extra key files, we do not support this at the local config level
	// TODO(GH-2620): Temporarily disabled by default until we fix the
	// key cleanup.
	if cfg, _ := config.FromContext(ctx); cfg.GetGlobal("recipients.remove-extra-keys") == "true" {
		f, e := s.removeExtraKeys(ctx, recipients)
		failed = failed || f
		exported = exported || e
	}

	if exported && ctxutil.IsGitCommit(ctx) {
		if err := s.storage.TryCommit(ctx, "Updated exported Public Keys"); err != nil {
			failed = true

			out.Errorf(ctx, "Failed to git commit: %s", err)
		}
	}

	if failed {
		return exported, fmt.Errorf("some keys failed")
	}

	return exported, nil
}

func (s *Store) addMissingKeys(ctx context.Context, exp keyExporter, recipients map[string]bool) (bool, bool) {
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
		if err := s.storage.TryAdd(ctx, path); err != nil {
			failed = true

			out.Errorf(ctx, "failed to add public key for %q to git: %s", r, err)

			continue
		}
	}

	return failed, exported
}

func extraKeys(recipients map[string]bool, keys []string) []string {
	extras := make([]string, 0, len(keys))
	for _, key := range keys {
		// do not use filepath, that would break on Windows. storage.List normalizes all paths
		// returned to normal (forward) slashes. Even on Windows.
		key := path.Base(key)

		if recipients[key] {
			debug.Log("Key %s found. Not removing", key)

			continue
		}
		extras = append(extras, key)
	}

	return extras
}

func (s *Store) removeExtraKeys(ctx context.Context, recipients map[string]bool) (bool, bool) {
	var failed, exported bool

	keys, err := s.storage.List(ctx, keyDir)
	if err != nil {
		failed = true

		out.Errorf(ctx, "Failed to list keys: %s", err)
	}

	debug.Log("Checking %q for extra keys that need to be removed", keys)
	for _, key := range extraKeys(recipients, keys) {
		debug.Log("Removing extra key %s", key)

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

	return failed, exported
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
	errSet := s.storage.Set(ctx, idf, buf)
	if errSet != nil && !errors.Is(errSet, store.ErrMeaninglessWrite) {
		return fmt.Errorf("failed to write recipients file: %w", errSet)
	}

	// always save recipients hash to global config
	cfg, _ := config.FromContext(ctx)
	if err := cfg.Set("", s.rhKey(), rs.Hash()); err != nil {
		out.Errorf(ctx, "Failed to update %s: %s", s.rhKey(), err)
	}

	// save all recipients public keys to the repo if wanted
	if config.AsBool(cfg.GetM(s.alias, "core.exportkeys")) {
		debug.Log("updating exported keys")
		if _, err := s.UpdateExportedPublicKeys(ctx); err != nil {
			out.Errorf(ctx, "Failed to export missing public keys: %s", err)
		}
	} else {
		debug.Log("updating exported keys not requested")
	}

	if errors.Is(errSet, store.ErrMeaninglessWrite) {
		debug.Log("no need to overwrite recipient file: ErrMeaninglessWrite")

		return nil
	}

	if err := s.storage.TryAdd(ctx, idf); err != nil {
		return fmt.Errorf("failed to add file %q to git: %w", idf, err)
	}

	if err := s.storage.TryCommit(ctx, msg); err != nil {
		return fmt.Errorf("failed to commit changes to git: %w", err)
	}

	if !config.AsBool(cfg.GetM(s.alias, "core.autopush")) {
		debug.Log("not pushing to git remote, core.autopush is false")

		return nil
	}

	// push to remote repo
	debug.Log("pushing changes to git remote")
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

	debug.Log("recipients saved")
	return nil
}

func (s *Store) rhKey() string {
	if s.alias == "" {
		return "recipients.hash"
	}

	return fmt.Sprintf("recipients.%s.hash", s.alias)
}
