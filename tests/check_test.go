package tests

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheck(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	ts.initStore()
	ts.initSecrets("")

	list := `Weak password for fixed/secret: it is too short`
	out, err := ts.run("check")
	assert.NoError(t, err)
	assert.Equal(t, strings.TrimSpace(list), out)
}
