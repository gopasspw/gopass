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
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/queue"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/fsutil"
	"github.com/gopasspw/gopass/pkg/termio"
)

// Convert will convert an existing store to a new store with possibly
// different set of crypto and storage backends. Please note that it
// will happily convert to the same set of backends if requested.
func (s *Store) Convert(ctx context.Context, cryptoBe backend.CryptoBackend, storageBe backend.StorageBackend, move bool) error {
	// use a temp queue so we can flush it before removing the old store
	q := queue.New(ctx)
	ctx = queue.WithQueue(ctx, q)

	// remove any previous attempts
	if pDir := filepath.Join(filepath.Dir(s.path), filepath.Base(s.path)+"-autoconvert"); fsutil.IsDir(pDir) {
		if err := os.RemoveAll(pDir); err != nil {
			return fmt.Errorf("failed to remove previous attempt %q: %w", pDir, err)
		}
	}

	// create temp path
	tmpPath := s.path + "-autoconvert"
	if err := os.MkdirAll(tmpPath, 0o700); err != nil {
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

	out.Printf(ctx, "Converting store ...")
	bar := termio.NewProgressBar(int64(len(entries)))
	bar.Hidden = ctxutil.IsHidden(ctx)
	if !ctxutil.IsTerminal(ctx) || ctxutil.IsHidden(ctx) {
		bar = nil
	}

	// Avoid network operations slowing down the bulk conversion.
	// We will sync with the remote later.
	ctx = ctxutil.WithNoNetwork(ctx, true)
	for _, e := range entries {
		e = strings.TrimPrefix(e, s.alias+Sep)
		debug.Log("converting %s", e)
		revs, err := s.ListRevisions(ctx, e)
		if err != nil {
			return err
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
			if err := tmpStore.Set(ctx, e, sec); err != nil {
				return err
			}
		}
		bar.Inc()
	}
	bar.Done()

	// flush queue
	_ = q.Close(ctx)

	if !move {
		return nil
	}

	// remove any previous backups
	bDir := filepath.Join(filepath.Dir(s.path), filepath.Base(s.path)+"-backup")
	if fsutil.IsDir(bDir) {
		if err := os.RemoveAll(bDir); err != nil {
			debug.Log("failed to remove previous backup %q: %s", bDir, err)
		}
	}

	// rename old to backup
	if err := os.Rename(s.path, bDir); err != nil {
		return err
	}

	// rename temp to old
	return os.Rename(tmpPath, s.path)
}
