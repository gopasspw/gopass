package tests

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestList(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	ts.initializeStore()

	out, err := ts.run("")
	assert.NoError(t, err)
	assert.Equal(t, "gopass", out)

	out, err = ts.run("list")
	assert.NoError(t, err)
	assert.Equal(t, "gopass", out)

	out, err = ts.run("ls")
	assert.NoError(t, err)
	assert.Equal(t, "gopass", out)

	ts.initializeSecrets()

	list := `
gopass
├── fixed
│   ├── secret
│   └── twoliner
├── foo
│   └── bar
└── baz
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
}
