package sub

import (
	"context"
	"sort"
	"strings"

	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/out"

	"github.com/pkg/errors"
)

// Fsck checks all entries matching the given prefix
func (s *Store) Fsck(ctx context.Context, path string) error {
	ctx = out.AddPrefix(ctx, "["+s.alias+"] ")
	out.Debug(ctx, "Fsck(%s)", path)

	// first let the storage backend check itself
	if err := s.storage.Fsck(ctx); err != nil {
		return errors.Wrapf(err, "storage backend found errors: %s", err)
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
		out.Debug(ctx, "sub.Fsck(%s) - Checking %s", path, name)
		if err := s.fsckCheckEntry(ctx, name); err != nil {
			return errors.Wrapf(err, "failed to check %s: %s", name, err)
		}
	}

	return nil
}

func (s *Store) fsckCheckEntry(ctx context.Context, name string) error {
	// make sure we can actually decode this secret
	// if this fails there is no way we could fix this
	if IsFsckDecrypt(ctx) {
		_, err := s.Get(ctx, name)
		if err != nil {
			return errors.Wrapf(err, "failed to decode secret %s: %s", name, err)
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
