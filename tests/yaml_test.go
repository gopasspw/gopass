package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestYAMLAndSecret(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	_, err := ts.run("show foo/bar baz")
	assert.Error(t, err)

	ts.initStore()

	out, err := ts.run("foo/bar baz")
	assert.Error(t, err)
	assert.Equal(t, "\nError: failed to retrieve key 'baz' from 'foo/bar': no YAML document marker found\n", out)

	_, err = ts.runCmd([]string{ts.Binary, "insert", "foo/bar"}, []byte("moar"))
	assert.NoError(t, err)

	_, err = ts.runCmd([]string{ts.Binary, "insert", "foo/bar", "baz"}, []byte("moar"))
	assert.NoError(t, err)

	out, err = ts.run("foo/bar baz")
	assert.NoError(t, err)
	assert.Equal(t, "moar", out)
}
