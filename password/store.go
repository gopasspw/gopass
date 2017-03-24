package password

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/fsutil"
	"github.com/justwatchcom/gopass/gpg"
)

const (
	gpgID = ".gpg-id"
)

var (
	// ErrExistsFailed is returend if we can't check for existence
	ErrExistsFailed = fmt.Errorf("Failed to check for existence")
	// ErrNotFound is returned if an entry was not found
	ErrNotFound = fmt.Errorf("Entry is not in the password store")
	// ErrEncrypt is returned if we failed to encrypt an entry
	ErrEncrypt = fmt.Errorf("Failed to encrypt")
	// ErrDecrypt is returned if we failed to decrypt and entry
	ErrDecrypt = fmt.Errorf("Failed to decrypt")
	// ErrSneaky is returned if the user passes a possible malicious path to gopass
	ErrSneaky = fmt.Errorf("you've attempted to pass a sneaky path to gopass. go home")
)

// RecipientCallback is a callback to verify the list of recipients
type RecipientCallback func(string, []string) ([]string, error)

// ImportCallback is a callback to ask the user if he wants to import
// a certain recipients public key into his keystore
type ImportCallback func(string) bool

// FsckCallback is a callback to ask the user to confirm certain fsck
// corrective actions
type FsckCallback func(string) bool

// Store is password store
type Store struct {
	recipients  []string
	alias       string
	path        string
	autoPush    bool
	autoPull    bool
	autoImport  bool
	persistKeys bool
	loadKeys    bool
	alwaysTrust bool
	importFunc  ImportCallback
	fsckFunc    FsckCallback
	debug       bool
}

// NewStore creates a new store, copying settings from the given root store
func NewStore(alias, path string, r *RootStore) (*Store, error) {
	if r == nil {
		r = &RootStore{}
	}
	if path == "" {
		return nil, fmt.Errorf("Need path")
	}
	s := &Store{
		alias:       alias,
		path:        path,
		autoPush:    r.AutoPush,
		autoPull:    r.AutoPull,
		autoImport:  r.AutoImport,
		persistKeys: r.PersistKeys,
		loadKeys:    r.LoadKeys,
		alwaysTrust: r.AlwaysTrust,
		importFunc:  r.ImportFunc,
		fsckFunc:    r.FsckFunc,
		debug:       r.Debug,
		recipients:  make([]string, 0, 5),
	}

	// only try to load recipients if the store / recipients file exist
	if fsutil.IsFile(s.idFile()) {
		keys, err := s.loadRecipients()
		if err != nil {
			return nil, err
		}
		s.recipients = keys
	}
	return s, nil
}

// Initialized returns true if the store is properly initialized
func (s *Store) Initialized() bool {
	return fsutil.IsFile(s.idFile())
}

// Init tries to initalize a new password store location matching the object
func (s *Store) Init(ids ...string) error {
	if s.Initialized() {
		return fmt.Errorf("Store is already initialized")
	}

	// initialize recipient list
	s.recipients = make([]string, 0, len(ids))

	for _, id := range ids {
		if id == "" {
			continue
		}
		kl, err := gpg.ListPublicKeys(id)
		if err != nil || len(kl) < 1 {
			fmt.Println("Failed to fetch public key:", id)
			continue
		}
		s.recipients = append(s.recipients, kl[0].Fingerprint)
	}

	if len(s.recipients) < 1 {
		return fmt.Errorf("failed to initialize store: no valid recipients given")
	}

	kl, err := gpg.ListPrivateKeys(s.recipients...)
	if err != nil {
		return fmt.Errorf("Failed to get available private keys: %s", err)
	}

	if len(kl) < 1 {
		return fmt.Errorf("None of the recipients has a secret key. You will not be able to decrypt the secrets you add")
	}

	if err := s.saveRecipients("Initialized Store for " + strings.Join(s.recipients, ", ")); err != nil {
		return fmt.Errorf("failed to initialize store: %v", err)
	}

	return nil
}

// idFile returns the path to the recipient list for this store
func (s *Store) idFile() string {
	return fsutil.CleanPath(filepath.Join(s.path, gpgID))
}

