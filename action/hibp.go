package action

import (
	"bufio"
	"context"
	"crypto/sha1"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/muesli/goprogressbar"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

// HIBP compares all entries from the store against the provided SHA1 sum dumps
func (s *Action) HIBP(ctx context.Context, c *cli.Context) error {
	force := c.Bool("force")

	fns := strings.Split(os.Getenv("HIBP_DUMPS"), ",")
	if len(fns) < 1 || fns[0] == "" {
		return errors.Errorf("Please provide the name(s) of the haveibeenpwned.com password dumps in HIBP_DUMPS. See https://haveibeenpwned.com/Passwords for more information")
	}

	if !force && !s.AskForConfirmation(ctx, fmt.Sprintf("This command is checking all your secrets against the haveibeenpwned.com hashes in %+v.\nYou will be asked to unlock all your secrets!\nDo you want to continue?", fns)) {
		return errors.Errorf("user aborted")
	}

	// build a map of all secrets sha sums to their names and also build a sorted (!)
	// list of this shasums. As the hibp dump is already sorted this allows for
	// a very efficient stream compare in O(n)
	t, err := s.Store.Tree()
	if err != nil {
		return s.exitError(ctx, ExitList, err, "failed to list store: %s", err)
	}

	pwList := t.List(0)
	// map sha1sum back to secret name for reporting
	shaSums := make(map[string]string, len(pwList))
	// build list of sha1sums (must be sorted later!) for stream comparison
	sortedShaSums := make([]string, 0, len(shaSums))
	// display progress bar
	bar := &goprogressbar.ProgressBar{
		Total: int64(len(pwList)),
		Width: 120,
	}
	fmt.Println("Computing SHA1 hashes of all your secrets ...")
	for _, secret := range pwList {
		// check for context cancelation
		select {
		case <-ctx.Done():
			return s.exitError(ctx, ExitAborted, nil, "user aborted")
		default:
		}

		bar.Current++
		bar.Text = fmt.Sprintf("%d of %d secrets computed", bar.Current, bar.Total)
		bar.LazyPrint()
		// only handle secrets / passwords, never the body
		// comparing the body is super hard, as every user may choose to use
		// the body of a secret differently. In the future we may support
		// go templates to extract and compare data from the body
		sec, err := s.Store.Get(ctx, secret)
		if err != nil {
			fmt.Println("\n" + color.YellowString("Failed to retrieve secret '%s': %s", secret, err))
			continue
		}
		// do not check empty passwords, there should be caught by `gopass audit`
		// anyway
		if len(sec.Password()) < 1 {
			continue
		}
		sum := sha1sum(sec.Password())
		shaSums[sum] = secret
		sortedShaSums = append(sortedShaSums, sum)
	}
	fmt.Println("")
	// IMPORTANT: sort after all entries have been added. without the sort
	// the stream compare will not work
	sort.Strings(sortedShaSums)

	fmt.Println("Checking pre-computed SHA1 hashes against the blacklists ...")
	matches := make(chan string, 1000)
	done := make(chan struct{})
	// compare the prepared list against all provided files. with a little more
	// code this could be parallelized
	for _, fn := range fns {
		go s.findHIBPMatches(ctx, fn, shaSums, sortedShaSums, matches, done)
	}
	matchList := make([]string, 0, 100)
	go func() {
		for match := range matches {
			matchList = append(matchList, match)
		}
	}()
	for range fns {
		<-done
	}

	if len(matchList) < 0 {
		fmt.Println(color.GreenString("Good news - No matches found!"))
		return nil
	}
	sort.Strings(matchList)
	fmt.Println(color.RedString("Oh no - Found some matches:"))
	for _, m := range matchList {
		fmt.Println(color.RedString("\t- %s", m))
	}
	fmt.Println(color.CyanString("The passwords in the listed secrets were included in public leaks in the past. This means they are likely included in many word-list attacks and provide only very little security. Strongly consider changing those passwords!"))
	return s.exitError(ctx, ExitAudit, nil, "weak passwords found")
}

func (s *Action) findHIBPMatches(ctx context.Context, fn string, shaSums map[string]string, sortedShaSums []string, matches chan<- string, done chan<- struct{}) {
	defer func() {
		done <- struct{}{}
	}()
	t0 := time.Now()

	fh, err := os.Open(fn)
	if err != nil {
		fmt.Println(color.RedString("Failed to open file %s: %s", fn, err))
		return
	}
	defer func() {
		_ = fh.Close()
	}()

	if ctxutil.IsDebug(ctx) {
		fmt.Printf("Checking file %s ...\n", fn)
	}

	// index in sortedShaSums
	i := 0
	lineNo := 0
	numMatches := 0
	scanner := bufio.NewScanner(fh)
	for scanner.Scan() {
		// check for context cancelation
		select {
		case <-ctx.Done():
			break
		default:
		}

		line := strings.TrimSpace(scanner.Text())
		lineNo++
		if i >= len(sortedShaSums) {
			break
		}
		if line == sortedShaSums[i] {
			matches <- shaSums[line]
			if ctxutil.IsDebug(ctx) {
				fmt.Printf("MATCH at line %d: %s / %s from %s\n", lineNo, line, shaSums[line], fn)
			}
			numMatches++
			// advance to next sha sum from store and next line in file
			i++
			continue
		}
		// advance in sha sums from store until we've reached the position in
		// the file
		for i < len(sortedShaSums) && line > sortedShaSums[i] {
			i++
		}
	}
	if ctxutil.IsDebug(ctx) {
		d0 := time.Since(t0)
		fmt.Printf("Found %d matches in %d lines from %s in %.2fs (%.2f lines / s)\n", numMatches, lineNo, fn, d0.Seconds(), float64(lineNo)/d0.Seconds())
	}
}

func sha1sum(data string) string {
	h := sha1.New()
	_, _ = h.Write([]byte(data))
	return strings.ToUpper(fmt.Sprintf("%x", h.Sum(nil)))
}
