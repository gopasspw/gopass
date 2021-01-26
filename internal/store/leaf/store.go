package leaf

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/pkg/debug"
)

// Store is password store
type Store struct {
	alias   string
	path    string
	crypto  backend.Crypto
	storage backend.Storage
}

// Init initializes this sub store
func Init(ctx context.Context, alias, path string) (*Store, error) {
	debug.Log("Initializing %s at %s", alias, path)
	s := &Store{
		alias: alias,
		path:  path,
	}

	st, err := backend.InitStorage(ctx, backend.GetStorageBackend(ctx), path)
	if err != nil {
		return nil, err
	}
	s.storage = st
	debug.Log("Storage initialized")

	crypto, err := backend.NewCrypto(ctx, backend.GetCryptoBackend(ctx))
	if err != nil {
		return nil, err
	}
	s.crypto = crypto
	debug.Log("Crypto initialized")

	return s, nil
}

// New creates a new store
func New(ctx context.Context, alias, path string) (*Store, error) {
	debug.Log("Instantiating %s at %s", alias, path)

	s := &Store{
		alias: alias,
		path:  path,
	}

	// init storage and rcs backend
	if err := s.initStorageBackend(ctx); err != nil {
		return nil, fmt.Errorf("failed to init storage backend: %w", err)
	}
	debug.Log("Storage initialized")

	// init crypto backend
	if err := s.initCryptoBackend(ctx); err != nil {
		return nil, fmt.Errorf("failed to init crypto backend: %w", err)
	}
	debug.Log("Crypto initialized")

	debug.Log("Instantiated %s at %s - storage: %+#v - crypto: %+#v", alias, path, s.storage, s.crypto)
	return s, nil
}

// idFile returns the path to the recipient list for this store
// it walks up from the given filename until it finds a directory containing
// a gpg id file or it leaves the scope of storage.
func (s *Store) idFile(ctx context.Context, name string) string {
	if s.crypto == nil {
		return ""
	}
	fn := name
	var cnt uint8
	for {
		cnt++
		if cnt > 100 {
			break
		}
		if fn == "" || fn == Sep {
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

// idFiles returns the path to all id files in this store.
func (s *Store) idFiles(ctx context.Context) []string {
	if s.crypto == nil {
		return nil
	}
	files, err := s.Storage().List(ctx, "")
	if err != nil {
		return nil
	}
	fileSet := make(map[string]struct{}, len(files))
	for _, file := range files {
		if strings.HasPrefix(filepath.Base(file), ".") {
			continue
		}
		idf := s.idFile(ctx, file)
		if s.storage.Exists(ctx, idf) {
			fileSet[idf] = struct{}{}
		}
	}
	out := make([]string, 0, len(fileSet))
	for file := range fileSet {
		out = append(out, file)
	}
	sort.Strings(out)
	return out
}

// Equals returns true if this.storage has the same on-disk path as the other
func (s *Store) Equals(other *Store) bool {
	if other == nil {
		return false
	}
	return s.Path() == other.Path()
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
		return nil, fmt.Errorf("failed to get recipients: %w", err)
	}

	if !IsCheckRecipients(ctx) {
		return rs, nil
	}

	kl, err := s.crypto.FindRecipients(ctx, rs...)
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
	return fmt.Sprintf("Store(Alias: %s, Path: %s)", s.alias, s.path)
}

// Path returns the value of path
func (s *Store) Path() string {
	return s.path
}

// Alias returns the value of alias
func (s *Store) Alias() string {
	return s.alias
}

// Storage returns the storage backend used by this.storage
func (s *Store) Storage() backend.Storage {
	return s.storage
}

// Valid returns true if this store is not nil
func (s *Store) Valid() bool {
	return s != nil
}
