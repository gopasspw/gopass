package tests

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGrep(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	ts.initStore()

	out, err := ts.run("grep")
	assert.Error(t, err)
	assert.Equal(t, "\nError: Usage: "+filepath.Base(ts.Binary)+" grep arg\n", out)

	out, err = ts.run("grep BOOM")
	assert.Error(t, err)
	assert.Equal(t, "\nError: no matches found\n", out)

	ts.initSecrets("")

	out, err = ts.run("grep moar")
	assert.NoError(t, err)
	assert.Equal(t, "fixed/secret:\nmoar\n\nScanned 4 secrets. 1 matches, 0 errors\n", out)
}
