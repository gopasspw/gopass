package tests

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBinaryCopy(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	_, err := ts.run("binary copy")
	assert.Error(t, err)

	ts.initStore()

	out, err := ts.run("binary copy")
	assert.Error(t, err)
	assert.Equal(t, "\nError: Usage: "+filepath.Base(ts.Binary)+" binary copy from to\n", out)

	fn := filepath.Join(ts.tempDir, "copy")
	dat := []byte("foobar")
	assert.NoError(t, ioutil.WriteFile(fn, dat, 0644))

	_, err = ts.run("binary copy " + fn + " foo/bar")
	assert.NoError(t, err)
	assert.NoError(t, os.Remove(fn))

	_, err = ts.run("binary copy foo/bar " + fn)
	assert.NoError(t, err)

	buf, err := ioutil.ReadFile(fn)
	assert.NoError(t, err)

	assert.Equal(t, buf, dat)

	_, err = ts.run("binary cat foo/bar")
	assert.NoError(t, err)
}

func TestBinaryMove(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	_, err := ts.run("binary move")
	assert.Error(t, err)

	ts.initStore()

	out, err := ts.run("binary move")
	assert.Error(t, err)
	assert.Equal(t, "\nError: Usage: "+filepath.Base(ts.Binary)+" binary move from to\n", out)

	fn := filepath.Join(ts.tempDir, "move")
	dat := []byte("foobar")
	assert.NoError(t, ioutil.WriteFile(fn, dat, 0644))

	_, err = ts.run("binary move " + fn + " foo/bar")
	assert.NoError(t, err)
	assert.Error(t, os.Remove(fn))

	_, err = ts.run("binary move foo/bar " + fn)
	assert.NoError(t, err)

	buf, err := ioutil.ReadFile(fn)
	assert.NoError(t, err)

	assert.Equal(t, buf, dat)

	_, err = ts.run("binary cat foo/bar")
	assert.Error(t, err)
}

func TestBinaryShasum(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	_, err := ts.run("binary sha256")
	assert.Error(t, err)

	ts.initStore()

	out, err := ts.run("binary sha256")
	assert.Error(t, err)
	assert.Equal(t, "\nError: Usage: "+filepath.Base(ts.Binary)+" binary sha256 name\n", out)

	fn := filepath.Join(ts.tempDir, "shasum")
	dat := []byte("foobar")
	assert.NoError(t, ioutil.WriteFile(fn, dat, 0644))

	_, err = ts.run("binary move " + fn + " foo/bar")
	assert.NoError(t, err)

	out, err = ts.run("binary sha256 foo/bar")
	assert.NoError(t, err)
	assert.Equal(t, out, "c3ab8ff13720e8ad9047dd39466b3c8974e592c2fa383d4a3960714caef0c4f2")
}
