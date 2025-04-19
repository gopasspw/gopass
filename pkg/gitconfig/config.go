package gitconfig

import (
	"bufio"
	"fmt"
	"io"
	"maps"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gopasspw/gopass/pkg/debug"
)

var keyValueTpl = "\t%s = %s%s"

// Config is a single parsed config file. It contains a reference of the input file, if any.
// It can only be populated only by reading the environment variables.
type Config struct {
	path     string
	readonly bool // do not allow modifying values (even in memory)
	noWrites bool // do not persist changes to disk (e.g. for tests)
	raw      strings.Builder
	vars     map[string][]string
}

// IsEmpty returns true if the config is empty (typically a newly initialized config, but still unused).
// Since gitconfig.New() already sets the global path to the globalConfigFile() one, we cannot rely on
// the path being set to checki this. We need to check the  raw length to be sure it wasn't just
// the default empty config struct.
func (c *Config) IsEmpty() bool {
	if c == nil || c.vars == nil {
		return true
	}

	if c.raw.Len() > 0 {
		return false
	}

	return true
}

// Unset deletes a key.
func (c *Config) Unset(key string) error {
	if c.readonly {
		return nil
	}

	_, present := c.vars[key]
	if !present {
		return nil
	}

	delete(c.vars, key)

	return c.rewriteRaw(key, "", func(fKey, key, value, comment, _ string) (string, bool) {
		return "", true
	})
}

// Get returns the first value of the key.
func (c *Config) Get(key string) (string, bool) {
	vs, found := c.vars[key]
	if !found || len(vs) < 1 {
		return "", false
	}

	return vs[0], true
}

// GetAll returns all values of the key.
func (c *Config) GetAll(key string) ([]string, bool) {
	vs, found := c.vars[key]
	if !found {
		return nil, false
	}

	return vs, true
}

// IsSet returns true if the key was set in this config.
func (c *Config) IsSet(key string) bool {
	_, present := c.vars[key]

	return present
}

// Set updates or adds a key in the config. If possible it will also update the underlying
// config file on disk.
func (c *Config) Set(key, value string) error {
	section, _, subkey := splitKey(key)
	if section == "" || subkey == "" {
		return fmt.Errorf("invalid key: %s", key)
	}

	// can't set env vars
	if c.readonly {
		debug.Log("can not write to a readonly config")

		return nil
	}

	if c.vars == nil {
		c.vars = make(map[string][]string, 16)
	}

	// already present at the same value, no need to rewrite the config
	if vs, found := c.vars[key]; found {
		for _, v := range vs {
			if v == value {
				debug.V(1).Log("key %q with value %q already present. Not re-writing.", key, value)

				return nil
			}
		}
	}

	vs, present := c.vars[key]
	if vs == nil {
		vs = make([]string, 1)
	}
	vs[0] = value
	c.vars[key] = vs

	debug.V(3).Log("set %q to %q", key, value)

	// a new key, insert it into an existing section, if any
	if !present {
		debug.V(3).Log("inserting value")

		return c.insertValue(key, value)
	}

	debug.V(3).Log("updating value")

	var updated bool

	return c.rewriteRaw(key, value, func(fKey, sKey, value, comment, line string) (string, bool) {
		if updated {
			return line, false
		}
		updated = true

		return fmt.Sprintf(keyValueTpl, sKey, value, comment), false
	})
}

func (c *Config) insertValue(key, value string) error {
	debug.V(3).Log("input (%s: %s): \n--------------\n%s\n--------------\n", key, value, strings.Join(strings.Split("- "+c.raw.String(), "\n"), "\n- "))

	wSection, wSubsection, wKey := splitKey(key)

	s := bufio.NewScanner(strings.NewReader(c.raw.String()))

	lines := make([]string, 0, 128)
	var section string
	var subsection string
	var written bool
	for s.Scan() {
		line := s.Text()

		lines = append(lines, line)

		if written {
			continue
		}
		if strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, "[") {
			s, subs, skip := parseSectionHeader(line)
			if skip {
				continue
			}
			section = s
			subsection = subs
		}

		if section != wSection {
			continue
		}
		if subsection != wSubsection {
			continue
		}

		lines = append(lines, fmt.Sprintf(keyValueTpl, wKey, value, ""))
		written = true
	}

	// not added to an existing section, so add it at the end
	if !written {
		sect := fmt.Sprintf("[%s]", wSection)
		if wSubsection != "" {
			sect = fmt.Sprintf("[%s \"%s\"]", wSection, wSubsection)
		}
		lines = append(lines, sect)
		lines = append(lines, fmt.Sprintf(keyValueTpl, wKey, value, ""))
	}

	c.raw = strings.Builder{}
	c.raw.WriteString(strings.Join(lines, "\n"))
	c.raw.WriteString("\n")

	debug.V(3).Log("output: \n--------------\n%s\n--------------\n", strings.Join(strings.Split("+ "+c.raw.String(), "\n"), "\n+ "))

	return c.flushRaw()
}

