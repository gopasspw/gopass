package leaf

import (
	"context"
	"sort"
	"strings"

	"github.com/gopasspw/gopass/internal/debug"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"

	"github.com/pkg/errors"
)

// Fsck checks all entries matching the given prefix
func (s *Store) Fsck(ctx context.Context, path string) error {
	ctx = out.AddPrefix(ctx, "["+s.alias+"] ")
	debug.Log("Checking %s", path)

	// first let the storage backend check itself
	if err := s.storage.Fsck(ctx); err != nil {
		return errors.Wrapf(err, "storage backend found errors: %s", err)
	}

	// second give a chance to the RCS backend
	if err := s.rcs.Compact(ctx); err != nil {
		return errors.Wrapf(err, "rcs backend compaction failed: %s", err)
	}

	pcb := ctxutil.GetProgressCallback(ctx)

	// then we'll make sure all the secrets are readable by us and every
	// valid recipient
	names, err := s.List(ctx, path)
	if err != nil {
		return errors.Wrapf(err, "failed to list entries: %s", err)
	}
	sort.Strings(names)
	for _, name := range names {
		pcb()
		if strings.HasPrefix(name, s.alias+"/") {
			name = strings.TrimPrefix(name, s.alias+"/")
		}
		ctx := ctxutil.WithNoNetwork(ctx, true)
		debug.Log("[%s] Checking %s", path, name)
		if err := s.fsckCheckEntry(ctx, name); err != nil {
			return errors.Wrapf(err, "failed to check %s: %s", name, err)
		}
	}

	if err := s.rcs.Push(ctx, "", ""); err != nil {
		out.Red(ctx, "RCS Push failed: %s", err)
	}

	return nil
}

func (s *Store) fsckCheckEntry(ctx context.Context, name string) error {
	// make sure we can actually decode this secret
	// if this fails there is no way we could fix this
	if IsFsckDecrypt(ctx) {
		oSec, err := s.Get(ctx, name)
		if err != nil {
			return errors.Wrapf(err, "failed to decode secret %s: %s", name, err)
		}
		// check if we need to convert the secret it self
		nSec := oSec.MIME()
		if string(nSec.Bytes()) != string(oSec.Bytes()) {
			if err := s.Set(ctx, name, nSec); err != nil {
				return errors.Wrapf(err, "failed to convert secret %s: %s", name, err)
			}
		}
	}

	// now compare the recipients this secret was encoded for and fix it if
	// if doesn't match
	ciphertext, err := s.storage.Get(ctx, s.passfile(name))
	if err != nil {
		return errors.Wrapf(err, "failed to get raw secret: %s", err)
	}
	itemRecps, err := s.crypto.RecipientIDs(ctx, ciphertext)
	if err != nil {
		return errors.Wrapf(err, "failed to read recipient IDs from raw secret: %s", err)
	}
	perItemStoreRecps, err := s.GetRecipients(ctx, name)
	if err != nil {
		return errors.Wrapf(err, "failed to get recipients from store: %s", err)
	}

	// check itemRecps matches storeRecps
	missing, extra := compareStringSlices(perItemStoreRecps, itemRecps)
	if len(missing) > 0 {
		out.Error(ctx, "Missing recipients on %s: %+v", name, missing)
	}
	if len(extra) > 0 {
		out.Error(ctx, "Extra recipients on %s: %+v", name, extra)
	}
	if len(missing) > 0 || len(extra) > 0 {
		sec, err := s.Get(ctx, name)
		if err != nil {
			return errors.Wrapf(err, "failed to decode secret: %s", err)
		}
		if err := s.Set(ctxutil.WithCommitMessage(ctx, "fsck fix recipients"), name, sec); err != nil {
			return errors.Wrapf(err, "failed to write secret: %s", err)
		}
	}

	return nil
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

	return missing, extra
}
