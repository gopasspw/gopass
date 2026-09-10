package dump

import (
	"compress/gzip"
	"context"
	"fmt"
	"os"
	"strings"
)

func (s *Scanner) Merge(ctx context.Context, outfile string) error { //nolint:cyclop
	for _, dump := range s.dumps {
		if !isSorted(dump) {
			return fmt.Errorf("merging unsorted input files is not supported")
		}
	}
	if len(s.dumps) != 2 {
		return fmt.Errorf("nothing to merge")
	}
	if !strings.HasSuffix(outfile, ".gz") {
		outfile += ".gz"
	}

	fmt.Printf("Merging %+v into %s\n", s.dumps, outfile)
	fh, err := os.OpenFile(outfile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer fh.Close() //nolint:errcheck

	gzw := gzip.NewWriter(fh)
	defer gzw.Close() //nolint:errcheck

	resLeft := make(chan string, 1024)
	resRight := make(chan string, 1024)
	go func() {
		s.scanSortedFile(ctx, s.dumps[0], nil, resLeft)
		close(resLeft)
	}()
	go func() {
		s.scanSortedFile(ctx, s.dumps[1], nil, resRight)
		close(resRight)
	}()

	for {
		lv, lok := <-resLeft
		rv, rok := <-resRight
		if !lok && !rok {
			// all done
			break
		}
		// left needs to catch up
		for lv[:40] < rv[:40] || !rok {
			fmt.Fprintln(gzw, lv)
			lv, lok = <-resLeft
			if !lok {
				break
			}
		}
		// right needs to catch up
		for rv[:40] < lv[:40] || !lok {
			fmt.Fprintln(gzw, rv)
			rv, rok = <-resRight
			if !rok {
				break
			}
		}
		if lv[:40] == rv[:40] {
			maxVal := max(rv[41:], lv[41:])
			fmt.Fprintf(gzw, "%s:%s\n", lv[:40], maxVal)

			continue
		}
		if lok && !rok {
			fmt.Fprintln(gzw, lv)

			continue
		}
		if !lok && rok {
			fmt.Fprintln(gzw, rv)

			continue
		}
		if lv[:40] < rv[:40] {
			fmt.Fprintln(gzw, lv)
			fmt.Fprintln(gzw, rv)

			continue
		}
		if lv[:40] < rv[:40] {
			fmt.Fprintln(gzw, rv)
			fmt.Fprintln(gzw, lv)

			continue
		}
	}

	return nil
}
