package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShow(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	_, err := ts.run("show")
	assert.Error(t, err)

	ts.initializeStore()

	out, err := ts.run("show")
	assert.Error(t, err)
	assert.Equal(t, "\nError: provide a secret name\n", out)

	out, err = ts.run("show foo")
	assert.Error(t, err)
	assert.Equal(t, "\nError: Entry is not in the password store\n", out)

	ts.initializeSecrets()

	_, err = ts.run("show foo")
	assert.NoError(t, err)
	_, err = ts.run("show -f foo")
	assert.NoError(t, err)
	_, err = ts.run("show foo -force")
	assert.NoError(t, err)

	out, err = ts.run("show fixed/secret")
	assert.Equal(t, "\nError: no safe content to display, you can force display with show -f\n", out)
	out, err = ts.run("show -f fixed/secret")
	assert.Equal(t, "moar", out)

	out, err = ts.run("show fixed/twoliner")
	assert.Equal(t, "more stuff", out)
	out, err = ts.run("show fixed/twoliner -f")
	assert.Equal(t, "and\nmore stuff", out)

	out, err = ts.run("config safecontent false")
	assert.NoError(t, err)
	out, err = ts.run("show fixed/twoliner")
	assert.Equal(t, "and\nmore stuff", out)
	out, err = ts.run("show fixed/secret")
	assert.Equal(t, "moar", out)
}
