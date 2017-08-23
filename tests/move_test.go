package tests

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMove(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	_, err := ts.run("move")
	assert.Error(t, err)

	ts.initStore()

	out, err := ts.run("move")
	assert.Error(t, err)
	assert.Equal(t, "\nError: Usage: "+filepath.Base(ts.Binary)+" mv old-path new-path\n", out)

	out, err = ts.run("move foo")
	assert.Error(t, err)
	assert.Equal(t, "\nError: Usage: "+filepath.Base(ts.Binary)+" mv old-path new-path\n", out)

	out, err = ts.run("move foo bar")
	assert.Error(t, err)
	assert.Equal(t, "\nError: failed to decrypt 'foo': Entry is not in the password store\n", out)

	ts.initSecrets("")

	_, err = ts.run("move foo bar")
	assert.NoError(t, err)

	out, _ = ts.run("move foo/bar foo/baz")
	assert.Equal(t, "\nError: failed to decrypt 'foo/bar': Entry is not in the password store\n", out)

	_, err = ts.run("show -f bar/bar")
	assert.NoError(t, err)

	_, err = ts.run("show -f baz")
	assert.NoError(t, err)
}
