package zsh

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

func formatFlag(name, usage string) string {
	return fmt.Sprintf("--%s[%s]", longName(name), usage)
}

func formatFlagFunc() func(cli.Flag) (string, error) {
	return func(f cli.Flag) (string, error) {
		switch ft := f.(type) {
		case cli.BoolFlag:
			return formatFlag(ft.Name, ft.Usage), nil
		case cli.Float64Flag:
			return formatFlag(ft.Name, ft.Usage), nil
		case cli.GenericFlag:
			return formatFlag(ft.Name, ft.Usage), nil
		case cli.Int64Flag:
			return formatFlag(ft.Name, ft.Usage), nil
		case cli.Int64SliceFlag:
			return formatFlag(ft.Name, ft.Usage), nil
		case cli.IntFlag:
			return formatFlag(ft.Name, ft.Usage), nil
		case cli.IntSliceFlag:
			return formatFlag(ft.Name, ft.Usage), nil
		case cli.StringFlag:
			return formatFlag(ft.Name, ft.Usage), nil
		case cli.StringSliceFlag:
			return formatFlag(ft.Name, ft.Usage), nil
		case cli.Uint64Flag:
			return formatFlag(ft.Name, ft.Usage), nil
		case cli.UintFlag:
			return formatFlag(ft.Name, ft.Usage), nil
		default:
			return "", fmt.Errorf("unknown type: '%T'", f)
		}
	}
}

// GetCompletion returns a zsh completion script
func GetCompletion(a *cli.App) (string, error) {
	tplFuncs := template.FuncMap{
		"formatFlag": formatFlagFunc(),
	}
	tpl, err := template.New("zsh").Funcs(tplFuncs).Parse(zshTemplate)
	if err != nil {
		return "", err
	}
	buf := &bytes.Buffer{}
	if err := tpl.Execute(buf, a); err != nil {
		return "", err
	}
	return buf.String(), nil
}
