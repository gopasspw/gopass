package tests

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBinaryCopy(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	t.Run("empty store", func(t *testing.T) {
		_, err := ts.run("fscopy")
		assert.Error(t, err)
	})

	ts.initStore()

	t.Run("no args", func(t *testing.T) {
		out, err := ts.run("fscopy")
		assert.Error(t, err)
		assert.Equal(t, "\nError: usage: gopass fscopy from to\n", out)
	})

	fn := filepath.Join(ts.tempDir, "copy")
	dat := []byte("foobar")
	require.NoError(t, ioutil.WriteFile(fn, dat, 0644))

	t.Run("copy file to store", func(t *testing.T) {
		_, err := ts.run("fscopy " + fn + " foo/bar")
		require.NoError(t, err)
		assert.NoError(t, os.Remove(fn))
	})

	t.Run("copy store to file", func(t *testing.T) {
		_, err := ts.run("fscopy foo/bar " + fn)
		assert.NoError(t, err)

		buf, err := ioutil.ReadFile(fn)
		require.NoError(t, err)

		assert.Equal(t, buf, dat)
	})

	t.Run("cat from store", func(t *testing.T) {
		_, err := ts.run("cat foo/bar")
		assert.NoError(t, err)
	})
}

func TestBinaryMove(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	t.Run("empty store", func(t *testing.T) {
		_, err := ts.run("fsmove")
		assert.Error(t, err)
	})

	ts.initStore()

	t.Run("no args", func(t *testing.T) {
		out, err := ts.run("fsmove")
		assert.Error(t, err)
		assert.Equal(t, "\nError: usage: gopass fsmove from to\n", out)
	})

	fn := filepath.Join(ts.tempDir, "move")
	dat := []byte("foobar")
	require.NoError(t, ioutil.WriteFile(fn, dat, 0644))

	t.Run("move fs to store", func(t *testing.T) {
		_, err := ts.run("fsmove " + fn + " foo/bar")
		assert.NoError(t, err)
		assert.Error(t, os.Remove(fn))
	})

	t.Run("move store to fs", func(t *testing.T) {
		_, err := ts.run("fsmove foo/bar " + fn)
		assert.NoError(t, err)

		buf, err := ioutil.ReadFile(fn)
		require.NoError(t, err)

		assert.Equal(t, buf, dat)
	})

	t.Run("cat secret", func(t *testing.T) {
		_, err := ts.run("cat foo/bar")
		assert.Error(t, err)
	})
}

func TestBinaryShasum(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	t.Run("shasum on empty store", func(t *testing.T) {
		_, err := ts.run("sha256")
		assert.Error(t, err)
	})

	ts.initStore()

	t.Run("shasum w/o args", func(t *testing.T) {
		out, err := ts.run("sha256")
		assert.Error(t, err)
		assert.Equal(t, "\nError: Usage: gopass sha256 name\n", out)
	})

	t.Run("populate store", func(t *testing.T) {
		fn := filepath.Join(ts.tempDir, "shasum")
		dat := []byte("foobar")
		require.NoError(t, ioutil.WriteFile(fn, dat, 0644))

		_, err := ts.run("fsmove " + fn + " foo/bar")
		assert.NoError(t, err)
	})

	t.Run("shasum on binary secret", func(t *testing.T) {
		out, err := ts.run("sha256 foo/bar")
		assert.NoError(t, err)
		assert.Equal(t, "c3ab8ff13720e8ad9047dd39466b3c8974e592c2fa383d4a3960714caef0c4f2", out)
	})
}