func parseSectionHeader(line string) (section, subsection string, skip bool) { //nolint:nonamedreturns
	line = strings.Trim(line, "[]")
	if line == "" {
		return "", "", true
	}
	wsp := strings.Index(line, " ")
	if wsp < 0 {
		return line, "", false
	}

	section = line[:wsp]
	subsection = line[wsp+1:]
	subsection = strings.ReplaceAll(subsection, "\\", "")
	subsection = strings.TrimPrefix(subsection, "\"")
	subsection = strings.TrimSuffix(subsection, "\"")

	return section, subsection, false
}

// rewriteRaw is used to rewrite the raw config copy. It is used for set and unset operations
// with different callbacks each.
func (c *Config) rewriteRaw(key, value string, cb parseFunc) error {
	debug.V(3).Log("input (%s: %s): \n--------------\n%s\n--------------\n", key, value, strings.Join(strings.Split("- "+c.raw.String(), "\n"), "\n- "))

	lines := parseConfig(strings.NewReader(c.raw.String()), key, value, cb)

	c.raw = strings.Builder{}
	c.raw.WriteString(strings.Join(lines, "\n"))
	c.raw.WriteString("\n")

	debug.V(3).Log("output: \n--------------\n%s\n--------------\n", strings.Join(strings.Split("+ "+c.raw.String(), "\n"), "\n+ "))

	return c.flushRaw()
}

func (c *Config) flushRaw() error {
	if c.noWrites || c.path == "" {
		debug.V(3).Log("not writing changes to disk (noWrites %t, path %q)", c.noWrites, c.path)

		return nil
	}

	if err := os.MkdirAll(filepath.Dir(c.path), 0o700); err != nil {
		return err
	}

	debug.V(3).Log("writing config to %s: \n--------------\n%s\n--------------", c.path, c.raw.String())

	if err := os.WriteFile(c.path, []byte(c.raw.String()), 0o600); err != nil {
		return fmt.Errorf("failed to write config to %s: %w", c.path, err)
	}

	debug.V(1).Log("wrote config to %s", c.path)

	return nil
}

type parseFunc func(fqkn, skn, value, comment, fullLine string) (newLine string, skipLine bool)

// parseConfig implements a simple parser for the gitconfig subset we support.
// The idea is to save all lines unaltered so we can reproduce the config
// almost exactly. Then we skip comments and extract section and subsection
// header. The next steps depend on the mode. Either we want to extract the
// values when loading (key and value empty, parseFunc adds the key-value pairs
// to the vars map), update a key (key is the target key, value the new value)
// or delete a key (parseFunc returns skip).
func parseConfig(in io.Reader, key, value string, cb parseFunc) []string {
	wSection, wSubsection, wKey := splitKey(key)

	s := bufio.NewScanner(in)

	lines := make([]string, 0, 128)
	var section string
	var subsection string
	for s.Scan() {
		fullLine := s.Text()

		lines = append(lines, fullLine)

		line := strings.TrimSpace(fullLine)
		// Handle full-line comments
		if strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, ";") {
			continue
		}

		// Handle section headers
		if strings.HasPrefix(line, "[") {
			s, subs, skip := parseSectionHeader(line)
			if skip {
				continue
			}
			section = s
			subsection = subs
		}

		if key != "" && (section != wSection && subsection != wSubsection) {
			continue
		}

		// TODO(gitconfig) This will skip over valid entries like this one:
		// [core]
		//  sslVerify
		// These are odd but we should still support them.
		// Check https://git-scm.com/docs/git-config#_syntax for more details.
		k, v, found := strings.Cut(line, "=")
		if !found {
			debug.V(3).Log("no valid KV-pair on line: %q", line)

			continue
		}
		// Remove whitespace from key and value that might be around the '='
		k = strings.TrimRight(k, " ")
		v = strings.TrimLeft(v, " ")

		fKey := section + "."
		if subsection != "" {
			fKey += subsection + "."
		}
		fKey += k
		if key == "" {
			wKey = k
		}

		oValue := v
		comment := ""

		// Handle inline comments
		if strings.ContainsAny(oValue, "#;") {
			comment = " " + oValue[strings.IndexAny(oValue, "#;"):]
			oValue = oValue[:strings.IndexAny(oValue, "#;")]
			oValue = strings.TrimSpace(oValue)
		}

		if key != "" && (key != fKey) {
			continue
		}
		if key != "" {
			oValue = value
		}

		newLine, skip := cb(fKey, wKey, oValue, comment, fullLine)
		if skip {
			// remove the last line
			lines = lines[:len(lines)-1]

			continue
		}
		lines[len(lines)-1] = newLine
	}

	return lines
}

