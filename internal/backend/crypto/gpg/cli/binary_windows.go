// +build windows

package cli

import (
	"context"
	"errors"
	"os/exec"
	"path/filepath"
	"sort"

	"golang.org/x/sys/windows/registry"

	"github.com/gopasspw/gopass/internal/debug"
	"github.com/gopasspw/gopass/pkg/fsutil"
)

func detectBinary(bin string) (string, error) {
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
				ver:  version(context.TODO(), p),
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
	return binary, nil
}

func detectBinaryCandidates(bin string) ([]string, error) {
	// gpg.exe for GPG4Win 3.0.0; would be gpg2.exe for 2.x
	bins := make([]string, 0, 4)

	bins, err := searchRegistry(bin, bins)
	if err != nil {
		return bins, err
	}

	bins, err = searchPath(bin, bins)
	if err != nil {
		return bins, err
	}

	return bins, nil
}

func searchRegistry(bin string, bins []string) ([]string, error) {
	// try to detect location of installed GPG4Win
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\GnuPG`, registry.QUERY_VALUE|registry.WOW64_32KEY)
	if err != nil {
		return bins, nil
	}

	if v, _, err := k.GetStringValue("Install Directory"); err == nil && v != "" {
		for _, b := range []string{bin, "gpg2.exe", "gpg.exe"} {
			if b == "" {
				continue
			}
			gpgPath := filepath.Join(v, "bin", b)
			if fsutil.IsFile(gpgPath) {
				bins = append(bins, gpgPath)
			}
		}
	}

	return bins, nil
}

func searchPath(bin string, bins []string) ([]string, error) {
	// try to detect location for GPG installed somewhere on the PATH
	for _, b := range []string{bin, "gpg2.exe", "gpg.exe"} {
		if b == "" {
			continue
		}
		gpgPath, err := exec.LookPath(b)
		if err != nil {
			continue
		}
		if fsutil.IsFile(gpgPath) {
			bins = append(bins, gpgPath)
		}
	}

	return bins, nil
}
