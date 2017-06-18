package root

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/justwatchcom/gopass/config"
	"github.com/justwatchcom/gopass/fsutil"
	"github.com/justwatchcom/gopass/store"
	"github.com/justwatchcom/gopass/store/sub"
	"github.com/justwatchcom/gopass/tree"
	"github.com/justwatchcom/gopass/tree/simple"
)

// Store is the public facing password store
type Store struct {
	alwaysTrust bool // always trust public keys when encrypting
	askForMore  bool
	autoImport  bool // import missing public keys w/o asking
	autoPull    bool // pull from git before push
	autoPush    bool // push to git remote after commit
	clipTimeout int  // clear clipboard after seconds
	debug       bool
	fsckFunc    store.FsckCallback
	importFunc  store.ImportCallback
	loadKeys    bool // load missing keys from store
	mounts      map[string]*sub.Store
	noColor     bool   // disable colors in output
	noConfirm   bool   // do not confirm recipients when encrypting
	path        string // path to the root store
	persistKeys bool   // store recipient keys in store
	safeContent bool   // avoid showing passwords in terminal
	store       *sub.Store
	version     string
}

// New creates a new store
func New(cfg *config.Config) (*Store, error) {
	if cfg == nil {
		cfg = &config.Config{}
	}
	if cfg.Path == "" {
		return nil, fmt.Errorf("need path")
	}
	r := &Store{
		alwaysTrust: cfg.AlwaysTrust,
		askForMore:  cfg.AskForMore,
		autoImport:  cfg.AutoImport,
		autoPull:    cfg.AutoPull,
		autoPush:    cfg.AutoPush,
		clipTimeout: cfg.ClipTimeout,
		debug:       cfg.Debug,
		fsckFunc:    cfg.FsckFunc,
		importFunc:  cfg.ImportFunc,
		loadKeys:    cfg.LoadKeys,
		mounts:      make(map[string]*sub.Store, len(cfg.Mounts)),
		noColor:     cfg.NoColor,
		noConfirm:   cfg.NoConfirm,
		path:        cfg.Path,
		persistKeys: cfg.PersistKeys,
		safeContent: cfg.SafeContent,
	}

	// TODO(dschulz) this should be passed down from main, not set here
	if d := os.Getenv("GOPASS_DEBUG"); d == "true" {
		r.debug = true
	}

	if r.autoImport {
		r.importFunc = nil
	}

	// create the base store
	subCfg := r.Config()
	subCfg.Path = fsutil.CleanPath(r.Path())
	s, err := sub.New("", subCfg)
	if err != nil {
		return nil, err
	}

	r.store = s

	// initialize all mounts
	for alias, path := range cfg.Mounts {
		path = fsutil.CleanPath(path)
		if err := r.addMount(alias, path); err != nil {
			fmt.Printf("Failed to initialized mount %s (%s): %s. Ignoring\n", alias, path, err)
			continue
		}
	}

	// check for duplicate mounts
	if err := r.checkMounts(); err != nil {
		return nil, fmt.Errorf("checking mounts failed: %s", err)
	}

	// set some defaults
	if r.clipTimeout < 1 {
		r.clipTimeout = 45
	}

	return r, nil
}

// Initialized checks on disk if .gpg-id was generated and thus returns true.
func (r *Store) Initialized() bool {
	return r.store.Initialized()
}

// Init tries to initalize a new password store location matching the object
func (r *Store) Init(alias, path string, ids ...string) error {
	cfg := r.Config()
	cfg.Path = fsutil.CleanPath(path)
	sub, err := sub.New(alias, cfg)
	if err != nil {
		return err
	}
	if !r.store.Initialized() && alias == "" {
		r.store = sub
	}

	return sub.Init(path, ids...)
}

// Format will pretty print all entries in this store and all substores
func (r *Store) Format(maxDepth int) (string, error) {
	t, err := r.Tree()
	if err != nil {
		return "", err
	}
	return t.Format(maxDepth), nil
}

// List will return a flattened list of all tree entries
func (r *Store) List(maxDepth int) ([]string, error) {
	t, err := r.Tree()
	if err != nil {
		return []string{}, err
	}
	return t.List(maxDepth), nil
}

