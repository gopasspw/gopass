package audit

import (
	"fmt"

	"github.com/google/cel-go/cel"
	"github.com/gopasspw/gopass/pkg/gopass"
)

func NewCELValidator(expr string) (Validator, error) {
	env, err := cel.NewEnv(
		cel.Variable("name", cel.StringType),
		cel.Variable("password", cel.StringType),
		cel.Variable("kvp", cel.MapType(cel.StringType, cel.StringType)),
	)
	if err != nil {
		return Validator{}, err
	}

	ast, issues := env.Compile(expr)
	if issues != nil && issues.Err() != nil {
		return Validator{}, fmt.Errorf("failed to compile CEL expression %q: %w", expr, issues.Err())
	}

	prg, err := env.Program(ast)
	if err != nil {
		return Validator{}, fmt.Errorf("failed to create CEL program for expression %q: %w", expr, err)
	}

	return Validator{
		Name:        "cel:" + expr,
		Description: "CEL expression validator",
		Validate: func(name string, sec gopass.Secret) error {
			kvp := make(map[string]string)
			for _, k := range sec.Keys() {
				kvp[k], _ = sec.Get(k)
			}
			data := map[string]any{
				"name":     name,
				"password": sec.Password(),
				"kvp":      kvp,
			}

			out, _, err := prg.Eval(data)
			if err != nil {
				return fmt.Errorf("failed to evaluate CEL expression %q: %w", expr, err)
			}
			result, ok := out.Value().(bool)
			if !ok {
				return fmt.Errorf("CEL expression %q did not return a boolean result", expr)
			}
			if !result {
				return fmt.Errorf("CEL expression %q evaluated to false", expr)
			}
			return nil
		},
	}, nil
}
