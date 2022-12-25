package leaf

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/gopasspw/gopass/internal/backend"
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
	debug.Log("Checking storage backend [leaf store fsck]")
	if err := s.storage.Fsck(ctx); err != nil {
		return fmt.Errorf("storage backend error: %w", err)
	}

	// then try to compact storage / rcs
	debug.Log("Compacting storage")
	if err := s.storage.Compact(ctx); err != nil {
		return fmt.Errorf("storage backend compaction failed: %w", err)
	}

	pcb := ctxutil.GetProgressCallback(ctx)

	// then we'll make sure all the secrets are readable by us and every
	// valid recipient
	if path != "" {
		out.Printf(ctx, "Checking all secrets matching %s", path)
	}

	names, err := s.List(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to list entries for %s: %w", path, err)
	}

	ctx = ctxutil.WithCommitMessage(ctx, "fsck --decrypt to fix recipients and format")

	err_collector := ""
	err_isfatal := false

	t := queue.GetQueue(ctx).Add(func(ctx2 context.Context) (context.Context,error) {
		if ctx==nil || ctx2==nil {
			panic("hahe")
		}
		if err := s.fsckUpdatePublicKeys(ctx); err != nil {
			out.Errorf(ctx, "Failed to update public keys: %s", err)
		}
		if ctx==nil || ctx2==nil {
			panic("haha")
		}
		return ctx2, nil
	})
	t(ctx)

	debug.Log("names (%d): %q", len(names), names)
	sort.Strings(names)
	for _, name := range names {
		pcb()
		if strings.HasPrefix(name, s.alias+"/") {
			name = strings.TrimPrefix(name, s.alias+"/")
		}
		ctx2 := ctxutil.WithNoNetwork(ctx, true)
		debug.Log("[%s] Checking %s", path, name)

		msg,err := s.fsckCheckEntry(ctx2, name)
		if err != nil {
			err_collector += fmt.Errorf("failed to check %q:\n    %w\n", name, err).Error()
			if msg == "F" {
				err_isfatal = true
			}
		} else {
			ctx = ctxutil.AddToCommitMessageBody(ctx, msg)
		}
	}


	if err_collector != "" && err_isfatal {
		return fmt.Errorf(err_collector)
	} else if err_collector != "" {
		out.Errorf(ctx, err_collector)
	} //else {
		//out.Warningf(ctx, "No visible error")
	//}
	if ctxutil.GetCommitMessageBody(ctx) == "" {
		out.Errorf(ctx, "Nothing to commit: all secrets seemed to have failed")
		return nil
	}

	t = queue.GetQueue(ctx).Add(func(ctx2 context.Context) (context.Context, error) {
		if err := s.storage.Commit(ctx, ctxutil.GetCommitMessageFull(ctx)); err != nil {
			switch {
			case errors.Is(err, store.ErrGitNotInit):
				out.Warning(ctx, "Cannot commit: git not initialized\nplease run `gopass git init` (and note that manual intervention might be needed)")
			case errors.Is(err, store.ErrGitNothingToCommit):
				debug.Log("commitAndPush - skipping git commit - nothing to commit")
			default:
				err := fmt.Errorf("failed to commit changes to git: %w", err)
				ctx2 = ctxutil.AddToErrorCollector(ctx2, err.Error())
				return ctx2, err
			}
		}

		if ctxutil.HasErrorCollector(ctx2) {
			out.Errorf(ctx2, ctxutil.GetErrorCollector(ctx2))
		}

		if err := s.storage.Push(ctx, "", ""); err != nil {
			if !errors.Is(err, store.ErrGitNoRemote) {
				ctx2 = ctxutil.AddToErrorCollector(ctx2, "RCS Push failed: "+ err.Error())
				out.Printf(ctx, "RCS Push failed: %s", err.Error())
			}
		}
		return ctx2, nil
	})
	ctx,err = t(ctx)
	return err
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

