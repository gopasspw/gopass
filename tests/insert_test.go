package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInsert(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	ts.initializeStore()

	out, err := ts.run("insert")
	assert.Error(t, err)
	assert.Equal(t, "\nError: provide a secret name\n", out)

	out, err = ts.runCmd([]string{ts.Binary, "insert", "some/secret"}, []byte("moar"))
	assert.NoError(t, err)
}
