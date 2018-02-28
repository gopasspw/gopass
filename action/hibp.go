package action

import (
	"bufio"
	"compress/gzip"
	"context"
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/utils/notify"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/justwatchcom/gopass/utils/termio"
	"github.com/muesli/goprogressbar"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

var hibpAPIURL = "https://api.pwnedpasswords.com"

// HIBP compares all entries from the store against the provided SHA1 sum dumps
func (s *Action) HIBP(ctx context.Context, c *cli.Context) error {
	force := c.Bool("force")
	api := c.Bool("api")

	if api {
		return s.hibpAPI(ctx, force)
	}

	out.Yellow(ctx, "WARNING: Using the HIBPv2 dumps is very expensive. If you can condone leaking a few bits of entropy per secret you should probably use the '--api' flag.")

	return s.hibpDump(ctx, force)
}

func (s *Action) hibpAPI(ctx context.Context, force bool) error {
	if !force && !termio.AskForConfirmation(ctx, fmt.Sprintf("This command is checking all your secrets against the haveibeenpwned.com API.\n\nThis will send five bytes of each passwords SHA1 hash to an untrusted server!\n\nYou will be asked to unlock all your secrets!\nDo you want to continue?")) {
		return exitError(ctx, ExitAborted, nil, "user aborted")
	}

	shaSums, sortedShaSums, err := s.hibpPrecomputeHashes(ctx)
	if err != nil {
		return err
	}

	out.Print(ctx, "Checking pre-computed SHA1 hashes against the HIBP API ...")

	// compare the prepared list against all provided files. with a little more
	// code this could be parallelized
	matchList := make([]string, 0, 100)
	for _, shaSum := range sortedShaSums {
		freq, err := s.hibpAPILookup(ctx, shaSum)
		if err != nil {
			out.Red(ctx, "Failed to check HIBP API: %s", err)
			continue
		}
		if freq < 1 {
			continue
		}
		if pw, found := shaSums[shaSum]; found {
			matchList = append(matchList, pw)
		}
	}

	return s.printHIBPMatches(ctx, matchList)
}

func (s *Action) hibpAPILookup(ctx context.Context, shaSum string) (uint64, error) {
	if len(shaSum) < 40 {
		return 0, errors.Errorf("invalid shasum")
	}

	prefix := strings.ToUpper(shaSum[:5])
	suffix := strings.ToUpper(shaSum[5:])

	var count uint64
	url := fmt.Sprintf("%s/range/%s", hibpAPIURL, prefix)

	op := func() error {
		out.Debug(ctx, "[%s] HTTP Request: %s", shaSum, url)
		resp, err := http.Get(url)
		if err != nil {
			return err
		}
		defer func() {
			_ = resp.Body.Close()
		}()

		if resp.StatusCode == http.StatusNotFound {
			return nil
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("HTTP request failed: %s %s", resp.Status, body)
		}

		for _, line := range strings.Split(string(body), "\n") {
			if len(line) < 37 {
				continue
			}
			if line[:35] != suffix {
				continue
			}
			if iv, err := strconv.ParseUint(line[36:], 10, 64); err == nil {
				count = iv
				return nil
			}
		}
		return nil
	}

	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = 10 * time.Second

	err := backoff.Retry(op, bo)
	return count, err
}

func (s *Action) hibpDump(ctx context.Context, force bool) error {
	fns := strings.Split(os.Getenv("HIBP_DUMPS"), ",")
	if len(fns) < 1 || fns[0] == "" {
		return errors.Errorf("Please provide the name(s) of the haveibeenpwned.com password dumps in HIBP_DUMPS. See https://haveibeenpwned.com/Passwords for more information. Use 7z to extract the dump: 7z x pwned-passwords-2.0.txt.7z, then sort them: cat pwned-passwords-2.0.txt | LANG=C sort -S 10G --parallel=4 | gzip --fast > pwned-passwords-2.0.txt.gz")
	}

	if !force && !termio.AskForConfirmation(ctx, fmt.Sprintf("This command is checking all your secrets against the haveibeenpwned.com hashes in %+v.\nYou will be asked to unlock all your secrets!\nDo you want to continue?", fns)) {
		return exitError(ctx, ExitAborted, nil, "user aborted")
	}

	shaSums, sortedShaSums, err := s.hibpPrecomputeHashes(ctx)
	if err != nil {
		return err
	}

	out.Print(ctx, "Checking pre-computed SHA1 hashes against the blacklists ...")
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

	return s.printHIBPMatches(ctx, matchList)
}

func (s *Action) hibpPrecomputeHashes(ctx context.Context) (map[string]string, []string, error) {
	// build a map of all secrets sha sums to their names and also build a sorted (!)
	// list of this shasums. As the hibp dump is already sorted this allows for
	// a very efficient stream compare in O(n)
	t, err := s.Store.Tree(ctx)
	if err != nil {
		return nil, nil, exitError(ctx, ExitList, err, "failed to list store: %s", err)
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
			return nil, nil, exitError(ctx, ExitAborted, nil, "user aborted")
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
			out.Print(ctx, "\n"+color.YellowString("Failed to retrieve secret '%s': %s", secret, err))
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
	out.Print(ctx, "")
	// IMPORTANT: sort after all entries have been added. without the sort
	// the stream compare will not work
	sort.Strings(sortedShaSums)

	return shaSums, sortedShaSums, nil
}

func (s *Action) printHIBPMatches(ctx context.Context, matchList []string) error {
	if len(matchList) < 1 {
		_ = notify.Notify("gopass - audit HIBP", "Good news - No matches found!")
		out.Green(ctx, "Good news - No matches found!")
		return nil
	}

	sort.Strings(matchList)
	_ = notify.Notify("gopass - audit HIBP", fmt.Sprintf("Oh no - found %d matches", len(matchList)))
	out.Red(ctx, "Oh no - Found some matches:")
	for _, m := range matchList {
		out.Red(ctx, "\t- %s", m)
	}
	out.Cyan(ctx, "The passwords in the listed secrets were included in public leaks in the past. This means they are likely included in many word-list attacks and provide only very little security. Strongly consider changing those passwords!")
	return exitError(ctx, ExitAudit, nil, "weak passwords found")
}

func (s *Action) findHIBPMatches(ctx context.Context, fn string, shaSums map[string]string, sortedShaSums []string, matches chan<- string, done chan<- struct{}) {
	defer func() {
		done <- struct{}{}
	}()
	t0 := time.Now()

	var rdr io.Reader
	fh, err := os.Open(fn)
	if err != nil {
		out.Red(ctx, "Failed to open file %s: %s", fn, err)
		return
	}
	defer func() {
		_ = fh.Close()
	}()
	rdr = fh

	if strings.HasSuffix(fn, ".gz") {
		gzr, err := gzip.NewReader(fh)
		if err != nil {
			out.Red(ctx, "Failed to open the file %s: %s", fn, err)
			return
		}
		defer func() {
			_ = gzr.Close()
		}()
		rdr = gzr
	}

	out.Debug(ctx, "Checking file %s ...\n", fn)

	// index in sortedShaSums
	i := 0
	lineNo := 0
	numMatches := 0
	scanner := bufio.NewScanner(rdr)
SCAN:
	for scanner.Scan() {
		// check for context cancelation
		select {
		case <-ctx.Done():
			break SCAN
		default:
		}

		line := strings.TrimSpace(scanner.Text())
		lineNo++
		if i >= len(sortedShaSums) {
			break
		}
		pw := line[:40]
		freq := ""
		if len(line) > 41 {
			freq = line[:41]
		}
		if pw == sortedShaSums[i] {
			matches <- shaSums[pw]
			out.Debug(ctx, "MATCH at line %d: %s (#%s) / %s from %s", lineNo, pw, freq, shaSums[freq], fn)
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

	d0 := time.Since(t0)
	out.Debug(ctx, "Found %d matches in %d lines from %s in %.2fs (%.2f lines / s)\n", numMatches, lineNo, fn, d0.Seconds(), float64(lineNo)/d0.Seconds())
}

func sha1sum(data string) string {
	h := sha1.New()
	_, _ = h.Write([]byte(data))
	return strings.ToUpper(fmt.Sprintf("%x", h.Sum(nil)))
}