// mkStoreWalkerFunc create a func to walk a (sub)store, i.e. list it's content
func mkStoreWalkerFunc(alias, folder string, fn func(...string)) func(string, os.FileInfo, error) error {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") && path != folder {
			return filepath.SkipDir
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasPrefix(info.Name(), ".") {
			return nil
		}
		if path == folder {
			return nil
		}
		if path == filepath.Join(folder, gpgID) {
			return nil
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return nil
		}
		s := strings.TrimPrefix(path, folder+"/")
		s = strings.TrimSuffix(s, ".gpg")
		if alias != "" {
			s = alias + "/" + s
		}
		fn(s)
		return nil
	}
}

// List will list all entries in this store
func (s *Store) List(prefix string) ([]string, error) {
	lst := make([]string, 0, 10)
	addFunc := func(in ...string) {
		for _, s := range in {
			lst = append(lst, s)
		}
	}

	if err := filepath.Walk(s.path, mkStoreWalkerFunc(prefix, s.path, addFunc)); err != nil {
		return lst, err
	}

	return lst, nil
}

// equals returns true if this store has the same on-disk path as the other
func (s *Store) equals(other *Store) bool {
	if other == nil {
		return false
	}
	return s.path == other.path
}

// Get returns the plaintext of a single key
func (s *Store) Get(name string) ([]byte, error) {
	p := s.passfile(name)

	if !strings.HasPrefix(p, s.path) {
		return []byte{}, ErrSneaky
	}

	if !fsutil.IsFile(p) {
		return []byte{}, ErrNotFound
	}

	content, err := gpg.Decrypt(p)
	if err != nil {
		return []byte{}, ErrDecrypt
	}

	return content, nil
}

// IsDir returns true if the entry is folder inside the store
func (s *Store) IsDir(name string) bool {
	return fsutil.IsDir(filepath.Join(s.path, name))
}

// Exists checks the existence of a single entry
func (s *Store) Exists(name string) (bool, error) {
	p := s.passfile(name)

	if !strings.HasPrefix(p, s.path) {
		return false, ErrSneaky
	}

	return fsutil.IsFile(p), nil
}

// Set encodes and write the ciphertext of one entry to disk
func (s *Store) Set(name string, content []byte, reason string) error {
	return s.SetConfirm(name, content, reason, nil)
}

// SetConfirm encodes and writes the cipertext of one entry to disk. This
// method can be passed a callback to confirm the recipients immedeately
// before encryption.
func (s *Store) SetConfirm(name string, content []byte, reason string, cb RecipientCallback) error {
	p := s.passfile(name)

	if !strings.HasPrefix(p, s.path) {
		return ErrSneaky
	}

	if s.IsDir(name) {
		return fmt.Errorf("a folder named %s already exists", name)
	}

	recipients := make([]string, len(s.recipients))
	copy(recipients, s.recipients)

	// confirm recipients
	if cb != nil {
		newRecipients, err := cb(name, recipients)
		if err != nil {
			return err
		}
		recipients = newRecipients
	}

	if err := gpg.Encrypt(p, content, recipients, s.alwaysTrust); err != nil {
		return ErrEncrypt
	}

	if err := s.gitAdd(p); err != nil {
		if err == ErrGitNotInit {
			return nil
		}
		return err
	}

	if err := s.gitCommit(fmt.Sprintf("Save secret to %s: %s", name, reason)); err != nil {
		if err == ErrGitNotInit {
			return nil
		}
		return err
	}

	if s.autoPush {
		if err := s.gitPush("", ""); err != nil {
			if err == ErrGitNotInit {
				msg := "Warning: git is not initialized for this store. Ignoring auto-push option\n" +
					"Run: gopass git init"
				fmt.Println(color.RedString(msg))
				return nil
			}
			if err == ErrGitNoRemote {
				msg := "Warning: git has not remote. Ignoring auto-push option\n" +
					"Run: gopass git remote add origin ..."
				fmt.Println(color.YellowString(msg))
				return nil
			}
			return err
		}
	}
	return nil
}

