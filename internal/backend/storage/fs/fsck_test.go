package fs

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/stretchr/testify/assert"
)

func TestFsck(t *testing.T) {
	ctx := context.Background()
	ctx = out.WithHidden(ctx, true)

	path, cleanup := newTempDir(t)
	defer cleanup()

	l := &loader{}
	s, err := l.Init(ctx, path)
	assert.NoError(t, err)
	assert.NoError(t, l.Handles(path))

	for _, fn := range []string{
		filepath.Join(path, ".plain-ids"),
		filepath.Join(path, "foo", "bar"),
		filepath.Join(path, "foo", "zen"),
	} {
		assert.NoError(t, os.MkdirAll(filepath.Dir(fn), 0777))
		assert.NoError(t, ioutil.WriteFile(fn, []byte(fn), 0663))
	}

	assert.NoError(t, s.Fsck(ctx))
}
