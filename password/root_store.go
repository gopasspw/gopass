package password

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/fsutil"
	"github.com/justwatchcom/gopass/gpg"
	"github.com/justwatchcom/gopass/tree"
)

// RootStore is the public facing password store
type RootStore struct {
	AutoPush        bool              `json:"autopush"`    // push to git remote after commit
	AutoPull        bool              `json:"autopull"`    // pull from git before push
	AutoImport      bool              `json:"autoimport"`  // import missing public keys w/o asking
	AlwaysTrust     bool              `json:"alwaystrust"` // always trust public keys when encrypting
	NoConfirm       bool              `json:"noconfirm"`   // do not confirm recipients when encrypting
	PersistKeys     bool              `json:"persistkeys"` // store recipient keys in store
	LoadKeys        bool              `json:"loadkeys"`    // load missing keys from store
	ClipTimeout     int               `json:"cliptimeout"` // clear clipboard after seconds
	NoColor         bool              `json:"nocolor"`     // disable colors in output
	Path            string            `json:"path"`        // path to the root store
	ShowSafeContent bool              `json:"safecontent"` // avoid showing passwords in terminal
	Mount           map[string]string `json:"mounts,omitempty"`
	Version         string            `json:"version"`
	ImportFunc      ImportCallback    `json:"-"`
	FsckFunc        FsckCallback      `json:"-"`
	Debug           bool              `json:"-"`
	store           *Store
	mounts          map[string]*Store
}

// NewRootStore creates a new store
func NewRootStore(path string) (*RootStore, error) {
	s := &RootStore{
		Path:   path,
		Mount:  make(map[string]string),
		mounts: make(map[string]*Store),
	}
	if err := s.init(); err != nil {
		return nil, err
	}
	return s, nil
}

// init checks internal consistency and initializes sub stores
// after unmarshaling
func (r *RootStore) init() error {
	if d := os.Getenv("GOPASS_DEBUG"); d == "true" {
		r.Debug = true
	}

	if r.Mount == nil {
		r.Mount = make(map[string]string)
	}
	if r.mounts == nil {
		r.mounts = make(map[string]*Store, len(r.Mount))
	}
	if r.Path == "" {
		return fmt.Errorf("Path must not be empty")
	}
	if r.AutoImport {
		r.ImportFunc = nil
	}

	// create the base store
	s, err := NewStore("", fsutil.CleanPath(r.Path), r)
	if err != nil {
		return err
	}

	r.store = s

	// initialize all mounts
	for alias, path := range r.Mount {
		path = fsutil.CleanPath(path)
		if err := r.addMount(alias, path); err != nil {
			fmt.Printf("Failed to initialized mount %s (%s): %s. Ignoring\n", alias, path, err)
			continue
		}
		r.Mount[alias] = path
	}

	// check for duplicate mounts
	if err := r.checkMounts(); err != nil {
		return fmt.Errorf("checking mounts failed: %s", err)
	}

	// set some defaults
	if r.ClipTimeout < 1 {
		r.ClipTimeout = 45
	}

	return nil
}

// Used to avoid recursion in UnmarshalJSON below
// http://attilaolah.eu/2013/11/29/json-decoding-in-go/
type rootStore RootStore

// UnmarshalJSON implements a custom JSON unmarshaler
// that will also make sure the store is properly initialized
// after loading
func (r *RootStore) UnmarshalJSON(b []byte) error {
	s := rootStore{}
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	*r = RootStore(s)
	return r.init()
}

// Initialized checks on disk if .gpg-id was generated and thus returns true.
func (r *RootStore) Initialized() bool {
	return r.store.Initialized()
}

// Init tries to initalize a new password store location matching the object
func (r *RootStore) Init(alias, path string, ids ...string) error {
	sub, err := NewStore(alias, fsutil.CleanPath(path), r)
	if err != nil {
		return err
	}

	sub.persistKeys = r.PersistKeys
	sub.loadKeys = r.LoadKeys
	sub.alwaysTrust = r.AlwaysTrust
	return sub.Init(ids...)
}

// AddMount adds a new mount
func (r *RootStore) AddMount(alias, path string, keys ...string) error {
	path = fsutil.CleanPath(path)
	if r.Mount == nil {
		r.Mount = make(map[string]string, 1)
	}
	if _, found := r.Mount[alias]; found {
		return fmt.Errorf("%s is already mounted", alias)
	}
	if err := r.addMount(alias, path, keys...); err != nil {
		return err
	}
	r.Mount[alias] = path

	// check for duplicate mounts
	return r.checkMounts()
}

