package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestYAMLAndSecret(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	_, err := ts.run("show foo/bar baz")
	assert.Error(t, err)

	ts.initStore()

	out, err := ts.run("foo/bar baz")
	assert.Error(t, err)
	assert.Equal(t, "Entry 'foo/bar' not found. Starting search...\n\nError: no results found\n", out)

	_, err = ts.runCmd([]string{ts.Binary, "insert", "foo/bar"}, []byte("moar"))
	assert.NoError(t, err)

	_, err = ts.runCmd([]string{ts.Binary, "insert", "foo/bar", "baz"}, []byte("moar"))
	assert.NoError(t, err)

	out, err = ts.run("foo/bar baz")
	assert.NoError(t, err)
	assert.Equal(t, "moar", out)
}

func TestInvalidYAML(t *testing.T) {
	var testBody = `somepasswd
---
Test / test.com
username: myuser@test.com
password: somepasswd
url: http://www.test.com/`

	ts := newTester(t)
	defer ts.teardown()

	_, err := ts.run("show foo/bar")
	assert.Error(t, err)

	ts.initStore()

	out, err := ts.run("foo/bar")
	assert.Error(t, err)
	assert.Equal(t, "Entry 'foo/bar' not found. Starting search...\n\nError: no results found\n", out)

	_, err = ts.runCmd([]string{ts.Binary, "insert", "foo/bar"}, []byte(testBody))
	assert.NoError(t, err)

	_, err = ts.run("show foo/bar")
	assert.NoError(t, err)
}
