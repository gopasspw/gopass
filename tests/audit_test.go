package tests

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAudit(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping test on windows.")
	}
	ts := newTester(t)
	defer ts.teardown()

	ts.initStore()
	ts.initSecrets("")

	out, err := ts.run("audit")
	assert.Error(t, err)
	assert.Contains(t, out, "No shared secrets found")
	assert.Contains(t, out, "Password is too short:")
	assert.Contains(t, out, "\t- fixed/secret")
}
