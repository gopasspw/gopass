package tests

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFind(t *testing.T) { //nolint:paralleltest
	ts := newTester(t)
	defer ts.teardown()

	ts.initStore()

	out, err := ts.run("find")
	assert.Error(t, err)
	assert.Equal(t, "\nError: Usage: "+filepath.Base(ts.Binary)+" find <NEEDLE>\n", out)

	_, err = ts.run("config core.showsafecontent false")
	require.NoError(t, err)

	out, err = ts.run("find bar")
	assert.Error(t, err)
	assert.Equal(t, "\nError: no results found\n", out)

	_, err = ts.runCmd([]string{ts.Binary, "insert", "foo/bar"}, []byte("baz"))
	assert.NoError(t, err)

	out, err = ts.run("find bar")
	assert.NoError(t, err)
	assert.Contains(t, "Found exact match in 'foo/bar'\nbaz", out)

	out, err = ts.run("find Bar")
	assert.NoError(t, err)
	assert.Contains(t, "Found exact match in 'foo/bar'\nbaz", out)

	out, err = ts.run("find b")
	assert.NoError(t, err)
	assert.Contains(t, "Found exact match in 'foo/bar'\nbaz", out)

	_, err = ts.run("config core.showsafecontent true")
	require.NoError(t, err)

	out, err = ts.run("find bar")
	assert.NoError(t, err)
	assert.Contains(t, out, "foo/bar")

	out, err = ts.run("find -f bar")
	assert.NoError(t, err)
	assert.Contains(t, out, "foo/bar")
}
