package tests

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInsert(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping test on windows.")
	}
	ts := newTester(t)
	defer ts.teardown()

	ts.initStore()

	out, err := ts.run("insert")
	assert.Error(t, err)
	assert.Equal(t, "\nError: Usage: "+filepath.Base(ts.Binary)+" insert name\n", out)

	_, err = ts.runCmd([]string{ts.Binary, "insert", "some/secret"}, []byte("moar"))
	assert.NoError(t, err)
}
