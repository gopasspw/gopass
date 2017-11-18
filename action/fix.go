package action

import (
	"context"
	"errors"
	"strings"

	"github.com/justwatchcom/gopass/store/sub"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/urfave/cli"
)

// Fix checks the secrets for fixable inconsistencies
func (s *Action) Fix(ctx context.Context, c *cli.Context) error {
	if c.IsSet("force") {
		ctx = sub.WithFsckForce(ctx, c.Bool("force"))
	}
	check := c.Bool("check")

	if !s.AskForConfirmation(ctx, "Do you want to introduce YAML document separators to all secrets that can be parsed as valid YAML?") {
		return errors.New("user aborted")
	}

	t, err := s.Store.Tree()
	if err != nil {
		return s.exitError(ctx, ExitList, err, "failed to list store: %s", err)
	}

	pwList := t.List(0)
	num := 0

	for _, name := range pwList {
		sec, err := s.Store.Get(ctx, name)
		if err != nil {
			out.Red(ctx, "Failed to decode secret %s: %s", name, err)
			continue
		}
		if sec.Data() != nil {
			// valid YAML already
			out.Debug(ctx, "%s - not fixing - valid YAML", name)
			continue
		}
		body := sec.Body()
		if strings.HasPrefix(body, "---\n") {
			// invalid YAML, but doc-sep, avoid duplicating the doc-sep
			out.Debug(ctx, "%s - not fixing - invalid YAML w/ doc-sep", name)
			continue
		}
		if strings.Trim(body, "\n\r\t ") == "" {
			// body contains only whitespaces
			out.Debug(ctx, "%s - not fixing - only whitepaces", name)
			continue
		}
		out.Debug(ctx, "Secret %s before fix: %s", name, sec.String())
		if err := sec.SetBody("---\n" + body); err != nil {
			out.Red(ctx, "Failed to add doc-sep to %s: %s", name, err)
			continue
		}
		out.Debug(ctx, "Secret after fix: %s", sec.String())
		num++
		if check {
			continue
		}
		if err := s.Store.Set(ctx, name, sec); err != nil {
			out.Red(ctx, "Failed to save secret %s: %s", name, err)
		}
	}
	out.Green(ctx, "Fixed %d secrets", num)

	return nil
}
