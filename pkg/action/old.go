package action

import (
	"context"
	"fmt"
	"io/ioutil"
	"sort"
	"time"

	"github.com/gopasspw/gopass/pkg/notify"
	"github.com/gopasspw/gopass/pkg/out"

	"github.com/muesli/goprogressbar"
	"github.com/urfave/cli"
)

// Old use last modification date for each entry and display it if it is too old
func (s *Action) Old(ctx context.Context, c *cli.Context) error {
	days := c.Int("days")
	// Password not changed after this date are considered to be old
	oldLimit := time.Now().AddDate(-0, -0, -days)

	// Gather secrets, to check their last modification date
	t, err := s.Store.Tree(ctx)
	if err != nil {
		return ExitError(ctx, ExitList, err, "failed to list store: %s", err)
	}
	pwList := t.List(0)
	matchList := make([]string, 0)

	bar := &goprogressbar.ProgressBar{
		Total: int64(len(pwList)),
		Width: 120,
	}
	if out.IsHidden(ctx) {
		old := goprogressbar.Stdout
		goprogressbar.Stdout = ioutil.Discard
		defer func() {
			goprogressbar.Stdout = old
		}()
	}

	out.Print(ctx, "Finding passwords not changed since %s ...", oldLimit.Format(time.RFC850))
	for _, secret := range pwList {
		// check for context cancelation
		select {
		case <-ctx.Done():
			return ExitError(ctx, ExitAborted, nil, "user aborted")
		default:
		}

		bar.Current++
		bar.Text = fmt.Sprintf("%d of %d secrets computed", bar.Current, bar.Total)
		bar.LazyPrint()

		// The first Revision in revs is the last commit for the secret in the store
		revs, err := s.Store.ListRevisions(ctx, secret)
		if err != nil {
			return ExitError(ctx, ExitUnknown, err, "Failed to get revisions: %s", err)
		}

		last := revs[0]
		if last.Date.Before(oldLimit) {
			matchList = append(matchList, secret)
		}

	}

	out.Print(ctx, "")
	return s.printOldMatches(ctx, matchList)
}

func (s *Action) printOldMatches(ctx context.Context, matchList []string) error {
	if len(matchList) < 1 {
		_ = notify.Notify(ctx, "gopass - audit old", "Good news - No matches found!")
		out.Green(ctx, "Good news - No matches found!")
		return nil
	}

	sort.Strings(matchList)
	_ = notify.Notify(ctx, "gopass - audit old", fmt.Sprintf("Oh no - found %d matches", len(matchList)))
	out.Error(ctx, "Oh no - Found some matches:")
	for _, m := range matchList {
		out.Error(ctx, "\t- %s", m)
	}
	out.Cyan(ctx, "The passwords in the listed secrets are too old.")
	return ExitError(ctx, ExitAudit, nil, "old passwords found")
}
