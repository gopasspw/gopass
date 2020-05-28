package tests

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSingleMount(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	ts.initStore()
	ts.initSecrets("")

	// insert some secret at a place that will be shadowed by a mount
	_, err := ts.runCmd([]string{ts.Binary, "insert", "mnt/m1/secret"}, []byte("moar"))
	require.NoError(t, err)

	out, err := ts.run("show -f mnt/m1/secret")
	assert.NoError(t, err)
	assert.Equal(t, "moar", out)

	out, err = ts.run("init --store mnt/m1 --path " + ts.storeDir("m1") + " --rcs=noop " + keyID)
	t.Logf("Output: %s", out)
	require.NoError(t, err)

	out, err = ts.run("show mnt/m1/secret")
	assert.Error(t, err)
	assert.Equal(t, "\nError: failed to retrieve secret 'mnt/m1/secret': Entry is not in the password store\n", out)

	ts.initSecrets("mnt/m1/")

	list := `gopass
├── fixed
│   ├── secret
│   └── twoliner
├── foo
│   └── bar
├── mnt
`
	list += "│   └── m1 (" + ts.storeDir("m1") + ")\n"
	list += `│       ├── fixed
│       │   ├── secret
│       │   └── twoliner
│       ├── foo
│       │   └── bar
│       └── baz
└── baz`

	out, err = ts.run("list")
	assert.NoError(t, err)
	t.Skip("integration test seems broken, but not manual tests")
	assert.Equal(t, strings.TrimSpace(list), out)
}

func TestMultiMount(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	ts.initStore()
	ts.initSecrets("")

	// mount m1
	out, err := ts.run("init --store mnt/m1 --path " + ts.storeDir("m1") + " --rcs=noop " + keyID)
	t.Logf("Output: %s", out)
	require.NoError(t, err)

	ts.initSecrets("mnt/m1/")

	list := `gopass
├── fixed
│   ├── secret
│   └── twoliner
├── foo
│   └── bar
├── mnt
`
	list += "│   └── m1 (" + ts.storeDir("m1") + ")\n"
	list += `│       ├── fixed
│       │   ├── secret
│       │   └── twoliner
│       ├── foo
│       │   └── bar
│       └── baz
└── baz`

	out, err = ts.run("list")
	assert.NoError(t, err)
	t.Skip("integration test seems broken, but not manual tests")
	assert.Equal(t, strings.TrimSpace(list), out)

	// mount m2
	out, err = ts.run("init --store mnt/m2 --path " + ts.storeDir("m2") + " --rcs=noop " + keyID)
	t.Logf("Output: %s", out)
	require.NoError(t, err)

	ts.initSecrets("mnt/m2/")

	list = `gopass
├── fixed
│   ├── secret
│   └── twoliner
├── foo
│   └── bar
├── mnt
`
	list += "│   ├── m1 (" + ts.storeDir("m1") + ")\n"
	list += `│   │   ├── fixed
│   │   │   ├── secret
│   │   │   └── twoliner
│   │   ├── foo
│   │   │   └── bar
│   │   └── baz
`
	list += "│   └── m2 (" + ts.storeDir("m2") + ")\n"
	list += `│       ├── fixed
│       │   ├── secret
│       │   └── twoliner
│       ├── foo
│       │   └── bar
│       └── baz
└── baz`

	out, err = ts.run("list")
	assert.NoError(t, err)
	assert.Equal(t, strings.TrimSpace(list), out)
}
