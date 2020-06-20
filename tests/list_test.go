package tests

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestList(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	ts.initStore()

	out, err := ts.run("list")
	assert.NoError(t, err)
	assert.Equal(t, "gopass", out)

	out, err = ts.run("ls")
	assert.NoError(t, err)
	assert.Equal(t, "gopass", out)

	ts.initSecrets("")

	list := `
gopass
├── baz
├── fixed
│   ├── secret
│   └── twoliner
└── foo
    └── bar
`
	out, err = ts.run("list")
	assert.NoError(t, err)
	assert.Equal(t, strings.TrimSpace(list), out)

	list = `
foo
└── bar
`
	out, err = ts.run("list foo")
	assert.NoError(t, err)
	assert.Equal(t, strings.TrimSpace(list), out)

	list = `fixed
foo
`
	out, err = ts.run("list --folders")
	assert.NoError(t, err)
	assert.Equal(t, strings.TrimSpace(list), out)

}
