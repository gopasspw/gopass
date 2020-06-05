package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestYAMLAndSecret(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	t.Run("show key from uninitialized store", func(t *testing.T) {
		_, err := ts.run("show foo/bar baz")
		assert.Error(t, err)
	})

	ts.initStore()

	t.Run("default action (show) from initialized store", func(t *testing.T) {
		out, err := ts.run("foo/bar baz")
		assert.Error(t, err)
		assert.Equal(t, "\nError: failed to retrieve secret 'foo/bar': Entry is not in the password store\n", out)
	})

	t.Run("insert key", func(t *testing.T) {
		_, err := ts.runCmd([]string{ts.Binary, "insert", "foo/bar", "password"}, []byte("moar"))
		require.NoError(t, err)
	})

	t.Run("insert another key", func(t *testing.T) {
		_, err := ts.runCmd([]string{ts.Binary, "insert", "foo/bar", "baz"}, []byte("moar"))
		require.NoError(t, err)
	})

	t.Run("insert into the body", func(t *testing.T) {
		out, err := ts.runCmd([]string{ts.Binary, "insert", "-a", "foo/bar"}, []byte("body"))
		assert.NoError(t, err, out)
	})

	t.Run("show a key", func(t *testing.T) {
		out, err := ts.run("foo/bar baz")
		assert.NoError(t, err)
		assert.Equal(t, "moar", out)
	})

	t.Run("show the whole secret", func(t *testing.T) {
		out, err := ts.run("foo/bar")
		assert.NoError(t, err)
		assert.Equal(t, "Baz: moar\nPassword: moar\n\nbody", out)
	})
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

	t.Run("show secret from uninitialized store", func(t *testing.T) {
		_, err := ts.run("show foo/bar")
		assert.Error(t, err)
	})

	ts.initStore()

	t.Run("show non-existing secret", func(t *testing.T) {
		out, err := ts.run("foo/bar")
		assert.Error(t, err)
		assert.Equal(t, "\nError: failed to retrieve secret 'foo/bar': Entry is not in the password store\n", out)
	})

	t.Run("insert new secret", func(t *testing.T) {
		_, err := ts.runCmd([]string{ts.Binary, "insert", "foo/bar"}, []byte(testBody))
		assert.NoError(t, err)
	})

	t.Run("show newly inserted secret", func(t *testing.T) {
		_, err := ts.run("show foo/bar")
		assert.NoError(t, err)
	})
}
