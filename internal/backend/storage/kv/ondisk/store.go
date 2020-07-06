// Package ondisk implements an encrypted on-disk storage backend with
// integrated revision control as well as automatic synchronization (soon).
package ondisk

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/blang/semver"
	"github.com/gopasspw/gopass/internal/backend/crypto/age"
	"github.com/gopasspw/gopass/internal/backend/storage/kv/ondisk/gpb"
	"github.com/gopasspw/gopass/internal/debug"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	idxFile    = "index.pb"
	idxBakFile = "index.pb.back"
	//lockFile   = "index.lock"
	maxRev = 256
	delTTL = time.Hour * 24 * 365
)

// OnDisk is an on disk key-value store
type OnDisk struct {
	dir string
	idx *gpb.Store
	age *age.Age
}

// New creates a new ondisk store
func New(baseDir string) (*OnDisk, error) {
	a, err := age.New()
	if err != nil {
		return nil, err
	}
	o := &OnDisk{
		dir: baseDir,
		age: a,
	}
	idx, err := o.loadOrCreate(baseDir)
	if err != nil {
		return nil, err
	}
	o.idx = idx
	return o, nil
}

func (o *OnDisk) loadOrCreate(path string) (*gpb.Store, error) {
	path = filepath.Join(path, idxFile)
	buf, err := ioutil.ReadFile(path)
	if os.IsNotExist(err) {
		return &gpb.Store{
			Name:    filepath.Base(path),
			Entries: make(map[string]*gpb.Entry),
		}, nil
	}
	buf, err = o.age.Decrypt(context.TODO(), buf)
	if err != nil {
		return nil, err
	}
	idx := &gpb.Store{}
	err = proto.Unmarshal(buf, idx)
	return idx, err
}

func (o *OnDisk) saveIndex() error {
	buf, err := proto.Marshal(o.idx)
	if err != nil {
		return err
	}
	os.Rename(filepath.Join(o.dir, idxFile), filepath.Join(o.dir, idxBakFile))
	buf, err = o.age.Encrypt(context.TODO(), buf, []string{})
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filepath.Join(o.dir, idxFile), buf, 0600)
}

// Get returns an entry
func (o *OnDisk) Get(ctx context.Context, name string) ([]byte, error) {
	e, err := o.getEntry(name)
	if err != nil {
		return nil, err
	}
	r := e.Latest()
	if r == nil {
		return nil, fmt.Errorf("not found")
	}
	path := filepath.Join(o.dir, r.GetFilename())
	debug.Log("Reading %s from %s", name, path)
	return ioutil.ReadFile(path)
}

func filename(buf []byte) string {
	sum := fmt.Sprintf("%x", sha256.Sum256(buf))
	return filepath.Join(sum[0:2], sum[2:])
}

// Set creates a new revision for an entry
func (o *OnDisk) Set(ctx context.Context, name string, value []byte) error {
	fn := filename(value)
	fp := filepath.Join(o.dir, filename(value))
	if err := os.MkdirAll(filepath.Dir(fp), 0700); err != nil {
		return err
	}
	if err := ioutil.WriteFile(fp, value, 0600); err != nil {
		return err
	}
	debug.Log("Wrote %s to %s", name, fp)
	e := o.getOrCreateEntry(name)
	msg := "Updated " + fn
	if cm := ctxutil.GetCommitMessage(ctx); cm != "" {
		msg = cm
	}
	e.Revisions = append(e.Revisions, &gpb.Revision{
		Created:  timestamppb.Now(),
		Message:  msg,
		Filename: fn,
	})
	debug.Log("Added Revision for %s: %+v", name, e)
	o.idx.Entries[name] = e
	return o.saveIndex()
}

// Exists checks if an entry exists
func (o *OnDisk) Exists(ctx context.Context, name string) bool {
	e, found := o.idx.Entries[name]
	if !found {
		return false
	}
	found = !e.IsDeleted()
	debug.Log("%s exists? %t in %+v", name, found, o.idx.Entries)
	return found
}

func (o *OnDisk) getEntry(name string) (*gpb.Entry, error) {
	em := o.idx.GetEntries()
	if em == nil {
		return nil, fmt.Errorf("%s not found (empty index)", name)
	}
	e, found := em[name]
	if !found {
		return nil, fmt.Errorf("%s not found", name)
	}
	return e, nil
}

func (o *OnDisk) getOrCreateEntry(name string) *gpb.Entry {
	if e, found := o.idx.Entries[name]; found && e != nil {
		return e
	}
	debug.Log("Created new Entry for %s", name)
	return &gpb.Entry{
		Name:      name,
		Revisions: make([]*gpb.Revision, 0, 1),
	}
}

// Delete removes an entry
func (o *OnDisk) Delete(ctx context.Context, name string) error {
	if !o.Exists(ctx, name) {
		debug.Log("Not adding tombstone for non-existing entry for %s", name)
		return nil
	}
	// add tombstone
	e := o.getOrCreateEntry(name)
	e.Delete(ctxutil.GetCommitMessage(ctx))
	o.idx.Entries[name] = e

	debug.Log("Added tombstone for %s", name)
	return o.saveIndex()
}

// List lists all entries
func (o *OnDisk) List(ctx context.Context, prefix string) ([]string, error) {
	res := make([]string, 0, len(o.idx.Entries))
	for k, v := range o.idx.Entries {
		if v.IsDeleted() {
			continue
		}
		if strings.HasPrefix(k, prefix) {
			res = append(res, k)
		}
	}
	return res, nil
}

// IsDir is not supported
func (o *OnDisk) IsDir(ctx context.Context, name string) bool {
	return false
}

// Prune removes all entries with a given prefix
func (o *OnDisk) Prune(ctx context.Context, prefix string) error {
	l, _ := o.List(ctx, name)
	for _, e := range l {
		if err := o.Delete(ctx, e); err != nil {
			return err
		}
	}
	return nil
}

// Name returns ondisk
func (o *OnDisk) Name() string {
	return name
}

// Version returns 1.0.0
func (o *OnDisk) Version(context.Context) semver.Version {
	return semver.Version{Major: 1}
}

// String returns the name and path
func (o *OnDisk) String() string {
	return fmt.Sprintf("%s(path: %s)", name, o.dir)
}

// Available always returns nil
func (o *OnDisk) Available(ctx context.Context) error {
	return nil
}

// Compact will prune all deleted entries and truncate every other entry
// to the last 10 revisions.
func (o *OnDisk) Compact(_ context.Context) error {
	for k, v := range o.idx.Entries {
		if v.IsDeleted() && time.Since(v.Latest().Time()) > delTTL {
			delete(o.idx.Entries, k)
			continue
		}
		sort.Sort(gpb.ByRevision(o.idx.Entries[k].Revisions))
		if len(o.idx.Entries[k].Revisions) > maxRev {
			o.idx.Entries[k].Revisions = o.idx.Entries[k].Revisions[0:maxRev]
		}
	}
	return o.saveIndex()
}
