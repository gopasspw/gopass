package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	out, err := ts.run("init")
	assert.Error(t, err)
	assert.Equal(t, "\nError: no interaction without terminal\n", out)

	ts = newTester(t)
	defer ts.teardown()

	out, err = ts.run("init " + keyID)
	assert.NoError(t, err)
	assert.Contains(t, out, "Please enter a user name for password store git config [John Doe]:")
}
