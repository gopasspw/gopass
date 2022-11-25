package tests

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBaseConfig(t *testing.T) { //nolint:paralleltest
	ts := newTester(t)
	defer ts.teardown()

	out, err := ts.run("config")
	assert.NoError(t, err)

	wanted := `core.autosync = true
core.cliptimeout = 45
core.exportkeys = false
core.notifications = true
core.parsing = true
`
	wanted += "mounts.path = " + ts.storeDir("root")

	assert.Equal(t, wanted, out)

	invertables := []string{
		"core.autoimport",
		"core.showsafecontent",
		"core.parsing",
	}

	for _, invert := range invertables { //nolint:paralleltest
		t.Run("invert "+invert, func(t *testing.T) {
			out, err = ts.run("config " + invert + " false")
			assert.NoError(t, err)
			assert.Equal(t, "false", out)

			out, err = ts.run("config " + invert)
			assert.NoError(t, err)
			assert.Equal(t, "false", out)
		})
	}

	t.Run("cliptimeout", func(t *testing.T) {
		out, err = ts.run("config core.cliptimeout 120")
		assert.NoError(t, err)
		assert.Equal(t, "120", out)

		out, err = ts.run("config core.cliptimeout")
		assert.NoError(t, err)
		assert.Equal(t, "120", out)
	})
}

func TestMountConfig(t *testing.T) { //nolint:paralleltest
	ts := newTester(t)
	defer ts.teardown()

	// we add a mount:
	_, err := ts.run("init --store mnt/m1 --path " + ts.storeDir("m1") + " --storage=fs " + keyID)
	require.NoError(t, err)

	_, err = ts.run("config")
	assert.NoError(t, err)

	wanted := `core.autosync = true
core.cliptimeout = 45
core.exportkeys = false
core.notifications = true
core.parsing = true
`
	wanted += "mounts.mnt/m1.path = " + ts.storeDir("m1") + "\n"
	wanted += "mounts.path = " + ts.storeDir("root") + "\n"

	out, err := ts.run("config")
	assert.NoError(t, err)
	assert.Equal(t, strings.TrimSpace(wanted), out)
}
