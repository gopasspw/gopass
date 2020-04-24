package tests

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestYAMLAndSecret(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping test on windows.")
	}
	ts := newTester(t)
	defer ts.teardown()

	_, err := ts.run("show foo/bar baz")
	assert.Error(t, err)

	ts.initStore()

	out, err := ts.run("foo/bar baz")
	assert.Error(t, err)
	assert.Equal(t, "\nError: failed to retrieve secret 'foo/bar': Entry is not in the password store\n", out)

	_, err = ts.runCmd([]string{ts.Binary, "insert", "foo/bar"}, []byte("moar"))
	require.NoError(t, err)

	_, err = ts.runCmd([]string{ts.Binary, "insert", "foo/bar", "baz"}, []byte("moar"))
	require.NoError(t, err)

	out, err = ts.run("foo/bar baz")
	assert.NoError(t, err)
	assert.Equal(t, "moar", out)
}

func TestInvalidYAML(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping test on windows.")
	}
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
	assert.Equal(t, "\nError: failed to retrieve secret 'foo/bar': Entry is not in the password store\n", out)

	_, err = ts.runCmd([]string{ts.Binary, "insert", "foo/bar"}, []byte(testBody))
	assert.NoError(t, err)

	_, err = ts.run("show foo/bar")
	assert.NoError(t, err)
}