// NewFromMap allows creating a new preset config from a map.
func NewFromMap(data map[string]string) *Config {
	c := &Config{
		readonly: true,
		vars:     make(map[string][]string, len(data)),
	}

	for k, v := range data {
		c.vars[k] = []string{v}
	}

	return c
}

// LoadConfig tries to load a gitconfig from the given path.
func LoadConfig(fn string) (*Config, error) {
	return loadConfigs(fn, "")
}

// LoadConfigWithWorkdir tries to load a gitconfig from the given path and
// a workdir. The workdir is used to resolve relative paths in the config.
func LoadConfigWithWorkdir(fn, workdir string) (*Config, error) {
	return loadConfigs(fn, workdir)
}

func getEffectiveIncludes(c *Config, workdir string) ([]string, bool) {
	includePaths, includeExists := c.GetAll("include.path")

	if cIncludes := getConditionalIncludes(c, workdir); len(cIncludes) > 0 {
		includePaths = append(includePaths, cIncludes...)
		includeExists = true
	}

	return includePaths, includeExists
}

func getConditionalIncludes(c *Config, workdir string) []string {
	candidates := []string{}
	for k := range c.vars {
		// must have the form includeIf.<condition>.path
		// e.g. includeIf."gitdir:/path/to/group/".path
		// see https://git-scm.com/docs/git-config#_conditional_includes
		if !strings.HasPrefix(k, "includeIf.") || !strings.HasSuffix(k, ".path") {
			continue
		}
		candidates = append(candidates, k)
	}

	out := make([]string, 0, len(candidates))
	for _, k := range filterCandidates(candidates, workdir) {
		path, found := c.GetAll(k)
		if !found {
			continue
		}
		out = append(out, path...)
	}

	return out
}

// filterCandidates filters the candidates for include paths.
// Currently only the gitdir condition is supported.
// Others might be added in the future.
func filterCandidates(candidates []string, workdir string) []string {
	out := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		sec, subsec, key := splitKey(candidate)
		if sec != "includeIf" || subsec == "" || key != "path" {
			debug.V(3).Log("skipping invalid include candidate %q", candidate)

			continue
		}

		// We only support gitdir: for now.
		if !strings.HasPrefix(subsec, "gitdir:") {
			debug.V(3).Log("skipping unsupported include candidate %q", candidate)

			continue
		}

		p := strings.Split(subsec, ":")
		// We have checked that there is a colon above.
		dir := p[1]

		// Either it is a full match or a prefix match.
		if strings.TrimSuffix(workdir, "/") != strings.TrimSuffix(dir, "/") && !prefixMatch(dir, workdir) {
			debug.V(3).Log("skipping include candidate %q, no exact match for workdir: %q == dir: %q and no prefix match for dir: %q, workdir: %q", candidate, workdir, dir, dir, workdir)

			continue
		}

		// We have a match, so we can add the path to the list.
		out = append(out, candidate)
	}

	return out
}

func prefixMatch(path, prefix string) bool {
	if !strings.HasSuffix(prefix, "/") {
		return false
	}

	return strings.HasPrefix(path, prefix)
}

