package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestYAMLAndSecret(t *testing.T) { //nolint:paralleltest
	ts := newTester(t)
	defer ts.teardown()

	t.Run("show key from uninitialized store", func(t *testing.T) { //nolint:paralleltest
		_, err := ts.run("show foo/bar baz")
		assert.Error(t, err)
	})

	ts.initStore()

	t.Run("default action (show) from initialized store", func(t *testing.T) { //nolint:paralleltest
		out, err := ts.run("foo/bar baz")
		assert.Error(t, err)
		assert.Contains(t, out, "entry is not in the password store")
	})

	t.Run("insert key", func(t *testing.T) { //nolint:paralleltest
		_, err := ts.runCmd([]string{ts.Binary, "insert", "foo/bar", "password"}, []byte("moar"))
		require.NoError(t, err)
	})

	t.Run("insert another key", func(t *testing.T) { //nolint:paralleltest
		_, err := ts.runCmd([]string{ts.Binary, "insert", "foo/bar", "baz"}, []byte("moar"))
		require.NoError(t, err)
	})

	t.Run("insert into the body", func(t *testing.T) { //nolint:paralleltest
		out, err := ts.runCmd([]string{ts.Binary, "insert", "-a", "foo/bar"}, []byte("body"))
		assert.NoError(t, err, out)
	})

	t.Run("show a key", func(t *testing.T) { //nolint:paralleltest
		out, err := ts.run("show foo/bar baz")
		assert.NoError(t, err)
		assert.Equal(t, "moar", out)
	})

	t.Run("show the whole secret", func(t *testing.T) { //nolint:paralleltest
		out, err := ts.run("show foo/bar")
		assert.NoError(t, err)
		assert.Equal(t, "password: moar\nbaz: moar\nbody", out)
	})
}

func TestInvalidYAML(t *testing.T) { //nolint:paralleltest
	testBody := `somepasswd
---
Test / test.com
username: myuser@test.com
password: someotherpasswd
url: http://www.test.com/`

	ts := newTester(t)
	defer ts.teardown()

	t.Run("show secret from uninitialized store", func(t *testing.T) { //nolint:paralleltest
		_, err := ts.run("show foo/bar")
		assert.Error(t, err)
	})

	ts.initStore()

	t.Run("show non-existing secret", func(t *testing.T) { //nolint:paralleltest
		out, err := ts.run("foo/bar")
		assert.Error(t, err)
		assert.Contains(t, out, "entry is not in the password store")
	})

	t.Run("insert new secret", func(t *testing.T) { //nolint:paralleltest
		_, err := ts.runCmd([]string{ts.Binary, "insert", "foo/bar"}, []byte(testBody))
		assert.NoError(t, err)
	})

	t.Run("show newly inserted secret", func(t *testing.T) { //nolint:paralleltest
		_, err := ts.run("show foo/bar")
		assert.NoError(t, err)
	})
}
