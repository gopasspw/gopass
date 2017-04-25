package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCopy(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	_, err := ts.run("copy")
	assert.Error(t, err)

	ts.initializeStore()

	out, err := ts.run("copy")
	assert.Error(t, err)
	assert.Equal(t, "\nError: Usage: gopass cp old-path new-path\n", out)

	out, err = ts.run("copy foo")
	assert.Error(t, err)
	assert.Equal(t, "\nError: Usage: gopass cp old-path new-path\n", out)

	out, err = ts.run("copy foo bar")
	assert.Error(t, err)
	assert.Equal(t, "\nError: foo doesn't exists\n", out)

	ts.initializeSecrets()

	//TODO: foo is a directory to be copied, which doesn't work
	out, err = ts.run("copy foo bar")
	assert.Error(t, err)
	assert.Equal(t, "\nError: foo doesn't exists\n", out)

	out, err = ts.run("copy foo/bar foo/baz")
	assert.NoError(t, err)
	assert.Equal(t, "", out)

	orig, err := ts.run("show -f foo/bar")
	assert.NoError(t, err)

	copy, err := ts.run("show -f foo/baz")
	assert.NoError(t, err)

	assert.Equal(t, orig, copy)
}
