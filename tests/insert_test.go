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
		assert.Equal(t, out, "Password: moar")

		out, err = ts.run("show -f some/newsecret")
		assert.NoError(t, err)
		assert.Equal(t, out, "Password: and\n\nmoar")

		_, err := ts.run("config mime false")
		assert.NoError(t, err)

		out, err = ts.run("show -f some/secret")
		assert.NoError(t, err)
		assert.Equal(t, out, "moar")

		out, err = ts.run("show -f some/newsecret")
		assert.NoError(t, err)
		assert.Equal(t, out, "and\n\nmoar")
	})
}
