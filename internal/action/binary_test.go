package action

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/gopasspw/gopass/internal/gptest"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBinary(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = out.WithHidden(ctx, true)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	act, err := newMock(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, act)

	infile := filepath.Join(u.Dir, "input.txt")
	assert.NoError(t, ioutil.WriteFile(infile, []byte("0xDEADBEEF"), 0644))
	assert.NoError(t, act.binaryCopy(ctx, gptest.CliCtx(ctx, t), infile, "bar", true))

	assert.Error(t, act.Cat(gptest.CliCtx(ctx, t)))
	assert.Error(t, act.BinaryCopy(gptest.CliCtx(ctx, t)))
	assert.Error(t, act.BinaryMove(gptest.CliCtx(ctx, t)))
	assert.Error(t, act.Sum(gptest.CliCtx(ctx, t)))
}

func TestBinaryCat(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = out.WithHidden(ctx, true)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	act, err := newMock(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, act)

	infile := filepath.Join(u.Dir, "input.txt")
	assert.NoError(t, ioutil.WriteFile(infile, []byte("0xDEADBEEF"), 0644))
	assert.NoError(t, act.binaryCopy(ctx, gptest.CliCtx(ctx, t), infile, "bar", true))

	// binary cat bar
	assert.NoError(t, act.Cat(gptest.CliCtx(ctx, t, "bar")))

	// binary cat baz from stdin
	stdinfile := filepath.Join(u.Dir, "stdin")
	assert.NoError(t, ioutil.WriteFile(stdinfile, []byte("foo"), 0644))
	fd, err := os.Open(stdinfile)
	assert.NoError(t, err)
	binstdin = fd
	defer func() {
		binstdin = os.Stdin
	}()

	assert.NoError(t, act.Cat(gptest.CliCtx(ctx, t, "baz")))
	sec, err := act.binaryGet(ctx, "baz")
	require.NoError(t, err)
	assert.Equal(t, "foo", string(sec))
}

func TestBinaryCopy(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = out.WithHidden(ctx, true)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	act, err := newMock(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, act)

	infile := filepath.Join(u.Dir, "input.txt")
	assert.NoError(t, ioutil.WriteFile(infile, []byte("0xDEADBEEF"), 0644))
	assert.NoError(t, act.binaryCopy(ctx, gptest.CliCtx(ctx, t), infile, "bar", true))

	outfile := filepath.Join(u.Dir, "output.txt")

	t.Run("binary copy bar tempdir/bar", func(t *testing.T) {
		assert.NoError(t, act.BinaryCopy(gptest.CliCtx(ctx, t, "bar", outfile)))
		buf.Reset()
	})

	t.Run("binary copy tempdir/bar tempdir/bar", func(t *testing.T) {
		assert.Error(t, act.BinaryCopy(gptest.CliCtx(ctx, t, "outfile, outfile")))
		buf.Reset()
	})

	t.Run("binary copy bar bar", func(t *testing.T) {
		assert.Error(t, act.BinaryCopy(gptest.CliCtx(ctx, t, "bar", "bar")))
		buf.Reset()
	})

	t.Run("binary move tempdir/bar bar2", func(t *testing.T) {
		assert.NoError(t, act.BinaryMove(gptest.CliCtx(ctx, t, outfile, "bar2")))
		buf.Reset()
	})
}

func TestBinarySum(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = out.WithHidden(ctx, true)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	act, err := newMock(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, act)

	infile := filepath.Join(u.Dir, "input.txt")
	assert.NoError(t, ioutil.WriteFile(infile, []byte("0xDEADBEEF"), 0644))
	assert.NoError(t, act.binaryCopy(ctx, gptest.CliCtx(ctx, t), infile, "bar", true))

	t.Run("binary sum bar", func(t *testing.T) {
		assert.NoError(t, act.Sum(gptest.CliCtx(ctx, t, "bar")))
		buf.Reset()
	})
}
