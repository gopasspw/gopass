package tests

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMove(t *testing.T) { //nolint:paralleltest
	ts := newTester(t)
	defer ts.teardown()

	t.Run("move before init", func(t *testing.T) { //nolint:paralleltest
		_, err := ts.run("move")
		assert.Error(t, err)
	})

	// init store so it does exist, but empty so far
	ts.initStore()

	t.Run("move w/o args", func(t *testing.T) { //nolint:paralleltest
		out, err := ts.run("move")
		assert.Error(t, err)
		assert.Equal(t, "\nError: Usage: "+filepath.Base(ts.Binary)+" mv old-path new-path\n", out)
	})

	t.Run("move w/o destination", func(t *testing.T) { //nolint:paralleltest
		out, err := ts.run("move foo")
		assert.Error(t, err)
		assert.Equal(t, "\nError: Usage: "+filepath.Base(ts.Binary)+" mv old-path new-path\n", out)
	})

	t.Run("move non existing source", func(t *testing.T) { //nolint:paralleltest
		out, err := ts.run("move foo bar")
		assert.Error(t, err)
		assert.Equal(t, "\nError: source foo does not exist in source store : entry is not in the password store\n", out)
	})

	// populate store with secrets
	ts.initSecrets("")

	t.Run("move a secret", func(t *testing.T) { //nolint:paralleltest
		_, err := ts.run("move foo bar")
		assert.NoError(t, err)
	})

	t.Run("move existing secret from non-existing destination", func(t *testing.T) { //nolint:paralleltest
		out, _ := ts.run("move foo/bar foo/baz")
		assert.Equal(t, "\nError: source foo/bar does not exist in source store : entry is not in the password store\n", out)
	})

	t.Run("show source secret", func(t *testing.T) { //nolint:paralleltest
		_, err := ts.run("show -f bar/bar")
		assert.NoError(t, err)
	})

	t.Run("show secret", func(t *testing.T) { //nolint:paralleltest
		_, err := ts.run("show -f baz")
		assert.NoError(t, err)
	})
}
