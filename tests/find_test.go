package tests

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFind(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping test on windows.")
	}
	ts := newTester(t)
	defer ts.teardown()

	ts.initStore()

	out, err := ts.run("find")
	assert.Error(t, err)
	assert.Equal(t, "\nError: Usage: "+filepath.Base(ts.Binary)+" find <NEEDLE>\n", out)

	_, err = ts.run("config safecontent false")
	require.NoError(t, err)

	out, err = ts.run("find bar")
	assert.Error(t, err)
	assert.Equal(t, "\nError: no results found\n", out)

	_, err = ts.runCmd([]string{ts.Binary, "insert", "foo/bar"}, []byte("baz"))
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

	_, err = ts.run("config safecontent true")
	require.NoError(t, err)

	out, err = ts.run("find bar")
	assert.NoError(t, err)
	assert.Contains(t, out, "no safe content to display")

	out, err = ts.run("find -f bar")
	assert.NoError(t, err)
	assert.Contains(t, "Found exact match in 'foo/bar'\nbaz", out)
}
