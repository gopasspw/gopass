package tests

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDelete(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	ts.initStore()

	out, err := ts.run("delete")
	require.Error(t, err)
	assert.Equal(t, "\nError: Usage: "+filepath.Base(ts.Binary)+" rm name\n", out)

	out, err = ts.run("delete foobarbaz")
	require.Error(t, err)
	assert.Contains(t, out, "does not exist", out)

	ts.initSecrets("")

	secrets := []string{"baz", "foo/bar"}
	for _, secret := range secrets {
		out, err = ts.run("delete -f " + secret)
		require.NoError(t, err)
		assert.Empty(t, out)

		out, err = ts.run("delete -f " + secret)
		require.Error(t, err)
		assert.Contains(t, out, "does not exist\n", out)
	}
}
