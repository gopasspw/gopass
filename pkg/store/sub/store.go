package sub

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gopasspw/gopass/pkg/agent/client"
	"github.com/gopasspw/gopass/pkg/backend"
	"github.com/gopasspw/gopass/pkg/backend/rcs/noop"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/out"
	"github.com/gopasspw/gopass/pkg/store"

	"github.com/muesli/goprogressbar"
	"github.com/pkg/errors"
)

type recipientHashStorer interface {
	CheckRecipientHash(string) bool
	GetRecipientHash(string, string) string
	SetRecipientHash(string, string, string) error
}

// Store is password store
type Store struct {
	alias   string
	url     *backend.URL
	crypto  backend.Crypto
	rcs     backend.RCS
	storage backend.Storage
	cfgdir  string
	agent   *client.Client
	sc      recipientHashStorer
}

// New creates a new store, copying settings from the given root store
func New(ctx context.Context, sc recipientHashStorer, alias string, u *backend.URL, cfgdir string, agent *client.Client) (*Store, error) {
	out.Debug(ctx, "sub.New - URL: %s", u.String())

	s := &Store{
		alias:  alias,
		url:    u,
		rcs:    noop.New(),
		cfgdir: cfgdir,
		agent:  agent,
		sc:     sc,
	}

	// init store backend
	if backend.HasStorageBackend(ctx) {
		s.url.Storage = backend.GetStorageBackend(ctx)
		out.Debug(ctx, "sub.New - Using storage backend from ctx: %s", backend.StorageBackendName(s.url.Storage))
	}
	if err := s.initStorageBackend(ctx); err != nil {
		return nil, errors.Wrapf(err, "failed to init storage backend: %s", err)
	}

	// init sync backend
	if backend.HasRCSBackend(ctx) {
		s.url.RCS = backend.GetRCSBackend(ctx)
		out.Debug(ctx, "sub.New - Using RCS backend from ctx: %s", backend.RCSBackendName(s.url.RCS))
	}
	if err := s.initRCSBackend(ctx); err != nil {
		return nil, errors.Wrapf(err, "failed to init RCS backend: %s", err)
	}

	// init crypto backend
	if backend.HasCryptoBackend(ctx) {
		s.url.Crypto = backend.GetCryptoBackend(ctx)
		out.Debug(ctx, "sub.New - Using Crypto backend from ctx: %s", backend.CryptoBackendName(s.url.Crypto))
	}
	if err := s.initCryptoBackend(ctx); err != nil {
		return nil, errors.Wrapf(err, "failed to init crypto backend: %s", err)
	}

	out.Debug(ctx, "sub.New - initialized - storage: %s (%p) - rcs: %s (%p) - crypto: %s (%p)", s.storage.Name(), s.storage, s.rcs.Name(), s.rcs, s.crypto.Name(), s.crypto)
	return s, nil
}

// idFile returns the path to the recipient list for this.storage
// it walks up from the given filename until it finds a directory containing
// a gpg id file or it leaves the scope of this.storage.
func (s *Store) idFile(ctx context.Context, name string) string {
	fn := name
	var cnt uint8
	for {
		cnt++
		if cnt > 100 {
			break
		}
		if fn == "" || fn == sep {
			break
		}
		gfn := filepath.Join(fn, s.crypto.IDFile())
		if s.storage.Exists(ctx, gfn) {
			return gfn
		}
		fn = filepath.Dir(fn)
	}
	return s.crypto.IDFile()
}

// Equals returns true if this.storage has the same on-disk path as the other
func (s *Store) Equals(other store.Store) bool {
	if other == nil {
		return false
	}
	return s.URL() == other.URL()
}

// IsDir returns true if the entry is folder inside the store
func (s *Store) IsDir(ctx context.Context, name string) bool {
	return s.storage.IsDir(ctx, name)
}

// Exists checks the existence of a single entry
func (s *Store) Exists(ctx context.Context, name string) bool {
	return s.storage.Exists(ctx, s.passfile(name))
}

func (s *Store) useableKeys(ctx context.Context, name string) ([]string, error) {
	rs, err := s.GetRecipients(ctx, name)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get recipients")
	}

	if !IsCheckRecipients(ctx) {
		return rs, nil
	}

	kl, err := s.crypto.FindPublicKeys(ctx, rs...)
	if err != nil {
		return rs, err
	}

	return kl, nil
}

// passfile returns the name of gpg file on disk, for the given key/name
func (s *Store) passfile(name string) string {
	return strings.TrimPrefix(name+"."+s.crypto.Ext(), "/")
}

// String implement fmt.Stringer
func (s *Store) String() string {
	return fmt.Sprintf("Store(Alias: %s, Path: %s)", s.alias, s.url.String())
}

