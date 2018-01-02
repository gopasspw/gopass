package root

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"path"

	gpgmock "github.com/justwatchcom/gopass/backend/gpg/mock"
	"github.com/justwatchcom/gopass/config"
	"github.com/justwatchcom/gopass/store/sub"
)

func TestSimpleList(t *testing.T) {
	ctx := context.Background()

	tempdir, err := ioutil.TempDir("", "gopass-")
	if err != nil {
		t.Fatalf("Failed to create tempdir: %s", err)
	}
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()

	_, ents, err := createStore(tempdir)
	if err != nil {
		t.Fatalf("Failed to create store directory: %s", err)
	}

	rs, err := New(
		ctx,
		&config.Config{
			Root: &config.StoreConfig{
				Path: tempdir,
			},
		},
		gpgmock.New(),
	)
	if err != nil {
		t.Fatalf("Failed to create root store: %s", err)
	}

	tree, err := rs.Tree()
	if err != nil {
		t.Fatalf("failed to list tree: %s", err)
	}

	compareLists(t, ents, tree.List(0))
}

func TestListMulti(t *testing.T) {
	ctx := context.Background()

	tempdir, err := ioutil.TempDir("", "gopass-")
	if err != nil {
		t.Fatalf("Failed to create tempdir: %s", err)
	}
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()

	// root store
	_, ents, err := createStore(path.Join(tempdir, "root"))
	if err != nil {
		t.Fatalf("Failed to init root store: %s", err)
	}

	// sub1 store
	_, sub1Ent, err := createStore(filepath.Join(tempdir, "sub1"))
	if err != nil {
		t.Fatalf("Failed to init sub1 store: %s", err)
	}
	for _, k := range sub1Ent {
		ents = append(ents, path.Join("sub1", k))
	}

	// sub2 store
	_, sub2Ent, err := createStore(filepath.Join(tempdir, "sub2"))
	if err != nil {
		t.Fatalf("Failed to init sub2 store: %s", err)
	}
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
	if err != nil {
		t.Fatalf("Failed to create root store: %s", err)
	}
	if err != nil {
		t.Fatalf("failed to create root store: %s", err)
	}
	if err := rs.AddMount(ctx, "sub1", filepath.Join(tempdir, "sub1")); err != nil {
		t.Fatalf("failed to add mount %s: %s", "sub1", err)
	}
	if err := rs.AddMount(ctx, "sub2", filepath.Join(tempdir, "sub2")); err != nil {
		t.Fatalf("failed to add mount %s: %s", "sub2", err)
	}
	tree, err := rs.Tree()
	if err != nil {
		t.Fatalf("failed to list tree: %s", err)
	}
	compareLists(t, ents, tree.List(0))
}

func TestListNested(t *testing.T) {
	ctx := context.Background()

	tempdir, err := ioutil.TempDir("", "gopass-")
	if err != nil {
		t.Fatalf("Failed to create tempdir: %s", err)
	}
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()
	// root store
	_, ents, err := createStore(filepath.Join(tempdir, "root"))
	if err != nil {
		t.Fatalf("Failed to init root store: %s", err)
	}
	// sub1 store
	_, sub1Ent, err := createStore(filepath.Join(tempdir, "sub1"))
	if err != nil {
		t.Fatalf("Failed to init sub1 store: %s", err)
	}
	for _, k := range sub1Ent {
		ents = append(ents, path.Join("sub1", k))
	}
	// sub2 store
	_, sub2Ent, err := createStore(filepath.Join(tempdir, "sub2"))
	if err != nil {
		t.Fatalf("Failed to init sub2 store: %s", err)
	}
	for _, k := range sub2Ent {
		ents = append(ents, path.Join("sub2", k))
	}
	// sub3 store
	_, sub3Ent, err := createStore(filepath.Join(tempdir, "sub3"))
	if err != nil {
		t.Fatalf("Failed to init sub3 store: %s", err)
	}
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
	if err != nil {
		t.Fatalf("Failed to create root store: %s", err)
	}
	if err := rs.AddMount(ctx, "sub1", filepath.Join(tempdir, "sub1")); err != nil {
		t.Fatalf("failed to add mount %s: %s", "sub1", err)
	}
	if err := rs.AddMount(ctx, "sub2", filepath.Join(tempdir, "sub2")); err != nil {
		t.Fatalf("failed to add mount %s: %s", "sub2", err)
	}
	if err := rs.AddMount(ctx, "sub2/sub3", filepath.Join(tempdir, "sub3")); err != nil {
		t.Fatalf("failed to add mount %s: %s", "sub2", err)
	}
	tree, err := rs.Tree()
	if err != nil {
		t.Fatalf("failed to list tree: %s", err)
	}
	t.Log(tree.Format(100))
	compareLists(t, ents, tree.List(0))
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

func maxLenStr(l []string) string {
	max := 10
	for _, e := range l {
		if len(e) > max {
			max = len(e)
		}
	}
	return strconv.Itoa(max)
}

func logLists(t *testing.T, l1, l2 []string) {
	tpl := "%3d | %-" + maxLenStr(l1) + "s | %-" + maxLenStr(l2) + "s"
	t.Logf(tpl, 0, "L1", "L2")
	max := len(l1)
	if len(l2) > max {
		max = len(l2)
	}
	for i := 0; i < max; i++ {
		e1 := "MISSING"
		e2 := "MISSING"
		if len(l1) > i {
			e1 = l1[i]
		}
		if len(l2) > i {
			e2 = l2[i]
		}
		t.Logf(tpl, i, e1, e2)
	}
}

func compareLists(t *testing.T, l1, l2 []string) {
	if len(l1) != len(l2) {
		t.Errorf("len(l1)=%d != len(l2)=%d", len(l1), len(l2))
		logLists(t, l1, l2)
		return
	}
	for i := 0; i < len(l1); i++ {
		if l1[i] != l2[i] {
			t.Errorf("Mismatch at pos %d: %s - %s", i, l1[i], l2[i])
			logLists(t, l1, l2)
			return
		}
	}
}
