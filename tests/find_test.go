package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFind(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	ts.initializeStore()

	out, err := ts.run("find")
	assert.Error(t, err)
	assert.Equal(t, "\nError: Usage: gopass find arg\n", out)

	out, err = ts.run("find bar")
	assert.NoError(t, err)
	assert.Zero(t, out)

	ts.initializeSecrets()

	out, err = ts.run("find bar")
	assert.NoError(t, err)
	assert.Equal(t, "foo/bar", out)

	out, err = ts.run("find b")
	assert.NoError(t, err)
	assert.Equal(t, "baz\nfoo/bar", out)
}