// reencrypt will re-encrypt all entries for the current recipients
func (s *Store) reencrypt(ctx context.Context) error {
	entries, err := s.List(ctx, "")
	if err != nil {
		return errors.Wrapf(err, "failed to list store")
	}

	// save original value of auto push
	{
		// shadow ctx in this block only
		ctx := WithAutoSync(ctx, false)
		ctx = ctxutil.WithGitCommit(ctx, false)

		// progress bar
		bar := &goprogressbar.ProgressBar{
			Total: int64(len(entries)),
			Width: 120,
		}
		if !ctxutil.IsTerminal(ctx) || out.IsHidden(ctx) {
			bar = nil
		}
		var wg sync.WaitGroup
		jobs := make(chan string)
		// We use a logger to write without race condition on stdout
		logger := log.New(os.Stdout, "", 0)
		fmt.Println("Starting rencrypt")
		// We spawn as many workers as we have set in the concurrency setting
		// GetConcurrency will return 1 if the concurrency setting is not set
		// or if it set to a value below 1.
		for i := 0; i < ctxutil.GetConcurrency(ctx); i++ {
			wg.Add(1) // we start a new job
			go func(workerId int) {
				// the workers are fed through an unbuffered channel
				for e := range jobs {
					content, err := s.Get(ctx, e)
					if err != nil {
						logger.Printf("Worker %d: Failed to get current value for %s: %s\n", workerId, e, err)
						continue
					}
					if err := s.Set(ctx, e, content); err != nil {
						logger.Printf("Worker %d: Failed to write %s: %s\n", workerId, e, err)
						continue
					}
				}
				wg.Done() // report the job as finished
			}(i)
		}
		for _, e := range entries {
			// check for context cancelation
			select {
			case <-ctx.Done():
				// We close the channel, so the worker will terminate
				close(jobs)
				// we wait for all workers to have finished
				wg.Wait()
				return errors.New("context canceled")
			default:
			}

			if bar != nil {
				bar.Current++
				bar.Text = fmt.Sprintf("%d of %d secrets reencrypted", bar.Current, bar.Total)
				bar.LazyPrint()
			}

			e = strings.TrimPrefix(e, s.alias)
			jobs <- e
		}
		// We close the channel, so the workers will terminate
		close(jobs)
		// we wait for all workers to have finished
		wg.Wait()
	}

	// if we were working concurrently, we couldn't git add during the process
	// to avoid a race condition on git .index.lock file, so we do it now.
	if ctxutil.HasConcurrency(ctx) {
		for _, name := range entries {
			p := s.passfile(name)
			if err := s.rcs.Add(ctx, p); err != nil {
				switch errors.Cause(err) {
				case store.ErrGitNotInit:
					out.Debug(ctx, "reencrypt - skipping git add - git not initialized")
					continue
				default:
					return errors.Wrapf(err, "failed to add '%s' to git", p)
				}
			}
			out.Debug(ctx, "reencrypt - added %s to git", p)
		}
	}

	if err := s.rcs.Commit(ctx, GetReason(ctx)); err != nil {
		switch errors.Cause(err) {
		case store.ErrGitNotInit:
			out.Debug(ctx, "reencrypt - skipping git commit - git not initialized")
		default:
			return errors.Wrapf(err, "failed to commit changes to git")
		}
	}

	if !IsAutoSync(ctx) {
		out.Debug(ctx, "reencrypt - auto sync is disabled")
		return nil
	}

	return s.reencryptGitPush(ctx)
}

func (s *Store) reencryptGitPush(ctx context.Context) error {
	if err := s.rcs.Push(ctx, "", ""); err != nil {
		if errors.Cause(err) == store.ErrGitNotInit {
			msg := "Warning: git is not initialized for this.storage. Ignoring auto-push option\n" +
				"Run: gopass git init"
			out.Red(ctx, msg)
			return nil
		}
		if errors.Cause(err) == store.ErrGitNoRemote {
			msg := "Warning: git has no remote. Ignoring auto-push option\n" +
				"Run: gopass git remote add origin ..."
			out.Yellow(ctx, msg)
			return nil
		}
		return errors.Wrapf(err, "failed to push change to git remote")
	}
	return nil
}

// Path returns the value of path
func (s *Store) Path() string {
	if s.url == nil {
		return ""
	}
	return s.url.Path
}

// Alias returns the value of alias
func (s *Store) Alias() string {
	return s.alias
}

// URL returns the store URL
func (s *Store) URL() string {
	return s.url.String()
}

// Storage returns the storage backend used by this.storage
func (s *Store) Storage() backend.Storage {
	return s.storage
}

// Valid returns true if this store is not nil
func (s *Store) Valid() bool {
	return s != nil
}
