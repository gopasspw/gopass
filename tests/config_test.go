package tests

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	out, err := ts.run("config")
	assert.NoError(t, err)
	assert.Contains(t, out, "askformore: false")
	assert.Contains(t, out, "autoimport: true")
	assert.Contains(t, out, "autosync: true")
	assert.Contains(t, out, "cliptimeout: 45")
	assert.Contains(t, out, "noconfirm: true")
	if runtime.GOOS == "darwin" {
		assert.Contains(t, out, "path: /var/folders")
	} else {
		assert.Contains(t, out, "path: /tmp")
	}
	assert.Contains(t, out, "safecontent: true")

	invertables := []string{
		"askformore",
		"autoimport",
		"autosync",
		"noconfirm",
		"safecontent",
	}

	for _, invert := range invertables {
		out, err = ts.run("config " + invert + " false")
		assert.NoError(t, err)
		assert.Equal(t, invert+": false", out)

		out, err = ts.run("config " + invert)
		assert.NoError(t, err)
		assert.Equal(t, invert+": false", out)
	}

	out, err = ts.run("config cliptimeout 120")
	assert.NoError(t, err)
	assert.Equal(t, "cliptimeout: 120", out)

	out, err = ts.run("config cliptimeout")
	assert.NoError(t, err)
	assert.Equal(t, "cliptimeout: 120", out)
}
