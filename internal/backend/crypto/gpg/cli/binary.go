package cli

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gopasspw/gopass/internal/debug"
	"github.com/gopasspw/gopass/pkg/appdir"
	"github.com/gopasspw/gopass/pkg/fsutil"
)

var (
	gpgBinC string
)

// Binary returns the GPG binary location
func (g *GPG) Binary() string {
	if g == nil {
		return ""
	}
	return g.binary
}

func binaryLocCacheFn() (string, error) {
	userCache := appdir.UserCache()
	if !fsutil.IsDir(userCache) {
		if err := os.MkdirAll(userCache, 0755); err != nil {
			return "", err
		}
	}
	return filepath.Join(userCache, "gpg-binary.loc"), nil
}

func readBinaryLocFromCache() (string, error) {
	fn, err := binaryLocCacheFn()
	if err != nil {
		return "", err
	}
	buf, err := ioutil.ReadFile(fn)
	if err != nil {
		return "", err
	}
	loc := strings.TrimSpace(string(buf))
	if loc == "" {
		return "", fmt.Errorf("empty location in cache")
	}
	return loc, nil
}

func writeBinaryLocToCache(fn string) error {
	binaryLoc, err := binaryLocCacheFn()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(binaryLoc, []byte(fn), 0644)
}

// Binary returns the GPG binary location
func Binary(ctx context.Context, bin string) (string, error) {
	if gpgBinC != "" {
		return gpgBinC, nil
	}
	if binLoc, err := readBinaryLocFromCache(); err == nil {
		gpgBinC = binLoc
		debug.Log("read binary from cache: %s", binLoc)
		return binLoc, nil
	}

	bins, err := detectBinaryCandidates(bin)
	if err != nil {
		return "", err
	}
	bv := make(byVersion, 0, len(bins))
	for _, b := range bins {
		debug.Log("Looking for '%s' ...", b)
		if p, err := exec.LookPath(b); err == nil {
			gb := gpgBin{
				path: p,
				ver:  version(ctx, p),
			}
			debug.Log("Found '%s' at '%s' (%s)", b, p, gb.ver.String())
			bv = append(bv, gb)
		}
	}
	if len(bv) < 1 {
		return "", errors.New("no gpg binary found")
	}
	sort.Sort(bv)
	binary := bv[len(bv)-1].path
	debug.Log("using '%s'", binary)
	gpgBinC = binary
	if err := writeBinaryLocToCache(binary); err != nil {
		debug.Log("failed to write binary location to cache file: %s", err)
	}
	return binary, nil
}
