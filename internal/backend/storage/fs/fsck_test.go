package fs

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/stretchr/testify/require"
)

func TestFsck(t *testing.T) {
	t.Parallel()

	ctx := config.NewNoWrites().WithConfig(context.Background())
	ctx = ctxutil.WithHidden(ctx, true)

	path := t.TempDir()

	l := &loader{}
	s, err := l.Init(ctx, path)
	require.NoError(t, err)
	require.NoError(t, l.Handles(ctx, path))

	for _, fn := range []string{
		filepath.Join(path, ".plain-ids"),
		filepath.Join(path, "foo", "bar"),
		filepath.Join(path, "foo", "zen"),
	} {
		require.NoError(t, os.MkdirAll(filepath.Dir(fn), 0o777))
		require.NoError(t, os.WriteFile(fn, []byte(fn), 0o663))
	}

	require.NoError(t, s.Fsck(ctx))
}
