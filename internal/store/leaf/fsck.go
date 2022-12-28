package leaf

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/diff"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/queue"
	"github.com/gopasspw/gopass/internal/store"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/gopass"
)

// Fsck checks all entries matching the given prefix.
func (s *Store) Fsck(ctx context.Context, path string) error {
	ctx = out.AddPrefix(ctx, "["+s.alias+"] ")
	debug.Log("Checking %s", path)

	// first let the storage backend check itself
	debug.Log("Checking storage backend")
	if err := s.storage.Fsck(ctx); err != nil {
		return fmt.Errorf("storage backend error: %w", err)
	}

	// then try to compact storage / rcs
	debug.Log("Compacting storage")
	if err := s.storage.Compact(ctx); err != nil {
		return fmt.Errorf("storage backend compaction failed: %w", err)
	}

	// make sure all recipients are valid
	debug.Log("Checking recipients")
	if err := s.CheckRecipients(ctx); err != nil {
		if IsCheckRecipients(ctx) {
			return fmt.Errorf("invalid recipients found: %w", err)
		}

		out.Errorf(ctx, "Invalid recipients found: %s", err)
	}

	// then we'll make sure all the secrets are readable by us and every
	// valid recipient
	if path != "" {
		out.Printf(ctx, "Checking all secrets matching %s", path)
	}

	if err := s.fsckLoop(ctx, path); err != nil {
		return err
	}

	if !config.Bool(ctx, "core.autosync") {
		debug.Log("not pushing to git remote, core.autosync is false")

		return nil
	}

	if err := s.storage.Push(ctx, "", ""); err != nil {
		if !errors.Is(err, store.ErrGitNoRemote) {
			out.Printf(ctx, "RCS Push failed: %s", err)
		}
	}

	return nil
}

func (s *Store) fsckLoop(ctx context.Context, path string) error {
	pcb := ctxutil.GetProgressCallback(ctx)

	// disable network ops, we will push at the end. pushing on possibly
	// every single secret could be terribly slow.
	ctx = ctxutil.WithNoNetwork(ctx, true)

	// disable the queue, for batch operations this is not necessary / wanted
	// since different git processes might step onto each others toes.
	ctx = queue.WithQueue(ctx, nil)

	names, err := s.List(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to list entries for %s: %w", path, err)
	}
	debug.Log("names (%d): %q", len(names), names)
	sort.Strings(names)

	for _, name := range names {
		pcb()
		if strings.HasPrefix(name, s.alias+"/") {
			name = strings.TrimPrefix(name, s.alias+"/")
		}

		debug.Log("[%s] Checking %s", path, name)

		if err := s.fsckCheckEntry(ctx, name); err != nil {
			return fmt.Errorf("failed to check %q: %w", name, err)
		}
	}

	if err := s.fsckUpdatePublicKeys(ctx); err != nil {
		out.Errorf(ctx, "Failed to update public keys: %s", err)
	}

	return nil
}

func (s *Store) fsckUpdatePublicKeys(ctx context.Context) error {
	ctx = WithPubkeyUpdate(ctx, true)
	rs := s.Recipients(ctx)

	// first import possibly new/updated keys to merge any changes
	// that might come from others.
	if err := s.ImportMissingPublicKeys(ctx, rs...); err != nil {
		return fmt.Errorf("failed to import new or updated pubkeys: %w", err)
	}

	// then export our (possibly updated) keys for consumption
	// by others.
	exported, err := s.UpdateExportedPublicKeys(ctx, rs)
	if err != nil {
		return fmt.Errorf("failed to update exported pubkeys: %w", err)
	}
	debug.Log("Updated exported public keys: %t", exported)

	return nil
}

type convertedSecret interface {
	gopass.Secret
	FromMime() bool
}

func (s *Store) fsckCheckEntry(ctx context.Context, name string) error {
	if err := s.fsckCheckRecipients(ctx, name); err != nil {
		out.Warningf(ctx, "Checking recipients for %s failed: %s", name, err)
	}

	// make sure we can actually decode this secret
	// if this fails there is no way we could fix this
	if !IsFsckDecrypt(ctx) {
		return nil
	}

	// we need to make sure Parsing is enabled in order to parse old Mime secrets
	ctx = ctxutil.WithShowParsing(ctx, true)
	sec, err := s.Get(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to decode secret %s: %w", name, err)
	}

	// check if this is still an old MIME secret.
	// Note: the secret was already converted when it was parsed during Get.
	// This is just checking if it was converted from MIME or not.
	// This branch is pretty much useless right now, but I'd like to add some
	// reporting on how many secrets were converted from MIME to new format.
	// TODO: report these stats
	if cs, ok := sec.(convertedSecret); ok && cs.FromMime() {
		debug.Log("leftover Mime secret: %s", name)
	}

	out.Printf(ctx, "Re-encrypting %s to fix recipients and storage format.", name)
	if err := s.Set(ctxutil.WithCommitMessage(ctx, "fsck --decrypt to fix recipients and format"), name, sec); err != nil {
		return fmt.Errorf("failed to write secret: %w", err)
	}

	return nil
}

func (s *Store) fsckCheckRecipients(ctx context.Context, name string) error {
	// now compare the recipients this secret was encoded for and fix it if
	// it doesn't match.
	ciphertext, err := s.storage.Get(ctx, s.Passfile(name))
	if err != nil {
		return fmt.Errorf("failed to get raw secret: %w", err)
	}

	itemRecps, err := s.crypto.RecipientIDs(ctx, ciphertext)
	if err != nil {
		return fmt.Errorf("failed to read recipient IDs from raw secret: %w", err)
	}

	itemRecps = fingerprints(ctx, s.crypto, itemRecps)

	rs, err := s.GetRecipients(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to get recipients from store: %w", err)
	}

	perItemStoreRecps := fingerprints(ctx, s.crypto, rs.IDs())

	// check itemRecps matches storeRecps
	extra, missing := diff.List(perItemStoreRecps, itemRecps)
	if len(missing) > 0 {
		out.Errorf(ctx, "Missing recipients on %s: %+v\nRun fsck with the --decrypt flag to re-encrypt it automatically, or edit this secret yourself.", name, missing)
	}

	if len(extra) > 0 {
		out.Errorf(ctx, "Extra recipients on %s: %+v\nRun fsck with the --decrypt flag to re-encrypt it automatically, or edit this secret yourself.", name, extra)
	}

	return nil
}

func fingerprints(ctx context.Context, crypto backend.Crypto, in []string) []string {
	out := make([]string, 0, len(in))
	for _, r := range in {
		out = append(out, crypto.Fingerprint(ctx, r))
	}

	return out
}

func compareStringSlices(want, have []string) ([]string, []string) {
	missing := []string{}
	extra := []string{}

	wantMap := make(map[string]struct{}, len(want))
	haveMap := make(map[string]struct{}, len(have))

	for _, w := range want {
		wantMap[w] = struct{}{}
	}

	for _, h := range have {
		haveMap[h] = struct{}{}
	}

	for k := range wantMap {
		if _, found := haveMap[k]; !found {
			missing = append(missing, k)
		}
	}

	for k := range haveMap {
		if _, found := wantMap[k]; !found {
			extra = append(extra, k)
		}
	}

	sort.Strings(missing)
	sort.Strings(extra)

	return missing, extra
}