// Tree returns the tree representation of the entries
func (r *Store) Tree() (tree.Tree, error) {
	root := simple.New("gopass")
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
		if err := root.AddMount(alias, substore.Path()); err != nil {
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
func (r *Store) Get(name string) ([]byte, error) {
	// forward to substore
	store := r.getStore(name)
	return store.Get(strings.TrimPrefix(name, store.Alias()))
}

// GetFirstLine returns the first line of the plaintext of a single key
func (r *Store) GetFirstLine(name string) ([]byte, error) {
	store := r.getStore(name)
	return store.GetFirstLine(strings.TrimPrefix(name, store.Alias()))
}

// GetBody returns everything but the first line from a key
func (r *Store) GetBody(name string) ([]byte, error) {
	store := r.getStore(name)
	return store.GetBody(strings.TrimPrefix(name, store.Alias()))
}

// Exists checks the existence of a single entry
func (r *Store) Exists(name string) bool {
	store := r.getStore(name)
	return store.Exists(strings.TrimPrefix(name, store.Alias()))
}

// IsDir checks if a given key is actually a folder
func (r *Store) IsDir(name string) bool {
	store := r.getStore(name)
	return store.IsDir(strings.TrimPrefix(name, store.Alias()))
}

// Set encodes and write the ciphertext of one entry to disk
func (r *Store) Set(name string, content []byte, reason string) error {
	store := r.getStore(name)
	return store.Set(strings.TrimPrefix(name, store.Alias()), content, reason)
}

// SetConfirm calls Set with confirmation callback
func (r *Store) SetConfirm(name string, content []byte, reason string, cb store.RecipientCallback) error {
	store := r.getStore(name)
	return store.SetConfirm(strings.TrimPrefix(name, store.Alias()), content, reason, cb)
}

// Copy will copy one entry to another location. Multi-store copies are
// supported. Each entry has to be decoded and encoded for the destination
// to make sure it's encrypted for the right set of recipients.
func (r *Store) Copy(from, to string) error {
	subFrom := r.getStore(from)
	subTo := r.getStore(to)

	from = strings.TrimPrefix(from, subFrom.Alias())
	to = strings.TrimPrefix(to, subFrom.Alias())

	// cross-store copy
	if !subFrom.Equals(subTo) {
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
func (r *Store) Move(from, to string) error {
	subFrom := r.getStore(from)
	subTo := r.getStore(to)

	from = strings.TrimPrefix(from, subFrom.Alias())

	// cross-store move
	if !subFrom.Equals(subTo) {
		to = strings.TrimPrefix(to, subTo.Alias())
		content, err := subFrom.Get(from)
		if err != nil {
			return fmt.Errorf("Source %s does not exist in source store %s: %s", from, subFrom.Alias(), err)
		}
		if err := subTo.Set(to, content, fmt.Sprintf("Moved from %s to %s", from, to)); err != nil {
			return err
		}
		if err := subFrom.Delete(from); err != nil {
			return err
		}
		return nil
	}

	to = strings.TrimPrefix(to, subFrom.Alias())
	return subFrom.Move(from, to)
}

// Delete will remove an single entry from the store
func (r *Store) Delete(name string) error {
	store := r.getStore(name)
	sn := strings.TrimPrefix(name, store.Alias())
	if sn == "" {
		return fmt.Errorf("can not delete a mount point. Use `gopass mount remove %s`", store.Alias())
	}
	return store.Delete(sn)
}

// Prune will remove a subtree from the Store
func (r *Store) Prune(tree string) error {
	for mp := range r.mounts {
		if strings.HasPrefix(mp, tree) {
			return fmt.Errorf("can not prune subtree with mounts. Unmount first: `gopass mount remove %s`", mp)
		}
	}

	store := r.getStore(tree)
	return store.Prune(strings.TrimPrefix(tree, store.Alias()))
}

func (r *Store) String() string {
	ms := make([]string, 0, len(r.mounts))
	for alias, sub := range r.mounts {
		ms = append(ms, alias+"="+sub.String())
	}
	return fmt.Sprintf("Store(Path: %s, Mounts: %+v)", r.store.Path(), strings.Join(ms, ","))
}
