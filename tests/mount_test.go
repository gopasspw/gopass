package tests

import (
	"fmt"
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

	out, err := ts.run("init --store mnt/m1 --path " + ts.storeDir("m1") + " --storage=fs " + keyID)
	t.Logf("Output: %s", out)
	require.NoError(t, err)

	out, err = ts.run("mounts")
	require.NoError(t, err)

	want := "gopass (" + ts.storeDir("root") + ")\n"
	want += "└── mnt/\n    └── m1 (" + ts.storeDir("m1") + ")"
	assert.Equal(t, strings.TrimSpace(want), out)

	out, err = ts.run("show mnt/m1/secret")
	require.Error(t, err)
	assert.Contains(t, out, "entry is not in the password store")

	ts.initSecrets("mnt/m1/")

	list := `
gopass
├── baz
├── fixed/
│   ├── secret
│   └── twoliner
├── foo/
│   └── bar
└── mnt/
    └── m1 (%s)
        ├── baz
        ├── fixed/
        │   ├── secret
        │   └── twoliner
        └── foo/
            └── bar
`
	list = fmt.Sprintf(list, ts.storeDir("m1"))

	out, err = ts.run("list")
	require.NoError(t, err)
	assert.Equal(t, strings.TrimSpace(list), out)
}

func TestMountShadowing(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	ts.initStore()
	ts.initSecrets("")

	// insert some secret at a place that will be shadowed by a mount
	_, err := ts.runCmd([]string{ts.Binary, "insert", "mnt/m1/secret"}, []byte("moar"))
	require.NoError(t, err)

	out, err := ts.run("show -u mnt/m1/secret")
	require.NoError(t, err)
	assert.Equal(t, "moar", out)

	out, err = ts.run("init --store mnt/m1 --path " + ts.storeDir("m1") + " --storage=fs " + keyID)
	t.Logf("Output: %s", out)
	require.NoError(t, err)

	// check the mount is there
	out, err = ts.run("mounts")
	require.NoError(t, err)

	want := "gopass (" + ts.storeDir("root") + ")\n"
	want += "└── mnt/\n    └── m1 (" + ts.storeDir("m1") + ")"
	assert.Equal(t, strings.TrimSpace(want), out)

	// check that the mount is not containing our shadowed secret
	out, err = ts.run("show -u mnt/m1/secret")
	require.Error(t, err)
	assert.Contains(t, out, "entry is not in the password store")

	// insert some secret at the place that is shadowed by the mount
	_, err = ts.runCmd([]string{ts.Binary, "insert", "mnt/m1/secret"}, []byte("food"))
	require.NoError(t, err)

	// check that the mount is containing our new secret shadowing the old one
	out, err = ts.run("show -u mnt/m1/secret")
	require.NoError(t, err)
	assert.Equal(t, "food", out)

	// add more secrets
	ts.initSecrets("mnt/m1/")

	// check that the mount is listed
	list := `
gopass
├── baz
├── fixed/
│   ├── secret
│   └── twoliner
├── foo/
│   └── bar
└── mnt/
    └── m1 (%s)
        ├── baz
        ├── fixed/
        │   ├── secret
        │   └── twoliner
        ├── foo/
        │   └── bar
        └── secret
`
	list = fmt.Sprintf(list, ts.storeDir("m1"))

	out, err = ts.run("list")
	require.NoError(t, err)
	assert.Equal(t, strings.TrimSpace(list), out)

	// check that unmounting works:
	_, err = ts.run("mounts rm mnt/m1")
	require.NoError(t, err)

	list = `
gopass
├── baz
├── fixed/
│   ├── secret
│   └── twoliner
├── foo/
│   └── bar
└── mnt/
    └── m1/
        └── secret
`

	out, err = ts.run("list")
	require.NoError(t, err)
	assert.Equal(t, strings.TrimSpace(list), out)

	out, err = ts.run("show -o mnt/m1/secret")
	require.NoError(t, err)
	assert.Equal(t, "moar", out)
}

func TestMultiMount(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	ts.initStore()
	ts.initSecrets("")

	// mount m1
	out, err := ts.run("init --store mnt/m1 --path " + ts.storeDir("m1") + " --storage=fs " + keyID)
	t.Logf("Output: %s", out)
	require.NoError(t, err)

	ts.initSecrets("mnt/m1/")

	list := `
gopass
├── baz
├── fixed/
│   ├── secret
│   └── twoliner
├── foo/
│   └── bar
└── mnt/
    └── m1 (%s)
        ├── baz
        ├── fixed/
        │   ├── secret
        │   └── twoliner
        └── foo/
            └── bar
`
	list = fmt.Sprintf(list, ts.storeDir("m1"))

	out, err = ts.run("list")
	require.NoError(t, err)
	assert.Equal(t, strings.TrimSpace(list), out)

	// mount m2
	out, err = ts.run("init --store mnt/m2 --path " + ts.storeDir("m2") + " --storage=fs " + keyID)
	t.Logf("Output: %s", out)
	require.NoError(t, err)

	ts.initSecrets("mnt/m2/")

	list = `
gopass
├── baz
├── fixed/
│   ├── secret
│   └── twoliner
├── foo/
│   └── bar
└── mnt/
    ├── m1 (%s)
    │   ├── baz
    │   ├── fixed/
    │   │   ├── secret
    │   │   └── twoliner
    │   └── foo/
    │       └── bar
    └── m2 (%s)
        ├── baz
        ├── fixed/
        │   ├── secret
        │   └── twoliner
        └── foo/
            └── bar
`
	list = fmt.Sprintf(list, ts.storeDir("m1"), ts.storeDir("m2"))

	out, err = ts.run("list")
	require.NoError(t, err)
	assert.Equal(t, strings.TrimSpace(list), out)
}
