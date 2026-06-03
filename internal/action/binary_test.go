package action

import (
	"bufio"
	"bytes"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/gopass/secrets"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBinary(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := config.NewContextInMemory()
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

	require.Error(t, act.Cat(ctx, gptest.CliCtx(ctx, t)))
	require.Error(t, act.BinaryCopy(ctx, gptest.CliCtx(ctx, t)))
	require.Error(t, act.BinaryMove(ctx, gptest.CliCtx(ctx, t)))
	require.Error(t, act.Sum(ctx, gptest.CliCtx(ctx, t)))
}

func TestBinaryCat(t *testing.T) {
	tSize := 1024

	u := gptest.NewUnitTester(t)

	ctx := config.NewContextInMemory()
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
		require.NoError(t, act.binaryCopy(ctx, gptest.CliCtx(ctx, t), infile, "bar", true))
	})

	t.Run("binary cat bar", func(t *testing.T) {
		require.NoError(t, act.Cat(ctx, gptest.CliCtx(ctx, t, "bar")))
	})

	stdinfile := filepath.Join(u.Dir, "stdin")
	t.Run("binary cat baz from stdin", func(t *testing.T) {
		writeBinfile(t, stdinfile, tSize)

		fd, err := os.Open(stdinfile)
		require.NoError(t, err)
		binstdin = fd
		defer func() {
			binstdin = os.Stdin
			_ = fd.Close()
		}()

		require.NoError(t, act.Cat(ctx, gptest.CliCtx(ctx, t, "baz")))
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

	ctx := config.NewContextInMemory()
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
		require.NoError(t, err)

		catFn := func() {
			binstdin = fd
			defer func() {
				binstdin = os.Stdin
				_ = fd.Close()
			}()

			require.NoError(t, act.Cat(ctx, gptest.CliCtx(ctx, t, "baz")))
		}
		catFn()

		// gopass cat baz and compare output with input, they should match
		buf, err := os.ReadFile(stdinfile)
		require.NoError(t, err)
		sec, err := act.binaryGet(ctx, "baz")
		require.NoError(t, err)

		if string(buf) != string(sec) {
			t.Fatalf("Input and output mismatch at tSize %d", tSize)
		}
		t.Logf("Input and Output match at tSize %d", tSize)
	}
}

func TestBinaryCopy(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := config.NewContextInMemory()
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
		require.NoError(t, os.WriteFile(infile, []byte("0xDEADBEEF\n"), 0o644))
		require.NoError(t, act.binaryCopy(ctx, gptest.CliCtx(ctx, t), infile, "txt", true))
	})

	infile := filepath.Join(u.Dir, "input.raw")
	outfile := filepath.Join(u.Dir, "output.raw")
	t.Run("copy binary file", func(t *testing.T) {
		defer buf.Reset()

		writeBinfile(t, infile, 1024)
		require.NoError(t, act.binaryCopy(ctx, gptest.CliCtx(ctx, t), infile, "bar", true))
	})

	t.Run("binary copy bar tempdir/bar", func(t *testing.T) {
		defer buf.Reset()
		require.NoError(t, act.BinaryCopy(ctx, gptest.CliCtx(ctx, t, "bar", outfile)))
	})

	t.Run("binary copy tempdir/bar tempdir/bar", func(t *testing.T) {
		defer buf.Reset()

		require.Error(t, act.BinaryCopy(ctx, gptest.CliCtx(ctx, t, outfile, outfile)))
	})

	t.Run("binary copy bar bar", func(t *testing.T) {
		defer buf.Reset()
		require.Error(t, act.BinaryCopy(ctx, gptest.CliCtx(ctx, t, "bar", "bar")))
	})

	t.Run("binary move tempdir/bar bar2", func(t *testing.T) {
		defer buf.Reset()
		require.NoError(t, act.BinaryMove(ctx, gptest.CliCtx(ctx, t, outfile, "bar2")))
	})

	t.Run("binary move bar2 tempdir/bar", func(t *testing.T) {
		defer buf.Reset()
		require.NoError(t, act.BinaryMove(ctx, gptest.CliCtx(ctx, t, "bar2", outfile)))
	})
}

