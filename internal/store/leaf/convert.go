package leaf

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/cui"
	"github.com/gopasspw/gopass/internal/debug"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
)

// Convert will convert an existing store to a new store with possibly
// different set of crypto and storage backends. Please note that it
// will happily convert to the same set of backends if requested.
func (s *Store) Convert(ctx context.Context, cryptoBe backend.CryptoBackend, storageBe backend.StorageBackend, move bool) error {

	// create temp path
	tmpPath := s.path + "-autoconvert"
	if err := os.MkdirAll(tmpPath, 0700); err != nil {
		return err
	}
	debug.Log("create temporary store path for conversion: %s", tmpPath)

	// init new store at temp path
	st, err := backend.InitStorage(ctx, storageBe, tmpPath)
	if err != nil {
		return err
	}
	debug.Log("initialized storage %s at %s", st, tmpPath)

	crypto, err := backend.NewCrypto(ctx, cryptoBe)
	if err != nil {
		return err
	}
	debug.Log("initialized Crypto %s", crypto)
	// TODO need to initialize recipients

	tmpStore := &Store{
		alias:   s.alias,
		path:    tmpPath,
		crypto:  crypto,
		storage: st,
	}

	// init new store
	key, err := cui.AskForPrivateKey(ctx, crypto, "Please select a private key")
	if err != nil {
		return err
	}
	if err := tmpStore.Init(ctx, tmpPath, key); err != nil {
		return err
	}

	// copy everything from old to temp, including all revisions
	entries, err := s.List(ctx, "")
	if err != nil {
		return err
	}

	out.Green(ctx, "Converting store ...")
	bar := out.NewProgressBar(ctx, int64(len(entries)))
	if !ctxutil.IsTerminal(ctx) || out.IsHidden(ctx) {
		bar = nil
	}

	ctx = ctxutil.WithNoNetwork(ctx, true)
	for _, e := range entries {
		e = strings.TrimPrefix(e, s.alias+sep)
		debug.Log("converting %s", e)
		revs, err := s.ListRevisions(ctx, e)
		if err != nil {
			return err
		}
		if len(revs) < 2 {
			debug.Log("entry %s has no revisions. convering latest", e)
			sec, err := s.Get(ctx, e)
			if err != nil {
				return err
			}
			if err := tmpStore.Set(ctx, e, sec.MIME()); err != nil {
				return err
			}
			continue
		}
		sort.Sort(sort.Reverse(backend.Revisions(revs)))
		for _, r := range revs {
			debug.Log("converting %s@%s", e, r.Hash)
			sec, err := s.GetRevision(ctx, e, r.Hash)
			if err != nil {
				return err
			}
			msg := fmt.Sprintf("%s\n%s\nCommitted as: %s\nDate: %s\nAuthor: %s <%s>",
				r.Subject,
				r.Body,
				r.Hash,
				r.Date.Format(time.RFC3339),
				r.AuthorName,
				r.AuthorEmail,
			)
			ctx := ctxutil.WithCommitMessage(ctx, msg)
			ctx = ctxutil.WithCommitTimestamp(ctx, r.Date)
			if err := tmpStore.Set(ctx, e, sec.MIME()); err != nil {
				return err
			}
		}
		bar.Inc()
	}
	bar.Done()

	if !move {
		return nil
	}

	// rename old to backup
	if err := os.Rename(s.path, filepath.Join(filepath.Dir(s.path), filepath.Base(s.path)+"-backup")); err != nil {
		return err
	}
	// rename temp to old
	return os.Rename(tmpPath, s.path)
}