// the first returned element is either something to add to the commit message,
// or (if there is an error) the consequence of the error: "F"atal, "NF" nonfatal.
func (s *Store) fsckCheckEntry(ctx context.Context, name string) (string,error) {
	err_collector := ""
	recipients_need_fix := false

	errc,err := s.fsckCheckRecipients(ctx, name)
	if err != nil {
		if errc == "F" {
			return "NF",fmt.Errorf("Checking recipients for %s failed:\n    %s", name, err)
		} else {
			// the only NF errorf from that function are missing/extra recipients,
			// which isn't much of an error since we have yet to correct that.
			//err_collector += err.Error()
			recipients_need_fix = true
		}
	}

	// make sure we are actually allowed to decode this secret
	// if this fails there is no way we could fix anything
	if !IsFsckDecrypt(ctx) {
		if err_collector == "" {
			return "NF", nil
		} else {
			return "NF", fmt.Errorf(err_collector)
		}
	}

	// we need to make sure Parsing is enabled in order to parse old Mime secrets
	ctx = ctxutil.WithShowParsing(ctx, true)
	sec, err := s.Get(ctx, name)
	if err != nil {
		return "NF",fmt.Errorf("failed to decode secret %s: %s", name, err.Error())
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

	if recipients_need_fix {
		out.Printf(ctx, "Re-encrypting %s to fix recipients and storage format. [leaf store]", name)
	} else {
		out.Printf(ctx, "Re-encrypting %s to fix storage format. [leaf store]", name)
	}
	if err := s.Set(ctxutil.WithGitCommit(ctx,false), name, sec); err != nil {
		return "NF",fmt.Errorf("failed to write secret %s: %s", name, err.Error())
	}


	t := queue.GetQueue(ctx).Add(func(ctx2 context.Context) (context.Context, error) {
		errc,err := s.fsckCheckRecipients(ctx, name)
		if err != nil {
			if errc == "F" {
				return ctx2,fmt.Errorf("Checking recipients for %s failed:\n    %s", name, err)
			} else {
				ctxutil.AddToErrorCollector(ctx2, err.Error())
			}
		}
		return ctx2,nil
	})
        ctx,err = t(ctx)
	if err != nil {
		err_collector += err.Error() + "\n"
	}

	if err_collector == "" {
		return fmt.Sprintf("- re-encrypt secret %s", name),nil
	} else {
		return "NF", fmt.Errorf(err_collector)
	}
}

func (s *Store) fsckCheckRecipients(ctx context.Context, name string) (string,error) {
	// now compare the recipients this secret was encoded for and fix it if
	// it doesn't match.
	ciphertext, err := s.storage.Get(ctx, s.Passfile(name))
	if err != nil {
		return "F",fmt.Errorf("failed to get raw secret: %w", err)
	}

	itemRecps, err := s.crypto.RecipientIDs(ctx, ciphertext)
	if err != nil {
		return "F",fmt.Errorf("failed to read recipient IDs from raw secret: %w", err)
	}

	itemRecps = fingerprints(ctx, s.crypto, itemRecps)

	perItemStoreRecps, err := s.GetRecipients(ctx, name)
	if err != nil {
		return "F",fmt.Errorf("failed to get recipients from store: %w", err)
	}

	perItemStoreRecps = fingerprints(ctx, s.crypto, perItemStoreRecps)

	// check itemRecps matches storeRecps
	extra, missing := diff.List(perItemStoreRecps, itemRecps)
	if len(missing) > 0 && len(extra)>0 {
		return "NF",fmt.Errorf("Missing/extra recipients on %s: %+v/%+v\nRun fsck with the --decrypt flag to re-encrypt it automatically, or edit this secret yourself.", name, missing, extra)
	} else if len(missing) > 0 {
		return "NF",fmt.Errorf("Missing recipients on %s: %+v\nRun fsck with the --decrypt flag to re-encrypt it automatically, or edit this secret yourself.", name, missing)
	} else if len(extra) > 0 {
		return "NF",fmt.Errorf("Extra recipients on %s: %+v\nRun fsck with the --decrypt flag to re-encrypt it automatically, or edit this secret yourself.", name, extra)
	} else {
		return "",nil
	}
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
