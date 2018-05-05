package action

import (
	"context"
	"time"

	"github.com/justwatchcom/gopass/pkg/out"

	"github.com/urfave/cli"
)

// History displays the history of a given secret
func (s *Action) History(ctx context.Context, c *cli.Context) error {
	name := c.Args().Get(0)
	showPassword := c.Bool("password")

	if name == "" {
		return ExitError(ctx, ExitUsage, nil, "Usage: %s history <NAME>", s.Name)
	}

	if !s.Store.Exists(ctx, name) {
		return ExitError(ctx, ExitNotFound, nil, "Secret not found")
	}

	revs, err := s.Store.ListRevisions(ctx, name)
	if err != nil {
		return ExitError(ctx, ExitUnknown, err, "Failed to get revisions: %s", err)
	}

	for _, rev := range revs {
		pw := ""
		if showPassword {
			sec, err := s.Store.GetRevision(ctx, name, rev.Hash)
			if err != nil {
				out.Debug(ctx, "Failed to get revision '%s' of '%s': %s", rev.Hash, name, err)
			}
			if err == nil {
				pw = " - " + sec.Password()
			}
		}
		out.Print(ctx, "%s - %s <%s> - %s - %s%s", rev.Hash[:8], rev.AuthorName, rev.AuthorEmail, rev.Date.Format(time.RFC3339), rev.Subject, pw)
	}
	return nil
}
