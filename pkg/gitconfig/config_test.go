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

func TestSubsection(t *testing.T) {
	t.Parallel()

	in := `[core]
	showsafecontent = true
	readonly = true
[aliases "subsection with spaces"]
	foo = bar
`
	c := ParseConfig(strings.NewReader(in))
	c.noWrites = true

	assert.Equal(t, c.vars["aliases.subsection with spaces.foo"], "bar")
}

func TestParseSection(t *testing.T) {
	t.Parallel()

	for in, out := range map[string]struct {
		section string
		subs    string
		skip    bool
	}{
		`[aliases]`: {
			section: "aliases",
		},
		`[aliases "subsection"]`: {
			section: "aliases",
			subs:    "subsection",
		},
		`[aliases "subsection with spaces"]`: {
			section: "aliases",
			subs:    "subsection with spaces",
		},
		`[aliases "subsection with spaces and \" \t \0 escapes"]`: {
			section: "aliases",
			subs:    `subsection with spaces and " t 0 escapes`,
		},
	} {
		section, subsection, skip := parseSectionHeader(in)
		assert.Equal(t, out.section, section, in)
		assert.Equal(t, out.subs, subsection, in)
		assert.Equal(t, out.skip, skip, in)
	}
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
	}
	for _, k := range set.Sorted(maps.Keys(updates)) {
		v := updates[k]
		assert.NoError(t, c.Set(k, v))
	}

	assert.Equal(t, `[core]
	showsafecontent = false
	readonly = true
[mounts]
	readonly = true
	path = /tmp/foo
[foo]
	bar = baz
`, c.raw.String())
}
