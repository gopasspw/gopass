package action

import (
	"context"
	"strings"

	"github.com/gopasspw/gopass/internal/cui"
	"github.com/urfave/cli/v2"
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
	args := make(argList, 0, c.Args().Len())
	kvps := make(map[string]string, c.Args().Len())
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
