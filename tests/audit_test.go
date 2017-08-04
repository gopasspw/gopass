package tests

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAudit(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	ts.initStore()
	ts.initSecrets("")

	list := `Detected weak password for fixed/secret: Password is too short`
	out, err := ts.run("audit")
	assert.NoError(t, err)
	assert.Equal(t, strings.TrimSpace(list), out)
}
