package analyzer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnalyzeGroupsEquivalentSSAWithDifferentNames(t *testing.T) {
	dir := writeModule(t, `package sample

func sum(a, b int) int {
	result := a + b
	return result
}

func total(left, right int) int {
	return left + right
}

func difference(left, right int) int {
	return left - right
}
`)

	report, err := Analyze(Config{Dir: dir, Patterns: []string{"./..."}, MinInstructions: 1})
	require.NoError(t, err)
	require.Equal(t, 3, report.FunctionsAnalyzed)
	require.Len(t, report.Groups, 1)

	names := []string{report.Groups[0].Functions[0].Name, report.Groups[0].Functions[1].Name}
	assert.ElementsMatch(t, []string{"sum", "total"}, names)
	assert.Empty(t, report.Groups[0].Effects)
}

func TestAnalyzePreservesConstantsAndSideEffects(t *testing.T) {
	dir := writeModule(t, `package sample

import "fmt"

func one() int { return 1 }
func two() int { return 2 }
func printA() { fmt.Println("a") }
func printB() { fmt.Println("a") }
`)

	report, err := Analyze(Config{Dir: dir, Patterns: []string{"./..."}, MinInstructions: 1})
	require.NoError(t, err)
	require.Len(t, report.Groups, 1)

	group := report.Groups[0]
	assert.ElementsMatch(t, []string{"call", "write"}, group.Effects)
	assert.ElementsMatch(t, []string{"printA", "printB"}, []string{
		group.Functions[0].Name,
		group.Functions[1].Name,
	})
}

func TestAnalyzeDoesNotNormalizeIdentifiersInsideConstants(t *testing.T) {
	dir := writeModule(t, `package sample

func literalA(a string) string { return "a" }
func literalB(x string) string { return "x" }
`)

	report, err := Analyze(Config{Dir: dir, Patterns: []string{"./..."}, MinInstructions: 1})
	require.NoError(t, err)
	assert.Empty(t, report.Groups)
}

func TestAnalyzeExcludesAnonymousClosures(t *testing.T) {
	dir := writeModule(t, `package sample

func outer() func() int { return func() int { return 1 } }
`)

	report, err := Analyze(Config{Dir: dir, Patterns: []string{"./..."}, MinInstructions: 1})
	require.NoError(t, err)
	assert.Equal(t, 1, report.FunctionsAnalyzed)
}

func writeModule(t *testing.T, source string) string {
	t.Helper()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/sample\n\ngo 1.25.0\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sample.go"), []byte(source), 0o600))

	return dir
}
