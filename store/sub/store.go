package sub

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/blang/semver"
	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/config"
	"github.com/justwatchcom/gopass/fsutil"
	"github.com/justwatchcom/gopass/gpg"
	gpgcli "github.com/justwatchcom/gopass/gpg/cli"
	"github.com/justwatchcom/gopass/store"
	"github.com/pkg/errors"
)

const (
	// GPGID is the name of the file containing the recipient ids
	GPGID = ".gpg-id"
)

type gpger interface {
	ListPublicKeys() (gpg.KeyList, error)
	FindPublicKeys(...string) (gpg.KeyList, error)
	ListPrivateKeys() (gpg.KeyList, error)
	FindPrivateKeys(...string) (gpg.KeyList, error)
	GetRecipients(string) ([]string, error)
	Encrypt(string, []byte, []string) error
	Decrypt(string) ([]byte, error)
	ExportPublicKey(string, string) error
	ImportPublicKey(string) error
	Version() semver.Version
}

// Store is password store
type Store struct {
	alias           string
	autoImport      bool
	autoSync        bool
	checkRecipients bool
	debug           bool
	fsckFunc        store.FsckCallback
	importFunc      store.ImportCallback
	path            string
	recipients      []string
	gpg             gpger
}

// New creates a new store, copying settings from the given root store
func New(alias string, cfg *config.Config) (*Store, error) {
	if cfg == nil {
		cfg = &config.Config{}
	}
	if cfg.Path == "" {
		return nil, errors.Errorf("Need path")
	}
	s := &Store{
		alias:           alias,
		autoImport:      cfg.AutoImport,
		autoSync:        cfg.AutoSync,
		checkRecipients: false,
		debug:           cfg.Debug,
		fsckFunc:        cfg.FsckFunc,
		importFunc:      cfg.ImportFunc,
		path:            cfg.Path,
		recipients:      make([]string, 0, 1),
		gpg: gpgcli.New(gpgcli.Config{
			Debug:       cfg.Debug,
			AlwaysTrust: true,
		}),
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
func (s *Store) Init(path string, ids ...string) error {
	if s.Initialized() {
		return errors.Errorf(`Found already initialized store at %s.
You can add secondary stores with gopass init --path <path to secondary store> --store <mount name>`, path)
	}

	// initialize recipient list
	s.recipients = make([]string, 0, len(ids))

	for _, id := range ids {
		if id == "" {
			continue
		}
		kl, err := s.gpg.FindPublicKeys(id)
		if err != nil || len(kl) < 1 {
			fmt.Println("Failed to fetch public key:", id)
			continue
		}
		s.recipients = append(s.recipients, kl[0].Fingerprint)
	}

	if len(s.recipients) < 1 {
		return errors.Errorf("failed to initialize store: no valid recipients given")
	}

	kl, err := s.gpg.FindPrivateKeys(s.recipients...)
	if err != nil {
		return errors.Errorf("Failed to get available private keys: %s", err)
	}

	if len(kl) < 1 {
		return errors.Errorf("None of the recipients has a secret key. You will not be able to decrypt the secrets you add")
	}

	if err := s.saveRecipients("Initialized Store for " + strings.Join(s.recipients, ", ")); err != nil {
		return errors.Errorf("failed to initialize store: %v", err)
	}

	return nil
}

// idFile returns the path to the recipient list for this store
func (s *Store) idFile() string {
	return fsutil.CleanPath(filepath.Join(s.path, GPGID))
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
		if path == filepath.Join(folder, GPGID) {
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
		lst = append(lst, in...)
	}

	err := filepath.Walk(s.path, mkStoreWalkerFunc(prefix, s.path, addFunc))
	return lst, err
}

// Equals returns true if this store has the same on-disk path as the other
func (s *Store) Equals(other *Store) bool {
	if other == nil {
		return false
	}
	return s.path == other.path
}

// Get returns the plaintext of a single key
func (s *Store) Get(name string) ([]byte, error) {
	p := s.passfile(name)

	if !strings.HasPrefix(p, s.path) {
		return []byte{}, store.ErrSneaky
	}

	if !fsutil.IsFile(p) {
		if s.debug {
			fmt.Printf("File %s not found\n", p)
		}
		return []byte{}, store.ErrNotFound
	}

	content, err := s.gpg.Decrypt(p)
	if err != nil {
		return []byte{}, store.ErrDecrypt
	}

	return content, nil
}

// GetFirstLine returns the first line of the plaintext of a single key
func (s *Store) GetFirstLine(name string) ([]byte, error) {
	content, err := s.Get(name)
	if err != nil {
		return nil, err
	}

	lines := bytes.Split(content, []byte("\n"))
	if len(lines) < 1 {
		return nil, store.ErrNoPassword
	}

	return bytes.TrimSpace(lines[0]), nil
}

// GetBody returns everything but the first line
func (s *Store) GetBody(name string) ([]byte, error) {
	content, err := s.Get(name)
	if err != nil {
		return nil, err
	}

	lines := bytes.SplitN(content, []byte("\n"), 2)
	if len(lines) < 2 || len(bytes.TrimSpace(lines[1])) < 1 {
		return nil, store.ErrNoBody
	}
	return lines[1], nil
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

// Set encodes and write the ciphertext of one entry to disk
func (s *Store) Set(name string, content []byte, reason string) error {
	return s.SetConfirm(name, content, reason, nil)
}

// SetPassword update a password in an already existing entry on the disk
func (s *Store) SetPassword(name string, password []byte) error {
	var err error
	body, err := s.GetBody(name)
	if err != nil && err != store.ErrNoBody {
		return errors.Wrapf(err, "failed to get existing secret")
	}
	first := append(password, '\n')
	return s.SetConfirm(name, append(first, body...), fmt.Sprintf("Updated password in %s", name), nil)
}

func (s *Store) useableKeys() ([]string, error) {
	recipients := make([]string, len(s.recipients))
	copy(recipients, s.recipients)

	if !s.checkRecipients {
		return recipients, nil
	}

	kl, err := s.gpg.FindPublicKeys(recipients...)
	if err != nil {
		return recipients, err
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

// SetConfirm encodes and writes the cipertext of one entry to disk. This
// method can be passed a callback to confirm the recipients immedeately
// before encryption.
func (s *Store) SetConfirm(name string, content []byte, reason string, cb store.RecipientCallback) error {
	p := s.passfile(name)

	if !strings.HasPrefix(p, s.path) {
		return store.ErrSneaky
	}

	if s.IsDir(name) {
		return errors.Errorf("a folder named %s already exists", name)
	}

	recipients, err := s.useableKeys()
	if err != nil {
		return errors.Wrapf(err, "failed to list useable keys")
	}

	// confirm recipients
	if cb != nil {
		newRecipients, err := cb(name, recipients)
		if err != nil {
			return errors.Wrapf(err, "user aborted")
		}
		recipients = newRecipients
	}

	if err := s.gpg.Encrypt(p, content, recipients); err != nil {
		return store.ErrEncrypt
	}

	if err := s.gitAdd(p); err != nil {
		if err == store.ErrGitNotInit {
			return nil
		}
		return errors.Wrapf(err, "failed to add '%s' to git", p)
	}

	if err := s.gitCommit(fmt.Sprintf("Save secret to %s: %s", name, reason)); err != nil {
		if err == store.ErrGitNotInit {
			return nil
		}
		return errors.Wrapf(err, "failed to commit changes to git")
	}

	if !s.autoSync {
		return nil
	}

	if err := s.gitPush("", ""); err != nil {
		if err == store.ErrGitNotInit {
			msg := "Warning: git is not initialized for this store. Ignoring auto-push option\n" +
				"Run: gopass git init"
			fmt.Println(color.RedString(msg))
			return nil
		}
		if err == store.ErrGitNoRemote {
			msg := "Warning: git has not remote. Ignoring auto-push option\n" +
				"Run: gopass git remote add origin ..."
			fmt.Println(color.YellowString(msg))
			return nil
		}
		return errors.Wrapf(err, "failed to push to git remote")
	}
	fmt.Println(color.GreenString("Pushed changes to git remote"))
	return nil
}

// Copy will copy one entry to another location. Multi-store copies are
// supported. Each entry has to be decoded and encoded for the destination
// to make sure it's encrypted for the right set of recipients.
func (s *Store) Copy(from, to string) error {
	// recursive copy?
	if s.IsDir(from) {
		if s.Exists(to) {
			return errors.Errorf("Can not copy dir to file")
		}
		sf, err := s.List("")
		if err != nil {
			return errors.Wrapf(err, "failed to list store")
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
		return errors.Wrapf(err, "failed to get '%s' from store", from)
	}
	if err := s.Set(to, content, fmt.Sprintf("Copied from %s to %s", from, to)); err != nil {
		return errors.Wrapf(err, "failed to save '%s' to store", to)
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
		if s.Exists(to) {
			return errors.Errorf("Can not move dir to file")
		}
		sf, err := s.List("")
		if err != nil {
			return errors.Wrapf(err, "failed to list store")
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
		return errors.Wrapf(err, "failed to decrypt '%s'", from)
	}
	if err := s.Set(to, content, fmt.Sprintf("Moved from %s to %s", from, to)); err != nil {
		return errors.Wrapf(err, "failed to write '%s'", to)
	}
	if err := s.Delete(from); err != nil {
		return errors.Wrapf(err, "failed to delete '%s'", from)
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
		return store.ErrNotFound
	}

	if recurse && !fsutil.IsFile(path) {
		path = filepath.Join(s.path, name)
		rf = os.RemoveAll
		if !fsutil.IsDir(path) {
			return store.ErrNotFound
		}
	}

	if err := rf(path); err != nil {
		return errors.Errorf("Failed to remove secret: %v", err)
	}

	if err := s.gitAdd(path); err != nil {
		if err == store.ErrGitNotInit {
			return nil
		}
		return errors.Wrapf(err, "failed to add '%s' to git", path)
	}
	if err := s.gitCommit(fmt.Sprintf("Remove %s from store.", name)); err != nil {
		if err == store.ErrGitNotInit {
			return nil
		}
		return errors.Wrapf(err, "failed to commit changes to git")
	}

	if s.autoSync {
		if err := s.gitPush("", ""); err != nil {
			if err == store.ErrGitNotInit || err == store.ErrGitNoRemote {
				return nil
			}
			return errors.Wrapf(err, "failed to push change to git remote")
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
		return errors.Wrapf(err, "failed to list store")
	}

	// save original value of auto push
	gitAutoSync := s.autoSync
	s.autoSync = false
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

	// restore value of auto push
	s.autoSync = gitAutoSync

	if s.autoSync {
		if err := s.gitPush("", ""); err != nil {
			if err == store.ErrGitNotInit {
				msg := "Warning: git is not initialized for this store. Ignoring auto-push option\n" +
					"Run: gopass git init"
				fmt.Println(color.RedString(msg))
				return nil
			}
			if err == store.ErrGitNoRemote {
				msg := "Warning: git has not remote. Ignoring auto-push option\n" +
					"Run: gopass git remote add origin ..."
				fmt.Println(color.YellowString(msg))
				return nil
			}
			return errors.Wrapf(err, "failed to push change to git remote")
		}
	}
	return nil
}
