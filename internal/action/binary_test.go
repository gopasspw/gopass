package action

import (
	"bytes"
	"context"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBinary(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithHidden(ctx, true)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	act, err := newMock(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, act)

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
	ctx = ctxutil.WithHidden(ctx, true)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	stdout = buf
	defer func() {
		out.Stdout = os.Stdout
		stdout = os.Stdout
	}()

	act, err := newMock(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, act)

	infile := filepath.Join(u.Dir, "input.txt")
	writeBinfile(t, infile)

	t.Run("populate store", func(t *testing.T) {
		assert.NoError(t, act.binaryCopy(ctx, gptest.CliCtx(ctx, t), infile, "bar", true))
	})

	t.Run("binary cat bar", func(t *testing.T) {
		assert.NoError(t, act.Cat(gptest.CliCtx(ctx, t, "bar")))
	})

	stdinfile := filepath.Join(u.Dir, "stdin")
	t.Run("binary cat baz from stdin", func(t *testing.T) {
		writeBinfile(t, stdinfile)

		fd, err := os.Open(stdinfile)
		assert.NoError(t, err)
		binstdin = fd
		defer func() {
			binstdin = os.Stdin
			fd.Close()
		}()

		assert.NoError(t, act.Cat(gptest.CliCtx(ctx, t, "baz")))
	})

	t.Run("compare output", func(t *testing.T) {
		buf, err := os.ReadFile(stdinfile)
		require.NoError(t, err)
		sec, err := act.binaryGet(ctx, "baz")
		require.NoError(t, err)
		assert.Equal(t, string(buf), string(sec))
	})
}

func TestBinaryCopy(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithHidden(ctx, true)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	act, err := newMock(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, act)

	t.Run("copy textfile", func(t *testing.T) {
		defer buf.Reset()

		infile := filepath.Join(u.Dir, "input.txt")
		assert.NoError(t, os.WriteFile(infile, []byte("0xDEADBEEF\n"), 0644))
		assert.NoError(t, act.binaryCopy(ctx, gptest.CliCtx(ctx, t), infile, "txt", true))
	})

	infile := filepath.Join(u.Dir, "input.raw")
	outfile := filepath.Join(u.Dir, "output.raw")
	t.Run("copy binary file", func(t *testing.T) {
		defer buf.Reset()

		writeBinfile(t, infile)
		assert.NoError(t, act.binaryCopy(ctx, gptest.CliCtx(ctx, t), infile, "bar", true))
	})

	t.Run("binary copy bar tempdir/bar", func(t *testing.T) {
		defer buf.Reset()
		assert.NoError(t, act.BinaryCopy(gptest.CliCtx(ctx, t, "bar", outfile)))
	})

	t.Run("binary copy tempdir/bar tempdir/bar", func(t *testing.T) {
		defer buf.Reset()

		assert.Error(t, act.BinaryCopy(gptest.CliCtx(ctx, t, outfile, outfile)))
	})

	t.Run("binary copy bar bar", func(t *testing.T) {
		defer buf.Reset()
		assert.Error(t, act.BinaryCopy(gptest.CliCtx(ctx, t, "bar", "bar")))
	})

	t.Run("binary move tempdir/bar bar2", func(t *testing.T) {
		defer buf.Reset()
		assert.NoError(t, act.BinaryMove(gptest.CliCtx(ctx, t, outfile, "bar2")))
	})

	t.Run("binary move bar2 tempdir/bar", func(t *testing.T) {
		defer buf.Reset()
		assert.NoError(t, act.BinaryMove(gptest.CliCtx(ctx, t, "bar2", outfile)))
	})
}

func TestBinarySum(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithHidden(ctx, true)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	act, err := newMock(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, act)

	infile := filepath.Join(u.Dir, "input.raw")

	t.Run("populate store", func(t *testing.T) {
		writeBinfile(t, infile)
		assert.NoError(t, act.binaryCopy(ctx, gptest.CliCtx(ctx, t), infile, "bar", true))
	})

	t.Run("binary sum bar", func(t *testing.T) {
		assert.NoError(t, act.Sum(gptest.CliCtx(ctx, t, "bar")))
		buf.Reset()
	})
}

func writeBinfile(t *testing.T, fn string) {
	// tests should be predicable
	rand.Seed(42)

	size := 1024
	buf := make([]byte, size)
	n, err := rand.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, size, n)
	assert.NoError(t, os.WriteFile(fn, buf, 0644))
}
