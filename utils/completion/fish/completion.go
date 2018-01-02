package fish

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"

	"github.com/urfave/cli"
)

func longName(name string) string {
	parts := strings.Split(name, ",")
	if len(parts) < 1 {
		return ""
	}
	return strings.TrimSpace(parts[0])
}

func shortName(name string) string {
	parts := strings.Split(name, ",")
	if len(parts) < 2 {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func formatFlag(name, usage, typ string) string {
	switch typ {
	case "short":
		return shortName(name)
	case "long":
		return longName(name)
	case "usage":
		return usage
	default:
		return ""
	}
}

func formatFlagFunc(typ string) func(cli.Flag) (string, error) {
	return func(f cli.Flag) (string, error) {
		switch ft := f.(type) {
		case cli.BoolFlag:
			return formatFlag(ft.Name, ft.Usage, typ), nil
		case cli.Float64Flag:
			return formatFlag(ft.Name, ft.Usage, typ), nil
		case cli.GenericFlag:
			return formatFlag(ft.Name, ft.Usage, typ), nil
		case cli.Int64Flag:
			return formatFlag(ft.Name, ft.Usage, typ), nil
		case cli.Int64SliceFlag:
			return formatFlag(ft.Name, ft.Usage, typ), nil
		case cli.IntFlag:
			return formatFlag(ft.Name, ft.Usage, typ), nil
		case cli.IntSliceFlag:
			return formatFlag(ft.Name, ft.Usage, typ), nil
		case cli.StringFlag:
			return formatFlag(ft.Name, ft.Usage, typ), nil
		case cli.StringSliceFlag:
			return formatFlag(ft.Name, ft.Usage, typ), nil
		case cli.Uint64Flag:
			return formatFlag(ft.Name, ft.Usage, typ), nil
		case cli.UintFlag:
			return formatFlag(ft.Name, ft.Usage, typ), nil
		default:
			return "", fmt.Errorf("unknown type: '%T'", f)
		}
	}
}

// GetCompletion returns a fish completion script
func GetCompletion(a *cli.App) (string, error) {
	tplFuncs := template.FuncMap{
		"formatShortFlag": formatFlagFunc("short"),
		"formatLongFlag":  formatFlagFunc("long"),
		"formatFlagUsage": formatFlagFunc("usage"),
	}
	tpl, err := template.New("fish").Funcs(tplFuncs).Parse(fishTemplate)
	if err != nil {
		return "", err
	}
	buf := &bytes.Buffer{}
	if err := tpl.Execute(buf, a); err != nil {
		return "", err
	}
	return buf.String(), nil
}
