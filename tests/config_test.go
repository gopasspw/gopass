package tests

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBaseConfig(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	out, err := ts.run("config")
	require.NoError(t, err)

	wanted := `core.autopush = true
core.autosync = true
core.cliptimeout = 45
core.exportkeys = false
core.notifications = true
core.follow-references = false
`
	wanted += "mounts.path = " + ts.storeDir("root") + "\n" +
		"pwgen.xkcd-lang = en"

	assert.Equal(t, wanted, out)

	invertables := []string{
		"core.autoimport",
		"show.safecontent",
	}

	for _, invert := range invertables {
		t.Run("invert "+invert, func(t *testing.T) {
			arg := "config " + invert + " false"
			out, err = ts.run(arg)
			require.NoError(t, err, "Running gopass "+arg)
			assert.Equal(t, "false", out, "Output of gopass "+arg)

			arg = "config " + invert
			out, err = ts.run(arg)
			require.NoError(t, err, "Running gopass "+arg)
			assert.Equal(t, "false", out, "Output of gopass "+arg)
		})
	}

	t.Run("cliptimeout", func(t *testing.T) {
		out, err = ts.run("config core.cliptimeout 120")
		require.NoError(t, err)
		assert.Equal(t, "120", out)

		out, err = ts.run("config core.cliptimeout")
		require.NoError(t, err)
		assert.Equal(t, "120", out)
	})
}

func TestMountConfig(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	// we add a mount:
	_, err := ts.run("init --store mnt/m1 --path " + ts.storeDir("m1") + " --storage=fs " + keyID)
	require.NoError(t, err)

	_, err = ts.run("config")
	require.NoError(t, err)

	wanted := `core.autopush = true
core.autosync = true
core.cliptimeout = 45
core.exportkeys = false
core.follow-references = false
core.notifications = true
`
	wanted += "mounts.mnt/m1.path = " + ts.storeDir("m1") + "\n"
	wanted += "mounts.path = " + ts.storeDir("root") + "\n"
	wanted += "pwgen.xkcd-lang = en\n"
	wanted += "recipients.mnt/m1.hash = 9a4c4b1e0eb9ade2e692ff948f43d9668145eca3df88ffff67e0e21426252907\n"

	out, err := ts.run("config")
	require.NoError(t, err)
	assert.Equal(t, strings.TrimSpace(wanted), out)
}