// TestBinaryCopyNameAmbiguity covers https://github.com/gopasspw/gopass/issues/3340:
// copying a file into the store must work even when the destination secret name
// matches the basename of an existing file in the current directory, while the
// genuine "both arguments are secrets" ambiguity is still rejected.
func TestBinaryCopyNameAmbiguity(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := config.NewContextInMemory()
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

	// Operate from a directory that holds a file whose name collides with the
	// destination secret name, reproducing the exact scenario from the issue.
	workdir := t.TempDir()
	old, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(workdir))
	defer func() { require.NoError(t, os.Chdir(old)) }()

	require.NoError(t, os.WriteFile(filepath.Join(workdir, "test"), []byte("0xDEADBEEF\n"), 0o644))

	t.Run("file into same-named store entry succeeds", func(t *testing.T) {
		defer buf.Reset()
		// "gopass fscopy test test": source is a real file, destination is the
		// store entry "test". This used to fail with an ambiguity error.
		require.NoError(t, act.BinaryCopy(ctx, gptest.CliCtx(ctx, t, "test", "test")))
		require.True(t, act.Store.Exists(ctx, "test"))
	})

	t.Run("file into nested same-basename store entry succeeds", func(t *testing.T) {
		defer buf.Reset()
		// "gopass fscopy test sub/test": the destination shares the basename of
		// the on-disk file but is a nested store path. This used to fail with
		// the ambiguity error too.
		require.NoError(t, act.BinaryCopy(ctx, gptest.CliCtx(ctx, t, "test", "sub/test")))
		require.True(t, act.Store.Exists(ctx, "sub/test"))
	})

	t.Run("file into rooted same-named path no longer reports ambiguity", func(t *testing.T) {
		defer buf.Reset()
		// "gopass fscopy test /test": the source is a file so this is routed to
		// a filesystem-to-store copy. The store rejects a leading-slash secret
		// name, but the user now gets that clear validation error instead of the
		// misleading "ambiguity detected" message from #3340.
		err := act.BinaryCopy(ctx, gptest.CliCtx(ctx, t, "test", "/test"))
		require.Error(t, err)
		require.NotContains(t, err.Error(), "ambiguity")
	})

	t.Run("genuine secret-to-secret ambiguity still errors", func(t *testing.T) {
		defer buf.Reset()
		// Set up two secrets that have no matching file on disk, then try to
		// fscopy between them: this must keep erroring and point at cp.
		require.NoError(t, act.Store.Set(ctx, "src-secret", secrets.NewAKV()))
		require.NoError(t, act.Store.Set(ctx, "dst-secret", secrets.NewAKV()))
		require.Error(t, act.BinaryCopy(ctx, gptest.CliCtx(ctx, t, "src-secret", "dst-secret")))
	})

	t.Run("unknown source errors", func(t *testing.T) {
		defer buf.Reset()
		// Neither a file on disk nor a secret in the store.
		require.Error(t, act.BinaryCopy(ctx, gptest.CliCtx(ctx, t, "does-not-exist", "dest")))
	})
}

func TestBinarySum(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := config.NewContextInMemory()
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
		require.NoError(t, act.binaryCopy(ctx, gptest.CliCtx(ctx, t), infile, "bar", true))
	})

	t.Run("binary sum bar", func(t *testing.T) {
		require.NoError(t, act.Sum(ctx, gptest.CliCtx(ctx, t, "bar")))
		buf.Reset()
	})
}

func TestBinaryGet(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithHidden(ctx, true)

	act, err := newMock(ctx, u.StoreDir(""))
	require.NoError(t, err)
	require.NotNil(t, act)
	ctx = act.cfg.WithConfig(ctx)

	data := []byte("1\n2\n3\n")
	require.NoError(t, act.insertStdin(ctx, "x", data, false))

	out, err := act.binaryGet(ctx, "x")
	require.NoError(t, err)
	assert.Equal(t, data, out)
}

func writeBinfile(t *testing.T, fn string, size int) {
	t.Helper()

	// tests should be predicable
	lr := rand.New(rand.NewSource(42))

	buf := make([]byte, size)
	n, err := lr.Read(buf)
	require.NoError(t, err)
	assert.Equal(t, size, n)
	require.NoError(t, os.WriteFile(fn, buf, 0o644))
}
