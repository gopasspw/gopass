package main

import (
	"context"
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"sort"

	"github.com/gopasspw/gopass/internal/notify"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/termio"
	"github.com/gopasspw/gopass/pkg/gopass"
	hibpapi "github.com/gopasspw/gopass/pkg/hibp/api"
	hibpdump "github.com/gopasspw/gopass/pkg/hibp/dump"

	"github.com/fatih/color"
	"github.com/muesli/goprogressbar"
)

type hibp struct {
	gp gopass.Store
}

// CheckAPI checks your secrets against the HIBPv2 API
func (s *hibp) CheckAPI(ctx context.Context, force bool) error {
	if !force && !termio.AskForConfirmation(ctx, "This command is checking all your secrets against the haveibeenpwned.com API.\n\nThis will send five bytes of each passwords SHA1 hash to an untrusted server!\n\nYou will be asked to unlock all your secrets!\nDo you want to continue?") {
		return fmt.Errorf("user aborted")
	}

	shaSums, sortedShaSums, err := s.precomputeHashes(ctx)
	if err != nil {
		return err
	}

	out.Print(ctx, "Checking pre-computed SHA1 hashes against the HIBP API ...")

	// compare the prepared list against all provided files
	matchList := make([]string, 0, len(sortedShaSums))
	for _, shaSum := range sortedShaSums {
		freq, err := hibpapi.Lookup(ctx, shaSum)
		if err != nil {
			out.Error(ctx, "Failed to check HIBP API: %s", err)
			continue
		}
		if freq < 1 {
			continue
		}
		if pw, found := shaSums[shaSum]; found {
			matchList = append(matchList, pw)
		}
	}

	return s.printMatches(ctx, matchList)
}

// CheckDump checks your secrets against the provided HIBPv2 Dumps
func (s *hibp) CheckDump(ctx context.Context, force bool, dumps []string) error {
	out.Yellow(ctx, "WARNING: Using the HIBPv2 dumps is very expensive. If you can condone leaking a few bits of entropy per secret you should probably use the '--api' flag.")

	if !force && !termio.AskForConfirmation(ctx, fmt.Sprintf("This command is checking all your secrets against the haveibeenpwned.com hashes in %+v.\nYou will be asked to unlock all your secrets!\nDo you want to continue?", dumps)) {
		return fmt.Errorf("user aborted")
	}

	shaSums, sortedShaSums, err := s.precomputeHashes(ctx)
	if err != nil {
		return err
	}

	scanner, err := hibpdump.New(dumps...)
	if err != nil {
		return fmt.Errorf("failed to create new HIBP Dump scanner: %s", err)
	}

	matchedSums := scanner.LookupBatch(ctx, sortedShaSums)
	out.Debug(ctx, "In: %+v - Out: %+v", sortedShaSums, matchedSums)
	matchList := make([]string, 0, len(matchedSums))
	for _, matchedSum := range matchedSums {
		if pw, found := shaSums[matchedSum]; found {
			matchList = append(matchList, pw)
		}
	}

	return s.printMatches(ctx, matchList)
}

func (s *hibp) precomputeHashes(ctx context.Context) (map[string]string, []string, error) {
	// build a map of all secrets sha sums to their names and also build a sorted (!)
	// list of this shasums. As the hibp dump is already sorted this allows for
	// a very efficient stream compare in O(n)
	pwList, err := s.gp.List(ctx)
	if err != nil {
		return nil, nil, err
	}
	// map sha1sum back to secret name for reporting
	shaSums := make(map[string]string, len(pwList))
	// build list of sha1sums (must be sorted later!) for stream comparison
	sortedShaSums := make([]string, 0, len(shaSums))
	// display progress bar
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

	out.Print(ctx, "Computing SHA1 hashes of all your secrets ...")
	for _, secret := range pwList {
		// check for context cancelation
		select {
		case <-ctx.Done():
			return nil, nil, fmt.Errorf("user aborted")
		default:
		}

		bar.Current++
		if bar.Current > bar.Total {
			bar.Total = bar.Current
		}
		bar.Text = fmt.Sprintf("%d of %d secrets computed", bar.Current, bar.Total)
		bar.LazyPrint()
		// only handle secrets / passwords, never the body
		// comparing the body is super hard, as every user may choose to use
		// the body of a secret differently. In the future we may support
		// go templates to extract and compare data from the body
		sec, err := s.gp.Get(ctx, secret)
		if err != nil {
			out.Print(ctx, "\n"+color.YellowString("Failed to retrieve secret '%s': %s", secret, err))
			continue
		}

		pw := sec.Password()
		// do not check empty passwords, there should be caught by `gopass audit`
		// anyway
		if len(pw) < 1 {
			continue
		}
		sum := sha1hex(pw)
		shaSums[sum] = secret
		sortedShaSums = append(sortedShaSums, sum)
	}
	out.Print(ctx, "")
	// IMPORTANT: sort after all entries have been added. without the sort
	// the stream compare will not work
	sort.Strings(sortedShaSums)

	return shaSums, sortedShaSums, nil
}

func (s *hibp) printMatches(ctx context.Context, matchList []string) error {
	if len(matchList) < 1 {
		_ = notify.Notify(ctx, name, "Good news - No matches found!")
		out.Green(ctx, "Good news - No matches found!")
		return nil
	}

	sort.Strings(matchList)
	_ = notify.Notify(ctx, name, fmt.Sprintf("Oh no - found %d matches", len(matchList)))
	out.Error(ctx, "Oh no - Found some matches:")
	for _, m := range matchList {
		out.Error(ctx, "\t- %s", m)
	}
	out.Cyan(ctx, "The passwords in the listed secrets were included in public leaks in the past. This means they are likely included in many word-list attacks and provide only very little security. Strongly consider changing those passwords!")
	return fmt.Errorf("weak passwords found")
}

func sha1hex(data string) string {
	h := sha1.New()
	_, _ = h.Write([]byte(data))
	return fmt.Sprintf("%X", h.Sum(nil))
}
