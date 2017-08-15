package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFind(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	ts.initStore()

	out, err := ts.run("find")
	assert.Error(t, err)
	assert.Equal(t, "\nError: Usage: gopass find arg\n", out)

	out, err = ts.run("config safecontent false")
	assert.NoError(t, err)

	out, err = ts.run("find bar")
	assert.NoError(t, err)
	assert.Zero(t, out)

	out, err = ts.runCmd([]string{ts.Binary, "insert", "foo/bar"}, []byte("baz"))
	assert.NoError(t, err)

	out, err = ts.run("find bar")
	assert.NoError(t, err)
	assert.Equal(t, "Found exact match in 'foo/bar'\nbaz", out)

	out, err = ts.run("find Bar")
	assert.NoError(t, err)
	assert.Equal(t, "Found exact match in 'foo/bar'\nbaz", out)

	out, err = ts.run("find b")
	assert.NoError(t, err)
	assert.Equal(t, "Found exact match in 'foo/bar'\nbaz", out)
}
