package fs

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/stretchr/testify/assert"
)

func TestFsck(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctx = ctxutil.WithHidden(ctx, true)

	path, cleanup := newTempDir(t)
	defer cleanup()

	l := &loader{}
	s, err := l.Init(ctx, path)
	assert.NoError(t, err)
	assert.NoError(t, l.Handles(ctx, path))

	for _, fn := range []string{
		filepath.Join(path, ".plain-ids"),
		filepath.Join(path, "foo", "bar"),
		filepath.Join(path, "foo", "zen"),
	} {
		assert.NoError(t, os.MkdirAll(filepath.Dir(fn), 0o777))
		assert.NoError(t, os.WriteFile(fn, []byte(fn), 0o663))
	}

	assert.NoError(t, s.Fsck(ctx))
}
