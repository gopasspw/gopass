package root

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"path"

	gpgmock "github.com/justwatchcom/gopass/backend/gpg/mock"
	"github.com/justwatchcom/gopass/config"
	"github.com/justwatchcom/gopass/store/sub"
	"github.com/stretchr/testify/assert"
)

func TestSimpleList(t *testing.T) {
	ctx := context.Background()

	tempdir, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()

	_, ents, err := createStore(tempdir)
	assert.NoError(t, err)

	rs, err := New(
		ctx,
		&config.Config{
			Root: &config.StoreConfig{
				Path: tempdir,
			},
		},
		gpgmock.New(),
	)
	assert.NoError(t, err)

	tree, err := rs.Tree(ctx)
	assert.NoError(t, err)
	assert.Equal(t, ents, tree.List(0))
}

func TestListMulti(t *testing.T) {
	ctx := context.Background()

	tempdir, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()

	// root store
	_, ents, err := createStore(path.Join(tempdir, "root"))
	assert.NoError(t, err)

	// sub1 store
	_, sub1Ent, err := createStore(filepath.Join(tempdir, "sub1"))
	assert.NoError(t, err)
	for _, k := range sub1Ent {
		ents = append(ents, path.Join("sub1", k))
	}

	// sub2 store
	_, sub2Ent, err := createStore(filepath.Join(tempdir, "sub2"))
	assert.NoError(t, err)
	for _, k := range sub2Ent {
		ents = append(ents, path.Join("sub2", k))
	}
	sort.Strings(ents)

	rs, err := New(
		ctx,
		&config.Config{
			Root: &config.StoreConfig{
				Path: filepath.Join(tempdir, "root"),
			},
		},
		gpgmock.New(),
	)
	assert.NoError(t, err)
	assert.NoError(t, rs.AddMount(ctx, "sub1", filepath.Join(tempdir, "sub1")))
	assert.NoError(t, rs.AddMount(ctx, "sub2", filepath.Join(tempdir, "sub2")))

	tree, err := rs.Tree(ctx)
	assert.NoError(t, err)
	assert.Equal(t, ents, tree.List(0))
}

func TestListNested(t *testing.T) {
	ctx := context.Background()

	tempdir, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()

	// root store
	_, ents, err := createStore(filepath.Join(tempdir, "root"))
	assert.NoError(t, err)

	// sub1 store
	_, sub1Ent, err := createStore(filepath.Join(tempdir, "sub1"))
	assert.NoError(t, err)
	for _, k := range sub1Ent {
		ents = append(ents, path.Join("sub1", k))
	}

	// sub2 store
	_, sub2Ent, err := createStore(filepath.Join(tempdir, "sub2"))
	assert.NoError(t, err)
	for _, k := range sub2Ent {
		ents = append(ents, path.Join("sub2", k))
	}

	// sub3 store
	_, sub3Ent, err := createStore(filepath.Join(tempdir, "sub3"))
	assert.NoError(t, err)
	for _, k := range sub3Ent {
		ents = append(ents, path.Join("sub2", "sub3", k))
	}

	sort.Strings(ents)

	rs, err := New(
		ctx,
		&config.Config{
			Root: &config.StoreConfig{
				Path: filepath.Join(tempdir, "root"),
			},
		},
		gpgmock.New(),
	)
	assert.NoError(t, err)
	if err != nil {
		t.Fatalf("Failed to create root store: %s", err)
	}
	assert.NoError(t, rs.AddMount(ctx, "sub1", filepath.Join(tempdir, "sub1")))
	assert.NoError(t, rs.AddMount(ctx, "sub2", filepath.Join(tempdir, "sub2")))
	assert.NoError(t, rs.AddMount(ctx, "sub2/sub3", filepath.Join(tempdir, "sub3")))

	tree, err := rs.Tree(ctx)
	assert.NoError(t, err)
	assert.Equal(t, ents, tree.List(0))
}

func allPathsToSlash(paths []string) []string {
	r := make([]string, len(paths))
	for i, p := range paths {
		r[i] = filepath.ToSlash(p)
	}
	return r
}

func createRootStore(ctx context.Context, dir string) (*Store, error) {
	sd := filepath.Join(dir, "root")
	_, _, err := createStore(sd)
	if err != nil {
		return nil, err
	}

	if err := os.Setenv("GOPASS_CONFIG", filepath.Join(dir, ".gopass.yml")); err != nil {
		return nil, err
	}
	if err := os.Setenv("GOPASS_HOMEDIR", dir); err != nil {
		return nil, err
	}
	if err := os.Unsetenv("PAGER"); err != nil {
		return nil, err
	}
	if err := os.Setenv("CHECKPOINT_DISABLE", "true"); err != nil {
		return nil, err
	}
	if err := os.Setenv("GOPASS_NO_NOTIFY", "true"); err != nil {
		return nil, err
	}
	gpgDir := filepath.Join(dir, ".gnupg")
	if err := os.Setenv("GNUPGHOME", gpgDir); err != nil {
		return nil, err
	}

	return New(
		ctx,
		&config.Config{
			Root: &config.StoreConfig{
				Path: sd,
			},
		},
		gpgmock.New(),
	)
}

func createStore(dir string) ([]string, []string, error) {
	recipients := []string{
		"0xDEADBEEF",
		"0xFEEDBEEF",
	}
	list := []string{
		filepath.Join("foo", "bar", "baz"),
		filepath.Join("baz", "ing", "a"),
	}
	sort.Strings(list)
	for _, file := range list {
		filename := filepath.Join(dir, file+".gpg")
		if err := os.MkdirAll(filepath.Dir(filename), 0700); err != nil {
			return recipients, allPathsToSlash(list), err
		}
		if err := ioutil.WriteFile(filename, []byte{}, 0644); err != nil {
			return recipients, allPathsToSlash(list), err
		}
	}
	err := ioutil.WriteFile(filepath.Join(dir, sub.GPGID), []byte(strings.Join(recipients, "\n")), 0600)
	return recipients, allPathsToSlash(list), err
}