func (r *RootStore) addMount(alias, path string, keys ...string) error {
	if alias == "" {
		return fmt.Errorf("alias must not be empty")
	}
	if r.mounts == nil {
		r.mounts = make(map[string]*Store, 1)
	}
	if _, found := r.mounts[alias]; found {
		return fmt.Errorf("%s is already mounted", alias)
	}

	s, err := NewStore(alias, fsutil.CleanPath(path), r)
	if err != nil {
		return err
	}

	if !s.Initialized() {
		if len(keys) < 1 {
			return fmt.Errorf("password store %s is not initialized. Try gopass init --alias %s --store %s", alias, alias, path)
		}
		if err := s.Init(keys...); err != nil {
			return err
		}
		fmt.Printf("Password store %s initialized for:", path)
		for _, r := range s.recipients {
			color.Yellow(r)
		}
	}

	r.mounts[alias] = s
	return nil
}

// RemoveMount removes and existing mount
func (r *RootStore) RemoveMount(alias string) error {
	if r.Mount == nil {
		r.Mount = make(map[string]string)
	}
	if _, found := r.Mount[alias]; !found {
		return fmt.Errorf("%s is not mounted", alias)
	}
	if _, found := r.mounts[alias]; !found {
		fmt.Println(color.YellowString("%s is not initialized", alias))
	}
	delete(r.Mount, alias)
	delete(r.mounts, alias)
	return nil
}

// mountPoints returns a sorted list of mount points. It encodes the logic that
// the longer a mount point the more specific it is. This allows to "shadow" a
// shorter mount point by a longer one.
func (r *RootStore) mountPoints() []string {
	mps := make([]string, 0, len(r.mounts))
	for k := range r.mounts {
		mps = append(mps, k)
	}
	sort.Sort(byLen(mps))
	return mps
}

// mountPoint returns the most-specific mount point for the given key
func (r *RootStore) mountPoint(name string) string {
	for _, mp := range r.mountPoints() {
		if strings.HasPrefix(name, mp) {
			return mp
		}
	}
	return ""
}

// getStore returns the Store object at the most-specific mount point for the
// given key
func (r *RootStore) getStore(name string) *Store {
	name = strings.TrimSuffix(name, "/")
	mp := r.mountPoint(name)
	if sub, found := r.mounts[mp]; found {
		return sub
	}
	return r.store
}

// checkMounts performs some sanity checks on our mounts. At the moment it
// only checks if some path is mounted twice.
func (r *RootStore) checkMounts() error {
	paths := make(map[string]string, len(r.mounts))
	for k, v := range r.mounts {
		if _, found := paths[v.path]; found {
			return fmt.Errorf("Doubly mounted path at %s: %s", v.path, k)
		}
		paths[v.path] = k
	}
	return nil
}

// Format will pretty print all entries in this store and all substores
func (r *RootStore) Format(maxDepth int) (string, error) {
	t, err := r.Tree()
	if err != nil {
		return "", err
	}
	return t.Format(maxDepth), nil
}

// List will return a flattened list of all tree entries
func (r *RootStore) List(maxDepth int) ([]string, error) {
	t, err := r.Tree()
	if err != nil {
		return []string{}, err
	}
	return t.List(maxDepth), nil
}

// Tree returns the tree representation of the entries
func (r *RootStore) Tree() (*tree.Folder, error) {
	root := tree.New("gopass")
	addFileFunc := func(in ...string) {
		for _, f := range in {
			ct := "text/plain"
			if strings.HasSuffix(f, ".yaml") {
				ct = "text/yaml"
				f = strings.TrimSuffix(f, ".yaml")
			} else if strings.HasSuffix(f, ".b64") {
				ct = "application/octet-stream"
				f = strings.TrimSuffix(f, ".b64")
			}
			if err := root.AddFile(f, ct); err != nil {
				fmt.Printf("Failed to add file %s to tree: %s\n", f, err)
				continue
			}
		}
	}
	addTplFunc := func(in ...string) {
		for _, f := range in {
			if err := root.AddTemplate(f); err != nil {
				fmt.Printf("Failed to add template %s to tree: %s\n", f, err)
				continue
			}
		}
	}
	mps := r.mountPoints()
	sort.Sort(sort.Reverse(byLen(mps)))
	for _, alias := range mps {
		substore := r.mounts[alias]
		if substore == nil {
			continue
		}
		if err := root.AddMount(alias, substore.path); err != nil {
			return nil, fmt.Errorf("failed to add mount: %s", err)
		}
		sf, err := substore.List(alias)
		if err != nil {
			return nil, fmt.Errorf("failed to add file: %s", err)
		}
		addFileFunc(sf...)
		addTplFunc(substore.ListTemplates(alias)...)
	}

	sf, err := r.store.List("")
	if err != nil {
		return nil, err
	}
	addFileFunc(sf...)
	addTplFunc(r.store.ListTemplates("")...)

	return root, nil
}

