package gitconfig

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gopasspw/gopass/internal/set"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	assert.Equal(t, []string{"bar"}, c.vars["aliases.subsection with spaces.foo"])
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

func TestUnsetSection(t *testing.T) {
	t.Parallel()

	in := `[core]
	showsafecontent = true
	readonly = true
[mounts]
	path = /tmp/foo
[foo]
	bar = baz
`
	c := ParseConfig(strings.NewReader(in))
	c.noWrites = true

	assert.NoError(t, c.Unset("core.readonly"))
	assert.Equal(t, `[core]
	showsafecontent = true
[mounts]
	path = /tmp/foo
[foo]
	bar = baz
`, c.raw.String())

	// should not exist
	assert.NoError(t, c.Unset("foo.bla"))

	// TODO: support remvoing sections
	t.Skip("removing sections is not supported, yet")

	assert.NoError(t, c.Unset("foo.bar"))
	assert.Equal(t, `[core]
	showsafecontent = false
	readonly = true
[mounts]
	readonly = true
	path = /tmp/foo
`, c.raw.String())
}

func TestNewFromMap(t *testing.T) {
	t.Parallel()

	tc := map[string]string{
		"core.foo":     "bar",
		"core.pager":   "false",
		"core.timeout": "10",
	}

	cfg := NewFromMap(tc)
	for k, v := range tc {
		assert.Equal(t, []string{v}, cfg.vars[k])
	}

	assert.True(t, cfg.IsSet("core.foo"))
	assert.False(t, cfg.IsSet("core.bar"))
	assert.NoError(t, cfg.Unset("core.foo"))
	assert.True(t, cfg.IsSet("core.foo"))
}

func TestLoadConfig(t *testing.T) {
	t.Parallel()

	td := t.TempDir()
	fn := filepath.Join(td, "config")
	assert.NoError(t, ioutil.WriteFile(fn, []byte(`[core]
	int = 7
	string = foo
	bar = false`), 0o600))

	cfg, err := LoadConfig(fn)
	require.NoError(t, err)

	assert.Equal(t, []string{"7"}, cfg.vars["core.int"])
	assert.Equal(t, []string{"foo"}, cfg.vars["core.string"])
	assert.Equal(t, []string{"false"}, cfg.vars["core.bar"])
}

func TestLoadFromEnv(t *testing.T) {
	tc := map[string]string{
		"core.foo":     "bar",
		"core.pager":   "false",
		"core.timeout": "10",
	}

	rand.Seed(time.Now().Unix())
	prefix := fmt.Sprintf("GPTEST%d", rand.Int31n(8192))

	i := 0
	for k, v := range tc {
		t.Setenv(fmt.Sprintf("%s_CONFIG_KEY_%d", prefix, i), k)
		t.Setenv(fmt.Sprintf("%s_CONFIG_VALUE_%d", prefix, i), v)
		i++
	}
	t.Setenv(fmt.Sprintf("%s_CONFIG_COUNT", prefix), strconv.Itoa(i))

	cfg := LoadConfigFromEnv(prefix)
	for k, v := range tc {
		assert.Equal(t, []string{v}, cfg.vars[k])
	}
}
