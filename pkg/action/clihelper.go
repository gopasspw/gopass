package action

import (
	"context"
	"strings"

	"github.com/justwatchcom/gopass/pkg/cui"
	"github.com/urfave/cli"
)

// ConfirmRecipients asks the user to confirm a given set of recipients
func (s *Action) ConfirmRecipients(ctx context.Context, name string, recipients []string) ([]string, error) {
	return cui.ConfirmRecipients(ctx, s.Store.Crypto(ctx, name), name, recipients)
}

type argList []string

func (a argList) Get(n int) string {
	if len(a) > n {
		return a[n]
	}
	return ""
}

func parseArgs(c *cli.Context) (argList, map[string]string) {
	args := make(argList, 0, len(c.Args()))
	kvps := make(map[string]string, len(c.Args()))
OUTER:
	for _, arg := range c.Args() {
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
			kvps[key] = p[1]
			continue OUTER
		}
		args = append(args, arg)
	}
	return args, kvps
}
