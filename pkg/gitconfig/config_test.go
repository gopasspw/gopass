package gitconfig

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/gopasspw/gopass/internal/set"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInsertOnce(t *testing.T) {
	t.Parallel()

	c := &Config{
		noWrites: true,
	}

	require.NoError(t, c.insertValue("foo.bar", "baz"))
	assert.Equal(t, `[foo]
	bar = baz
`, c.raw.String())
}

func TestInsertMultipleSameKey(t *testing.T) {
	t.Parallel()

	c := &Config{
		noWrites: true,
	}

	require.NoError(t, c.Set("foo.bar", "baz"))
	assert.Equal(t, `[foo]
	bar = baz
`, c.raw.String())
	require.NoError(t, c.Set("foo.bar", "zab"))
	assert.Equal(t, `[foo]
	bar = zab
`, c.raw.String())
}

func TestGetAll(t *testing.T) {
	t.Parallel()

	r := bytes.NewReader([]byte(`[core]
	foo = bar
	foo = zab
	foo = 123
`))

	c := ParseConfig(r)
	require.NotNil(t, c)
	vs, found := c.GetAll("core.foo")
	assert.True(t, found)
	assert.Equal(t, []string{"bar", "zab", "123"}, vs)

	require.NoError(t, c.Set("core.foo", "456"))
	vs, found = c.GetAll("core.foo")
	assert.True(t, found)
	assert.Equal(t, []string{"456", "zab", "123"}, vs)

	assert.Equal(t, `[core]
	foo = 456
	foo = zab
	foo = 123
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

	for _, k := range set.SortedKeys(updates) {
		v := updates[k]
		require.NoError(t, c.insertValue(k, v))
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
	autoimport = true
	readonly = true
[mounts]
	path = /tmp/foo
`
	c := ParseConfig(strings.NewReader(in))
	c.noWrites = true

	updates := map[string]string{
		"foo.bar":          "baz",
		"mounts.readonly":  "true",
		"show.safecontent": "false",
		"core.autoimport":  "false",
	}
	for _, k := range set.SortedKeys(updates) {
		v := updates[k]
		require.NoError(t, c.Set(k, v))
	}

	assert.Equal(t, `[core]
	autoimport = false
	readonly = true
[mounts]
	readonly = true
	path = /tmp/foo
[foo]
	bar = baz
[show]
	safecontent = false
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

	require.NoError(t, c.Unset("core.readonly"))
	assert.Equal(t, `[core]
	showsafecontent = true
[mounts]
	path = /tmp/foo
[foo]
	bar = baz
`, c.raw.String())

	// should not exist
	require.NoError(t, c.Unset("foo.bla"))

	// TODO: support remvoing sections
	t.Skip("removing sections is not supported, yet")

	require.NoError(t, c.Unset("foo.bar"))
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
	require.NoError(t, cfg.Unset("core.foo"))
	assert.True(t, cfg.IsSet("core.foo"))
}

func TestLoadConfig(t *testing.T) {
	t.Parallel()

	td := t.TempDir()
	fn := filepath.Join(td, "config")
	require.NoError(t, os.WriteFile(fn, []byte(`[core]
	int = 7
	string = foo
	bar = false`), 0o600))

	cfg, err := LoadConfig(fn)
	require.NoError(t, err)

	v, ok := cfg.Get("core.int")
	assert.True(t, ok)
	assert.Equal(t, "7", v)

	v, ok = cfg.Get("core.string")
	assert.True(t, ok)
	assert.Equal(t, "foo", v)

	v, ok = cfg.Get("core.bar")
	assert.True(t, ok)
	assert.Equal(t, "false", v)
}

func TestLoadConfigWithInclude(t *testing.T) {
	t.Parallel()

	td := t.TempDir()
	fn := filepath.Join(td, "config")

	tdBar := t.TempDir()
	fnBar := path.Join(tdBar, "bar.config")

	require.NoError(t, os.WriteFile(fn, []byte(`[core]
	int = 7
	string = foo
	bar = false
  [include]
    path = foo.config
    path = foo.config`), 0o600))
	fnFoo := filepath.Join(td, "foo.config")

	require.NoError(t, os.WriteFile(fnFoo, []byte(fmt.Sprintf(`[core]
	int = 8
  [include]
    path = config
    path = %s`, fnBar)), 0o600))
	assert.NoError(t, os.WriteFile(fnBar, []byte(`[core]
	int = 9`), 0o600))

	cfg, err := LoadConfig(fn)
	require.NoError(t, err)

	v, ok := cfg.Get("core.int")
	assert.True(t, ok)
	assert.Equal(t, "7", v)

	vs, ok := cfg.GetAll("core.int")
	assert.True(t, ok)
	assert.Equal(t, []string{"7", "8", "9"}, vs)

	v, ok = cfg.Get("core.string")
	assert.True(t, ok)
	assert.Equal(t, "foo", v)

	v, ok = cfg.Get("core.bar")
	assert.True(t, ok)
	assert.Equal(t, "false", v)
}

func TestLoadFromEnv(t *testing.T) {
	tc := map[string]string{
		"core.foo":     "bar",
		"core.pager":   "false",
		"core.timeout": "10",
	}

	prefix := fmt.Sprintf("GPTEST%d", rand.Int31n(8192))

	i := 0
	for k, v := range tc {
		t.Setenv(fmt.Sprintf("%s_KEY_%d", prefix, i), k)
		t.Setenv(fmt.Sprintf("%s_VALUE_%d", prefix, i), v)
		i++
	}
	t.Setenv(fmt.Sprintf("%s_COUNT", prefix), strconv.Itoa(i))

	cfg := LoadConfigFromEnv(prefix)
	for k, v := range tc {
		got, ok := cfg.Get(k)
		assert.True(t, ok)
		assert.Equal(t, v, got)
	}
}

func TestGetPathsForNestedConfig(t *testing.T) {
	t.Setenv("HOME", "/home/user")
	tc := map[string][3]string{
		"relative": {"/home/user/config", "foo.config", "/home/user/foo.config"},
		"~":        {"/home/user/config", "~/foo.config", "/home/user/foo.config"},
		"absolute": {"/home/user/config", "/home/user/foo.config", "/home/user/foo.config"},
	}

	for _, v := range tc {
		got := getPathsForNestedConfig([]string{v[1]}, v[0])
		assert.Equal(t, []string{v[2]}, got)
	}
}

func TestMergeConfigs(t *testing.T) {
	t.Parallel()

	baseConfig := Config{path: "/home/user/config", noWrites: true, readonly: true, raw: strings.Builder{}, vars: map[string][]string{"core.bar": {"1"}}}
	baseConfig.raw.WriteString("base")
	extensionConfig := Config{path: "/home/user/config.foo", noWrites: false, readonly: false, raw: strings.Builder{}, vars: map[string][]string{"core.bar": {"2"}}}
	extensionConfig.raw.WriteString("foo")

	mergedConfig := mergeConfigs(&baseConfig, &extensionConfig)
	assert.NotSame(t, &baseConfig, mergedConfig)
	assert.NotSame(t, &baseConfig.raw, &mergedConfig.raw)
	assert.NotSame(t, &extensionConfig, mergedConfig)
	assert.NotSame(t, &extensionConfig.raw, &mergedConfig.raw)
	assert.Equal(t, baseConfig.noWrites, mergedConfig.noWrites)
	assert.Equal(t, baseConfig.readonly, mergedConfig.readonly)
	assert.Equal(t, baseConfig.path, mergedConfig.path)
	assert.Equal(t, map[string][]string{"core.bar": {"1", "2"}}, mergedConfig.vars)
}

func TestMultiInclude(t *testing.T) {
	t.Parallel()

	td := t.TempDir()
	fn := filepath.Join(td, "config")
	require.NoError(t, os.WriteFile(fn, []byte(`[core]
	int = 7
	string = foo
	bar = false
  [include]
	path = foo.config`), 0o600))
	fnFoo := filepath.Join(td, "foo.config")
	require.NoError(t, os.WriteFile(fnFoo, []byte(`[core]
	int = 8
  [include]
	path = bar.config`), 0o600))
	fnBar := filepath.Join(td, "bar.config")
	require.NoError(t, os.WriteFile(fnBar, []byte(`[core]
	int = 9
  [include]
	path = baz.config`), 0o600))
	fnBaz := filepath.Join(td, "baz.config")
	require.NoError(t, os.WriteFile(fnBaz, []byte(`[core]
	int = 10`), 0o600))

	cfg, err := LoadConfig(fn)
	require.NoError(t, err)
	v, ok := cfg.Get("core.int")
	assert.True(t, ok)
	assert.Equal(t, "7", v)
	vs, ok := cfg.GetAll("core.int")
	assert.True(t, ok)
	assert.Equal(t, []string{"7", "8", "9", "10"}, vs)
	v, ok = cfg.Get("core.string")
	assert.True(t, ok)
	assert.Equal(t, "foo", v)
	v, ok = cfg.Get("core.bar")
	assert.True(t, ok)
	assert.Equal(t, "false", v)
}

func TestIncludeWrite(t *testing.T) {
	t.Parallel()

	td := t.TempDir()
	fn := filepath.Join(td, "config")
	require.NoError(t, os.WriteFile(fn, []byte(`[core]
	int = 7
	string = foo
	bar = false
  [include]
	path = foo.config`), 0o600))
	fnFoo := filepath.Join(td, "foo.config")
	require.NoError(t, os.WriteFile(fnFoo, []byte(`[core]
	int = 8`), 0o600))

	cfg, err := LoadConfig(fn)
	require.NoError(t, err)

	require.NoError(t, cfg.Set("core.int", "9"))
	require.NoError(t, cfg.Set("core.string", "bar"))
	require.NoError(t, cfg.Set("core.bar", "true"))

	cfg, err = LoadConfig(fn)
	require.NoError(t, err)
	v, ok := cfg.Get("core.int")
	assert.True(t, ok)
	assert.Equal(t, "9", v)
	vs, ok := cfg.GetAll("core.int")
	assert.True(t, ok)
	assert.Equal(t, []string{"9", "8"}, vs)
	v, ok = cfg.Get("core.string")
	assert.True(t, ok)
	assert.Equal(t, "bar", v)
	v, ok = cfg.Get("core.bar")
	assert.True(t, ok)
	assert.Equal(t, "true", v)

	// Check if the config was written correctly
	expected := `[core]
	int = 9
	string = bar
	bar = true
  [include]
	path = foo.config
`

	actual, err := os.ReadFile(fn)
	require.NoError(t, err)
	assert.Equal(t, expected, string(actual))
}