// Get returns the plaintext of a single key
func (r *RootStore) Get(name string) ([]byte, error) {
	// forward to substore
	store := r.getStore(name)
	return store.Get(strings.TrimPrefix(name, store.alias))
}

// GetKey will return a single named entry from a structured document (YAML)
// in secret name. If no such key exists or yaml decoding fails it will
// return an error
func (r *RootStore) GetKey(name, key string) ([]byte, error) {
	store := r.getStore(name)
	return store.GetKey(strings.TrimPrefix(name, store.alias), key)
}

// First returns the first line of the plaintext of a single key
func (r *RootStore) First(name string) ([]byte, error) {
	store := r.getStore(name)
	return store.First(strings.TrimPrefix(name, store.alias))
}

// SafeContent returns everything but the first line from a key
func (r *RootStore) SafeContent(name string) ([]byte, error) {
	store := r.getStore(name)
	return store.SafeContent(strings.TrimPrefix(name, store.alias))
}

// Exists checks the existence of a single entry
func (r *RootStore) Exists(name string) (bool, error) {
	store := r.getStore(name)
	return store.Exists(strings.TrimPrefix(name, store.alias))
}

// IsDir checks if a given key is actually a folder
func (r *RootStore) IsDir(name string) bool {
	store := r.getStore(name)
	return store.IsDir(strings.TrimPrefix(name, store.alias))
}

// Set encodes and write the ciphertext of one entry to disk
func (r *RootStore) Set(name string, content []byte, reason string) error {
	store := r.getStore(name)
	return store.Set(strings.TrimPrefix(name, store.alias), content, reason)
}

// SetKey sets a single key in structured document (YAML) to the given
// value. If the secret name is non-empty but no YAML it will return an error.
func (r *RootStore) SetKey(name, key, value string) error {
	store := r.getStore(name)
	return store.SetKey(strings.TrimPrefix(name, store.alias), key, value)
}

// SetConfirm calls Set with confirmation callback
func (r *RootStore) SetConfirm(name string, content []byte, reason string, cb RecipientCallback) error {
	store := r.getStore(name)
	return store.SetConfirm(strings.TrimPrefix(name, store.alias), content, reason, cb)
}

// Copy will copy one entry to another location. Multi-store copies are
// supported. Each entry has to be decoded and encoded for the destination
// to make sure it's encrypted for the right set of recipients.
func (r *RootStore) Copy(from, to string) error {
	subFrom := r.getStore(from)
	subTo := r.getStore(to)

	from = strings.TrimPrefix(from, subFrom.alias)
	to = strings.TrimPrefix(to, subFrom.alias)

	// cross-store copy
	if !subFrom.equals(subTo) {
		content, err := subFrom.Get(from)
		if err != nil {
			return err
		}
		if err := subTo.Set(to, content, fmt.Sprintf("Copied from %s to %s", from, to)); err != nil {
			return err
		}
		return nil
	}

	return subFrom.Copy(from, to)
}

// Move will move one entry from one location to another. Cross-store moves are
// supported. Moving an entry will decode it from the old location, encode it
// for the destination store with the right set of recipients and remove it
// from the old location afterwards.
func (r *RootStore) Move(from, to string) error {
	subFrom := r.getStore(from)
	subTo := r.getStore(to)

	from = strings.TrimPrefix(from, subFrom.alias)

	// cross-store move
	if !subFrom.equals(subTo) {
		to = strings.TrimPrefix(to, subTo.alias)
		content, err := subFrom.Get(from)
		if err != nil {
			return fmt.Errorf("Source %s does not exist in source store %s: %s", from, subFrom.alias, err)
		}
		if err := subTo.Set(to, content, fmt.Sprintf("Moved from %s to %s", from, to)); err != nil {
			return err
		}
		if err := subFrom.Delete(from); err != nil {
			return err
		}
		return nil
	}

	to = strings.TrimPrefix(to, subFrom.alias)
	return subFrom.Move(from, to)
}

// Delete will remove an single entry from the store
func (r *RootStore) Delete(name string) error {
	store := r.getStore(name)
	sn := strings.TrimPrefix(name, store.alias)
	if sn == "" {
		return fmt.Errorf("can not delete a mount point. Use `gopass mount remove %s`", store.alias)
	}
	return store.Delete(sn)
}

