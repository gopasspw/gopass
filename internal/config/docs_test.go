package config

import (
	"bufio"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/gopasspw/gopass/internal/set"
	"golang.org/x/exp/maps"
)

// ignoredEnvs is a list of environment variables that are used by gopass
// but originate from elsewhere. They should be well known and properly
// documented already.
var ignoredEnvs = set.Map([]string{
	"APPDATA",
	"GIT_AUTHOR_EMAIL",
	"GIT_AUTHOR_NAME",
	"GNUPGHOME",
	"GOPATH",
	"GOPASS_CONFIG_NOSYSTEM", // name assembled, tests can't catch it
	"GOPASS_DEBUG_FILES",     // indirect usage
	"GOPASS_DEBUG_FUNCS",     // indirect usage
	"GOPASS_GPG_OPTS",        // indirect usage
	"GOPASS_UMASK",           // indirect usage
	"PASSWORD_STORE_UMASK",   // indirect usage
	"GPG_TTY",
	"HOME",
	"LOCALAPPDATA",
	"XDG_CACHE_HOME",
	"XDG_CONFIG_HOME",
	"XDG_DATA_HOME",
})

// ignoredOptions is a list of config options that are used by gopass
// but originate elsewhere (e.g. git). They should not be documented
// here as well.
var ignoredOptions = set.Map([]string{
	"core.pre-hook",
	"core.post-hook",
	"user.email",
	"user.name",
})

func TestConfigOptsInDocs(t *testing.T) {
	t.Parallel()

	documented := documentedOpts(t)
	used := usedOpts(t)

	t.Logf("Config options documented in doc: %+v", documented)
	t.Logf("Config options used in the code: %+v", used)

	for _, k := range set.SortedKeys(documented) {
		if !used[k] {
			t.Errorf("Documented but not used: %s", k)
		}
	}
	for _, k := range set.SortedKeys(used) {
		if !documented[k] {
			t.Errorf("Used but not documented: %s", k)
		}
	}
}

func usedOpts(t *testing.T) map[string]bool {
	t.Helper()

	optRE := regexp.MustCompile(`(?:\.Get(?:|Int|Bool)\(\"([a-z]+\.[a-z-]+)\"\)|\.GetM\([^,]+, \"([a-z]+\.[a-z-]+)\"\)|config\.(?:Bool|Int|String)\((?:ctx|c\.Context), \"([a-z]+\.[a-z-]+)\"\)|hook\.Invoke(?:Root)?\(ctx, \"([a-z]+\.[a-z-]+)\")`)
	opts := make(map[string]bool, 42)

	dir := filepath.Join("..", "..")
	if err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") && path != dir {
			return filepath.SkipDir
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(info.Name(), "_test.go") {
			return nil
		}
		if !strings.HasSuffix(info.Name(), ".go") {
			return nil
		}

		return usedOptsInFile(t, path, opts, optRE)
	}); err != nil {
		t.Errorf("failed to walk %s: %s", dir, err)
	}

	return opts
}

func usedOptsInFile(t *testing.T, fn string, opts map[string]bool, re *regexp.Regexp) error {
	t.Helper()

	fh, err := os.Open(fn)
	if err != nil {
		return err
	}
	defer fh.Close() //nolint:errcheck

	scanner := bufio.NewScanner(fh)
	for scanner.Scan() {
		line := scanner.Text()

		if !re.MatchString(line) {
			continue
		}

		found := re.FindStringSubmatch(line)
		// t.Logf("found: %q", found)
		if len(found) < 4 {
			continue
		}

		for i := 1; i < 10; i++ {
			if found[i] == "" {
				continue
			}
			if ignoredOptions[found[i]] {
				break
			}
			opts[found[i]] = true

			break
		}
	}

	return nil
}

func documentedOpts(t *testing.T) map[string]bool {
	t.Helper()

	fn := filepath.Join("..", "..", "docs", "config.md")
	fh, err := os.Open(fn)
	if err != nil {
		t.Fatalf("failed to open %s: %s", fn, err)
	}
	defer fh.Close() //nolint:errcheck

	optRE := regexp.MustCompile(`^\| .([a-z]+\.[a-z-]+).`)

	opts := make(map[string]bool, 42)
	scanner := bufio.NewScanner(fh)
	for scanner.Scan() {
		line := scanner.Text()

		if !optRE.MatchString(line) {
			continue
		}
		found := optRE.FindStringSubmatch(line)
		if len(found) < 2 {
			continue
		}
		if _, found := ignoredOptions[found[1]]; found {
			continue
		}
		opts[found[1]] = true
	}

	return opts
}

func TestEnvVarsInDocs(t *testing.T) {
	t.Parallel()

	documented := documentedEnvs(t)
	used := usedEnvs(t)

	t.Logf("env options documented in doc: %+v", documented)
	t.Logf("env options used in the code: %+v", used)

	for _, k := range set.Sorted(maps.Keys(documented)) {
		if !used[k] {
			t.Errorf("Documented but not used: %s", k)
		}
	}
	for _, k := range set.Sorted(maps.Keys(used)) {
		if !documented[k] {
			t.Errorf("Used but not documented: %s", k)
		}
	}
}

func usedEnvs(t *testing.T) map[string]bool {
	t.Helper()

	optRE := regexp.MustCompile(`os\.(?:Getenv|LookupEnv)\(\"([^"]+)\"\)`)
	opts := make(map[string]bool, 42)

	dir := filepath.Join("..", "..")
	if err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") && path != dir {
			return filepath.SkipDir
		}
		if info.IsDir() && (info.Name() == "helpers" || info.Name() == "tests") {
			return filepath.SkipDir
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(info.Name(), "_test.go") {
			return nil
		}
		if !strings.HasSuffix(info.Name(), ".go") {
			return nil
		}

		return usedEnvsInFile(t, path, opts, optRE)
	}); err != nil {
		t.Errorf("failed to walk %s: %s", dir, err)
	}

	return opts
}

func usedEnvsInFile(t *testing.T, fn string, opts map[string]bool, re *regexp.Regexp) error {
	t.Helper()

	fh, err := os.Open(fn)
	if err != nil {
		return err
	}
	defer fh.Close() //nolint:errcheck

	scanner := bufio.NewScanner(fh)
	for scanner.Scan() {
		line := scanner.Text()

		if !re.MatchString(line) {
			continue
		}

		found := re.FindStringSubmatch(line)
		// t.Logf("found: %q", found)
		if len(found) < 2 {
			continue
		}

		v := found[1]

		if ignoredEnvs[v] {
			continue
		}

		opts[v] = true
	}

	return nil
}

func documentedEnvs(t *testing.T) map[string]bool {
	t.Helper()

	fn := filepath.Join("..", "..", "docs", "config.md")
	fh, err := os.Open(fn)
	if err != nil {
		t.Fatalf("failed to open %s: %s", fn, err)
	}
	defer fh.Close() //nolint:errcheck

	optRE := regexp.MustCompile(`^\| .([A-Z0-9_]+).`)

	opts := make(map[string]bool, 42)
	scanner := bufio.NewScanner(fh)
	for scanner.Scan() {
		line := scanner.Text()

		if !optRE.MatchString(line) {
			continue
		}
		found := optRE.FindStringSubmatch(line)
		if len(found) < 2 {
			continue
		}

		v := found[1]

		if ignoredEnvs[v] {
			continue
		}

		opts[v] = true
	}

	return opts
}
