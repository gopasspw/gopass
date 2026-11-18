package root

import (
	"testing"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/require"
)

func TestFsck(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithHidden(ctx, true)

	rs, err := createRootStore(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, rs)

	require.NoError(t, rs.Fsck(ctx, "", "", nil))
}

func TestFsckLimitsToStoreAndPath(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithHidden(ctx, true)

	rs, err := createRootStore(ctx, u)
	require.NoError(t, err)

	// Entry counts are powers of two so the number of checked entries
	// identifies the stores: root 1, sub1 2, sub2 4.
	u.Entries = []string{"a", "b"}
	require.NoError(t, u.InitStore("sub1"))
	u.Entries = []string{"foo", "b", "c", "d"}
	require.NoError(t, u.InitStore("sub2"))
	require.NoError(t, rs.AddMount(ctx, "sub1", u.StoreDir("sub1")))
	require.NoError(t, rs.AddMount(ctx, "sub2", u.StoreDir("sub2")))

	for _, tc := range []struct {
		name    string
		store   string
		path    string
		want    int
		wantErr bool
	}{
		{name: "everything", want: 7},
		{name: "store", store: "sub1", want: 2},
		{name: "path in mount", path: "sub2/foo", want: 1},
		{name: "mount as path", path: "sub2", want: 4},
		{name: "mount as path with trailing slash", path: "sub2/", want: 4},
		{name: "store and path in the same mount", store: "sub2", path: "sub2/foo", want: 1},
		{name: "path in root store", path: "foo", want: 1},
		{name: "unknown store", store: "nope", wantErr: true},
		{name: "path outside store", store: "sub1", path: "sub2/foo", wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var checked int
			err := rs.Fsck(ctx, tc.store, tc.path, func() { checked++ })
			if tc.wantErr {
				require.Error(t, err)

				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.want, checked)
		})
	}
}
