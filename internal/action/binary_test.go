package action

import (
	"bufio"
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

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithHidden(ctx, true)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	act, err := newMock(ctx, u.StoreDir(""))
	require.NoError(t, err)
	require.NotNil(t, act)
	ctx = act.cfg.WithConfig(ctx)

	assert.Error(t, act.Cat(gptest.CliCtx(ctx, t)))
	assert.Error(t, act.BinaryCopy(gptest.CliCtx(ctx, t)))
	assert.Error(t, act.BinaryMove(gptest.CliCtx(ctx, t)))
	assert.Error(t, act.Sum(gptest.CliCtx(ctx, t)))
}

func TestBinaryCat(t *testing.T) {
	tSize := 1024

	u := gptest.NewUnitTester(t)

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

	act, err := newMock(ctx, u.StoreDir(""))
	require.NoError(t, err)
	require.NotNil(t, act)
	ctx = act.cfg.WithConfig(ctx)

	infile := filepath.Join(u.Dir, "input.txt")
	writeBinfile(t, infile, tSize)

	t.Run("populate store", func(t *testing.T) {
		assert.NoError(t, act.binaryCopy(ctx, gptest.CliCtx(ctx, t), infile, "bar", true))
	})

	t.Run("binary cat bar", func(t *testing.T) {
		assert.NoError(t, act.Cat(gptest.CliCtx(ctx, t, "bar")))
	})

	stdinfile := filepath.Join(u.Dir, "stdin")
	t.Run("binary cat baz from stdin", func(t *testing.T) {
		writeBinfile(t, stdinfile, tSize)

		fd, err := os.Open(stdinfile)
		assert.NoError(t, err)
		binstdin = fd
		defer func() {
			binstdin = os.Stdin
			_ = fd.Close()
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

func TestBinaryCatSizes(t *testing.T) {
	u := gptest.NewUnitTester(t)

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

	act, err := newMock(ctx, u.StoreDir(""))
	require.NoError(t, err)
	require.NotNil(t, act)
	ctx = act.cfg.WithConfig(ctx)

	for tSize := 1024; tSize < bufio.MaxScanTokenSize*2; tSize += 1024 {
		// cat stdinfile | gopass cat baz
		stdinfile := filepath.Join(u.Dir, "stdin")
		writeBinfile(t, stdinfile, tSize)

		fd, err := os.Open(stdinfile)
		assert.NoError(t, err)

		catFn := func() {
			binstdin = fd
			defer func() {
				binstdin = os.Stdin
				_ = fd.Close()
			}()

			assert.NoError(t, act.Cat(gptest.CliCtx(ctx, t, "baz")))
		}
		catFn()

		// gopass cat baz and compare output with input, they should match
		buf, err := os.ReadFile(stdinfile)
		require.NoError(t, err)
		sec, err := act.binaryGet(ctx, "baz")
		require.NoError(t, err)

		if string(buf) != string(sec) {
			t.Fatalf("Input and output mismatch at tSize %d", tSize)

			break
		}
		t.Logf("Input and Output match at tSize %d", tSize)
	}
}

func TestBinaryCopy(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithHidden(ctx, true)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	act, err := newMock(ctx, u.StoreDir(""))
	require.NoError(t, err)
	require.NotNil(t, act)
	ctx = act.cfg.WithConfig(ctx)

	t.Run("copy textfile", func(t *testing.T) {
		defer buf.Reset()

		infile := filepath.Join(u.Dir, "input.txt")
		assert.NoError(t, os.WriteFile(infile, []byte("0xDEADBEEF\n"), 0o644))
		assert.NoError(t, act.binaryCopy(ctx, gptest.CliCtx(ctx, t), infile, "txt", true))
	})

	infile := filepath.Join(u.Dir, "input.raw")
	outfile := filepath.Join(u.Dir, "output.raw")
	t.Run("copy binary file", func(t *testing.T) {
		defer buf.Reset()

		writeBinfile(t, infile, 1024)
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

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithHidden(ctx, true)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	act, err := newMock(ctx, u.StoreDir(""))
	require.NoError(t, err)
	require.NotNil(t, act)
	ctx = act.cfg.WithConfig(ctx)

	infile := filepath.Join(u.Dir, "input.raw")

	t.Run("populate store", func(t *testing.T) {
		writeBinfile(t, infile, 1024)
		assert.NoError(t, act.binaryCopy(ctx, gptest.CliCtx(ctx, t), infile, "bar", true))
	})

	t.Run("binary sum bar", func(t *testing.T) {
		assert.NoError(t, act.Sum(gptest.CliCtx(ctx, t, "bar")))
		buf.Reset()
	})
}

func TestBinaryGet(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithHidden(ctx, true)

	act, err := newMock(ctx, u.StoreDir(""))
	require.NoError(t, err)
	require.NotNil(t, act)
	ctx = act.cfg.WithConfig(ctx)

	data := []byte("1\n2\n3\n")
	assert.NoError(t, act.insertStdin(ctx, "x", data, false))

	out, err := act.binaryGet(ctx, "x")
	assert.NoError(t, err)
	assert.Equal(t, data, out)
}

func writeBinfile(t *testing.T, fn string, size int) {
	t.Helper()

	// tests should be predicable
	lr := rand.New(rand.NewSource(42))

	buf := make([]byte, size)
	n, err := lr.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, size, n)
	assert.NoError(t, os.WriteFile(fn, buf, 0o644))
}