// Copy will copy one entry to another location. Multi-store copies are
// supported. Each entry has to be decoded and encoded for the destination
// to make sure it's encrypted for the right set of recipients.
func (s *Store) Copy(from, to string) error {
	// recursive copy?
	if s.IsDir(from) {
		if found, err := s.Exists(to); err != nil || found {
			return fmt.Errorf("Can not copy dir to file")
		}
		sf, err := s.List("")
		if err != nil {
			return err
		}
		destPrefix := to
		if s.IsDir(to) {
			destPrefix = filepath.Join(to, filepath.Base(from))
		}
		for _, e := range sf {
			if !strings.HasPrefix(e, strings.TrimSuffix(from, "/")+"/") {
				continue
			}
			et := filepath.Join(destPrefix, strings.TrimPrefix(e, from))
			if err := s.Copy(e, et); err != nil {
				fmt.Println(err)
			}
		}
		return nil
	}
	content, err := s.Get(from)
	if err != nil {
		return err
	}
	if err := s.Set(to, content, fmt.Sprintf("Copied from %s to %s", from, to)); err != nil {
		return err
	}
	return nil
}

// Move will move one entry from one location to another. Cross-store moves are
// supported. Moving an entry will decode it from the old location, encode it
// for the destination store with the right set of recipients and remove it
// from the old location afterwards.
func (s *Store) Move(from, to string) error {
	// recursive move?
	if s.IsDir(from) {
		if found, err := s.Exists(to); err != nil || found {
			return fmt.Errorf("Can not move dir to file")
		}
		sf, err := s.List("")
		if err != nil {
			return err
		}
		destPrefix := to
		if s.IsDir(to) {
			destPrefix = filepath.Join(to, filepath.Base(from))
		}
		for _, e := range sf {
			if !strings.HasPrefix(e, strings.TrimSuffix(from, "/")+"/") {
				continue
			}
			et := filepath.Join(destPrefix, strings.TrimPrefix(e, from))
			if err := s.Move(e, et); err != nil {
				fmt.Println(err)
			}
		}
		return nil
	}

	content, err := s.Get(from)
	if err != nil {
		return err
	}
	if err := s.Set(to, content, fmt.Sprintf("Moved from %s to %s", from, to)); err != nil {
		return err
	}
	if err := s.Delete(from); err != nil {
		return err
	}
	return nil
}

// Delete will remove an single entry from the store
func (s *Store) Delete(name string) error {
	return s.delete(name, false)
}

// Prune will remove a subtree from the Store
func (s *Store) Prune(tree string) error {
	return s.delete(tree, true)
}

// delete will either delete one file or an directory tree depending on the
// RemoveFunc given. Use nil or os.Remove for the single-file mode and
// os.RemoveAll for the recursive mode.
func (s *Store) delete(name string, recurse bool) error {
	path := s.passfile(name)
	rf := os.Remove

	if !recurse && !fsutil.IsFile(path) {
		return ErrNotFound
	}

	if recurse && !fsutil.IsFile(path) {
		path = filepath.Join(s.path, name)
		rf = os.RemoveAll
		if !fsutil.IsDir(path) {
			return ErrNotFound
		}
	}

	if err := rf(path); err != nil {
		return fmt.Errorf("Failed to remove secret: %v", err)
	}

	if err := s.gitAdd(path); err != nil {
		if err == ErrGitNotInit {
			return nil
		}
		return err
	}
	if err := s.gitCommit(fmt.Sprintf("Remove %s from store.", name)); err != nil {
		if err == ErrGitNotInit {
			return nil
		}
		return err
	}

	if s.autoPush {
		if err := s.gitPush("", ""); err != nil {
			if err == ErrGitNotInit || err == ErrGitNoRemote {
				return nil
			}
			return err
		}
	}

	return nil
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
func (s *Store) reencrypt(reason string) error {
	entries, err := s.List("")
	if err != nil {
		return err
	}
	for _, e := range entries {
		content, err := s.Get(e)
		if err != nil {
			fmt.Printf("Failed to get current value for %s: %s\n", e, err)
			continue
		}
		if err := s.Set(e, content, reason); err != nil {
			fmt.Printf("Failed to write %s: %s\n", e, err)
		}
	}
	return nil
}
