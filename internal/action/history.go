package action

import (
	"time"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"

	"github.com/urfave/cli/v2"
)

// History displays the history of a given secret
func (s *Action) History(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	name := c.Args().Get(0)
	showPassword := c.Bool("password")

	if name == "" {
		return ExitError(ExitUsage, nil, "Usage: %s history <NAME>", s.Name)
	}

	if !s.Store.Exists(ctx, name) {
		return ExitError(ExitNotFound, nil, "Secret not found")
	}

	revs, err := s.Store.ListRevisions(ctx, name)
	if err != nil {
		return ExitError(ExitUnknown, err, "Failed to get revisions: %s", err)
	}

	for _, rev := range revs {
		pw := ""
		if showPassword {
			_, sec, err := s.Store.GetRevision(ctx, name, rev.Hash)
			if err != nil {
				debug.Log("Failed to get revision %q of %q: %s", rev.Hash, name, err)
			}
			if err == nil {
				pw = " - " + sec.Password()
			}
		}
		out.Printf(ctx, "%s - %s <%s> - %s - %s%s\n", rev.Hash, rev.AuthorName, rev.AuthorEmail, rev.Date.Format(time.RFC3339), rev.Subject, pw)
	}
	return nil
}
