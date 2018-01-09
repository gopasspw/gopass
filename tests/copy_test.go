package tests

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCopy(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	_, err := ts.run("copy")
	assert.Error(t, err)

	ts.initStore()

	out, err := ts.run("copy")
	assert.Error(t, err)
	assert.Equal(t, "\nError: Usage: "+filepath.Base(ts.Binary)+" cp old-path new-path\n", out)

	out, err = ts.run("copy foo")
	assert.Error(t, err)
	assert.Equal(t, "\nError: Usage: "+filepath.Base(ts.Binary)+" cp old-path new-path\n", out)

	out, err = ts.run("copy foo bar")
	assert.Error(t, err)
	assert.Equal(t, "\nError: foo does not exist\n", out)

	ts.initSecrets("")

	out, err = ts.run("copy foo bar")
	assert.Error(t, err)
	assert.Equal(t, "\nError: foo does not exist\n", out)

	out, err = ts.run("copy foo/bar foo/baz")
	assert.NoError(t, err)
	assert.Equal(t, "", out)

	orig, err := ts.run("show -f foo/bar")
	assert.NoError(t, err)

	copy, err := ts.run("show -f foo/baz")
	assert.NoError(t, err)

	assert.Equal(t, orig, copy)
}
