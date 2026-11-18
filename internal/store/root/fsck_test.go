package root

import (
	"bytes"
	"context"
	"testing"

	"github.com/fatih/color"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/store/leaf"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/gopass/secrets"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
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

func setupMultiMountFsckStore(t *testing.T) (*Store, context.Context) {
	t.Helper()

	u := gptest.NewUnitTester(t)
	color.NoColor = true

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithTerminal(ctx, false)
	ctx = ctxutil.WithHidden(ctx, false)
	ctx = leaf.WithFsckDecrypt(ctx, true)

	rs, err := createRootStore(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, rs)
	ctx = rs.WithStoreConfig(ctx)

	for _, dir := range []string{"sub1", "sub2", "sub1-nested"} {
		require.NoError(t, u.InitStore(dir))
	}

	require.NoError(t, rs.AddMount(ctx, "sub1", u.StoreDir("sub1")))
	require.NoError(t, rs.AddMount(ctx, "sub2", u.StoreDir("sub2")))
	require.NoError(t, rs.AddMount(ctx, "sub1/nested", u.StoreDir("sub1-nested")))

	for _, name := range []string{"rootonly", "sub1/only1", "sub1/dir/item", "sub2/only2", "sub1/nested/only3"} {
		sec := secrets.NewAKV()
		sec.SetPassword("pw-" + name)
		require.NoError(t, rs.Set(ctx, name, sec))
	}

	return rs, ctx
}

func TestFsckMultiMount(t *testing.T) {
	rs, ctx := setupMultiMountFsckStore(t)

	for _, tc := range []struct {
		name       string
		store      string
		path       string
		wantErr    bool
		want       []string
		wantAbsent []string
	}{
		{
			name:  "all stores",
			store: "",
			path:  "",
			want: []string{
				"[] Re-encrypting rootonly",
				"[sub1] Re-encrypting only1",
				"[sub2] Re-encrypting only2",
				"[sub1/nested] Re-encrypting only3",
			},
		},
		{
			name:  "explicit sub1 mount only",
			store: "sub1",
			path:  "",
			want: []string{
				"[sub1] Re-encrypting only1",
				"[sub1] Re-encrypting dir/item",
			},
			wantAbsent: []string{
				"[] Re-encrypting",
				"rootonly",
				"[sub2]",
				"only2",
				"[sub1/nested]",
				"only3",
			},
		},
		{
			name:  "explicit sub1 mount with trailing slash",
			store: "sub1/",
			path:  "",
			want: []string{
				"[sub1] Re-encrypting only1",
			},
			wantAbsent: []string{
				"[] Re-encrypting",
				"[sub2]",
				"[sub1/nested]",
			},
		},
		{
			name:  "explicit root store only",
			store: "<root>",
			path:  "",
			want: []string{
				"[] Re-encrypting rootonly",
				"[] Re-encrypting foo",
			},
			wantAbsent: []string{
				"[sub1]",
				"[sub2]",
				"[sub1/nested]",
			},
		},
		{
			name:  "explicit sub1 with relative filter",
			store: "sub1",
			path:  "dir/item",
			want: []string{
				"[sub1] Re-encrypting dir/item",
			},
			wantAbsent: []string{
				"[sub1] Re-encrypting only1",
				"[] Re-encrypting",
				"[sub2]",
				"[sub1/nested]",
			},
		},
		{
			name:  "explicit sub1 with mount-prefixed filter",
			store: "sub1",
			path:  "sub1/dir/item",
			want: []string{
				"[sub1] Re-encrypting dir/item",
			},
			wantAbsent: []string{
				"[sub1] Re-encrypting only1",
				"[] Re-encrypting",
				"[sub2]",
				"[sub1/nested]",
			},
		},
		{
			name:    "unknown store returns error",
			store:   "nonexistent",
			path:    "",
			wantErr: true,
		},
		{
			name:    "slash-only store returns error",
			store:   "/",
			path:    "",
			wantErr: true,
		},
		{
			name:  "path filter matching sub2 without trailing slash",
			store: "",
			path:  "sub2",
			want: []string{
				"[sub2] Re-encrypting only2",
			},
			wantAbsent: []string{
				"[] Re-encrypting",
				"[sub1]",
				"[sub1/nested]",
			},
		},
		{
			name:  "path filter matching sub2 with trailing slash",
			store: "",
			path:  "sub2/",
			want: []string{
				"[sub2] Re-encrypting only2",
			},
			wantAbsent: []string{
				"[] Re-encrypting",
				"[sub1]",
				"[sub1/nested]",
			},
		},
		{
			name:  "path filter matching shadowed nested mount",
			store: "",
			path:  "sub1/nested/only3",
			want: []string{
				"[sub1/nested] Re-encrypting only3",
			},
			wantAbsent: []string{
				"[] Re-encrypting",
				"[sub1] ",
				"[sub2]",
			},
		},
		{
			name:  "path filter on parent mount includes nested mount but not root or siblings",
			store: "",
			path:  "sub1",
			want: []string{
				"[sub1] Re-encrypting only1",
				"[sub1/nested] Re-encrypting only3",
			},
			wantAbsent: []string{
				"[] Re-encrypting",
				"[sub2]",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			runCtx := out.WithWriter(ctx, buf, buf)

			err := rs.Fsck(runCtx, tc.store, tc.path, nil)
			if tc.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			output := buf.String()

			for _, w := range tc.want {
				assert.Contains(t, output, w)
			}

			for _, a := range tc.wantAbsent {
				assert.NotContains(t, output, a)
			}
		})
	}
}
