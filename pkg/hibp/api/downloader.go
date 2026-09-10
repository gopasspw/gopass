package api

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/fsutil"
	"github.com/gopasspw/gopass/pkg/termio"
)

// Download will download the list of all hashes from the API to a single, gzipped txt file.
// This is inspired by the "official" .NET based download tool. It does exactly 16⁵ / 1024*1024 (1M) requests
// to fetch all the possible prefixes.
func Download(ctx context.Context, path string, keep bool) error {
	if path == "" {
		return fmt.Errorf("need output path")
	}
	if fsutil.IsDir(path) {
		path = filepath.Join(path, fmt.Sprintf("pwned-passwords-sha1-ordered-by-hash-%s.txt.gz", time.Now().Format("2006-01-02")))
	}
	if !strings.HasSuffix(path, ".gz") {
		path += ".gz"
	}

	dir := filepath.Join(filepath.Dir(path), ".hibp-dl")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	fmt.Printf("Downloading hashes to %s ...", dir)

	maxVal := 1024 * 1024
	bar := termio.NewProgressBar(int64(maxVal))
	bar.Hidden = ctxutil.IsHidden(ctx)

	sem := make(chan struct{}, runtime.NumCPU()*4)
	wg := &sync.WaitGroup{}
	for i := range maxVal {
		wg.Add(1)
		go func() {
			sem <- struct{}{}
			defer func() {
				bar.Inc()
				<-sem
				wg.Done()
			}()
			// check for context cancelation
			select {
			case <-ctx.Done():
				return
			default:
			}
			if err := downloadChunk(i, dir, keep); err != nil {
				fmt.Printf("Chunk %d failed: %s", i, err)
			}
		}()
	}
	wg.Wait()
	bar.Done()

	// check for context cancelation
	select {
	case <-ctx.Done():
		return fmt.Errorf("user aborted")
	default:
	}

	fmt.Println("Download done.")

	fmt.Println("Assembling chunks ...")

	if err := joinChunks(dir, path, keep); err != nil {
		return err
	}

	fmt.Printf("Chunks assembled at %s\n", path)

	return nil
}

func joinChunks(dir, path string, keep bool) error {
	dirs, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	fh, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer fh.Close() //nolint:errcheck

	gzw := gzip.NewWriter(fh)
	defer gzw.Close() //nolint:errcheck

	bar := termio.NewProgressBar(int64(len(dirs)))

	for _, de := range dirs {
		// XXXXX.gz
		if len(de.Name()) != 8 {
			continue
		}
		if strings.HasPrefix(de.Name(), ".") {
			continue
		}
		if !strings.HasSuffix(de.Name(), ".gz") {
			continue
		}
		if err := copyChunk(gzw, filepath.Join(dir, de.Name())); err != nil {
			return err
		}

		bar.Inc()
	}
	bar.Done()

	if keep {
		return nil
	}

	return os.RemoveAll(dir)
}

func copyChunk(w io.Writer, fn string) error {
	fh, err := os.Open(fn)
	if err != nil {
		return err
	}
	defer fh.Close() //nolint:errcheck

	gzr, err := gzip.NewReader(fh)
	if err != nil {
		return err
	}
	defer gzr.Close() //nolint:errcheck

	n, err := io.Copy(w, gzr)
	debug.Log("Copied %d bytes from %s", n, fn)

	return err
}

func downloadChunk(chunk int, dir string, keep bool) error {
	hex := fmt.Sprintf("%X", chunk)
	prefix := strings.Repeat("0", 5-len(hex)) + hex

	fn := filepath.Join(dir, prefix+".gz")
	if keep && fsutil.IsFile(fn) {
		debug.Log("re-using existing file for chunk #%d: %s", chunk, fn)

		return nil
	}

	fh, err := os.OpenFile(fn, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer fh.Close() //nolint:errcheck

	gzw := gzip.NewWriter(fh)
	defer gzw.Close() //nolint:errcheck

	url := URL + "/range/" + prefix

	op := func() error {
		debug.Log("HTTP Request: %s", url)
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

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("HTTP request failed: %s %s", resp.Status, body)
		}

		for _, line := range strings.Split(string(body), "\n") {
			line = strings.TrimSpace(line)
			if len(line) < 37 {
				continue
			}
			fmt.Fprintf(gzw, "%s%s\n", prefix, line)
		}

		return nil
	}

	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = 10 * time.Second

	return backoff.Retry(op, bo)
}
