package tests

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGrep(t *testing.T) { //nolint:paralleltest
	ts := newTester(t)
	defer ts.teardown()

	ts.initStore()

	out, err := ts.run("grep")
	assert.Error(t, err)
	assert.Equal(t, "\nError: Usage: "+filepath.Base(ts.Binary)+" grep arg\n", out)

	out, err = ts.run("grep BOOM")
	assert.NoError(t, err)
	assert.Contains(t, out, "Scanned 0 secrets. 0 matches, 0 errors")

	ts.initSecrets("")

	out, err = ts.run("grep moar")
	assert.NoError(t, err)
	assert.Contains(t, out, "fixed/secret matches")
}