func loadConfigs(fn, workdir string) (*Config, error) {
	c, err := loadConfig(fn)
	if err != nil {
		return nil, err
	}
	c.path = fn

	loadedConfigs := map[string]struct{}{
		fn: {},
	}
	configsToLoad := []string{}

	includePaths, includeExists := getEffectiveIncludes(c, workdir)
	if includeExists {
		configsToLoad = append(configsToLoad, getPathsForNestedConfig(includePaths, c.path)...)
	}

	// load all nested configs
	// this is using a slice as a stack because when we load a config
	// it may include other configs
	// so we need to load them in the order they are found.
	for len(configsToLoad) > 0 {
		head := configsToLoad[0]
		configsToLoad = configsToLoad[1:]

		// check if we already loaded this config
		// this is needed to avoid infinite loops when loading nested configs
		_, ignore := loadedConfigs[head]
		if ignore {
			debug.V(3).Log("skipping already loaded config %q", head)

			continue
		}

		nc, err := loadConfig(head)
		if err != nil {
			return nil, err
		}

		c = mergeConfigs(c, nc)
		loadedConfigs[head] = struct{}{}

		includePaths, includeExists := getEffectiveIncludes(nc, workdir)
		if includeExists {
			configsToLoad = append(configsToLoad, getPathsForNestedConfig(includePaths, nc.path)...)
		}
	}

	return c, nil
}

func loadConfig(fn string) (*Config, error) {
	fh, err := os.Open(fn)
	if err != nil {
		return nil, err
	}
	defer fh.Close() //nolint:errcheck

	c := ParseConfig(fh)
	c.path = fn

	return c, nil
}

// mergeConfigs merge two configs, using first config as a base config extending it with vars, raw fields from the latter.
func mergeConfigs(base *Config, extension *Config) *Config {
	newConfig := Config{path: base.path, readonly: base.readonly, noWrites: base.noWrites, raw: strings.Builder{}, vars: map[string][]string{}}
	newConfig.raw.WriteString(base.raw.String())
	// Note: We can not append the included config raw to the base config raw, because it will
	// write the included config to the base config file when we write the base config.

	// populate the new config with the base config
	maps.Copy(newConfig.vars, base.vars)

	for k, v := range extension.vars {
		_, existing := newConfig.vars[k]
		if !existing {
			newConfig.vars[k] = []string{}
		}
		newConfig.vars[k] = append(newConfig.vars[k], v...)
	}

	return &newConfig
}

// getPathsForNestedConfig tries to convert paths of nested configs ('/absolute', '~/from/home', 'relative/to/base') to absolute paths.
func getPathsForNestedConfig(nestedConfigs []string, baseConfig string) []string {
	absolutePaths := []string{}
	for _, nc := range nestedConfigs {
		if path.IsAbs(nc) {
			absolutePaths = append(absolutePaths, nc)

			continue
		}
		if strings.HasPrefix(nc, "~/") {
			home, exists := os.LookupEnv("HOME")
			if !exists {
				// cannot resolve home directory
				debug.V(3).Log("cannot resolve home directory, skipping %q", nc)

				continue
			}
			absolutePaths = append(absolutePaths, path.Join(home, strings.Replace(nc, "~/", "", 1)))

			continue
		}
		absolutePaths = append(absolutePaths, path.Clean(path.Join(path.Dir(baseConfig), nc)))
	}

	return absolutePaths
}

// ParseConfig will try to parse a gitconfig from the given io.Reader. It never fails.
// Invalid configs will be silently rejected.
func ParseConfig(r io.Reader) *Config {
	c := &Config{
		vars: make(map[string][]string, 42),
	}

	lines := parseConfig(r, "", "", func(fk, k, v, comment, _ string) (string, bool) {
		c.vars[fk] = append(c.vars[fk], v)

		return fmt.Sprintf(keyValueTpl, k, v, comment), false
	})

	c.raw.WriteString(strings.Join(lines, "\n"))
	c.raw.WriteString("\n")

	debug.V(3).Log("processed config: %s\nvars: %+v", c.raw.String(), c.vars)

	return c
}

// LoadConfigFromEnv will try to parse an overlay config from the environment variables.
// If no environment variables are set the resulting config will be valid but empty.
// Either way it will not be writeable.
func LoadConfigFromEnv(envPrefix string) *Config {
	c := &Config{
		noWrites: true,
	}

	count, err := strconv.Atoi(os.Getenv(envPrefix + "_COUNT"))
	if err != nil || count < 1 {
		return &Config{
			noWrites: true,
		}
	}

	c.vars = make(map[string][]string, count)

	for i := range count {
		keyVar := fmt.Sprintf("%s%d", envPrefix+"_KEY_", i)
		key := os.Getenv(keyVar)

		valVar := fmt.Sprintf("%s%d", envPrefix+"_VALUE_", i)
		value, found := os.LookupEnv(valVar)

		if key == "" || !found {
			return &Config{
				noWrites: true,
			}
		}

		c.vars[key] = append(c.vars[key], value)
		debug.V(3).Log("added %s from env", key)
	}

	return c
}
