package gitconfig

import (
	"strings"
	"testing"

	"github.com/gopasspw/gopass/internal/set"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/maps"
)

func TestInsertOnce(t *testing.T) {
	t.Parallel()

	c := &Config{
		noWrites: true,
	}

	assert.NoError(t, c.insertValue("foo.bar", "baz"))
	assert.Equal(t, `[foo]
    bar = baz
`, c.raw.String())
}

func TestInsertMultiple(t *testing.T) {
	t.Parallel()

	c := &Config{
		noWrites: true,
	}

	updates := map[string]string{
		"foo.bar":     "baz",
		"core.show":   "true",
		"core.noshow": "true",
	}

	for _, k := range set.Sorted(maps.Keys(updates)) {
		v := updates[k]
		assert.NoError(t, c.insertValue(k, v))
	}

	assert.Equal(t, `[core]
    show = true
    noshow = true
[foo]
    bar = baz
`, c.raw.String())
}

func TestRewriteRaw(t *testing.T) {
	t.Parallel()

	in := `[core]
    showsafecontent = true
	parsing = false
	readonly = true
[mounts]
    path = /tmp/foo
`
	c := ParseConfig(strings.NewReader(in))
	c.noWrites = true

	updates := map[string]string{
		"foo.bar":              "baz",
		"mounts.readonly":      "true",
		"core.showsafecontent": "false",
		"core.parsing":         "true",
	}
	for _, k := range set.Sorted(maps.Keys(updates)) {
		v := updates[k]
		assert.NoError(t, c.Set(k, v))
	}

	assert.Equal(t, `[core]
    showsafecontent = false
    parsing = true
    readonly = true
[mounts]
    readonly = true
    path = /tmp/foo
[foo]
    bar = baz
`, c.raw.String())
}
