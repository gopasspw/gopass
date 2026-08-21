// semdup reports production functions with identical normalized SSA.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/gopasspw/gopass/helpers/semdup/analyzer"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("semdup", flag.ContinueOnError)
	flags.SetOutput(stderr)
	dir := flags.String("dir", ".", "module directory")
	jsonOutput := flags.Bool("json", false, "write JSON")
	minimum := flags.Int("min-instructions", 2, "minimum SSA instruction count")
	exclude := flags.String("exclude", "", "comma-separated import path prefixes")
	if err := flags.Parse(args); err != nil {
		return 2
	}

	patterns := flags.Args()
	if len(patterns) == 0 {
		patterns = []string{"./..."}
	}

	report, err := analyzer.Analyze(analyzer.Config{
		Dir:             *dir,
		Patterns:        patterns,
		ExcludePrefixes: splitNonEmpty(*exclude),
		MinInstructions: *minimum,
	})
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "semdup: %v\n", err)
		return 1
	}

	if *jsonOutput {
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(report); err != nil {
			_, _ = fmt.Fprintf(stderr, "semdup: encode report: %v\n", err)
			return 1
		}
		return 0
	}

	_, _ = fmt.Fprintf(stdout, "packages=%d functions=%d groups=%d\n", report.PackagesLoaded, report.FunctionsAnalyzed, len(report.Groups))
	for _, group := range report.Groups {
		_, _ = fmt.Fprintf(stdout, "\n%s effects=%s\n", group.Fingerprint[:12], strings.Join(group.Effects, ","))
		for _, function := range group.Functions {
			_, _ = fmt.Fprintf(stdout, "  %s.%s %s instructions=%d\n", function.Package, function.Name, function.Position, function.Instructions)
		}
	}

	return 0
}

func splitNonEmpty(value string) []string {
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
