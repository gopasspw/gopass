package sub

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	gitcli "github.com/justwatchcom/gopass/backend/git/cli"
	"github.com/justwatchcom/gopass/store"
	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/justwatchcom/gopass/utils/fsutil"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/muesli/goprogressbar"
	"github.com/pkg/errors"
)

const (
	// GPGID is the name of the file containing the recipient ids
	GPGID = ".gpg-id"
)

// Store is password store
type Store struct {
	alias string
	path  string
	gpg   gpger
	git   giter
}

// New creates a new store, copying settings from the given root store
func New(alias string, path string, gpg gpger) *Store {
	path = fsutil.CleanPath(path)
	return &Store{
		alias: alias,
		path:  path,
		gpg:   gpg,
		git:   gitcli.New(path, gpg.Binary()),
	}
}

// idFile returns the path to the recipient list for this store
// it walks up from the given filename until it finds a directory containing
// a gpg id file or it leaves the scope of this store.
func (s *Store) idFile(name string) string {
	fn, err := filepath.Abs(filepath.Join(s.path, name))
	if err != nil {
		panic(err)
	}
	var cnt uint8
	for {
		cnt++
		if cnt > 100 {
			break
		}
		if fn == "" || fn == sep {
			break
		}
		if !strings.HasPrefix(fn, s.path) {
			break
		}
		gfn := filepath.Join(fn, GPGID)
		fi, err := os.Stat(gfn)
		if err == nil && !fi.IsDir() {
			return gfn
		}
		fn = filepath.Dir(fn)
	}
	return fsutil.CleanPath(filepath.Join(s.path, GPGID))
}

// Equals returns true if this store has the same on-disk path as the other
func (s *Store) Equals(other *Store) bool {
	if other == nil {
		return false
	}
	return s.path == other.path
}

// IsDir returns true if the entry is folder inside the store
func (s *Store) IsDir(name string) bool {
	return fsutil.IsDir(filepath.Join(s.path, name))
}

// Exists checks the existence of a single entry
func (s *Store) Exists(name string) bool {
	p := s.passfile(name)

	if !strings.HasPrefix(p, s.path) {
		return false
	}

	return fsutil.IsFile(p)
}

func (s *Store) useableKeys(ctx context.Context, name string) ([]string, error) {
	rs, err := s.GetRecipients(ctx, name)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get recipients")
	}

	if !IsCheckRecipients(ctx) {
		return rs, nil
	}

	kl, err := s.gpg.FindPublicKeys(ctx, rs...)
	if err != nil {
		return rs, err
	}

	unusable := kl.UnusableKeys()
	if len(unusable) > 0 {
		out.Red(ctx, "Unusable public keys detected (IGNORING FOR ENCRYPTION):")
		for _, k := range unusable {
			out.Red(ctx, "  - %s", k.OneLine())
		}
	}
	return kl.UseableKeys().Recipients(), nil
}

// passfile returns the name of gpg file on disk, for the given key/name
func (s *Store) passfile(name string) string {
	return fsutil.CleanPath(filepath.Join(s.path, name) + ".gpg")
}

// String implement fmt.Stringer
func (s *Store) String() string {
	return fmt.Sprintf("Store(Alias: %s, Path: %s)", s.alias, s.path)
}

func (s *Store) filenameToName(fn string) string {
	return strings.TrimPrefix(strings.TrimSuffix(fn, ".gpg"), s.path+sep)
}

// reencrypt will re-encrypt all entries for the current recipients
func (s *Store) reencrypt(ctx context.Context) error {
	entries, err := s.List("")
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
		if !ctxutil.IsTerminal(ctx) {
			bar = nil
		}
		for _, e := range entries {
			// check for context cancelation
			select {
			case <-ctx.Done():
				return errors.New("context canceled")
			default:
			}

			if bar != nil {
				bar.Current++
				bar.Text = fmt.Sprintf("%d of %d secrets reencrypted", bar.Current, bar.Total)
				bar.LazyPrint()
			}

			content, err := s.Get(ctx, e)
			if err != nil {
				fmt.Printf("Failed to get current value for %s: %s\n", e, err)
				continue
			}
			if err := s.Set(ctx, e, content); err != nil {
				fmt.Printf("Failed to write %s: %s\n", e, err)
			}
		}
	}

	if err := s.git.Commit(ctx, GetReason(ctx)); err != nil {
		if errors.Cause(err) != store.ErrGitNotInit {
			return errors.Wrapf(err, "failed to commit changes to git")
		}
	}

	if !IsAutoSync(ctx) {
		return nil
	}

	if err := s.git.Push(ctx, "", ""); err != nil {
		if errors.Cause(err) == store.ErrGitNotInit {
			msg := "Warning: git is not initialized for this store. Ignoring auto-push option\n" +
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
	return s.path
}

// Alias returns the value of alias
func (s *Store) Alias() string {
	return s.alias
}
