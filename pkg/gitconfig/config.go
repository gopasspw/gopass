package gitconfig

import (
	"bufio"
	"fmt"
	"io"
	"os"
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
				debug.Log("key %q with value %q already present. Not re-writing.", key, value)

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
		debug.Log("not writing changes to disk (noWrites %t, path %q)", c.noWrites, c.path)

		return nil
	}

	if err := os.MkdirAll(filepath.Dir(c.path), 0o700); err != nil {
		return err
	}

	debug.V(3).Log("writing config to %s: \n--------------\n%s\n--------------", c.path, c.raw.String())

	if err := os.WriteFile(c.path, []byte(c.raw.String()), 0o600); err != nil {
		return fmt.Errorf("failed to write config to %s: %w", c.path, err)
	}

	debug.Log("wrote config to %s", c.path)

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

		if key != "" && (section != wSection && subsection != wSubsection) {
			continue
		}

		// TODO(gitconfig) This will skip over valid entries like this one:
		// [core]
		//  sslVerify
		// These are odd but we should still support them.
		k, v, found := strings.Cut(line, " = ")
		if !found {
			debug.V(3).Log("no valid KV-pair on line: %q", line)

			continue
		}

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
	fh, err := os.Open(fn)
	if err != nil {
		return nil, err
	}
	defer fh.Close() //nolint:errcheck

	c := ParseConfig(fh)
	c.path = fn

	return c, nil
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
		debug.Log("added %s from env", key)
	}

	return c
}
