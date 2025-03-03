package fossilfs

import (
	"context"
	"testing"

	"github.com/gopasspw/gopass/internal/backend/storage/fs"
	"github.com/stretchr/testify/assert"
)

func TestFossil_Get(t *testing.T) {
	td := t.TempDir()
	fossil := &Fossil{fs: fs.New(td)}
	ctx := context.Background()
	name := "test"

	fossil.fs.Set(ctx, name, []byte("content"))

	content, err := fossil.Get(ctx, name)
	assert.NoError(t, err)
	assert.Equal(t, []byte("content"), content)
}

func TestFossil_Set(t *testing.T) {
	td := t.TempDir()
	fossil := &Fossil{fs: fs.New(td)}
	ctx := context.Background()
	name := "test"
	value := []byte("content")

	err := fossil.Set(ctx, name, value)
	assert.NoError(t, err)
}

func TestFossil_Delete(t *testing.T) {
	t.Skip("needs fossil binary")

	td := t.TempDir()
	fossil := &Fossil{fs: fs.New(td)}
	ctx := context.Background()
	name := "test"

	fossil.fs.Set(ctx, name, []byte("content"))

	err := fossil.Delete(ctx, name)
	assert.NoError(t, err)
}

func TestFossil_Exists(t *testing.T) {
	td := t.TempDir()
	fossil := &Fossil{fs: fs.New(td)}
	ctx := context.Background()
	name := "test"

	fossil.fs.Set(ctx, name, []byte("content"))

	exists := fossil.Exists(ctx, name)
	assert.True(t, exists)
}

func TestFossil_List(t *testing.T) {
	td := t.TempDir()
	fossil := &Fossil{fs: fs.New(td)}
	ctx := context.Background()
	prefix := "test"

	fossil.fs.Set(ctx, "test/foo", []byte("content"))
	fossil.fs.Set(ctx, "test/bar", []byte("content"))
	fossil.fs.Set(ctx, "foo/bar", []byte("content"))

	list, err := fossil.List(ctx, prefix)
	assert.NoError(t, err)
	assert.Equal(t, []string{"test/bar", "test/foo"}, list)
}

func TestFossil_IsDir(t *testing.T) {
	td := t.TempDir()
	fossil := &Fossil{fs: fs.New(td)}
	ctx := context.Background()
	name := "test"

	fossil.fs.Set(ctx, "test/foo", []byte("content"))

	assert.True(t, fossil.IsDir(ctx, name))
}

func TestFossil_Prune(t *testing.T) {
	td := t.TempDir()
	fossil := &Fossil{fs: fs.New(td)}
	ctx := context.Background()
	prefix := "test"

	fossil.fs.Set(ctx, "test/foo", []byte("content"))
	fossil.fs.Set(ctx, "test/bar", []byte("content"))

	err := fossil.Prune(ctx, prefix)
	assert.NoError(t, err)
}

func TestFossil_String(t *testing.T) {
	td := t.TempDir()
	fossil := &Fossil{fs: fs.New(td)}

	str := fossil.String()
	assert.Contains(t, str, "fossilfs(")
	assert.Contains(t, str, "path:/path/to/storage")
}

func TestFossil_Path(t *testing.T) {
	td := t.TempDir()
	fossil := &Fossil{fs: fs.New(td)}

	path := fossil.Path()
	assert.Equal(t, "/path/to/storage", path)
}

func TestFossil_Fsck(t *testing.T) {
	td := t.TempDir()
	fossil := &Fossil{fs: fs.New(td)}
	ctx := context.Background()

	err := fossil.Fsck(ctx)
	assert.NoError(t, err)
}

func TestFossil_Link(t *testing.T) {
	td := t.TempDir()
	fossil := &Fossil{fs: fs.New(td)}
	ctx := context.Background()
	from := "from"
	to := "to"

	fossil.fs.Set(ctx, "from", []byte("content"))

	err := fossil.Link(ctx, from, to)
	assert.NoError(t, err)
}

func TestFossil_Move(t *testing.T) {
	td := t.TempDir()
	fossil := &Fossil{fs: fs.New(td)}
	ctx := context.Background()
	src := "src"
	dst := "dst"
	del := true

	fossil.fs.Set(ctx, "src", []byte("content"))

	err := fossil.Move(ctx, src, dst, del)
	assert.NoError(t, err)
}
