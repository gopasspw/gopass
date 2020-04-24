package tests

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDelete(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping test on windows.")
	}
	ts := newTester(t)
	defer ts.teardown()

	ts.initStore()

	out, err := ts.run("delete")
	assert.Error(t, err)
	assert.Equal(t, "\nError: Usage: "+filepath.Base(ts.Binary)+" rm name\n", out)

	out, err = ts.run("delete foobarbaz")
	assert.Error(t, err)
	assert.Equal(t, "\nError: Can not delete 'foobarbaz': Entry is not in the password store\n", out)

	ts.initSecrets("")

	secrets := []string{"baz", "foo/bar"}
	for _, secret := range secrets {
		out, err = ts.run("delete -f " + secret)
		assert.NoError(t, err)
		assert.Equal(t, "", out)

		out, err = ts.run("delete -f " + secret)
		assert.Error(t, err)
		assert.Equal(t, "\nError: Can not delete '"+secret+"': Entry is not in the password store\n", out)
	}
}
