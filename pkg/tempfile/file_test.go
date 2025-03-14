package tempfile

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Example() {
	ctx := config.NewContextInMemory()

	tempfile, err := New(ctx, "gopass-secure-")
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := tempfile.Remove(ctx); err != nil {
			panic(err)
		}
	}()

	fmt.Fprintln(tempfile, "foobar")

	if err := tempfile.Close(); err != nil {
		panic(err)
	}

	out, err := os.ReadFile(tempfile.Name())
	if err != nil {
		panic(err)
	}

	fmt.Println(string(out))

	// Output: foobar
}

func TestTempdirBase(t *testing.T) {
	t.Parallel()

	tempdir := t.TempDir()
	require.NotEmpty(t, tempdir)

	defer func() {
		_ = os.RemoveAll(tempdir)
	}()
}

func TestTempdirBaseEmpty(t *testing.T) {
	oldShm := shmDir
	defer func() {
		shmDir = oldShm
	}()

	shmDir = "/this/should/not/exist"

	assert.Equal(t, "", tempdirBase())
}

func TestTempFiler(t *testing.T) {
	t.Parallel()

	ctx := config.NewContextInMemory()

	// regular tempfile
	tf, err := New(ctx, "gp-test-")
	require.NoError(t, err)

	defer func() {
		require.NoError(t, tf.Close())
	}()

	t.Logf("Name: %s", tf.Name())
	_, err = fmt.Fprintf(tf, "foobar")
	require.NoError(t, err)

	// uninitialized tempfile
	utf := File{}
	assert.Equal(t, "", utf.Name())
	_, err = utf.Write([]byte("foo"))
	require.Error(t, err)
	require.NoError(t, utf.Remove(ctx))
	require.NoError(t, utf.Close())
}

func TestGlobalPrefix(t *testing.T) {
	assertPrefix := func(file *File, prefix string) {
		requirePrefix := filepath.Join(tempdirBase(), prefix)
		fileOrDirName := file.Name()

		if runtime.GOOS != "linux" {
			dir := filepath.Dir(fileOrDirName)
			fileOrDirName = filepath.Base(dir)
		}

		assert.True(t, strings.HasPrefix(fileOrDirName, requirePrefix))
	}
	ctx := config.NewContextInMemory()

	assert.Equal(t, "", globalPrefix)

	// without global prefix
	withoutGlobalPrefix, err := New(ctx, "some-prefix")
	require.NoError(t, err)

	defer func() {
		require.NoError(t, withoutGlobalPrefix.Close())
	}()

	assertPrefix(withoutGlobalPrefix, "some-prefix")

	// with global prefix
	globalPrefix = "global-prefix."
	withGlobalPrefix, err := New(ctx, "some-prefix")
	require.NoError(t, err)

	defer func() {
		globalPrefix = ""

		require.NoError(t, withGlobalPrefix.Close())
	}()

	assertPrefix(withGlobalPrefix, "global-prefix.some-prefix")
}
