package root

import (
	"context"
	"path"
	"sort"
	"testing"

	"github.com/gopasspw/gopass/internal/backend"
	_ "github.com/gopasspw/gopass/internal/backend/crypto"
	_ "github.com/gopasspw/gopass/internal/backend/storage"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/tree"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSimpleList(t *testing.T) {
	ctx := context.Background()

	u := gptest.NewUnitTester(t)
	defer u.Remove()

	rs, err := createRootStore(ctx, u)
	require.NoError(t, err)

	st, err := rs.Tree(ctx)
	require.NoError(t, err)
	assert.Equal(t, []string{"foo"}, st.List(tree.INF))
}

func TestListMulti(t *testing.T) {
	ctx := context.Background()
	ctx = backend.WithCryptoBackend(ctx, backend.Plain)

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

	st, err := rs.Tree(ctx)
	require.NoError(t, err)

	sort.Strings(ents)
	lst := st.List(tree.INF)
	sort.Strings(lst)
	assert.Equal(t, ents, lst)

	assert.Contains(t, rs.String(), "Store(Path:")
}

func TestListNested(t *testing.T) {
	ctx := context.Background()
	ctx = backend.WithCryptoBackend(ctx, backend.Plain)

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

	st, err := rs.Tree(ctx)
	assert.NoError(t, err)

	sort.Strings(ents)
	lst := st.List(tree.INF)
	sort.Strings(lst)
	assert.Equal(t, ents, lst)

	assert.False(t, rs.Exists(ctx, "sub1"))
	assert.True(t, rs.IsDir(ctx, "sub1"))
	assert.Equal(t, "", rs.Alias())
	assert.NotNil(t, rs.Storage(ctx, "sub1"))
}

func createRootStore(ctx context.Context, u *gptest.Unit) (*Store, error) {
	ctx = backend.WithCryptoBackendString(ctx, "plain")
	s := New(
		&config.Config{
			Path: u.StoreDir(""),
		},
	)
	if _, err := s.IsInitialized(ctx); err != nil {
		return nil, err
	}
	return s, nil
}
