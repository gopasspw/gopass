package tests

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDelete(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	ts.initStore()

	out, err := ts.run("delete")
	assert.Error(t, err)
	assert.Equal(t, "\nError: Usage: "+filepath.Base(ts.Binary)+" rm name\n", out)

	out, err = ts.run("delete foobarbaz")
	assert.Error(t, err)
	assert.Contains(t, out, "does not exist", out)

	ts.initSecrets("")

	secrets := []string{"baz", "foo/bar"}
	for _, secret := range secrets {
		out, err = ts.run("delete -f " + secret)
		assert.NoError(t, err)
		assert.Equal(t, "", out)

		out, err = ts.run("delete -f " + secret)
		assert.Error(t, err)
		assert.Contains(t, out, "does not exist\n", out)
	}
}
