package sub

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/blang/semver"
	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/backend/gpg"
	gpgcli "github.com/justwatchcom/gopass/backend/gpg/cli"
	"github.com/justwatchcom/gopass/store"
	"github.com/justwatchcom/gopass/utils/fsutil"
	"github.com/pkg/errors"
)

const (
	// GPGID is the name of the file containing the recipient ids
	GPGID = ".gpg-id"
)

type gpger interface {
	ListPublicKeys(context.Context) (gpg.KeyList, error)
	FindPublicKeys(context.Context, ...string) (gpg.KeyList, error)
	ListPrivateKeys(context.Context) (gpg.KeyList, error)
	FindPrivateKeys(context.Context, ...string) (gpg.KeyList, error)
	GetRecipients(context.Context, string) ([]string, error)
	Encrypt(context.Context, string, []byte, []string) error
	Decrypt(context.Context, string) ([]byte, error)
	ExportPublicKey(context.Context, string, string) error
	ImportPublicKey(context.Context, string) error
	Version(context.Context) semver.Version
}

// Store is password store
type Store struct {
	alias string
	path  string
	gpg   gpger
}

// New creates a new store, copying settings from the given root store
func New(alias string, path string) *Store {
	return &Store{
		alias: alias,
		path:  fsutil.CleanPath(path),
		gpg:   gpgcli.New(gpgcli.Config{}),
	}
}

// idFile returns the path to the recipient list for this store
// it walks up from the given filename until it finds a directoy containing
// a gpg id file or it leaves the scope of this store.
func (s *Store) idFile(fn string) string {
	fn, err := filepath.Abs(filepath.Join(s.path, fn))
	if err != nil {
		panic(err)
	}
	var cnt uint8
	for {
		cnt++
		if cnt > 100 {
			break
		}
		if fn == "" || fn == "/" {
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

func (s *Store) useableKeys(ctx context.Context, file string) ([]string, error) {
	rs, err := s.GetRecipients(ctx, file)
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

	unuseable := kl.UnuseableKeys()
	if len(unuseable) > 0 {
		fmt.Println(color.RedString("Unuseable public keys detected (IGNORING FOR ENCRYPTION):"))
		for _, k := range unuseable {
			fmt.Println(color.RedString("  - %s", k.OneLine()))
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
	return strings.TrimPrefix(strings.TrimSuffix(fn, ".gpg"), s.path+"/")
}

// reencrypt will re-encrypt all entries for the current recipients
func (s *Store) reencrypt(ctx context.Context) error {
	entries, err := s.List("")
	if err != nil {
		return errors.Wrapf(err, "failed to list store")
	}

	// save original value of auto push
	ctx2 := WithAutoSync(ctx, false)
	for _, e := range entries {
		content, err := s.Get(ctx2, e)
		if err != nil {
			fmt.Printf("Failed to get current value for %s: %s\n", e, err)
			continue
		}
		if err := s.Set(ctx2, e, content); err != nil {
			fmt.Printf("Failed to write %s: %s\n", e, err)
		}
	}

	if !IsAutoSync(ctx) {
		return nil
	}

	if err := s.GitPush(ctx, "", ""); err != nil {
		if errors.Cause(err) == store.ErrGitNotInit {
			msg := "Warning: git is not initialized for this store. Ignoring auto-push option\n" +
				"Run: gopass git init"
			fmt.Println(color.RedString(msg))
			return nil
		}
		if errors.Cause(err) == store.ErrGitNoRemote {
			msg := "Warning: git has not remote. Ignoring auto-push option\n" +
				"Run: gopass git remote add origin ..."
			fmt.Println(color.YellowString(msg))
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
