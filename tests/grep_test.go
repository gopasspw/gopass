package tests

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGrep(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping test on windows.")
	}
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
	assert.Error(t, err)
	assert.Equal(t, "fixed/secret:\nmoar\n\nError: no matches found\n", out)
}
