package tests

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAudit(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	ts.initStore()
	ts.initSecrets("")

	out, err := ts.run("audit")
	assert.Error(t, err)
	assert.Contains(t, out, "No shared secrets found")
	assert.Contains(t, out, "Password is too short:")
	assert.Contains(t, out, "\t- "+filepath.Join("fixed", "secret"))
}
