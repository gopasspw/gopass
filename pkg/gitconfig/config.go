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

// Config is a single parsed config file. It contains a reference of the input file, if any.
// It can only be populated only by reading the environment variables.
type Config struct {
	path     string
	readonly bool // do not allow modifying values (even in memory)
	noWrites bool // do not persist changes to disk (e.g. for tests)
	raw      strings.Builder
	vars     map[string]string

	// TODO(gitconfig) keep a checksum of the parsed file and re-parse before over-writing a changed file?
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

	return c.rewriteRaw(key, "", func(fKey, key, value, comment string) (string, bool) {
		return "", true
	})
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
		c.vars = make(map[string]string, 16)
	}

	// already present at the same value, no need to rewrite the config
	if v, found := c.vars[key]; found && v == value {
		debug.Log("key %q with value %q already present. No re-writing.")

		return nil
	}

	_, present := c.vars[key]
	c.vars[key] = value

	debug.Log("set %q to %q", key, value)

	// a new key, insert it into an existing section, if any
	if !present {
		debug.Log("inserting value")

		return c.insertValue(key, value)
	}

	debug.Log("updating value")

	return c.rewriteRaw(key, value, func(fKey, sKey, value, comment string) (string, bool) {
		return fmt.Sprintf("    %s = %s%s", sKey, value, comment), false
	})
}

func (c *Config) insertValue(key, value string) error {
	debug.Log("input (%s: %s): ---------\n%s\n-----------\n", key, value, c.raw.String())

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
			line = strings.Trim(line, "[]")
			p := strings.Fields(line)
			if len(p) < 1 {
				continue
			}
			section = p[0]
			subsection = ""
			if len(p) > 1 {
				subsection = strings.ReplaceAll(p[1], "\\", "")
				subsection = strings.TrimPrefix(subsection, "\"")
				subsection = strings.TrimSuffix(subsection, "\"")
			}
		}

		if section != wSection {
			continue
		}
		if subsection != wSubsection {
			continue
		}

		lines = append(lines, fmt.Sprintf("    %s = %s", wKey, value))
		written = true
	}

	// not added to an existing section, so add it at the end
	if !written {
		sect := fmt.Sprintf("[%s]", wSection)
		if wSubsection != "" {
			sect = fmt.Sprintf("[%s \"%s\"]", wSection, wSubsection)
		}
		lines = append(lines, sect)
		lines = append(lines, fmt.Sprintf("    %s = %s", wKey, value))
	}

	c.raw = strings.Builder{}
	c.raw.WriteString(strings.Join(lines, "\n"))
	c.raw.WriteString("\n")

	debug.Log("output: ---------\n%s\n-----------\n", c.raw.String())

	return c.flushRaw()
}

// rewriteRaw is used to rewrite the raw config copy. It is used for set and unset operations
// with different callbacks each.
func (c *Config) rewriteRaw(key, value string, cb parseFunc) error {
	debug.Log("input (%s: %s): ---------\n%s\n-----------\n", key, value, c.raw.String())

	lines := parseConfig(strings.NewReader(c.raw.String()), key, value, cb)

	c.raw = strings.Builder{}
	c.raw.WriteString(strings.Join(lines, "\n"))
	c.raw.WriteString("\n")

	debug.Log("output: ---------\n%s\n-----------\n", c.raw.String())

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

	debug.Log("writing config to %s: -----------\n%s\n--------------", c.path, c.raw.String())

	return os.WriteFile(c.path, []byte(c.raw.String()), 0o600)
}

type parseFunc func(fqkn, skn, value, comment string) (newLine string, skipLine bool)

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
		line := s.Text()

		lines = append(lines, line)

		if strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, "[") {
			line = strings.Trim(line, "[]")
			p := strings.Fields(line)
			if len(p) < 1 {
				continue
			}
			section = p[0]
			subsection = ""
			if len(p) > 1 {
				subsection = strings.ReplaceAll(p[1], "\\", "")
				subsection = strings.TrimPrefix(subsection, "\"")
				subsection = strings.TrimSuffix(subsection, "\"")
			}
		}

		if key != "" && (section != wSection && subsection != wSubsection) {
			continue
		}

		kvp := strings.Split(line, "=")
		trim(kvp)
		if len(kvp) < 2 {
			continue
		}

		fKey := section + "."
		if subsection != "" {
			fKey += subsection + "."
		}
		fKey += kvp[0]
		if key == "" {
			wKey = kvp[0]
		}

		oValue := kvp[1]
		comment := ""

		if strings.ContainsAny(oValue, "#;") {
			comment = " " + oValue[strings.IndexAny(oValue, "#;"):]
			oValue = oValue[:strings.IndexAny(oValue, "#;")]
			oValue = strings.TrimSpace(oValue)
		}

		// debug.Log("got: %q - want: %q\n", key, fKey)
		if key != "" && (key != fKey) {
			continue
		}
		if key != "" {
			oValue = value
		}

		newLine, skip := cb(fKey, wKey, oValue, comment)
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
		vars:     make(map[string]string, len(data)),
	}

	for k, v := range data {
		c.vars[k] = v
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
// Invalid configs will be silently rejceted.
//
// Warning: Error handling is subject to change!
func ParseConfig(r io.Reader) *Config {
	c := &Config{
		vars: make(map[string]string, 42),
	}

	lines := parseConfig(r, "", "", func(fk, k, v, comment string) (string, bool) {
		// debug.Log("setting %q to %q (%s)", fk, v, comment)
		c.vars[fk] = v

		return fmt.Sprintf("    %s = %s%s", k, v, comment), false
	})

	c.raw.WriteString(strings.Join(lines, "\n"))
	c.raw.WriteString("\n")

	debug.Log("processed config: %s\nvars: %+v", c.raw.String(), c.vars)

	return c
}

// LoadConfigFromEnv will try to parse an overlay config from the environment variables.
// If no environment variables are set the resulting config will be valid but empty.
// Either way it will not be writeable.
func LoadConfigFromEnv(envPrefix string) *Config {
	c := &Config{
		noWrites: true,
	}

	count, err := strconv.Atoi(os.Getenv(envPrefix + "_CONFIG_COUNT"))
	if err != nil || count < 1 {
		return &Config{
			noWrites: true,
		}
	}

	for i := 0; i < count; i++ {
		keyVar := fmt.Sprintf("%s%d", envPrefix+"_CONFIG_KEY_", i)
		key := os.Getenv(keyVar)

		valVar := fmt.Sprintf("%s%d", envPrefix+"_CONFIG_VALUE_", i)
		value, found := os.LookupEnv(valVar)

		if key == "" || !found {
			return &Config{
				noWrites: true,
			}
		}

		c.vars[key] = value
		debug.Log("added %s from env", key)
	}

	return c
}
