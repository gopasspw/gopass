package tests

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFind(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	ts.initStore()

	out, err := ts.run("find")
	require.Error(t, err)
	assert.Equal(t, "\nError: Usage: "+filepath.Base(ts.Binary)+" find <pattern>\n", out)

	_, err = ts.run("config show.safecontent false")
	require.NoError(t, err)

	out, err = ts.run("find bar")
	require.Error(t, err)
	assert.Equal(t, "\nError: no results found\n", out)

	_, err = ts.runCmd([]string{ts.Binary, "insert", "foo/bar"}, []byte("baz"))
	require.NoError(t, err)

	out, err = ts.run("find bar")
	require.NoError(t, err)
	assert.Contains(t, "Found exact match in 'foo/bar'\nbaz", out)

	out, err = ts.run("find Bar")
	require.NoError(t, err)
	assert.Contains(t, "Found exact match in 'foo/bar'\nbaz", out)

	out, err = ts.run("find b")
	require.NoError(t, err)
	assert.Contains(t, "Found exact match in 'foo/bar'\nbaz", out)

	_, err = ts.run("config show.safecontent true")
	require.NoError(t, err)

	out, err = ts.run("find bar")
	require.NoError(t, err)
	assert.Contains(t, out, "foo/bar")

	out, err = ts.run("find -f bar")
	require.NoError(t, err)
	assert.Contains(t, out, "foo/bar")
}
