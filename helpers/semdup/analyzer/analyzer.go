// Package analyzer finds production functions with identical normalized SSA.
package analyzer

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"go/token"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

// Config controls package loading and report thresholds.
type Config struct {
	Dir             string
	Patterns        []string
	ExcludePrefixes []string
	MinInstructions int
}

// Report contains exact normalized-SSA duplicate groups.
type Report struct {
	PackagesLoaded    int     `json:"packages_loaded"`
	FunctionsAnalyzed int     `json:"functions_analyzed"`
	Groups            []Group `json:"groups"`
}

// Group contains functions with the same normalized SSA fingerprint.
type Group struct {
	Fingerprint string     `json:"fingerprint"`
	Effects     []string   `json:"effects"`
	Functions   []Function `json:"functions"`
}

// Function identifies one analyzed function.
type Function struct {
	Package      string `json:"package"`
	Name         string `json:"name"`
	Position     string `json:"position"`
	Instructions int    `json:"instructions"`
}

// Analyze loads the configured production packages and groups exact SSA matches.
func Analyze(cfg Config) (Report, error) {
	if len(cfg.Patterns) == 0 {
		cfg.Patterns = []string{"./..."}
	}
	if cfg.MinInstructions < 1 {
		cfg.MinInstructions = 2
	}

	pkgs, err := packages.Load(&packages.Config{
		Dir: cfg.Dir,
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles |
			packages.NeedImports | packages.NeedDeps | packages.NeedTypes |
			packages.NeedSyntax | packages.NeedTypesInfo,
		Tests: false,
	}, cfg.Patterns...)
	if err != nil {
		return Report{}, fmt.Errorf("load packages: %w", err)
	}
	if err := packageErrors(pkgs); err != nil {
		return Report{}, err
	}

	prog, _ := ssautil.AllPackages(pkgs, ssa.InstantiateGenerics)
	prog.Build()

	owned := make(map[string]struct{}, len(pkgs))
	for _, pkg := range pkgs {
		if pkg.Types != nil {
			owned[pkg.Types.Path()] = struct{}{}
		}
	}

	report := Report{PackagesLoaded: len(owned)}
	candidates := make(map[string][]functionRecord)
	for fn := range ssautil.AllFunctions(prog) {
		if !eligible(fn, owned, cfg.ExcludePrefixes) {
			continue
		}

		fingerprint, effects, instructions := fingerprint(fn)
		if instructions < cfg.MinInstructions {
			continue
		}

		report.FunctionsAnalyzed++
		pos := prog.Fset.Position(fn.Pos())
		record := functionRecord{
			Function: Function{
				Package:      fn.Pkg.Pkg.Path(),
				Name:         displayName(fn),
				Position:     position(pos),
				Instructions: instructions,
			},
			Effects: effects,
		}
		candidates[fingerprint] = append(candidates[fingerprint], record)
	}

	for fp, records := range candidates {
		if len(records) < 2 {
			continue
		}
		sort.Slice(records, func(i, j int) bool {
			if records[i].Package == records[j].Package {
				return records[i].Name < records[j].Name
			}
			return records[i].Package < records[j].Package
		})

		group := Group{Fingerprint: fp, Effects: records[0].Effects}
		for _, record := range records {
			group.Functions = append(group.Functions, record.Function)
		}
		report.Groups = append(report.Groups, group)
	}

	sort.Slice(report.Groups, func(i, j int) bool {
		if len(report.Groups[i].Functions) == len(report.Groups[j].Functions) {
			return report.Groups[i].Fingerprint < report.Groups[j].Fingerprint
		}
		return len(report.Groups[i].Functions) > len(report.Groups[j].Functions)
	})

	return report, nil
}

type functionRecord struct {
	Function
	Effects []string
}

func eligible(fn *ssa.Function, owned map[string]struct{}, excludes []string) bool {
	if fn == nil || fn.Pkg == nil || fn.Pkg.Pkg == nil || len(fn.Blocks) == 0 {
		return false
	}
	if fn.Synthetic != "" || fn.Name() == "init" || strings.Contains(fn.Name(), "$") {
		return false
	}
	pkgPath := fn.Pkg.Pkg.Path()
	if _, ok := owned[pkgPath]; !ok {
		return false
	}
	for _, prefix := range excludes {
		if strings.HasPrefix(pkgPath, prefix) {
			return false
		}
	}
	return true
}

