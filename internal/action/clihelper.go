package action

import (
	"strings"

	"github.com/urfave/cli/v2"
)

type argList []string

func (a argList) Get(n int) string {
	if len(a) > n {
		return a[n]
	}

	return ""
}

func parseArgs(c *cli.Context) (argList, map[string]string) {
	args := make(argList, 0, c.Args().Len())
	kvps := make(map[string]string, c.Args().Len())
	if c.Args().Len() == 1 {
		// If there is only one arg, assume it is
		// the secret name, so don't attempt to
		// parse into args and kvps
		args = append(args, c.Args().Get(0))

		return args, kvps
	}
OUTER:
	for _, arg := range c.Args().Slice() {
		for _, sep := range []string{":", "="} {
			if !strings.Contains(arg, sep) {
				continue
			}
			p := strings.Split(arg, sep)
			if len(p) < 2 {
				args = append(args, arg)

				continue OUTER
			}
			key := p[0]
			kvps[key] = strings.Join(p[1:], ":")

			continue OUTER
		}
		args = append(args, arg)
	}

	return args, kvps
}
