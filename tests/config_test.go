package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	out, err := ts.run("config")
	assert.NoError(t, err)
	assert.Contains(t, out, "autoimport: true")
	assert.Contains(t, out, "cliptimeout: 45")
	assert.Contains(t, out, "confirm: false")
	assert.Contains(t, out, "path: ")
	assert.Contains(t, out, "safecontent: false")

	invertables := []string{
		"autoimport",
		"confirm",
		"safecontent",
	}

	for _, invert := range invertables {
		t.Run("invert "+invert, func(t *testing.T) {
			out, err = ts.run("config " + invert + " false")
			assert.NoError(t, err)
			assert.Equal(t, invert+": false", out)

			out, err = ts.run("config " + invert)
			assert.NoError(t, err)
			assert.Equal(t, invert+": false", out)
		})
	}

	t.Run("cliptimeout", func(t *testing.T) {
		out, err = ts.run("config cliptimeout 120")
		assert.NoError(t, err)
		assert.Equal(t, "cliptimeout: 120", out)

		out, err = ts.run("config cliptimeout")
		assert.NoError(t, err)
		assert.Equal(t, "cliptimeout: 120", out)
	})
}