func fingerprint(fn *ssa.Function) (string, []string, int) {
	replacements := make(map[string]string)
	next := 0
	register := func(value ssa.Value) {
		if value == nil || value.Name() == "" {
			return
		}
		if _, exists := replacements[value.Name()]; exists {
			return
		}
		replacements[value.Name()] = fmt.Sprintf("v%d", next)
		next++
	}
	for _, param := range fn.Params {
		register(param)
	}
	for _, free := range fn.FreeVars {
		register(free)
	}
	for _, block := range fn.Blocks {
		for _, instruction := range block.Instrs {
			if value, ok := instruction.(ssa.Value); ok {
				register(value)
			}
		}
	}

	var normalized strings.Builder
	effectSet := make(map[string]struct{})
	instructions := 0
	for blockIndex, block := range fn.Blocks {
		fmt.Fprintf(&normalized, "b%d:", blockIndex)
		for _, instruction := range block.Instrs {
			instructions++
			line := fmt.Sprintf("%T:%s", instruction, instruction.String())
			line = replaceIdentifiers(line, replacements)
			normalized.WriteString(line)
			normalized.WriteByte(';')
			classifyEffect(effectSet, instruction)
		}
		normalized.WriteString("succ=")
		for _, successor := range block.Succs {
			fmt.Fprintf(&normalized, "%d,", successor.Index)
		}
	}

	sum := sha256.Sum256([]byte(normalized.String()))
	effects := make([]string, 0, len(effectSet))
	for effect := range effectSet {
		effects = append(effects, effect)
	}
	sort.Strings(effects)

	return hex.EncodeToString(sum[:]), effects, instructions
}

func replaceIdentifiers(input string, replacements map[string]string) string {
	var output strings.Builder
	output.Grow(len(input))

	var quote byte
	escaped := false
	for i := 0; i < len(input); {
		current := input[i]
		if quote != 0 {
			output.WriteByte(current)
			i++
			if escaped {
				escaped = false
				continue
			}
			if current == '\\' && quote != '`' {
				escaped = true
				continue
			}
			if current == quote {
				quote = 0
			}
			continue
		}

		if current == '\'' || current == '"' || current == '`' {
			quote = current
			output.WriteByte(current)
			i++
			continue
		}

		if isIdentifierByte(current) {
			end := i + 1
			for end < len(input) && isIdentifierByte(input[end]) {
				end++
			}
			identifier := input[i:end]
			if replacement, ok := replacements[identifier]; ok {
				output.WriteString(replacement)
			} else {
				output.WriteString(identifier)
			}
			i = end
			continue
		}

		output.WriteByte(current)
		i++
	}

	return output.String()
}

func isIdentifierByte(value byte) bool {
	return value == '_' || value == '$' || value >= 'a' && value <= 'z' ||
		value >= 'A' && value <= 'Z' || value >= '0' && value <= '9'
}

func classifyEffect(effects map[string]struct{}, instruction ssa.Instruction) {
	switch instruction.(type) {
	case *ssa.Call, *ssa.Defer, *ssa.Go:
		effects["call"] = struct{}{}
	case *ssa.Store, *ssa.MapUpdate, *ssa.Send:
		effects["write"] = struct{}{}
	case *ssa.Panic:
		effects["panic"] = struct{}{}
	}
}

func displayName(fn *ssa.Function) string {
	if fn.Signature.Recv() == nil {
		return fn.Name()
	}
	return fmt.Sprintf("(%s).%s", fn.Signature.Recv().Type(), fn.Name())
}

func position(pos token.Position) string {
	if !pos.IsValid() {
		return ""
	}
	return fmt.Sprintf("%s:%d", pos.Filename, pos.Line)
}

func packageErrors(pkgs []*packages.Package) error {
	var messages []string
	packages.Visit(pkgs, nil, func(pkg *packages.Package) {
		for _, pkgErr := range pkg.Errors {
			messages = append(messages, pkgErr.Error())
		}
	})
	if len(messages) == 0 {
		return nil
	}
	sort.Strings(messages)
	return fmt.Errorf("package loading failed:\n%s", strings.Join(messages, "\n"))
}
