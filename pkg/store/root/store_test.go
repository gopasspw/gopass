package root

import (
	"context"
	"path"
	"sort"
	"testing"

	"github.com/gopasspw/gopass/pkg/backend"
	"github.com/gopasspw/gopass/pkg/config"
	"github.com/gopasspw/gopass/tests/gptest"

	_ "github.com/gopasspw/gopass/pkg/backend/crypto"
	_ "github.com/gopasspw/gopass/pkg/backend/rcs"
	_ "github.com/gopasspw/gopass/pkg/backend/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSimpleList(t *testing.T) {
	ctx := context.Background()

	u := gptest.NewUnitTester(t)
	defer u.Remove()

	rs, err := createRootStore(ctx, u)
	require.NoError(t, err)

	tree, err := rs.Tree(ctx)
	require.NoError(t, err)
	assert.Equal(t, []string{"foo"}, tree.List(0))
}

func TestListMulti(t *testing.T) {
	ctx := context.Background()
	ctx = backend.WithCryptoBackend(ctx, backend.Plain)
	ctx = backend.WithRCSBackend(ctx, backend.Noop)

	u := gptest.NewUnitTester(t)
	defer u.Remove()

	// root store
	rs, err := createRootStore(ctx, u)
	require.NoError(t, err)

	ents := make([]string, 0, 3*len(u.Entries))
	ents = append(ents, u.Entries...)

	// sub1 store
	assert.NoError(t, u.InitStore("sub1"))
	for _, k := range u.Entries {
		ents = append(ents, path.Join("sub1", k))
	}

	// sub2 store
	assert.NoError(t, u.InitStore("sub2"))
	for _, k := range u.Entries {
		ents = append(ents, path.Join("sub2", k))
	}

	assert.NoError(t, rs.AddMount(ctx, "sub1", u.StoreDir("sub1")))
	assert.NoError(t, rs.AddMount(ctx, "sub2", u.StoreDir("sub2")))

	tree, err := rs.Tree(ctx)
	require.NoError(t, err)

	sort.Strings(ents)
	lst := tree.List(0)
	sort.Strings(lst)
	assert.Equal(t, ents, lst)
}

func TestListNested(t *testing.T) {
	ctx := context.Background()
	ctx = backend.WithCryptoBackend(ctx, backend.Plain)
	ctx = backend.WithRCSBackend(ctx, backend.Noop)

	u := gptest.NewUnitTester(t)
	defer u.Remove()

	// root store
	rs, err := createRootStore(ctx, u)
	assert.NoError(t, err)

	ents := make([]string, 0, 3*len(u.Entries))
	ents = append(ents, u.Entries...)

	// sub1 store
	assert.NoError(t, u.InitStore("sub1"))
	for _, k := range u.Entries {
		ents = append(ents, path.Join("sub1", k))
	}

	// sub2 store
	assert.NoError(t, u.InitStore("sub2"))
	for _, k := range u.Entries {
		ents = append(ents, path.Join("sub2", k))
	}

	// sub3 store
	assert.NoError(t, u.InitStore("sub3"))
	for _, k := range u.Entries {
		ents = append(ents, path.Join("sub2", "sub3", k))
	}

	assert.NoError(t, rs.AddMount(ctx, "sub1", u.StoreDir("sub1")))
	assert.NoError(t, rs.AddMount(ctx, "sub2", u.StoreDir("sub2")))
	assert.NoError(t, rs.AddMount(ctx, "sub2/sub3", u.StoreDir("sub3")))

	tree, err := rs.Tree(ctx)
	assert.NoError(t, err)

	sort.Strings(ents)
	lst := tree.List(0)
	sort.Strings(lst)
	assert.Equal(t, ents, lst)

	assert.Equal(t, false, rs.Exists(ctx, "sub1"))
	assert.Equal(t, true, rs.IsDir(ctx, "sub1"))
	assert.Equal(t, "", rs.Alias())
	assert.NotNil(t, rs.Storage(ctx, "sub1"))
}

func createRootStore(ctx context.Context, u *gptest.Unit) (*Store, error) {
	ctx = backend.WithRCSBackendString(ctx, "noop")
	ctx = backend.WithCryptoBackendString(ctx, "plain")
	s, err := New(
		ctx,
		&config.Config{
			Root: &config.StoreConfig{
				Path: backend.FromPath(u.StoreDir("")),
			},
		},
	)
	if err != nil {
		return nil, err
	}
	if _, err := s.Initialized(ctx); err != nil {
		return nil, err
	}
	return s, nil
}
