package leaf

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/set"
	"github.com/gopasspw/gopass/pkg/debug"
)

// Store is a password store.
type Store struct {
	alias   string
	path    string
	crypto  backend.Crypto
	storage backend.Storage
}

// Init initializes this sub store.
func Init(ctx context.Context, alias, path string) (*Store, error) {
	debug.Log("Initializing %s at %s", alias, path)

	s := &Store{
		alias: alias,
		path:  path,
	}

	st, err := backend.InitStorage(ctx, backend.GetStorageBackend(ctx), path)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage for %s at %s: %w", alias, path, err)
	}

	s.storage = st
	debug.Log("Storage for %s => %s initialized as %s", alias, path, st.Name())

	crypto, err := backend.NewCrypto(ctx, backend.GetCryptoBackend(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize crypto for %s at %s: %w", alias, path, err)
	}

	s.crypto = crypto
	debug.Log("Crypto for %q => %q initialized as %s", alias, path, crypto.Name())

	return s, nil
}

// New creates a new store.
func New(ctx context.Context, alias, path string) (*Store, error) {
	debug.Log("Instantiating %q at %q", alias, path)

	s := &Store{
		alias: alias,
		path:  path,
	}

	// init storage and rcs backend
	if err := s.initStorageBackend(ctx); err != nil {
		return nil, fmt.Errorf("failed to init storage backend: %w", err)
	}

	debug.Log("Storage for %s => %s initialized as %v", alias, path, s.storage)

	// init crypto backend
	if err := s.initCryptoBackend(ctx); err != nil {
		return nil, fmt.Errorf("failed to init crypto backend: %w", err)
	}

	debug.Log("Crypto for %s => %s initialized as %v", alias, path, s.crypto)

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
	if s == nil || s.crypto == nil {
		return nil
	}

	files, err := s.Storage().List(ctx, "")
	if err != nil {
		return nil
	}

	// we need to transform the list of files into a list of id files so we can't use
	// set.SortedFiltered as it doesn't support transformations
	idfs := make([]string, 0, len(files))

	for _, f := range files {
		if strings.HasPrefix(filepath.Base(f), ".") {
			continue
		}

		idf := s.idFile(ctx, f)
		debug.Log("checking for if %q has an idf: %q", f, idf)
		if s.storage.Exists(ctx, idf) {
			idfs = append(idfs, idf)
		}
	}

	debug.Log("idFiles: %q", idfs)

	return set.Sorted(idfs)
}

// Equals returns true if this storage has the same on-disk path as the other.
func (s *Store) Equals(other *Store) bool {
	if other == nil {
		return false
	}

	return s.Path() == other.Path()
}

// IsDir returns true if the entry is folder inside the store.
func (s *Store) IsDir(ctx context.Context, name string) bool {
	return s.storage.IsDir(ctx, name)
}

// Exists checks the existence of a single entry.
func (s *Store) Exists(ctx context.Context, name string) bool {
	return s.storage.Exists(ctx, s.Passfile(name))
}

func (s *Store) useableKeys(ctx context.Context, name string) ([]string, error) {
	rs, err := s.GetRecipients(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get recipients: %w", err)
	}

	if !IsCheckRecipients(ctx) {
		return rs.IDs(), nil
	}

	kl, err := s.crypto.FindRecipients(ctx, rs.IDs()...)
	if err != nil {
		return rs.IDs(), err
	}

	return kl, nil
}

// Passfile returns the name of gpg file on disk, for the given key/name.
func (s *Store) Passfile(name string) string {
	return strings.TrimPrefix(name+"."+s.crypto.Ext(), "/")
}

// String implement fmt.Stringer.
func (s *Store) String() string {
	return fmt.Sprintf("Store(Alias: %s, Path: %s)", s.alias, s.path)
}

// Path returns the value of path.
func (s *Store) Path() string {
	return s.path
}

// Alias returns the value of alias.
func (s *Store) Alias() string {
	return s.alias
}

// Storage returns the storage backend used by this store.
func (s *Store) Storage() backend.Storage {
	return s.storage
}

// Valid returns true if this store is not nil.
func (s *Store) Valid() bool {
	return s != nil
}

// Concurrency returns the number of concurrent operations allowed
// by this stores crypto implementation (e.g. usually 1 for GPG).
func (s *Store) Concurrency() int {
	return s.Crypto().Concurrency()
}
