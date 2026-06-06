package leaf

import (
	"context"
	"errors"
	"fmt"
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
		out.Errorf(ctx, "Failed to read recipient list: %s", err)
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

// JoinTeam is the unified join/clone post-processing step. It imports all
// recipient keys that are already in .public-keys/, checks whether the
// current user can decrypt the store, and — if not — exports only the
// user's own public key additively (never removing other recipients).
//
// Returns a boolean indicating whether the user's key was newly exported.
func (s *Store) JoinTeam(ctx context.Context) (bool, error) {
	// 1. Import every recipient key that the store already ships.
	if err := s.ImportMissingPublicKeys(ctx); err != nil {
		return false, fmt.Errorf("failed to import missing public keys: %w", err)
	}

	// 2. Can we decrypt?
	if s.hasDecryptionKey(ctx) {
		debug.Log("JoinTeam: user can already decrypt this store")

		return false, nil
	}

	// 3. No access yet: export ONLY our own key, additively.
	ourIDs, err := s.crypto.FindIdentities(ctx)
	if err != nil || len(ourIDs) == 0 {
		return false, fmt.Errorf("cannot find own encryption keys: %w", err)
	}

	ourID := ourIDs[0]
	debug.Log("JoinTeam: exporting our own key %q to .public-keys/", ourID)

	// Write our public key to .public-keys/<ourID> if it does not exist.
	exp, ok := s.crypto.(keyExporter)
	if !ok {
		return false, fmt.Errorf("crypto backend %T cannot export public keys", s.crypto)
	}

	if _, err := s.exportPublicKey(ctx, exp, ourID); err != nil {
		return false, fmt.Errorf("failed to export own public key: %w", err)
	}

	// Stage the new key file in git.
	pubKeyPath := filepath.Join(keyDir, ourID)
	if err := s.storage.TryAdd(ctx, pubKeyPath); err != nil {
		debug.Log("JoinTeam: failed to stage %s: %s", pubKeyPath, err)
	}

	return true, nil
}

// hasDecryptionKey returns true when at least one of the local keyring's
// identities matches a recipient in the .gpg-id list.
func (s *Store) hasDecryptionKey(ctx context.Context) bool {
	recp := s.Recipients(ctx)
	ids, err := s.crypto.FindIdentities(ctx, recp...)

	return err == nil && len(ids) > 0
}

// GuardPartialViewWrite ensures the operator can resolve all current
// recipients (either via the local keyring or via .public-keys/) before a
// write that regenerates the exported key set. If some recipients are
// unresolvable, it imports them from .public-keys/; if that still fails, it
// returns an error and prints actionable guidance instead of silently
// pushing a reduced set.
func (s *Store) GuardPartialViewWrite(ctx context.Context) error {
	rs, err := s.GetRecipients(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to read recipient list: %w", err)
	}

	var unresolved []string
	for _, id := range rs.IDs() {
		kl, findErr := s.crypto.FindRecipients(ctx, id)
		keyringOK := findErr == nil && len(kl) > 0

		pubkeyPath := filepath.Join(keyDir, id)
		pubkeyExists := s.storage.Exists(ctx, pubkeyPath)

		if !keyringOK && !pubkeyExists {
			unresolved = append(unresolved, id)

			continue
		}

		if !keyringOK && pubkeyExists {
			// Try to import from .public-keys/.
			debug.Log("GuardPartialViewWrite: importing %s from .public-keys/", id)
		}
	}

	// Try importing from .public-keys/ for all recipients not in the keyring.
	if err := s.ImportMissingPublicKeys(ctx); err != nil {
		debug.Log("GuardPartialViewWrite: ImportMissingPublicKeys failed: %s", err)
	}

	// Re-check after import.
	unresolved = unresolved[:0]
	for _, id := range rs.IDs() {
		kl, findErr := s.crypto.FindRecipients(ctx, id)
		keyringOK := findErr == nil && len(kl) > 0

		if !keyringOK {
			unresolved = append(unresolved, id)
		}
	}

	if len(unresolved) > 0 {
		return fmt.Errorf("cannot resolve all recipients locally and cannot import them from .public-keys/; "+
			"unresolved: %v. Run 'gopass sync' to fetch missing keys, or ask a team owner to add your key with 'gopass recipients add <your-key>'", unresolved)
	}

	return nil
}

// RecipientDiagnosticLevel encodes the severity of a recipient finding.
type RecipientDiagnosticLevel int

const (
	// DiagInfo is an informational note (already canonical, key is fine).
	DiagInfo RecipientDiagnosticLevel = iota
	// DiagWarning is a non-fatal issue (non-canonical ID, key only in
	// .public-keys, expired key in keyring).
	DiagWarning
	// DiagError is a critical issue (recipient unresolvable anywhere).
	DiagError
)

// String returns a one-word label for the diagnostic level.
func (l RecipientDiagnosticLevel) String() string {
	switch l {
	case DiagError:
		return "ERROR"
	case DiagWarning:
		return "WARN"
	case DiagInfo:
		return "INFO"
	}

	return "INFO"
}

// RecipientDiagnostic describes a single finding about a store recipient.
type RecipientDiagnostic struct {
	Level     RecipientDiagnosticLevel
	Recipient string // the ID as stored in .gpg-id
	Store     string // mount alias (empty for root store)
	Message   string // human-readable description
}

// RecipientDiagnostics is a sorted list of recipient findings.
type RecipientDiagnostics []RecipientDiagnostic

// HasErrors returns true when at least one finding is at Error level.
func (rd RecipientDiagnostics) HasErrors() bool {
	for _, d := range rd {
		if d.Level == DiagError {
			return true
		}
	}

	return false
}

// DiagnoseRecipients performs a read-only diagnostic of the recipient list
// for this store. It checks for non-canonical IDs (GH-2762), unresolvable
// recipients, key-expiry state (ADR A-13), and the .public-keys/ file
// availability. This does not require decryption.
func (s *Store) DiagnoseRecipients(ctx context.Context) RecipientDiagnostics {
	storeLabel := s.alias

	rs, err := s.GetRecipients(ctx, "")
	if err != nil {
		return RecipientDiagnostics{{
			Level:   DiagError,
			Store:   storeLabel,
			Message: fmt.Sprintf("cannot read recipient list: %s", err),
		}}
	}

	diags := make(RecipientDiagnostics, 0, len(rs.IDs())*2)

	for _, rawID := range rs.IDs() {
		// 1. Check whether the ID is canonical.
		canon := s.canonicalizeRecipient(ctx, rawID)
		if canon != rawID {
			diags = append(diags, RecipientDiagnostic{
				Level:     DiagWarning,
				Recipient: rawID,
				Store:     storeLabel,
				Message:   fmt.Sprintf("non-canonical ID (canonical is %q); run 'gopass recipients canonicalize' to migrate", canon),
			})
		}

		// 2. Check keyring availability.
		kl, findErr := s.crypto.FindRecipients(ctx, rawID)
		keyringOK := findErr == nil && len(kl) > 0

		// 3. Check .public-keys/ availability.
		pubkeyPath := filepath.Join(keyDir, rawID)
		pubkeyExists := s.storage.Exists(ctx, pubkeyPath)

		switch {
		case !keyringOK && !pubkeyExists:
			diags = append(diags, RecipientDiagnostic{
				Level:     DiagError,
				Recipient: rawID,
				Store:     storeLabel,
				Message:   "key not found in keyring and not in .public-keys/; cannot encrypt for this recipient",
			})

		case !keyringOK && pubkeyExists:
			// Key not in (usable) keyring, but present in .public-keys/.
			// Check if the fingerprint matches a key in the keyring
			// (expiry / staleness detection, ADR A-13 R-4).
			pk, pkErr := s.getPublicKey(ctx, rawID)
			expiredMsg := "key only available via .public-keys/ (not in local keyring); run 'gopass sync' to import it"
			if pkErr == nil && len(pk) > 0 {
				fp, fpErr := s.crypto.GetFingerprint(ctx, pk)
				if fpErr == nil && fp != "" {
					fpKL, fpFindErr := s.crypto.FindRecipients(ctx, fp)
					if fpFindErr == nil && len(fpKL) > 0 {
						// Key is in keyring by fingerprint but not
						// usable by direct ID lookup — expired.
						expiredMsg = "key in keyring appears expired or unusable; run 'gopass recipients update' to refresh it in the store, then run 'gopass sync'"
					}
				}
			}
			diags = append(diags, RecipientDiagnostic{
				Level:     DiagWarning,
				Recipient: rawID,
				Store:     storeLabel,
				Message:   expiredMsg,
			})

		case keyringOK:
			// Key is in the keyring. Log a brief info message; add expiry
			// warning if the keyring copy is expired (ADR A-13 R-1).
			if canon == rawID {
				diags = append(diags, RecipientDiagnostic{
					Level:     DiagInfo,
					Recipient: rawID,
					Store:     storeLabel,
					Message:   "key is in local keyring",
				})
			}

			// TODO Stage 4: check keyring key for expiry and compare
			// store vs. keyring versions.
		}
	}

	return diags
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

// canonicalizeRecipient resolves a raw recipient identifier (email, short key
// ID, or fingerprint) to its canonical form (full fingerprint) by querying
// the crypto backend. The canonical form is stored in .gpg-id and used as the
// .public-keys/ filename so that both are always consistent (GH-2762).
//
// Resolution order:
//  1. crypto.FindRecipients — if exactly one key matches, use its fingerprint.
//  2. .public-keys/ store copy — parse the armored key to obtain its fingerprint.
//  3. Fall back to the input unchanged so that the plain test backend and
//     --force operations continue to work.
//
// If multiple keys match, the first is used and a warning is emitted; the
// action layer is expected to have disambiguated before calling AddRecipient.
func (s *Store) canonicalizeRecipient(ctx context.Context, id string) string {
	kl, err := s.crypto.FindRecipients(ctx, id)
	if err != nil {
		debug.Log("canonicalizeRecipient: FindRecipients(%q) error: %s", id, err)
	}

	switch len(kl) {
	case 1:
		debug.Log("canonicalizeRecipient: %q -> %q", id, kl[0])

		return kl[0]
	case 0:
		// Key not in local keyring; try to read the fingerprint from the
		// .public-keys/ copy inside the store.
		pk, err := s.getPublicKey(ctx, id)
		if err != nil {
			debug.Log("canonicalizeRecipient: no .public-keys entry for %q, using as-is", id)

			return id
		}

		fp, err := s.crypto.GetFingerprint(ctx, pk)
		if err != nil || fp == "" {
			debug.Log("canonicalizeRecipient: GetFingerprint(%q) error or empty: %v, using as-is", id, err)

			return id
		}

		debug.Log("canonicalizeRecipient: %q -> %q (via .public-keys/)", id, fp)

		return fp
	default:
		// Ambiguous: multiple keys match. Warn and use the first.
		out.Warningf(ctx, "Recipient %q matched %d keys; using the first: %q. "+
			"Run 'gpg -k %s' to confirm the correct key.", id, len(kl), kl[0], id)

		return kl[0]
	}
}

// AddRecipient adds a new recipient to the list.
func (s *Store) AddRecipient(ctx context.Context, id string) error {
	// Resolve the user-supplied identifier to its canonical form (full
	// fingerprint) before storing it. This ensures that the .gpg-id entry
	// and the .public-keys/<id> filename always match and are unambiguous.
	// See GH-2762.
	canonID := s.canonicalizeRecipient(ctx, id)
	if canonID != id {
		out.Printf(ctx, "Resolved %q to canonical key ID %q", id, canonID)
	}

	rs, err := s.GetRecipients(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to read recipient list: %w", err)
	}

	debug.Log("new recipient: %q (canonical: %q) - existing: %+v", id, canonID, rs)

	idAlreadyInStore := rs.Has(canonID)
	if idAlreadyInStore {
		if !termio.AskForConfirmation(ctx, fmt.Sprintf("key %q already in store. Do you want to re-encrypt with public key? This is useful if you changed your public key (e.g. added subkeys).", canonID)) {
			return nil
		}
	} else {
		rs.Add(canonID)

		if err := s.saveRecipients(ctx, rs, "Added Recipient "+canonID); err != nil {
			return fmt.Errorf("failed to save recipients: %w", err)
		}
	}

	out.Printf(ctx, "Reencrypting existing secrets. This may take some time ...")

	commitMsg := "Recipient " + canonID
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
//
// Stage 3 (GH-2620): After removing the recipient from .gpg-id, this
// method performs recipient-scoped cleanup of the corresponding
// .public-keys/ and legacy .gpg-keys/ files. Only the explicitly
// removed recipient's key files are deleted — never an unrelated key.
func (s *Store) RemoveRecipient(ctx context.Context, key string) error {
	rs, err := s.GetRecipients(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to read recipient list: %w", err)
	}

	var removed int
	var removedIDs []string // track which recipient IDs were removed for key file cleanup
RECIPIENTS:
	for _, k := range rs.IDs() { //nolint:whitespace
		debug.V(1).Log("testing key: %q", k)
		// First lets try a simple match of the stored ids
		if k == key {
			debug.Log("removing recipient based on id match %s", k)
			if rs.Remove(k) {
				removed++
				removedIDs = append(removedIDs, k)
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
				removedIDs = append(removedIDs, k)
			}

			continue RECIPIENTS
		}

		for _, recipientID := range recipientIds {
			if recipientID == key {
				debug.Log("removing recipient based on recipient id match %s", recipientID)
				if rs.Remove(k) {
					removed++
					removedIDs = append(removedIDs, k)
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

	// Stage 3 (GH-2620): recipient-scoped key file cleanup.
	// Only delete the .public-keys/ and legacy .gpg-keys/ files for the
	// explicitly removed recipient. This replaces the old blanket
	// removeExtraKeys which could delete unrelated key files.
	for _, id := range removedIDs {
		pubKeyPath := filepath.Join(keyDir, id)
		if s.storage.Exists(ctx, pubKeyPath) {
			if err := s.storage.Delete(ctx, pubKeyPath); err != nil {
				debug.Log("RemoveRecipient: failed to delete %s: %s", pubKeyPath, err)
			} else {
				debug.Log("RemoveRecipient: deleted %s", pubKeyPath)
			}
		}

		// Legacy .gpg-keys/ directory.
		legacyPath := filepath.Join(oldKeyDir, id)
		if s.storage.Exists(ctx, legacyPath) {
			if err := s.storage.Delete(ctx, legacyPath); err != nil {
				debug.Log("RemoveRecipient: failed to delete legacy %s: %s", legacyPath, err)
			} else {
				debug.Log("RemoveRecipient: deleted legacy %s", legacyPath)
			}
		}
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
// stores .public-keys directory. This operation is strictly additive: it never
// removes key files. Cleanup is the responsibility of RemoveRecipient (Stage 3).
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

	// NOTE(GH-2620): removeExtraKeys was disabled by default and is now
	// removed entirely. Key cleanup will move to explicit recipient-scoped
	// removal in RemoveRecipient (Stage 3).

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

// UpdateRecipientKeys re-exports the named recipients' public keys from the
// local keyring into .public-keys/, overwriting stale copies. If no IDs are
// provided, the current user's own identity is used. This is the command
// backing 'gopass recipients update' (Stage 4 / GH-1430).
//
// The context must carry the PubkeyUpdate flag (via WithPubkeyUpdate) so
// exportPublicKey overwrites existing files rather than skipping them.
func (s *Store) UpdateRecipientKeys(ctx context.Context, ids []string) error {
	exp, ok := s.crypto.(keyExporter)
	if !ok {
		return fmt.Errorf("crypto backend %T cannot export public keys", s.crypto)
	}

	// Default to our own identity.
	if len(ids) == 0 {
		ourIDs, err := s.crypto.FindIdentities(ctx)
		if err != nil {
			return fmt.Errorf("cannot find own identities: %w", err)
		}
		if len(ourIDs) == 0 {
			return fmt.Errorf("no local identities found; provide explicit recipient IDs to update")
		}
		ids = ourIDs
	}

	// Resolve each ID to canonical form.
	canonIDs := make([]string, 0, len(ids))
	for _, id := range ids {
		canon := s.canonicalizeRecipient(ctx, id)
		if canon != id {
			out.Printf(ctx, "Resolved %q to canonical key ID %q", id, canon)
		}
		canonIDs = append(canonIDs, canon)
	}

	// Force overwrite in exportPublicKey.
	ctx = WithPubkeyUpdate(ctx, true)

	var failed bool
	exported := make([]string, 0, len(canonIDs))
	for _, id := range canonIDs {
		path, err := s.exportPublicKey(ctx, exp, id)
		if err != nil {
			out.Errorf(ctx, "Failed to update public key for %q: %s", id, err)
			failed = true

			continue
		}
		if path == "" {
			out.Printf(ctx, "Public key for %q is already up-to-date.", id)

			continue
		}

		out.Printf(ctx, "Updated public key for %q.", id)
		exported = append(exported, path)

		if err := s.storage.TryAdd(ctx, path); err != nil {
			debug.Log("UpdateRecipientKeys: failed to stage %s: %s", path, err)
		}
	}

	if len(exported) > 0 && ctxutil.IsGitCommit(ctx) {
		if err := s.storage.TryCommit(ctx, "Updated recipient public keys"); err != nil {
			out.Errorf(ctx, "Failed to git commit: %s", err)
			failed = true
		}
	}

	if failed {
		return fmt.Errorf("some key updates failed")
	}

	return nil
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

			out.Errorf(ctx, "Failed to export public key for %q: %s", r, err)

			continue
		}
		if path == "" {
			continue
		}
		// at least one key has been exported
		exported = true
		if err := s.storage.TryAdd(ctx, path); err != nil {
			failed = true

			out.Errorf(ctx, "Failed to add public key for %q to git: %s", r, err)

			continue
		}
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
		return fmt.Errorf("cannot remove all recipients")
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

// CanonicalizeRecipients rewrites the .gpg-id file so that every recipient
// ID is in its canonical (full-fingerprint) form and renames the corresponding
// .public-keys/ files to match. This migration is safe: it does not change
// which keys the secrets are encrypted to, it only makes the identifiers
// unambiguous. It should be run once on stores that contain non-canonical IDs
// (e.g. email addresses or short key IDs). After running this command, use
// 'gopass sync' to publish the changes.
func (s *Store) CanonicalizeRecipients(ctx context.Context) error {
	rs, err := s.GetRecipients(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to get recipients: %w", err)
	}

	newRS := recipients.New()
	changed := false

	for _, id := range rs.IDs() {
		canon := s.canonicalizeRecipient(ctx, id)
		if canon == id {
			newRS.Add(id)
			out.Printf(ctx, "  %s (already canonical)", id)

			continue
		}

		out.Printf(ctx, "  %s -> %s", id, canon)

		// Rename .public-keys/<id> -> .public-keys/<canon> so that the
		// filename is consistent with the new .gpg-id entry.
		if err := s.renamePublicKeyFile(ctx, id, canon, keyDir); err != nil {
			out.Errorf(ctx, "Failed to rename public key file for %s: %s. Keeping original ID.", id, err)
			newRS.Add(id) // revert: keep original to avoid ID/file mismatch

			continue
		}

		// Also handle the legacy .gpg-keys/ directory, best-effort.
		_ = s.renamePublicKeyFile(ctx, id, canon, oldKeyDir)

		newRS.Add(canon)
		changed = true
	}

	if !changed {
		out.Printf(ctx, "All recipients are already in canonical form.")

		return nil
	}

	return s.saveRecipients(ctx, newRS, "Canonicalized recipient IDs")
}

// renamePublicKeyFile moves the key file from dir/<oldID> to dir/<newID> and
// stages both the new file (add) and the old path (deletion) in git.
// It is a no-op when the source file does not exist.
func (s *Store) renamePublicKeyFile(ctx context.Context, oldID, newID, dir string) error {
	oldPath := filepath.Join(dir, oldID)
	if !s.storage.Exists(ctx, oldPath) {
		return nil
	}

	newPath := filepath.Join(dir, newID)

	pk, err := s.storage.Get(ctx, oldPath)
	if err != nil {
		return fmt.Errorf("read %s: %w", oldPath, err)
	}

	if err := s.storage.Set(ctx, newPath, pk); err != nil {
		if !errors.Is(err, store.ErrMeaninglessWrite) {
			return fmt.Errorf("write %s: %w", newPath, err)
		}
	}

	if err := s.storage.Delete(ctx, oldPath); err != nil {
		return fmt.Errorf("delete %s: %w", oldPath, err)
	}

	// Stage new file and old path (removal) for git.
	if err := s.storage.TryAdd(ctx, newPath); err != nil {
		debug.Log("renamePublicKeyFile: failed to stage %s: %s", newPath, err)
	}

	if err := s.storage.TryAdd(ctx, oldPath); err != nil {
		debug.Log("renamePublicKeyFile: failed to stage deletion of %s: %s", oldPath, err)
	}

	return nil
}

func (s *Store) rhKey() string {
	if s.alias == "" {
		return "recipients.hash"
	}

	return fmt.Sprintf("recipients.%s.hash", s.alias)
}
