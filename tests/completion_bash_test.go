package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// bashCompletionEnv is the environment a bash completion request runs in.
type bashCompletionEnv struct {
	// words is COMP_WORDS, cword is the index of the word being completed.
	words []string
	cword int
	// compType mirrors COMP_TYPE; 63 means bash is listing matches.
	compType string
}

// runBashCompletion sources the completion script produced by
// `gopass completion bash` and invokes the completion function the way bash
// would on TAB, returning the COMPREPLY entries.
//
// COMP_WORDS is passed through a file, one word per line, so that words
// containing quotes or spaces need no shell escaping.
func runBashCompletion(t *testing.T, binary string, env bashCompletionEnv) []string {
	t.Helper()

	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Skipf("bash not found: %s", err)
	}

	dir := t.TempDir()
	binCopy := filepath.Join(dir, "gopass")
	require.NoError(t, copyFile(binary, binCopy))

	out, err := exec.Command(binCopy, "completion", "bash").Output()
	require.NoError(t, err)

	script := filepath.Join(dir, "completion.sh")
	require.NoError(t, os.WriteFile(script, out, 0o600))

	wordsFile := filepath.Join(dir, "words.txt")
	require.NoError(t, os.WriteFile(wordsFile, []byte(strings.Join(env.words, "\n")+"\n"), 0o600))

	// Drive the generated _gopass_bash_autocomplete function with a
	// COMP_WORDS/COMP_CWORD setup mirroring what bash does on TAB.
	driver := filepath.Join(dir, "driver.sh")
	driverSrc := `source "` + script + `"
COMP_WORDS=()
while IFS= read -r w; do
  COMP_WORDS+=("$w")
done < "` + wordsFile + `"
COMP_CWORD=` + strconv.Itoa(env.cword) + `
COMP_TYPE=` + env.compType + `
_gopass_bash_autocomplete
i=0
while [ $i -lt ${#COMPREPLY[@]} ]; do
  printf '%s\n' "${COMPREPLY[$i]}"
  i=$((i+1))
done
`
	require.NoError(t, os.WriteFile(driver, []byte(driverSrc), 0o600))

	cmd := exec.Command(bash, driver)
	cmd.Dir = dir
	result, err := cmd.CombinedOutput()
	require.NoError(t, err, "bash harness failed: %s", string(result))

	lines := strings.Split(strings.TrimRight(string(result), "\n"), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return nil
	}

	return lines
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, data, 0o700)
}

// TestCompletionBashTokensNoDescriptions pins the fix for the regression
// reported in #3468: since the urfave/cli v3 migration the completion output
// uses a "token:description" format. Completing `gopass sh<TAB>` must offer
// the bare command name, not "show:Display the content of a secret".
func TestCompletionBashTokensNoDescriptions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping test on windows.")
	}

	ts := newTester(t)
	defer ts.teardown()

	ts.initStore()

	t.Run("completes bare command names", func(t *testing.T) {
		got := runBashCompletion(t, ts.Binary, bashCompletionEnv{
			words:    []string{ts.Binary, "sh"},
			cword:    1,
			compType: "37",
		})
		require.NotEmpty(t, got)
		assert.Equal(t, []string{"show"}, got)
	})

	t.Run("completes secret names", func(t *testing.T) {
		out, err := ts.runCmd([]string{ts.Binary, "insert", "foo"}, []byte("bar"))
		require.NoError(t, err, "failed to insert password:\n%s", out)

		got := runBashCompletion(t, ts.Binary, bashCompletionEnv{
			words:    []string{ts.Binary, "show", "fo"},
			cword:    2,
			compType: "37",
		})
		require.Equal(t, []string{"foo"}, got)
	})

	t.Run("completes secret names containing a colon", func(t *testing.T) {
		// gopass must not cut a secret whose name contains a colon at the
		// first colon even though urfave/cli uses colons to separate
		// descriptions (#3468).
		// The secret goes through insert so it is encrypted and listed
		// like any secret a user would have.
		out, err := ts.runCmd([]string{ts.Binary, "insert", "websites:github"}, []byte("secret"))
		require.NoError(t, err, "failed to insert password:\n%s", out)

		got := runBashCompletion(t, ts.Binary, bashCompletionEnv{
			words:    []string{ts.Binary, "show", "web"},
			cword:    2,
			compType: "37",
		})
		// Byte-exact: the entry must carry no leftover escape from the
		// token:description protocol, i.e. no backslash before the colon.
		require.Equal(t, []string{"websites:github"}, got)
	})

	t.Run("completes flags without descriptions", func(t *testing.T) {
		got := runBashCompletion(t, ts.Binary, bashCompletionEnv{
			words:    []string{ts.Binary, "--c"},
			cword:    1,
			compType: "37",
		})
		require.NotEmpty(t, got)
		for _, c := range got {
			assert.True(t, strings.HasPrefix(c, "--"),
				"flag completion must only offer flags, got %q", c)
			assert.NotContains(t, c, ":",
				"flag completion must not contain the usage description, got %q", c)
		}
		assert.Contains(t, got, "--clip")
	})

	t.Run("lists descriptions on second tab", func(t *testing.T) {
		got := runBashCompletion(t, ts.Binary, bashCompletionEnv{
			words:    []string{ts.Binary, "sh"},
			cword:    1,
			compType: "63",
		})
		require.NotEmpty(t, got)
		assert.Contains(t, got[0], "show")
		assert.Contains(t, got[0], "--  ")
		assert.Contains(t, got[0], "secret")
	})
}
