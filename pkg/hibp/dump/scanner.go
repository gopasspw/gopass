package dump

import (
	"bufio"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"sort"
	"strings"

	"github.com/gopasspw/gopass/pkg/fsutil"
	"github.com/gopasspw/gopass/pkg/out"
)

// Scanner is a HIBP dump scanner
type Scanner struct {
	dumps []string
}

// New creates a new scanner
func New(dumps ...string) (*Scanner, error) {
	ok := make([]string, 0, len(dumps))
	for _, dump := range dumps {
		if !fsutil.IsFile(dump) {
			continue
		}
		ok = append(ok, dump)
	}
	if len(ok) < 1 {
		return nil, fmt.Errorf("no valid dumps given")
	}
	return &Scanner{
		dumps: ok,
	}, nil
}

// LookupBatch takes a slice SHA1 hashes and matches them against
// the provided dumps
func (s *Scanner) LookupBatch(ctx context.Context, in []string) []string {
	if len(in) < 1 {
		return nil
	}

	sort.Strings(in)
	for i, hash := range in {
		in[i] = strings.ToUpper(hash)
	}

	out := make([]string, 0, len(in))
	results := make(chan string, len(in))
	done := make(chan struct{}, len(s.dumps))

	for _, fn := range s.dumps {
		go s.scanFile(ctx, fn, in, results, done)
	}
	go func() {
		for result := range results {
			out = append(out, result)
		}
		done <- struct{}{}
	}()
	for range s.dumps {
		<-done
	}
	close(results)
	<-done

	return out
}

func (s *Scanner) scanFile(ctx context.Context, fn string, in []string, results chan string, done chan struct{}) {
	defer func() {
		done <- struct{}{}
	}()

	if isSorted(fn) {
		out.Debug(ctx, "file %s appears to be sorted", fn)
		s.scanSortedFile(ctx, fn, in, results)
		return
	}
	out.Debug(ctx, "file %s is not sorted", fn)
	s.scanUnsortedFile(ctx, fn, in, results)
}

func isSorted(fn string) bool {
	var rdr io.Reader
	fh, err := os.Open(fn)
	if err != nil {
		return false
	}
	defer func() {
		_ = fh.Close()
	}()
	rdr = fh

	if strings.HasSuffix(fn, ".gz") {
		gzr, err := gzip.NewReader(fh)
		if err != nil {
			return false
		}
		defer func() {
			_ = gzr.Close()
		}()
		rdr = gzr
	}

	lineNo := 0
	lastLine := ""
	scanner := bufio.NewScanner(rdr)
	for scanner.Scan() {
		lineNo++
		if lineNo > 100 {
			return true
		}

		line := scanner.Text()
		if len(line) > 40 {
			line = line[:40]
		}
		if line < lastLine {
			return false
		}
		lastLine = line
	}
	return true
}

func (s *Scanner) scanSortedFile(ctx context.Context, fn string, in []string, results chan string) {
	var rdr io.Reader
	fh, err := os.Open(fn)
	if err != nil {
		out.Error(ctx, "Failed to open file %s: %s", fn, err)
		return
	}
	defer func() {
		_ = fh.Close()
	}()
	rdr = fh

	if strings.HasSuffix(fn, ".gz") {
		gzr, err := gzip.NewReader(fh)
		if err != nil {
			out.Error(ctx, "Failed to open the file %s: %s", fn, err)
			return
		}
		defer func() {
			_ = gzr.Close()
		}()
		rdr = gzr
	}

	out.Debug(ctx, "Checking file %s ...\n", fn)

	// index in input (sorted SHA sums)
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

		lineNo++
		if i >= len(in) {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		hash := line[:40]

		if hash == in[i] {
			results <- hash
			out.Debug(ctx, "[%s] MATCH at line %d: %s", fn, lineNo, hash)
			numMatches++
			// advance to next sha sum from store and next line in file
			i++
			continue
		}
		// advance in sha sums from store until we've reached the position in
		// the file
		for i < len(in) && line > in[i] {
			i++
		}
	}

	out.Debug(ctx, "Finished checking file %s", fn)
}

func (s *Scanner) scanUnsortedFile(ctx context.Context, fn string, in []string, results chan string) {
	var rdr io.Reader
	fh, err := os.Open(fn)
	if err != nil {
		out.Error(ctx, "Failed to open file %s: %s", fn, err)
		return
	}
	defer func() {
		_ = fh.Close()
	}()
	rdr = fh

	if strings.HasSuffix(fn, ".gz") {
		gzr, err := gzip.NewReader(fh)
		if err != nil {
			out.Error(ctx, "Failed to open the file %s: %s", fn, err)
			return
		}
		defer func() {
			_ = gzr.Close()
		}()
		rdr = gzr
	}

	lines := make(chan string, 1024)
	worker := runtime.NumCPU()
	done := make(chan struct{}, worker)
	for i := 0; i < worker; i++ {
		out.Debug(ctx, "[%d] Starting matcher ...", i)
		go s.matcher(ctx, in, lines, results, done)
	}

	out.Debug(ctx, "Checking file %s ...\n", fn)
	scanner := bufio.NewScanner(rdr)
SCAN:
	for scanner.Scan() {
		// check for context cancelation
		select {
		case <-ctx.Done():
			break SCAN
		default:
		}

		lines <- scanner.Text()
	}
	close(lines)

	for i := 0; i < worker; i++ {
		<-done
	}

	out.Debug(ctx, "Finished checking file %s", fn)
}

func (s *Scanner) matcher(ctx context.Context, in []string, lines chan string, results chan string, done chan struct{}) {
	defer func() {
		done <- struct{}{}
	}()

LINE:
	for line := range lines {
		// check for context cancelation
		select {
		case <-ctx.Done():
			break LINE
		default:
		}

		line := strings.ToUpper(strings.TrimSpace(line))
		hash := line[:40]
		for _, candidate := range in {
			if candidate == hash {
				results <- hash
				continue LINE
			}
		}
	}
}
