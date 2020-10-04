package tests

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInsert(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	ts.initStore()

	out, err := ts.run("insert")
	assert.Error(t, err)
	assert.Equal(t, "\nError: Usage: "+filepath.Base(ts.Binary)+" insert name\n", out)

	_, err = ts.runCmd([]string{ts.Binary, "insert", "some/secret"}, []byte("moar"))
	assert.NoError(t, err)

	_, err = ts.runCmd([]string{ts.Binary, "insert", "some/newsecret"}, []byte("and\nmoar"))
	assert.NoError(t, err)

	t.Run("Regression test for #1573 without actual pipes", func(t *testing.T) {
		out, err = ts.run("show -f some/secret")
		assert.NoError(t, err)
		assert.Equal(t, "moar", out)

		out, err = ts.run("show -f some/newsecret")
		assert.NoError(t, err)
		assert.Equal(t, "and\nmoar", out)

		out, err = ts.run("show -f some/secret")
		assert.NoError(t, err)
		assert.Equal(t, "moar", out)

		out, err = ts.run("show -f some/newsecret")
		assert.NoError(t, err)
		assert.Equal(t, "and\nmoar", out)
	})

	t.Run("Regression test for #1595", func(t *testing.T) {
		t.Skip("TODO")

		_, err = ts.runCmd([]string{ts.Binary, "insert", "some/other"}, []byte("nope"))
		assert.NoError(t, err)

		out, err = ts.run("insert some/other")
		assert.Error(t, err)
		assert.Equal(t, "\nError: not overwriting your current secret\n", out)

		out, err = ts.run("show -o some/other")
		assert.NoError(t, err)
		assert.Equal(t, "nope", out)

		out, err = ts.run("--yes insert some/other")
		assert.NoError(t, err)
		assert.Equal(t, "Warning: Password is empty or all whitespace", out)

		out, err = ts.run("insert -f some/other")
		assert.NoError(t, err)
		assert.Equal(t, "Warning: Password is empty or all whitespace", out)

		out, err = ts.run("show -o some/other")
		assert.Error(t, err)
		assert.Equal(t, "\nError: empty secret\n", out)

		_, err = ts.runCmd([]string{ts.Binary, "insert", "-f", "some/other"}, []byte("final"))
		assert.NoError(t, err)

		out, err = ts.run("show -o some/other")
		assert.NoError(t, err)
		assert.Equal(t, "final", out)

		// This is arguably not a good behaviour: it should not overwrite the password when we are only working on a key:value.
		out, err = ts.run("insert -f some/other test:inline")
		assert.NoError(t, err)
		assert.Equal(t, "", out)

		out, err = ts.run("show some/other test")
		assert.NoError(t, err)
		assert.Equal(t, "inline", out)

		out, err = ts.run("insert some/other test:inline2")
		assert.Error(t, err)
		assert.Equal(t, "\nError: not overwriting your current secret\n", out)

		out, err = ts.run("show some/other Test")
		assert.NoError(t, err)
		assert.Equal(t, "inline", out)

		out, err = ts.run("--yes insert some/other test:inline2")
		assert.NoError(t, err)
		assert.Equal(t, "", out)

		out, err = ts.run("show some/other Test")
		assert.NoError(t, err)
		assert.Equal(t, "inline2", out)
	})
}