// Prune will remove a subtree from the Store
func (r *RootStore) Prune(tree string) error {
	for mp := range r.mounts {
		if strings.HasPrefix(mp, tree) {
			return fmt.Errorf("can not prune subtree with mounts. Unmount first: `gopass mount remove %s`", mp)
		}
	}

	store := r.getStore(tree)
	return store.Prune(strings.TrimPrefix(tree, store.alias))
}

func (r *RootStore) String() string {
	ms := make([]string, 0, len(r.mounts))
	for alias, sub := range r.mounts {
		ms = append(ms, alias+"="+sub.String())
	}
	return fmt.Sprintf("RootStore(Path: %s, Mounts: %+v)", r.store.path, strings.Join(ms, ","))
}

// GitInit initializes the git repo
func (r *RootStore) GitInit(store, sk string) error {
	return r.getStore(store).GitInit(sk)
}

// Git runs arbitrary git commands on this store and all substores
func (r *RootStore) Git(store string, args ...string) error {
	return r.getStore(store).Git(args...)
}

// Fsck checks the stores integrity
func (r *RootStore) Fsck(check, force bool) error {
	sh := make(map[string]string, 100)
	for _, alias := range r.mountPoints() {
		// check sub-store integrity
		counts, err := r.mounts[alias].Fsck(alias, check, force)
		if err != nil {
			return err
		}
		fmt.Println(color.GreenString("[%s] Store (%s) checked (%d OK, %d warnings, %d errors)", alias, r.Mount[alias], counts["ok"], counts["warn"], counts["err"]))
		// check shadowing
		lst, err := r.mounts[alias].List(alias)
		if err != nil {
			return err
		}
		for _, e := range lst {
			if a, found := sh[e]; found {
				fmt.Println(color.YellowString("Entry %s is being shadowed by %s", e, a))
			}
			sh[e] = alias
		}
	}

	counts, err := r.store.Fsck("root", check, force)
	if err != nil {
		return err
	}
	fmt.Println(color.GreenString("[%s] Store checked (%d OK, %d warnings, %d errors)", r.store.path, counts["ok"], counts["warn"], counts["err"]))
	// check shadowing
	lst, err := r.store.List("")
	if err != nil {
		return err
	}
	for _, e := range lst {
		if a, found := sh[e]; found {
			fmt.Println(color.YellowString("Entry %s is being shadowed by %s", e, a))
		}
		sh[e] = ""
	}
	return nil
}

// ListRecipients lists all recipients for the given store
func (r *RootStore) ListRecipients(store string) []string {
	return r.getStore(store).recipients
}

// AddRecipient adds a single recipient to the given store
func (r *RootStore) AddRecipient(store, rec string) error {
	return r.getStore(store).AddRecipient(rec)
}

// RemoveRecipient removes a single recipient from the given store
func (r *RootStore) RemoveRecipient(store, rec string) error {
	return r.getStore(store).RemoveRecipient(rec)
}

// RecipientsTree returns a tree view of all stores' recipients
func (r *RootStore) RecipientsTree(pretty bool) (*tree.Folder, error) {
	root := tree.New("gopass")
	mps := r.mountPoints()
	sort.Sort(sort.Reverse(byLen(mps)))
	for _, alias := range mps {
		substore := r.mounts[alias]
		if substore == nil {
			continue
		}
		if err := root.AddMount(alias, substore.path); err != nil {
			return nil, fmt.Errorf("failed to add mount: %s", err)
		}
		for _, r := range substore.recipients {
			key := fmt.Sprintf("%s (missing public key)", r)
			kl, err := gpg.ListPublicKeys(r)
			if err == nil {
				if len(kl) > 0 {
					if pretty {
						key = kl[0].OneLine()
					} else {
						key = kl[0].Fingerprint
					}
				}
			}
			if err := root.AddFile(alias+"/"+key, "gopass/recipient"); err != nil {
				fmt.Println(err)
			}
		}
	}

	for _, r := range r.store.recipients {
		kl, err := gpg.ListPublicKeys(r)
		if err != nil {
			fmt.Println(err)
			continue
		}
		if len(kl) < 1 {
			fmt.Println("key not found", r)
			continue
		}
		key := kl[0].Fingerprint
		if pretty {
			key = kl[0].OneLine()
		}
		if err := root.AddFile(key, "gopass/recipient"); err != nil {
			fmt.Println(err)
		}
	}
	return root, nil
}
